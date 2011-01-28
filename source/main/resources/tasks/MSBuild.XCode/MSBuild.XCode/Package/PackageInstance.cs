using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using System.Security.Cryptography;
using Ionic.Zip;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public enum ELocation
    {
        Remote,     ///< Remote package repository
        Cache,      ///< Cache package repository (on local machine)
        Local,      ///< Local package, a 'Created' package of the Root
        Target,     ///< Target package, an 'Extracted' package in the target folder of a root package
        Share,      ///< Share package, an 'Extracted' package in the shared package repo
        Root,       ///< Root package
    }

    public partial class PackageInstance
    {
        private PackageResource mResource;

        private string mRootURL = string.Empty;

        private string mBranch;
        private string mPlatform;
        private bool mIsRoot;

        private PomInstance mPom;

        public Group Group { get { return mResource.Group; } }
        public string Name { get { return mResource.Name; } }
        public string Branch { get { return mBranch; } set { mBranch = value; } }
        public string Platform { get { return mPlatform; } set { mPlatform = value; } }

        public bool IsValid { get { return mResource.IsValid; } }

        public DateTime RemoteSignature { get; set; }
        public DateTime CacheSignature { get; set; }
        public DateTime ShareSignature { get; set; }
        public DateTime TargetSignature { get; set; }
        public DateTime LocalSignature { get; set; }

        public IPackageFilename RemoteFilename { get; set; }
        public IPackageFilename CacheFilename { get; set; }
        public IPackageFilename ShareFilename { get; set; }
        public IPackageFilename TargetFilename { get; set; }
        public IPackageFilename LocalFilename { get; set; }

        public ComparableVersion RemoteVersion { get; set; }
        public ComparableVersion CacheVersion { get; set; }
        public ComparableVersion ShareVersion { get; set; }
        public ComparableVersion TargetVersion { get; set; }
        public ComparableVersion LocalVersion { get; set; }
        public ComparableVersion RootVersion { get; set; }

        public bool RemoteExists { get { return !String.IsNullOrEmpty(RemoteURL); } }
        public bool CacheExists { get { return !String.IsNullOrEmpty(CacheURL); } }
        public bool LocalExists { get { return !String.IsNullOrEmpty(LocalURL); } }
        public bool ShareExists { get { return !String.IsNullOrEmpty(ShareURL); } }
        public bool TargetExists { get { return !String.IsNullOrEmpty(TargetURL); } }
        public bool RootExists { get { return !String.IsNullOrEmpty(RootURL); } }

        public string RemoteURL { get; set; }
        public string CacheURL { get; set; }
        public string LocalURL { get; set; }
        public string ShareURL { get; set; }
        public string TargetURL { get; set; }
        public string RootURL { get { return mRootURL; } }

        public bool IsRootPackage { get { return mIsRoot; } }

        public PomInstance Pom { get { return mPom; } }
        public List<DependencyResource> Dependencies { get { return Pom.Dependencies; } }

        public bool IsCpp { get { return Pom.IsCpp; } }
        public bool IsCs { get { return Pom.IsCs; } }

        private PackageInstance(bool isRoot)
        {
            mIsRoot = isRoot;
        }

        internal PackageInstance(PackageResource resource)
        {
            mResource = resource;
        }

        internal PackageInstance(PackageResource resource, PomInstance pom)
        {
            mResource = resource;
            mPom = pom;
        }

        public static PackageInstance From(string name, string group, string branch, string platform)
        {
            PackageResource resource = PackageResource.From(name, group);
            PackageInstance instance = resource.CreateInstance(false);
            instance.mBranch = branch;
            instance.mPlatform = platform;
            return instance;
        }

        public static PackageInstance LoadFromRoot(string dir)
        {
            PackageResource resource = PackageResource.LoadFromFile(dir);
            PackageInstance instance = resource.CreateInstance(true);
            instance.mRootURL = dir;
            return instance;
        }

        public static PackageInstance LoadFromTarget(string dir)
        {
            PackageResource resource = PackageResource.LoadFromFile(dir);
            PackageInstance instance = resource.CreateInstance(false);
            instance.TargetURL = dir;
            return instance;
        }

        public static PackageInstance LoadFromLocal(string rootURL, IPackageFilename filename)
        {
            string subDir = "target\\" + filename.Name + "\\build\\";
            PackageResource resource = PackageResource.LoadFromPackage(rootURL + subDir, filename);
            PackageInstance instance = resource.CreateInstance(false);
            instance.mRootURL = rootURL;
            instance.LocalURL = rootURL + subDir;
            instance.LocalFilename = filename;
            instance.LocalVersion = filename.Version;
            return instance;
        }

        public void SetPlatform(string platform)
        {
            mPlatform = platform;
            RootVersion = Pom.Versions.GetForPlatform(Platform);
        }

        public bool Load()
        {
            // Load it from the best location (Root, Target or Cache)
            if (IsRootPackage)
            {
                PackageResource resource = PackageResource.LoadFromFile(RootURL);
                mPom = resource.CreatePomInstance(true);
                mResource = resource;
                return true;
            }
            else if (TargetExists)
            {
                // Target actually is a dummy, it doesn't contain the content, the actual content
                // is in the Share repository. The Target only contains the full filename .
                if (ShareExists)
                {
                    PackageResource resource = PackageResource.LoadFromFile(ShareURL);
                    mPom = resource.CreatePomInstance(false);
                    mResource = resource;
                    return true;
                }
            }
            else if (CacheExists)
            {
                PackageResource resource = PackageResource.LoadFromPackage(CacheURL, CacheFilename);
                mPom = resource.CreatePomInstance(false);
                mResource = resource;
                return true;
            }
            return false;
        }

        public void SetURL(ELocation location, string url)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteURL = url; break;
                case ELocation.Cache: CacheURL = url; break;
                case ELocation.Local: LocalURL = url; break;
                case ELocation.Share: ShareURL = url; break;
                case ELocation.Target: TargetURL = url; break;
                case ELocation.Root: mRootURL = url; break;
            }
        }

        public void SetFilename(ELocation location, IPackageFilename filename)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteFilename = filename; break;
                case ELocation.Cache: CacheFilename = filename; break;
                case ELocation.Local: LocalFilename = filename; break;
                case ELocation.Share: ShareFilename = filename; break;
                case ELocation.Target: TargetFilename = filename; break;
                case ELocation.Root: break;
            }
        }

        public void SetSignature(ELocation location, DateTime signature)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteSignature = signature; break;
                case ELocation.Cache: CacheSignature = signature; break;
                case ELocation.Local: LocalSignature = signature; break;
                case ELocation.Share: ShareSignature = signature; break;
                case ELocation.Target: TargetSignature = signature; break;
                case ELocation.Root: break;
            }
        }


        public void SetVersion(ELocation location, ComparableVersion version)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteVersion = version; break;
                case ELocation.Cache: CacheVersion = version; break;
                case ELocation.Local: LocalVersion = version; break;
                case ELocation.Share: ShareVersion = version; break;
                case ELocation.Target: TargetVersion = version; break;
                case ELocation.Root: RootVersion = version; break;
            }
        }

        public string GetURL(ELocation location)
        {
            string url = string.Empty;
            switch (location)
            {
                case ELocation.Remote: url = RemoteURL; break;
                case ELocation.Cache: url = CacheURL; break;
                case ELocation.Local: url = LocalURL; break;
                case ELocation.Share: url = ShareURL; break;
                case ELocation.Target: url = TargetURL; break;
                case ELocation.Root: url = RootURL; break;
            }
            return url;
        }

        public bool HasURL(ELocation location)
        {
            bool has = false;
            switch (location)
            {
                case ELocation.Remote: has = RemoteExists; break;
                case ELocation.Cache: has = CacheExists; break;
                case ELocation.Local: has = LocalExists; break;
                case ELocation.Share: has = ShareExists; break;
                case ELocation.Target: has = TargetExists; break;
                case ELocation.Root: has = RootExists; break;
            }
            return has;
        }

        public IPackageFilename GetFilename(ELocation location)
        {
            IPackageFilename filename = null;
            switch (location)
            {
                case ELocation.Remote: filename = RemoteFilename; break;
                case ELocation.Cache: filename = CacheFilename; break;
                case ELocation.Local: filename = LocalFilename; break;
                case ELocation.Share: filename = ShareFilename; break;
                case ELocation.Target: filename = TargetFilename; break;
                case ELocation.Root: break;
            }
            return filename;
        }

        public DateTime GetSignature(ELocation location)
        {
            DateTime signature = DateTime.MinValue;
            switch (location)
            {
                case ELocation.Remote: signature = RemoteSignature; break;
                case ELocation.Cache: signature = CacheSignature; break;
                case ELocation.Local: signature = LocalSignature; break;
                case ELocation.Share: signature = ShareSignature; break;
                case ELocation.Target: signature = TargetSignature; break;
                case ELocation.Root: break;
            }
            return signature;
        }

        public ComparableVersion GetVersion(ELocation location)
        {
            ComparableVersion version = null;
            switch (location)
            {
                case ELocation.Remote: version = RemoteVersion; break;
                case ELocation.Cache: version = CacheVersion; break;
                case ELocation.Local: version = LocalVersion; break;
                case ELocation.Share: version = ShareVersion; break;
                case ELocation.Target: version = TargetVersion; break;
                case ELocation.Root: version = RootVersion; break;
            }
            return version;
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
                        string[] projectDependencies = rootProject.DependsOn.Split(new char[] { ';' }, StringSplitOptions.RemoveEmptyEntries);
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

                                    string[] projectDependencies = project.DependsOn.Split(new char[] { ';' }, StringSplitOptions.RemoveEmptyEntries);
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
            // The format of DependsOn (seperated with a ';'):
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
                        // Variable projectDependency is a full key at thist stage
                        string projectDependency = projectDependenciesQueue.Dequeue();
                        if (!projectsMerged.Contains(projectDependency))
                        {
                            ProjectInstance dependencyProject;
                            if (projectMap.TryGetValue(projectDependency, out dependencyProject))
                            {
                                ///<<<<<<< PROJECT MERGE <<<<<<<<<<
                                rootProject.MergeWithDependencyProject(dependencyProject);


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

        public bool GenerateSolution()
        {
            MsDev.ISolution solution = null;
            if (IsCpp)
                solution = new MsDev.CppSolution(MsDev.CppSolution.EVersion.VS2010);
            else
                solution = new MsDev.CsSolution(MsDev.CsSolution.EVersion.VS2010);

            List<string> projectFilenames = new List<string>();
            foreach (ProjectInstance prj in Pom.Projects)
            {
                string path = prj.Location.Replace("/", "\\");
                path = path.EndWith('\\');
                path = path + prj.Name + prj.Extension;
                projectFilenames.Add(path);

                string[] dependencies = prj.DependsOn.Split(new char[] { ';' }, StringSplitOptions.RemoveEmptyEntries);
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

            string solutionFilename = RootURL + Name + solution.Extension;
            if (solution.Save(solutionFilename, projectFilenames) < 0)
                return false;

            return true;
        }

    }
}