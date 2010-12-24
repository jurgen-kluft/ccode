using System;
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

        public string Name { get; set; }
        public ComparableVersion Version { get; set; }
        public Pom Package { get; set; }
        public List<Dependency> Dependencies { get; set; }

        public bool Build(string Platform)
        {
            Queue<DependencyTreeNode> dependencyQueue = new Queue<DependencyTreeNode>();
            Dictionary<string, DependencyTreeNode> dependencyFlatMap = new Dictionary<string, DependencyTreeNode>();
            foreach (Dependency d in Dependencies)
            {
                DependencyTreeNode depNode = new DependencyTreeNode(d, 1);
                dependencyQueue.Enqueue(depNode);
                dependencyFlatMap.Add(depNode.Name, depNode);
            }

            // These are the root nodes of the tree
            mRootNodes = new List<DependencyTreeNode>();
            foreach (DependencyTreeNode node in dependencyFlatMap.Values)
                mRootNodes.Add(node);

            // Breadth-First 
            while (dependencyQueue.Count > 0)
            {
                DependencyTreeNode node = dependencyQueue.Dequeue();
                List<DependencyTreeNode> newDepNodes = node.Build(dependencyFlatMap, Platform);
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
            mAllNodes = new List<DependencyTreeNode>();
            foreach (DependencyTreeNode node in dependencyFlatMap.Values)
                mAllNodes.Add(node);

            return true;
        }

        // Synchronize dependencies
        public bool Sync(string Platform)
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

        public void Print()
        {
            string indent = "+";
            Loggy.Add(String.Format("{0} {1}, version={2}, type=Main", indent, Name, Version.ToString()));
            foreach (DependencyTreeNode node in mRootNodes)
                node.Print(indent);
        }
    }
}


