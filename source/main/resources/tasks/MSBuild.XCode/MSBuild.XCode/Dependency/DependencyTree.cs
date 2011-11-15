using System;
using System.IO;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyTree
    {
        private List<DependencyTreeNode> mRootNodes;
        private Dictionary<string, DependencyTreeNode> mAllNodesMap;
        
        private Queue<DependencyTreeNode> mCompileQueue;
        private int mCompileIteration;

        private string mPlatform;
        private PackageInstance mPackage;
        private List<DependencyInstance> mDependencies;

        public PackageInstance Package { get { return mPackage; } }
        public string Platform { get { return mPlatform;  } }
        public List<DependencyInstance> Dependencies { get { return mDependencies; } }

        public DependencyTree(PackageInstance package, string platform, List<DependencyInstance> dependencies)
        {
            mPackage = package;
            mPlatform = platform;
            mDependencies = new List<DependencyInstance>();
            
            // Filter the dependencies:
            //  We need to cross-reference the dependency list with all the projects specified
            //  in the POM. Every project might have a different sub-set of platforms and this
            //  determines the platforms that this dependency package is required for.
            //
            //  For example, package A might be a dependency for the first project in the POM
            //  and in that project only 'Win32' is specified. It is obvious that package A only
            //  needs to be synchronized for platform 'Win32'.
            //  
            foreach(DependencyInstance di in dependencies)
            {
                foreach (ProjectInstance pi in Package.Pom.Projects)
                {
                    // If this project does not define this platform and does not
                    // specify this dependency package than we do not add it to
                    // the dependency list for this dependency tree.
                    if (pi.HasPlatform(platform) && pi.IsDependentOn(di.Name))
                    {
                        mDependencies.Add(di);
                        break;
                    }
                }
            }

            mRootNodes = new List<DependencyTreeNode>();
            mAllNodesMap = new Dictionary<string, DependencyTreeNode>();
            mCompileQueue = new Queue<DependencyTreeNode>();
            mCompileIteration = 0;
        }

        public List<PackageInstance> GetAllDependencyPackages()
        {
            List<PackageInstance> allDependencyPackages = new List<PackageInstance>();
            foreach (DependencyTreeNode node in mAllNodesMap.Values)
                allDependencyPackages.Add(node.Package);
            return allDependencyPackages;
        }

        public bool ContainsDependencyForPlatform(string DependencyName, string Platform)
        {
            foreach (DependencyTreeNode node in mAllNodesMap.Values)
            {
                if (String.Compare(node.Name, DependencyName, true) == 0)
                    return true;
            }
            return false;
        }

        public bool HasNode(string name)
        {
            return mAllNodesMap.ContainsKey(name);
        }

        public DependencyTreeNode FindNode(string name)
        {
            DependencyTreeNode depNode;
            if (!mAllNodesMap.TryGetValue(name, out depNode))
            {
                return null;
            }
            return depNode;
        }

        public bool AddNode(DependencyTreeNode node)
        {
            if (!mAllNodesMap.ContainsKey(node.Name))
            {
                mAllNodesMap.Add(node.Name, node);
                return true;
            }
            return false;
        }

        public bool EnqueueForCompile(DependencyTreeNode node)
        {
            if (node.Iteration != mCompileIteration)
            {
                mCompileQueue.Enqueue(node);
                return true;
            }
            return false;
        }

        // Incremental compilation of the dependency tree
        // Return value:
        // 1 = A package in the tree needs to be updated (it is not available in the repo)
        // 0 = The tree has been compiled
        // -1 = A package failed to load
        public int Compile()
        {
            ProgressTracker progress = ProgressTracker.Instance;

            List<double> progress_percentages = new List<double>();
            foreach (DependencyInstance d in Dependencies)
                progress_percentages.Add(100);
            
            ProgressTracker.Step progress_step = progress.Add(progress_percentages.ToArray());

            mCompileQueue.Clear();
            mCompileIteration++;

            // Process all nodes again
            foreach (DependencyInstance d in Dependencies)
            {
                if (!mAllNodesMap.ContainsKey(d.Name))
                {
                    DependencyTreeNode depNode = new DependencyTreeNode(Platform, d, 1);
                    EnqueueForCompile(depNode);
                    mAllNodesMap.Add(depNode.Name, depNode);
                    mRootNodes.Add(depNode);
                }
                else
                {
                    DependencyTreeNode depNode;
                    mAllNodesMap.TryGetValue(d.Name, out depNode);
                    EnqueueForCompile(depNode);
                }
                progress.Next();
            }

            int result = 0;
            if (mCompileQueue.Count > 0)
            {
                progress_percentages.Clear();
                for (int i = 0; i < mCompileQueue.Count; ++i)
                    progress_percentages.Add(500);
                progress_step.Add(progress_percentages.ToArray());

                // Breadth-First 
                while (mCompileQueue.Count > 0 && result == 0)
                {
                    DependencyTreeNode node = mCompileQueue.Dequeue();
                    node.Iteration = mCompileIteration;

                    int added_to_queue = mCompileQueue.Count;
                    result = node.Compile(this);
                    added_to_queue = mCompileQueue.Count - added_to_queue;

                    if (added_to_queue > 0)
                    {
                        progress_percentages.Clear();
                        for (int i = 0; i < added_to_queue; ++i)
                            progress_percentages.Add(500);
                        progress_step.Add(progress_percentages.ToArray());
                    }
                }
            }

            return result;
        }

        public int Update(DependencyTreeNode node)
        {
            int result = 0;

            PackageInstance pi = node.Package;
            PackageState p = pi.Package;
            result = PackageInstance.RepoActor.Update(p, node.Dependency.VersionRange);

            if (result == -1)
                Loggy.Error(String.Format("Dependency {0} in group {1} doesn't exist at the remote and cache package repositories", node.Dependency.Name, node.Dependency.Group.ToString()));

            return result;
        }

        public bool SaveInfo(xFilename filepath)
        {
            try
            {
                if (!Directory.Exists(filepath.AbsolutePath.Full))
                    Directory.CreateDirectory(filepath.AbsolutePath.Full);

                FileStream stream = new FileStream(filepath.ToString(), FileMode.Create, FileAccess.Write);
                StreamWriter writer = new StreamWriter(stream);

                HashSet<string> register = new HashSet<string>();

                ComparableVersion version = Package.Pom.Versions.GetForPlatform(Platform);
                string versionStr = version != null ? version.ToString() : "?";
                writer.WriteLine(String.Format("name={0}, group={1}, language={2}, version={3}, platform={4}", Package.Name, Package.Group.ToString(), Package.Language, versionStr, Platform));
                foreach (DependencyTreeNode node in mRootNodes)
                    node.SaveInfo(writer, register);

                writer.Close();
                stream.Close();
                return true;
            }
            catch (SystemException)
            {
                return false;
            }
        }

        public void Info()
        {
            foreach (DependencyTreeNode node in mAllNodesMap.Values)
            {
                PackageState p = node.Package.Package;
                Loggy.Info(String.Format("Name                       : {0}", p.Name));
                Loggy.Info(String.Format("Version                    : {0}", p.TargetVersion));
            }
        }

        public void Print()
        {
            string indent = "+";
            ComparableVersion version = Package.Pom.Versions.GetForPlatform(Platform);
            string versionStr = version != null ? version.ToString() : "?";
            Loggy.Info(String.Format("{0} {1}, version={2}, type={3}", indent, Package.Name, versionStr, "Root"));
            foreach (DependencyTreeNode node in mRootNodes)
                node.Print(indent);
        }
    }
}


