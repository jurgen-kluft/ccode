using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;
using FileDirectoryPath;

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
            mDependencies = dependencies;
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
            }

            // Breadth-First 
            int result = 0;
            while (mCompileQueue.Count > 0 && result == 0)
            {
                DependencyTreeNode node = mCompileQueue.Dequeue();
                node.Iteration = mCompileIteration;
                result = node.Compile(this);
            }

            return result;
        }

        public int Update(DependencyTreeNode node)
        {
            int result = 0;

            PackageInstance pi = node.Package;
            Package p = pi.Package;
            result = PackageInstance.RepoActor.Update(p, node.Dependency.VersionRange);

            if (result == -1)
                Loggy.Error(String.Format("Dependency {0} in group {1} doesn't exist at the remote and cache package repositories", node.Dependency.Name, node.Dependency.Group.ToString()));

            return result;
        }

        public bool SaveInfo(FileDirectoryPath.FilePathAbsolute filepath)
        {
            try
            {
                if (!Directory.Exists(filepath.ParentDirectoryPath.ToString()))
                    Directory.CreateDirectory(filepath.ParentDirectoryPath.ToString());

                FileStream stream = new FileStream(filepath.ToString(), FileMode.Create, FileAccess.Write);
                StreamWriter writer = new StreamWriter(stream);

                HashSet<string> register = new HashSet<string>();

                ComparableVersion version = Package.Pom.Versions.GetForPlatform(Platform);
                string versionStr = version != null ? version.ToString() : "?";
                writer.WriteLine(String.Format("{0}, version={1}, platform={2}", Package.Name, versionStr, Platform));
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
                Package p = node.Package.Package;
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


