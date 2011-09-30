using System;
using System.IO;
using System.Collections.Generic;
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
            ProgressTracker progress = ProgressTracker.Instance;

            progress.Add(new int[] { 75, 20, 5 });
            if (dependencyTree.Update(this) == -1)
            {
                progress.Next();
                progress.Next();
                progress.Next();
                return -1;
            }

            progress.Next();
            progress.ToConsole();

            if (!Package.Load())
            {
                progress.Next();
                progress.Next();
                return -1;
            }

            progress.Next();
            progress.ToConsole();

            Package.Pom.OnlyKeepPlatformSpecifics(Platform);

            foreach (DependencyResource dependencyResource in Package.Dependencies)
            {
                if (dependencyResource.IsForPlatform(Platform))
                {
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
            }

            progress.Next();
            progress.ToConsole();

            return 0;
        }

        public void SaveInfo(StreamWriter writer, HashSet<string> register)
        {
            if (!register.Contains(Name))
            {
                register.Add(Name);

                PackageState p = Package.Package;

                string versionStr = p.TargetVersion == null ? "?" : p.TargetVersion.ToString();
                writer.WriteLine(String.Format("name={0}, group={1}, language={2}, version={3}, type={4}", Package.Name, Package.Group.ToString(), Package.Language, versionStr, Dependency.Type));

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

            PackageState p = Package.Package;

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
