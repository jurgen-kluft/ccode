using System;
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
        public bool Done { get; set; }
        public int Depth { get; set; }
        public Package Package { get; set; }
        public Dependency Dependency { get; set; }
        public Dictionary<string, DependencyTreeNode> Children { get; set; }

        public DependencyTreeNode(string platform, Dependency dep, int depth)
        {
            Platform = platform;
            Done = false;
            Depth = depth;
            Dependency = dep;
            Children = null;
            Package = null;
        }

        public List<DependencyTreeNode> Build(Dictionary<string, DependencyTreeNode> dependencyFlatMap)
        {
            // Sync remote repo to cache repo which will cache the best version in our cache repo
            // Obtain the package from the cache repo of the best version
            // Get the dependencies of that package and add them as children
            // - Some dependencies already have been processed, maybe resulting in a different best version due to a different branch of version range
            List<DependencyTreeNode> newDepNodes = null;

            if (!Done)
            {
                Package = new Package();
                Package.IsRoot = false;
                Package.RootDir = Global.RootDir;
                Package.LocalURL = Global.RootDir;
                Package.Name = Dependency.Name;
                Package.Group = Dependency.Group;
                Package.Branch = Dependency.GetBranch(Platform);
                Package.Platform = Platform;

                if (!Global.RemoteRepo.Update(Package, Dependency.GetVersionRange(Package.Platform)))
                {
                    if (!Global.CacheRepo.Update(Package, Dependency.GetVersionRange(Package.Platform)))
                    {
                        Loggy.Add(String.Format("Error, Dependency Tree : Build, Node={0}, Group={1}, reason: unable to find package in remote and cache repository", Package.Name, Package.Group.ToString()));
                        Done = true;
                        return newDepNodes;
                    }
                }
                else
                {
                    if (!Global.CacheRepo.Add(Package, ELocation.Remote))
                    {
                        if (!Global.CacheRepo.Update(Package, Dependency.GetVersionRange(Package.Platform)))
                        {
                            Loggy.Add(String.Format("Error, Dependency Tree : Build, Node={0}, Group={1}, reason: unable to add package from remote to cache repository and unable to find a suitable package in the cache repository", Package.Name, Package.Group.ToString()));
                        }
                    }
                }

                if (Package.LoadPom())
                {
                    Package.Pom.OnlyKeepPlatformSpecifics(Platform);

                    newDepNodes = new List<DependencyTreeNode>();

                    Children = new Dictionary<string, DependencyTreeNode>();
                    Dictionary<string, DependencyTreeNode> dependencyTreeMap = Children;
                    foreach (Dependency d in Package.Pom.Dependencies)
                    {
                        DependencyTreeNode depNode;
                        if (!dependencyFlatMap.TryGetValue(d.Name, out depNode))
                        {
                            depNode = new DependencyTreeNode(Platform, d, Depth + 1);
                            newDepNodes.Add(depNode);
                            dependencyTreeMap.Add(depNode.Name, depNode);
                            dependencyFlatMap.Add(depNode.Name, depNode);
                        }
                        else
                        {
                            // Check if we need to process it again, the criteria are:
                            // - If ((Depth + 1) < depNode.Depth)
                            //   - Replace Dependency with this one
                            // - If ((Depth + 1) == depNode.Depth)
                            //   - prefer default branch
                            //   - prefer latest version
                            if (depNode.Depth > (Depth + 1))
                            {
                                // Take this dependency
                                if (depNode.ReplaceDependency(d, Depth + 1))
                                {
                                    // Dependency is modified, we have to process it again
                                    newDepNodes.Add(depNode);
                                }
                            }
                            else if (depNode.Depth == (Depth + 1))
                            {
                                // If merging these dependencies results in a modified dependency then we have to build it again
                                if (depNode.Dependency.Merge(d))
                                {
                                    // Name is still the same
                                    depNode.Depth = Depth + 1;
                                    depNode.Children = null;
                                    depNode.Done = false;

                                    // For now register it as a new DepNode
                                    newDepNodes.Add(depNode);
                                }
                            }
                            else
                            {
                                dependencyTreeMap.Add(depNode.Name, depNode);
                            }
                        }
                    }
                }
                else
                {
                    Loggy.Add(String.Format("Error, Dependency Tree : Build, Node={0}, Group={1}, reason: unable to load pom of package", Package.Name, Package.Group.ToString()));
                }

                Done = true;
                return newDepNodes;
            }
            return newDepNodes;
        }

        public bool ReplaceDependency(Dependency dependency, int depth)
        {
            Dependency = dependency;
            if (!Dependency.IsEqual(dependency))
            {
                bool processAgain = Done;
                Depth = depth;
                Children = null;
                Done = false;
                return processAgain;
            }
            return false;
        }

        public void Print(string indent)
        {
            if (String.IsNullOrEmpty(indent)) indent = "+";
            else if (indent == "+") indent = "|----+";
            else indent = "     " + indent;

            string versionStr = Package.Version == null ? "?" : Package.Version.ToString();

            Loggy.Add(String.Format("{0} {1}, version={2}, type={3}", indent, Name, versionStr, Dependency.Type));

            if (Children != null)
            {
                foreach (DependencyTreeNode child in Children.Values)
                    child.Print(indent);
            }
        }

    }
}
