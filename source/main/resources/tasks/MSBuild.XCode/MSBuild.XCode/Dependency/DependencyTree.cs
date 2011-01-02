using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyTree
    {
        private List<DependencyTreeNode> mRootNodes;
        private List<DependencyTreeNode> mAllNodes;
        private Dictionary<string, DependencyTreeNode> mAllNodesMap;

        public PackageInstance Package { get; set; }
        public string Platform { get; set; }
        public List<DependencyInstance> Dependencies { get; set; }

        public List<PackageInstance> GetAllDependencyPackages()
        {
            List<PackageInstance> allDependencyPackages = new List<PackageInstance>();
            foreach (DependencyTreeNode node in mAllNodes)
            {
                allDependencyPackages.Add(node.Package);
            }
            return allDependencyPackages;
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

        public void AddNode(DependencyTreeNode node)
        {
            if (!mAllNodesMap.ContainsKey(node.Name))
            {
                mAllNodesMap.Add(node.Name, node);
            }
        }

        public bool Build()
        {
            Queue<DependencyTreeNode> dependencyQueue = new Queue<DependencyTreeNode>();
            mAllNodesMap = new Dictionary<string, DependencyTreeNode>();
            foreach (DependencyInstance d in Dependencies)
            {
                DependencyTreeNode depNode = new DependencyTreeNode(Platform, d, 1);
                dependencyQueue.Enqueue(depNode);
                mAllNodesMap.Add(depNode.Name, depNode);
            }


            // These are the root nodes of the tree
            mRootNodes = new List<DependencyTreeNode>(mAllNodesMap.Values);

            // Breadth-First 
            while (dependencyQueue.Count > 0)
            {
                DependencyTreeNode node = dependencyQueue.Dequeue();
                List<DependencyTreeNode> newDepNodes = node.Build(this);
                if (newDepNodes != null)
                {
                    foreach (DependencyTreeNode n in newDepNodes)
                        dependencyQueue.Enqueue(n);
                }
                else
                {
                    // Error building dependencies !!
                }
            }

            // Store all dependency nodes in a list
            mAllNodes = new List<DependencyTreeNode>(mAllNodesMap.Values);

            return true;
        }

        // Synchronize dependencies
        public bool Sync()
        {
            bool result = true;

            // Checkout all dependencies
            foreach (DependencyTreeNode depNode in mAllNodes)
            {
                if (!Global.CacheRepo.Update(depNode.Package))
                {
                    // Failed to checkout!
                    result = false;
                    break;
                }
                else
                {
                    if (!depNode.Package.VerifyBeforeExtract())
                    {
                        result = false;
                        break;
                    }
                }
            }

            return result;
        }

        public void SaveInfo(FileDirectoryPath.FilePathAbsolute filepath)
        {
            FileStream stream = new FileStream(filepath.ToString(), FileMode.Create, FileAccess.Write);
            StreamWriter writer = new StreamWriter(stream);

            HashSet<string> register = new HashSet<string>();

            ComparableVersion version = Package.Pom.Versions.GetForPlatform(Platform);
            string versionStr = version != null ? version.ToString() : "?";
            foreach (DependencyTreeNode node in mRootNodes)
                node.SaveInfo(writer, register);
        }

        public void Info()
        {
            foreach (DependencyTreeNode node in mAllNodes)
            {
                Loggy.Add(String.Format("Name                       : {0}", node.Package.Name));
                Loggy.Add(String.Format("Version                    : {0}", node.Package.Version));
            }
        }

        public void Print()
        {
            string indent = "+";
            ComparableVersion version = Package.Pom.Versions.GetForPlatform(Platform);
            string versionStr = version != null ? version.ToString() : "?";
            Loggy.Add(String.Format("{0} {1}, version={2}, type={3}", indent, Package.Name, versionStr, "Root"));
            foreach (DependencyTreeNode node in mRootNodes)
                node.Print(indent);
        }
    }
}


