﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

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

        private void CollectPlatformInformation(Project MainProject, Platform Platform, string Config)
        {
            var e1 = new { GroupName = "ClCompile", ElementName = "PreprocessorDefinitions", Index = 0 };
            var e2 = new { GroupName = "ClCompile", ElementName = "AdditionalIncludeDirectories", Index = 1 };
            var e3 = new { GroupName = "Link", ElementName = "AdditionalLibraryDirectories", Index = 2 };
            var e4 = new { GroupName = "Link", ElementName = "AdditionalDependencies", Index = 3 };

            var el = new[] { e1, e2, e3, e4 };

            Config config;
            if (Platform.configs.TryGetValue(Config, out config))
            {
                foreach (var v in el)
                {
                    Element e = config.FindElement(v.GroupName, v.ElementName);
                    if (e != null)
                    {
                        switch (v.Index)
                        {
                            case 0: MainProject.AddPreprocessorDefinitions(Platform.Name, Config, e.Value, true, e.Separator); break;
                            case 1: MainProject.AddIncludeDir(Platform.Name, Config, e.Value, true, e.Separator); break;
                            case 2: MainProject.AddLibraryDir(Platform.Name, Config, e.Value, true, e.Separator); break;
                            case 3: MainProject.AddLibraryDep(Platform.Name, Config, e.Value, true, e.Separator); break;
                        }
                    }
                }
            }
        }

        public void CollectProjectInformation(string Category, string Platform, string Config)
        {
            Project mainProject = Package.GetProjectByCategory(Category);
            Platform mainPlatform;
            if (mainProject.Platforms.TryGetValue(Platform, out mainPlatform))
            {
                Config mainConfig;
                if (mainPlatform.configs.TryGetValue(Config, out mainConfig))
                {
                    foreach (XDepNode node in mAllNodes)
                    {
                        Project dep_project = node.Package.Pom.GetProjectByCategory(Category);

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
                            Platform platform;
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

