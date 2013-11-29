using System;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public partial class PackageInstance
    {
        private PackageResource mResource;
        private PomInstance mPom;
        private PackageState mPackage;

        public Group Group { get { return mResource.Group; } }
        public string Name { get { return mResource.Name; } }
        public string Branch { get { return mPackage.Branch; } set { mPackage.Branch = value; } }

        public string Language
        {
            get
            {
				if (IsMixed)  return "Mixed";
                else if (IsCpp) return "C++";
                else if (IsCs) return "C#";
                else return "C++";
            }
        }

        public string IDE { set; get; }
        public string ToolSet { set; get; }

        public bool IsValid { get { return mResource.IsValid; } }

        public string RootURL { get; set; }
        public bool IsRootPackage { get; private set; }

        public PomInstance Pom { get { return mPom; } }
        public PackageState Package { get { return mPackage; } }
        public List<DependencyResource> Dependencies { get { return Pom.Dependencies; } }

		public bool IsMixed { get { return Pom.IsCpp && Pom.IsCs; } }
		public bool IsCpp { get { return Pom.IsCpp; } }
        public bool IsCs { get { return Pom.IsCs; } }

        private PackageInstance(bool isRoot)
        {
            IDE = "vs2012";
            ToolSet = "v110";
            IsRootPackage = isRoot;
            mPackage = new PackageState();
        }

        internal PackageInstance(bool isRoot, PackageResource resource)
            : this(isRoot)
        {
            mResource = resource;
            InitPackage();
        }

        internal PackageInstance(bool isRoot, PackageResource resource, PomInstance pom)
            : this(isRoot)
        {
            mResource = resource;
            mPom = pom;
            mPom.Package = this;
            InitPackage();
        }

        public static PackageInstance From(bool isRoot, string name, string group, string branch, string platform)
        {
            PackageResource resource = PackageResource.From(name, group);
            PackageInstance instance = resource.CreateInstance(isRoot);
            instance.Branch = branch;
            instance.SetPlatform(platform);
            instance.InitPackage();
            return instance;
        }

        public static PackageInstance LoadFromRoot(string dir)
        {
            PackageResource resource = PackageResource.LoadFromFile(dir);
            PackageInstance instance = resource.CreateInstance(true);
            instance.RootURL = dir;

            instance.InitPackage();
            instance.Package.RootURL = dir;

            if (resource.Platforms.Count > 0)
            {
                string platform = resource.Platforms[0];
                instance.Package.Platform = platform;
                instance.Package.RootVersion = instance.Pom.Versions.GetForPlatform(platform);
            }

            return instance;
        }

        public static PackageInstance LoadFromTarget(string dir)
        {
            PackageResource resource = PackageResource.LoadFromFile(dir);
            PackageInstance instance = resource.CreateInstance(false);
            instance.InitPackage();
            instance.Package.TargetURL = dir;
            
            if (resource.Platforms.Count > 0)
            {
                string platform = resource.Platforms[0];
                instance.Package.Platform = platform;
                instance.Package.RootVersion = instance.Pom.Versions.GetForPlatform(platform);
            }
            return instance;
        }

        private void InitPackage()
        {
            mPackage.Name = Name;
            mPackage.Group = Group.ToString();
            mPackage.Branch = Branch;
            if (String.IsNullOrEmpty(mPackage.Platform))
                mPackage.Platform = "?";
            mPackage.Language = Language;
        }

        public bool HasPlatform(string platform)
        {
            if (!ContainsPlatform(Pom.Platforms.ToArray(), platform))
                return false;
            
            return true;
        }

        public void SetPlatform(string platform)
        {
            mPackage.Platform = platform;
            mPackage.RootVersion = Pom.Versions.GetForPlatform(platform);
        }

        public bool Load()
        {
            // Load it from the best location (Root, Target or Cache)
            if (IsRootPackage)
            {
                PackageResource resource = PackageResource.LoadFromFile(RootURL);
                mPom = resource.CreatePomInstance(true);
                mPom.Package = this;
                mResource = resource;
                return true;
            }

            if (mPackage.TargetExists)
            {
                // Target actually is a dummy, it doesn't contain the content, the actual content
                // is in the Share repository. The Target only contains the full filename .
                if (mPackage.ShareExists)
                {
                    PackageResource resource = PackageResource.LoadFromFile(mPackage.ShareURL);
                    mPom = resource.CreatePomInstance(false);
                    mPom.Package = this;
                    mResource = resource;
                    return true;
                }
            }
            else if (mPackage.CacheExists)
            {
                PackageResource resource = PackageResource.LoadFromPackage(mPackage.CacheURL, mPackage.CacheFilename);
                mPom = resource.CreatePomInstance(false);
                mPom.Package = this;
                mResource = resource;
                return true;
            }
            return false;
        }

        public bool Info()
        {
            return Pom.Info();
        }

        private static bool ContainsPlatform(string[] platforms, string platform)
        {
            foreach (string p in platforms)
                if (String.Compare(p, platform, true) == 0)
                    return true;
            return false;
        }

        public void GenerateProjects(PackageDependencies dependencies, List<string> platforms)
        {
            // Generate the project .props file for every platform
            foreach (string platform in platforms)
            {
                List<PackageInstance> packageInstances = dependencies.GetAllDependencyPackages(platform);
                packageInstances.Add(this);

                foreach (PackageInstance pi in packageInstances)
                {
                    bool isRoot = pi.IsRootPackage;
                    string path = RootURL.EndWith('\\') + "target\\" + pi.Name + "\\";
                    string filename = path + pi.Name + "." + platform + ".props";
                    string type = isRoot ? "Root" : "Package";
                    string location = isRoot ? pi.Package.RootURL : pi.Package.ShareURL;
                    pi.Pom.ProjectProperties.Write(filename, platform, type, location);
                }
            }

            // Generating project files is a bit complex in that it has to merge project definitions on a per platform basis.
            // Every platform is considered to have its own package (zip -> pom.xml) containing the project settings for that platform.
            // For every platform we have to merge in only the conditional xml elements into the final project file.
            foreach (ProjectInstance rootProject in Pom.Projects)
                rootProject.ConstructFullMsDevProject(platforms);

            // First, build all the dictionaries:
            // Dictionary<PackageName,PackageInstance>
            // Dictionary<ProjectName,ProjectInstance>
            // Stack<ProjectInstance> dependencyProjects;

            // PackageInstances and ProjectInstances are registered with a key like:
            //      Key = platform:package_name:project_name
            Dictionary<string, PackageInstance> packageMap = new Dictionary<string, PackageInstance>();
            Dictionary<string, ProjectInstance> projectMap = new Dictionary<string, ProjectInstance>();
            Dictionary<string, List<string>> projectDependsOnMap = new Dictionary<string, List<string>>();
            Dictionary<string, bool> projectDependenciesToResolve = new Dictionary<string, bool>();

            // Add ourselves in the package map
            packageMap.Add(Name.ToLower(), this);

            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                // Also put the root projects in the map for every platform
                foreach (string platform in platforms)
                {
                    string key = (platform + ":" + Name + ":" + rootProject.Name).ToLower();
                    if (!projectMap.ContainsKey(key))
                    {
                        projectMap.Add(key, rootProject);

                        // Store the root project DependsOn items and also schedule them for resolving them
                        string[] projectDependencies = rootProject.DependsOn;
                        for (int i = 0; i < projectDependencies.Length; ++i)
                        {
                            if (!projectDependencies[i].Contains(":"))
                                projectDependencies[i] = projectDependencies[i] + ":" + projectDependencies[i];

                            projectDependencies[i] = (platform + ":" + projectDependencies[i]).ToLower();

                            // See if this 'dependency' is part of this platform
                            string[] platform_package_project = projectDependencies[i].Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);
                            if (dependencies.IsDependencyForPlatform(platform_package_project[1], platform_package_project[0]))
                            {
                                if (!projectDependenciesToResolve.ContainsKey(projectDependencies[i]))
                                    projectDependenciesToResolve.Add(projectDependencies[i], true);
                            }
                        }

                        // Store the DependsOn array for future use
                        projectDependsOnMap.Add(key, new List<string>(projectDependencies));
                    }

                    string[] projectPlatforms = rootProject.GetPlatforms();
                    if (ContainsPlatform(projectPlatforms, platform))
                    {
                        List<PackageInstance> allDependencyPackages = dependencies.GetAllDependencyPackages(platform);
                        foreach (PackageInstance dependencyPackage in allDependencyPackages)
                        {
                            if (!packageMap.ContainsKey(dependencyPackage.Name.ToLower()))
                                packageMap.Add(dependencyPackage.Name.ToLower(), dependencyPackage);

                            foreach (ProjectInstance project in dependencyPackage.Pom.Projects)
                            {
                                string projectKey = (platform + ":" + dependencyPackage.Name + ":" + project.Name).ToLower();
                                if (!projectMap.ContainsKey(projectKey))
                                {
                                    projectMap.Add(projectKey, project);

                                    string[] projectDependencies = project.DependsOn;
                                    for (int i = 0; i < projectDependencies.Length; ++i)
                                    {
                                        if (!projectDependencies[i].Contains(":"))
                                            projectDependencies[i] = projectDependencies[i] + ":" + projectDependencies[i];

                                        projectDependencies[i] = (platform + ":" + projectDependencies[i]).ToLower();

                                        // See if this 'dependency' is part of this platform
                                        string[] platform_package_project = projectDependencies[i].Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);
                                        if (dependencies.IsDependencyForPlatform(platform_package_project[1], platform_package_project[0]))
                                        {
                                            if (!projectDependenciesToResolve.ContainsKey(projectDependencies[i]))
                                                projectDependenciesToResolve.Add(projectDependencies[i], true);
                                        }
                                    }

                                    // Store the DependsOn array for future use
                                    projectDependsOnMap.Add(projectKey, new List<string>(projectDependencies));
                                }
                            }
                        }
                    }
                }
            }

            foreach (string projectDependency in projectDependenciesToResolve.Keys)
            {
                string[] platform_package_project = projectDependency.Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);

                PackageInstance packageInstance;
                if (packageMap.TryGetValue(platform_package_project[1], out packageInstance))
                {
                    ProjectInstance project = packageInstance.Pom.GetProjectByName(platform_package_project[2]);
                    if (project != null)
                    {
                        string key = (platform_package_project[0] + ":" + packageInstance.Name + ":" + project.Name).ToLower();
                        if (!projectMap.ContainsKey(key))
                            projectMap.Add(key, project);
                    }
                    else
                    {
                        // Error; unable to find the project in the dependency package
                        Loggy.Error(String.Format("Error: unable to find project {0} in package {1} for platform {2}", platform_package_project[2], platform_package_project[1], platform_package_project[0]));
                    }
                }
                else
                {
                    // Error; unable to find the dependency package
                    Loggy.Error(String.Format("Error: unable to find package {0} for platform {1}", platform_package_project[1], platform_package_project[0]));
                }
            }

            // We need to change the way things are merged, from now we only use 'DependsOn' and we do not magically
            // merge groups anymore. Group is to be removed from the project settings.
            // The format of DependsOn (separated with a ';'):
            // - 'PackageNameA:ProjectNameA';'PackageNameA:ProjectNameB';'PackageNameB:ProjectNameA'

            // Second, now that we have the dictionaries setup, we can start to merge dependency projects
            // 1. Merge with the dependency projects
            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                Queue<ProjectInstance> dependencyProjects = new Queue<ProjectInstance>();
                {
                    Queue<string> projectDependenciesQueue = new Queue<string>();
                    HashSet<string> projectsMerged = new HashSet<string>();

                    foreach (string platform in platforms)
                    {
                        List<string> dependsOn;
                        string key = (platform + ":" + Name + ":" + rootProject.Name).ToLower();
                        if (projectDependsOnMap.TryGetValue(key, out dependsOn))
                        {
                            foreach (string dependencyProjectDependency in dependsOn)
                            {
                                string[] platform_package_project = dependencyProjectDependency.Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);
                                if (dependencies.IsDependencyForPlatform(platform_package_project[1], platform_package_project[0]))
                                {
                                    if (!projectsMerged.Contains(dependencyProjectDependency))
                                        projectDependenciesQueue.Enqueue(dependencyProjectDependency);
                                }
                            }
                        }
                    }

                    while (projectDependenciesQueue.Count > 0)
                    {
                        // Variable projectDependency is a full key at this stage
                        string projectDependency = projectDependenciesQueue.Dequeue();
                        if (!projectsMerged.Contains(projectDependency))
                        {
                            ProjectInstance dependencyProject;
                            if (projectMap.TryGetValue(projectDependency, out dependencyProject))
                            {
                                ///<<<<<<< PROJECT MERGE <<<<<<<<<<
                                rootProject.MergeWithDependencyProject(dependencyProject);

#if WRITE_PROJECT_PROPS_FILE_HERE
                                // Write the project properties .props file of this dependency project
                                foreach (string platform in platforms)
                                {
                                    Package dependencyPackage = dependencyProject.Pom.Package.GetPackageForPlatform(platform);
                                    bool isRoot = dependencyProject.Pom.Package.IsRootPackage;

                                    string name = dependencyProject.Pom.Name;
                                    string packageLocation = isRoot ? dependencyPackage.RootURL : dependencyPackage.ShareURL;
                                    string path = RootURL.EndWith('\\') + "target\\" + name + "\\";
                                    string filename = path + name + "." + platform + ".props";
                                    string type = isRoot ? "Root" : "Package";
                                    dependencyProject.Pom.ProjectProperties.Write(filename, platform, type, packageLocation);
                                }
#endif

                                /// Now queue all the dependencies of this dependencyProject, since
                                /// we also have to merge them and also the dependencies of those
                                /// dependencies etc..
                                projectsMerged.Add(projectDependency);
                                List<string> dependsOn;
                                if (projectDependsOnMap.TryGetValue(projectDependency, out dependsOn))
                                {
                                    foreach (string dependencyProjectDependency in dependsOn)
                                    {
                                        // See if this 'dependency' is part of this platform
                                        string[] platform_package_project = dependencyProjectDependency.Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);
                                        if (dependencies.IsDependencyForPlatform(platform_package_project[1], platform_package_project[0]))
                                        {
                                            if (!projectsMerged.Contains(dependencyProjectDependency))
                                                projectDependenciesQueue.Enqueue(dependencyProjectDependency);
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }


            foreach (ProjectInstance p in Pom.Projects)
            {
                string path = p.Location.Replace("/", "\\");
                path = path.EndWith('\\');
                string filename = path + p.Name + p.Extension;
                p.Save(RootURL, filename);
            }
        }

        public bool GenerateSolution(List<string> platforms)
        {
            MsDev.ISolution solution = null;
			solution = new MsDev.MixedSolution(MsDev.MixedSolution.EVersion.VS2010);

            List<string> projectFilenames = new List<string>();
            foreach (ProjectInstance prj in Pom.Projects)
            {
                bool sln_should_include_project = false;
                foreach (string platform in platforms)
                {
                    if (prj.HasPlatform(platform))
                    {
                        sln_should_include_project = true;
                        break;
                    }
                }
                if (sln_should_include_project)
                {
                    string path = prj.Location.Replace("/", "\\");
                    path = path.EndWith('\\');
                    path = path + prj.Name + prj.Extension;

                    // Add it to the list of projects that should be included in the solution
                    projectFilenames.Add(path);
					solution.SetProjectLanguage(prj.Name + prj.Extension, prj.Language);

                    string[] dependencies = prj.DependsOn;
                    if (dependencies.Length > 0)
                    {
                        for (int i = 0; i < dependencies.Length; ++i)
                        {
                            string[] package_project = dependencies[i].Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);
                            if (package_project.Length == 0)
                                dependencies[i] = package_project[0] + prj.Extension;
                            else
                                dependencies[i] = package_project[1] + prj.Extension;
                        }
                        solution.AddDependencies(prj.Name + prj.Extension, dependencies);
                    }
                }
            }

            string solutionFilename = RootURL + Name + solution.Extension;
            if (solution.Save(solutionFilename, projectFilenames) < 0)
                return false;

            return true;
        }
    }
}