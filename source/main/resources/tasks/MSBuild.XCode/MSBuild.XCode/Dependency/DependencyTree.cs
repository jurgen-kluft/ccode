using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyTree
    {
        private List<XDepNode> mRootNodes;
        private List<XDepNode> mAllNodes;

        public string Name { get; set; }
        public ComparableVersion Version { get; set; }
        public Pom Package { get; set; }
        public List<Dependency> Dependencies { get; set; }

        public bool Build(string Platform)
        {
            Queue<XDepNode> dependencyQueue = new Queue<XDepNode>();
            Dictionary<string, XDepNode> dependencyFlatMap = new Dictionary<string, XDepNode>();
            foreach (Dependency d in Dependencies)
            {
                XDepNode depNode = new XDepNode(d, 1);
                dependencyQueue.Enqueue(depNode);
                dependencyFlatMap.Add(depNode.Name, depNode);
            }

            // These are the root nodes of the tree
            mRootNodes = new List<XDepNode>();
            foreach (XDepNode node in dependencyFlatMap.Values)
                mRootNodes.Add(node);

            // Breadth-First 
            while (dependencyQueue.Count > 0)
            {
                XDepNode node = dependencyQueue.Dequeue();
                List<XDepNode> newDepNodes = node.Build(dependencyFlatMap, Platform);
                if (newDepNodes != null)
                {
                    foreach (XDepNode n in newDepNodes)
                        dependencyQueue.Enqueue(n);
                }
                else
                {
                    // Error building dependencies !!
                }
            }

            // Store all dependency nodes in a list
            mAllNodes = new List<XDepNode>();
            foreach (XDepNode node in dependencyFlatMap.Values)
                mAllNodes.Add(node);

            return true;
        }

        // Synchronize dependencies
        public bool Sync(string Platform, PackageRepository localRepo)
        {
            bool result = true;

            // Checkout all dependencies
            foreach (XDepNode depNode in mAllNodes)
            {
                if (!localRepo.Update(depNode.Package))
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
            foreach (XDepNode node in mRootNodes)
                node.Print(indent);
        }
    }
}


