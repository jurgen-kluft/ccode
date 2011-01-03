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
        public bool Done { get; set; }
        public int Depth { get; set; }
        public PackageInstance Package { get; set; }
        public DependencyInstance Dependency { get; set; }
        public Dictionary<string, DependencyTreeNode> Children { get; set; }

        public DependencyTreeNode(string platform, DependencyInstance dep, int depth)
        {
            Platform = platform;
            Done = false;
            Depth = depth;
            Dependency = dep;
            Children = null;
            Package = null;
        }

        public List<DependencyTreeNode> Build(DependencyTree dependencyTree)
        {
            // Sync remote repository to cache repository which will cache the best version in our cache repository
            // Obtain the package from the cache repository of the best version
            // Get the dependencies of that package and add them as children
            // - Some dependencies already have been processed, maybe resulting in a different best version due to a different branch of version range
            List<DependencyTreeNode> unprocessedDependencyNodes = null;

            if (!Done)
            {
                Package = PackageInstance.From(Dependency.Name, Dependency.Group.ToString(), Dependency.Branch, Platform);

                // @TODO: This procedure should be:
                //    Negotiate with cache
                //    If not found in cache
                //       Negotiate with remote
                //       If not found at remote
                //           return error
                //       Endif
                //       Cache it
                //    Else
                //       Negotiate with remote if it has a better version than we found in the cache (Version and VersionRange)
                //       If remote has better version
                //           Cache it
                //       Endif
                //    Endif

                if (!Global.RemoteRepo.Update(Package, Dependency.VersionRange))
                {
                    if (!Global.CacheRepo.Update(Package, Dependency.VersionRange))
                    {
                        Loggy.Add(String.Format("Error, Dependency Tree : Build, Node={0}, Group={1}, reason: unable to find package in remote and cache repository", Package.Name, Package.Group.ToString()));
                        Done = true;
                        return unprocessedDependencyNodes;
                    }
                }
                else
                {
                    if (!Global.CacheRepo.Add(Package, ELocation.Remote))
                    {
                        if (!Global.CacheRepo.Update(Package, Dependency.VersionRange))
                        {
                            Loggy.Add(String.Format("Error, Dependency Tree : Build, Node={0}, Group={1}, reason: unable to add package from remote to cache repository and unable to find a suitable package in the cache repository", Package.Name, Package.Group.ToString()));
                        }
                    }
                }

                if (Package.Load())
                {
                    Package.Pom.OnlyKeepPlatformSpecifics(Platform);

                    unprocessedDependencyNodes = new List<DependencyTreeNode>();

                    Children = new Dictionary<string, DependencyTreeNode>();
                    Dictionary<string, DependencyTreeNode> dependencyTreeMap = Children;
                    foreach (DependencyResource dependencyResource in Package.Dependencies)
                    {
                        if (!dependencyTree.HasNode(dependencyResource.Name))
                        {
                            DependencyInstance dependencyInstance = new DependencyInstance(Platform, dependencyResource);
                            DependencyTreeNode depNode = new DependencyTreeNode(Platform, dependencyInstance, Depth + 1);
                            unprocessedDependencyNodes.Add(depNode);
                            dependencyTreeMap.Add(depNode.Name, depNode);
                            dependencyTree.AddNode(depNode);
                        }
                        else
                        {
                            DependencyTreeNode depNode = dependencyTree.FindNode(dependencyResource.Name);

                            // Check if we need to process it again, the criteria are:
                            // - If ((Depth + 1) < depNode.Depth)
                            //   - Replace Dependency with this one
                            // - If ((Depth + 1) == depNode.Depth)
                            //   - prefer default branch
                            //   - prefer latest version
                            if (depNode.Depth > (Depth + 1))
                            {
                                // Take this dependency
                                DependencyInstance dependencyInstance = new DependencyInstance(Platform, dependencyResource);
                                if (depNode.ReplaceDependency(dependencyInstance, Depth + 1))
                                {
                                    // Dependency is modified, we have to process it again
                                    unprocessedDependencyNodes.Add(depNode);
                                }
                            }
                            else if (depNode.Depth == (Depth + 1))
                            {
                                // If merging these dependencies results in a modified dependency then we have to build it again
                                DependencyInstance dependencyInstance = new DependencyInstance(Platform, dependencyResource);
                                if (depNode.Dependency.Merge(dependencyInstance))
                                {
                                    // Name is still the same
                                    depNode.Depth = Depth + 1;
                                    depNode.Children = null;
                                    depNode.Done = false;

                                    // For now register it as a new DepNode
                                    unprocessedDependencyNodes.Add(depNode);
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
                return unprocessedDependencyNodes;
            }
            return unprocessedDependencyNodes;
        }

        public bool ReplaceDependency(DependencyInstance dependency, int depth)
        {
            if (!Dependency.IsEqual(dependency))
            {
                Dependency = dependency;
                bool processAgain = Done;
                Depth = depth;
                Children = null;
                Done = false;
                return processAgain;
            }
            return false;
        }

        public void SaveInfo(StreamWriter writer, HashSet<string> register)
        {
            if (!register.Contains(Name))
            {
                register.Add(Name);

                string versionStr = Package.Version == null ? "?" : Package.Version.ToString();
                writer.WriteLine(String.Format("{0} {1}, version={2}, type={3}", Name, versionStr, Dependency.Type));

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
