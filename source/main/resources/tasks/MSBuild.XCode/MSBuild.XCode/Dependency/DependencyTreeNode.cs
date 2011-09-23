using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyTreeNode
    {
        public string Name { get { return Dependency.Name; } }
        public string Platform { get; set; }
        public int Depth { get; set; }
        public int Iteration { get; set; }
        public PackageInstance Package { get; set; }
        public DependencyInstance Dependency { get; set; }
        public Dictionary<string, DependencyTreeNode> Children { get; set; }

        public DependencyTreeNode(string platform, DependencyInstance dep, int depth)
        {
            Platform = platform;
            Depth = depth;
            Iteration = -1;
            Dependency = dep;
            Children = null;

            Package = PackageInstance.From(false, Dependency.Name, Dependency.Group.ToString(), Dependency.Branch, Platform);
            Children = new Dictionary<string, DependencyTreeNode>();
        }

        public int Compile(DependencyTree dependencyTree)
        {
            if (dependencyTree.Update(this) == -1)
                return -1;
            if (!Package.Load())
                return -1;

            Package.Pom.OnlyKeepPlatformSpecifics(Platform);

            foreach (DependencyResource dependencyResource in Package.Dependencies)
            {
                if (!dependencyResource.IsForPlatform(Platform))
                    continue;

                if (!dependencyTree.HasNode(dependencyResource.Name))
                {
                    DependencyInstance dependencyInstance = new DependencyInstance(Platform, dependencyResource);
                    DependencyTreeNode depNode = new DependencyTreeNode(Platform, dependencyInstance, Depth + 1);

                    Children.Add(depNode.Name, depNode);
                    dependencyTree.AddNode(depNode);
                    dependencyTree.EnqueueForCompile(depNode);
                }
                else
                {
                    DependencyTreeNode depNode = dependencyTree.FindNode(dependencyResource.Name);
                    Children.Add(depNode.Name, depNode);
                }
            }

            return 0;
        }

        public void SaveInfo(StreamWriter writer, HashSet<string> register)
        {
            if (!register.Contains(Name))
            {
                register.Add(Name);

                Package p = Package.Package;

                string versionStr = p.TargetVersion == null ? "?" : p.TargetVersion.ToString();
                writer.WriteLine(String.Format("{0}, version={1}, type={2}", Name, versionStr, Dependency.Type));

                if (Children != null)
                {
                    foreach (DependencyTreeNode child in Children.Values)
                        child.SaveInfo(writer, register);
                }
            }
        }

        public void Print(string indent)
        {
            if (String.IsNullOrEmpty(indent)) indent = "+";
            else if (indent == "+") indent = "|----+";
            else indent = "     " + indent;

            Package p = Package.Package;

            string versionStr = p.TargetVersion == null ? "?" : p.TargetVersion.ToString();

            Loggy.Info(String.Format("{0}{1}, version={2}, type={3}", indent, Name, versionStr, Dependency.Type));

            if (Children != null)
            {
                foreach (DependencyTreeNode child in Children.Values)
                    child.Print(indent);
            }
        }

    }
}
