using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode
{
    public class XDependencyTree
    {
        private List<XDepNode> mRootNodes;
        private List<XDepNode> mAllNodes;

        public string Name { get; set; }
        public XVersion Version { get; set; }
        public XPom Package { get; set; }
        public List<XDependency> Dependencies { get; set; }

        public bool Build(string Platform)
        {
            Queue<XDepNode> dependencyQueue = new Queue<XDepNode>();
            Dictionary<string, XDepNode> dependencyFlatMap = new Dictionary<string, XDepNode>();
            foreach (XDependency d in Dependencies)
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
            }

            // Store all dependency nodes in a list
            mAllNodes = new List<XDepNode>();
            foreach (XDepNode node in dependencyFlatMap.Values)
                mAllNodes.Add(node);

            return true;
        }


        public bool Checkout(string Path, string Platform, XPackageRepository localRepo)
        {
            bool result = true;

            // Checkout all dependencies
            foreach (XDepNode depNode in mAllNodes)
            {
                XDependency dependency = depNode.Dependency;

                XPackage package = new XPackage();
                package.Group = new XGroup(dependency.Group);
                package.Name = dependency.Name;
                package.Branch = dependency.GetBranch(Platform);
                package.Version = new XVersion(depNode.Version);
                package.Platform = Platform;

                if (!localRepo.Checkout(package))
                {
                    // Failed to checkout!
                    result = false;
                    break;
                }
            }

            return result;
        }

        private void CollectPlatformInformation(XProject MainProject, XPlatform Platform, string Config)
        {
            var e1 = new { GroupName = "ClCompile", ElementName = "AdditionalIncludeDirectories", Index = 0 };
            var e2 = new { GroupName = "Link", ElementName = "AdditionalLibraryDirectories", Index = 1 };
            var e3 = new { GroupName = "Link", ElementName = "AdditionalDependencies", Index = 2 };

            var el = new[] { e1, e2, e3 };

            XConfig config;
            if (Platform.configs.TryGetValue(Config, out config))
            {
                foreach (var v in el)
                {
                    XElement e = config.FindElement(v.GroupName, v.ElementName);
                    if (e != null)
                    {
                        switch (v.Index)
                        {
                            case 0: MainProject.AddIncludeDir(Platform.Name, Config, e.Value, true, e.Separator); break;
                            case 1: MainProject.AddLibraryDir(Platform.Name, Config, e.Value, true, e.Separator); break;
                            case 2: MainProject.AddLibraryDep(Platform.Name, Config, e.Value, true, e.Separator); break;
                        }
                    }
                }
            }
        }

        public void CollectProjectInformation(string Category, string Platform, string Config)
        {
            XProject mainProject = Package.GetProjectByCategory(Category);
            XPlatform mainPlatform;
            if (mainProject.Platforms.TryGetValue(Platform, out mainPlatform))
            {
                XConfig mainConfig;
                if (mainPlatform.configs.TryGetValue(Config, out mainConfig))
                {
                    foreach (XDepNode node in mAllNodes)
                    {
                        XProject dep_project = node.Package.Pom.GetProjectByCategory(Category);

                        // Prepend with $(SolutionDir)target\package_name\platform\
                        // Where should this be configured ? in the pom.xml ?
                        string includeDir = node.Package.Pom.IncludePath.Replace("${Platform}", Platform).Replace("${Config}", Config);
                        mainProject.AddIncludeDir(Platform, Config, includeDir, true, ";");
                        string libraryDir = node.Package.Pom.LibraryPath.Replace("${Platform}", Platform).Replace("${Config}", Config);
                        mainProject.AddLibraryDir(Platform, Config, libraryDir, true, ";");
                        string libraryDep = node.Package.Pom.LibraryDep.Replace("${Platform}", Platform).Replace("${Config}", Config);
                        mainProject.AddLibraryDep(Platform, Config, libraryDep, true, ";");

                        if (dep_project != null)
                        {
                            XPlatform platform;
                            if (dep_project.Platforms.TryGetValue(Platform, out platform))
                            {
                                CollectPlatformInformation(mainProject, platform, Config);
                            }
                        }
                    }
                }
            }
        }

        public void Print()
        {
            string indent = "+";
            Console.WriteLine(String.Format("{0} {1}, version={2}, type=Main", indent, Name, Version.ToString()));
            foreach (XDepNode node in mRootNodes)
                node.Print(indent);
        }
    }
}


