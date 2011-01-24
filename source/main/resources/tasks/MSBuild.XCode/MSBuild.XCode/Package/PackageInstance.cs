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

        private static bool ContainsPlatform(List<string> platforms, string platform)
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

            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                // Every platform has a dependency tree and every dependency package for that platform has filtered their
                // project to only keep their platform specific xml elements.

                string[] projectPlatforms = rootProject.GetPlatforms();
                foreach (string projectPlatform in projectPlatforms)
                {
                    if (ContainsPlatform(platforms, projectPlatform))
                    {
                        List<PackageInstance> allDependencyPackages = dependencies.GetAllDependencyPackages(projectPlatform);

                        // Merge in all Projects of those dependency packages which are already filtered on the platform
                        foreach (PackageInstance dependencyPackage in allDependencyPackages)
                        {
                            ProjectInstance dependencyProject = dependencyPackage.Pom.GetProjectByGroup(rootProject.Group);
                            if (dependencyProject != null && !dependencyProject.IsPrivate)
                                rootProject.MergeWithDependencyProject(dependencyProject);
                        }
                    }
                }
            }

            // And the root package UnitTest project generally merges with Main, since UnitTest will use Main.
            // Although the user could specify more projects and different internal project dependencies with 'ProjectName'.
            // It is also possible to define external project references, these are specified by 'PackageName:ProjectGroup'.
            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                if (!String.IsNullOrEmpty(rootProject.DependsOn))
                {
                    string[] projectDependencies = rootProject.DependsOn.Split(new char[] { ';' }, StringSplitOptions.RemoveEmptyEntries);

                    // External project dependencies (Format is "PackageName:ProjectGroup;PackageName:ProjectGroup;PackageName:ProjectGroup;etc..")
                    foreach (string dependency in projectDependencies)
                    {
                        if (!dependency.Contains(":")) ///< This is not an external project reference
                            continue;

                        // Here the reference is PackageName:ProjectGroup
                        string[] package_projectgroup = dependency.Split(new char[] { ':' }, StringSplitOptions.RemoveEmptyEntries);
                        if (package_projectgroup.Length == 2 && !String.IsNullOrEmpty(package_projectgroup[0]) && !String.IsNullOrEmpty(package_projectgroup[1]))
                        {
                            string[] projectPlatforms = rootProject.GetPlatforms();
                            foreach (string projectPlatform in projectPlatforms)
                            {
                                if (ContainsPlatform(platforms, projectPlatform))
                                {
                                    List<PackageInstance> allDependencyPackages = dependencies.GetAllDependencyPackages(projectPlatform);
                                    foreach (PackageInstance dependencyPackage in allDependencyPackages)
                                    {
                                        if (String.Compare(dependencyPackage.Name, package_projectgroup[0], true) == 0)
                                        {
                                            ProjectInstance dependencyProject = dependencyPackage.Pom.GetProjectByGroup(package_projectgroup[1]);
                                            if (dependencyProject != null)
                                                rootProject.MergeWithDependencyProject(dependencyProject);
                                        }
                                    }
                                }
                            }
                        }
                        else
                        {
                            Loggy.Error(String.Format("Error: PackageInstance:GenerateProjects, invalid project reference {0}", dependency));
                        }
                    }
                }
            }

            // And the root package UnitTest project generally merges with Main, since UnitTest will use Main.
            // Although the user could specify more project and different internal dependencies.
            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                if (!String.IsNullOrEmpty(rootProject.DependsOn))
                {
                    string[] projectDependencies = rootProject.DependsOn.Split(new char[] { ';' }, StringSplitOptions.RemoveEmptyEntries);

                    // Internal pom project dependencies (Format is "ProjectName;ProjectName;etc..")
                    foreach (string dependency in projectDependencies)
                    {
                        if (dependency.Contains(":")) ///< This is an external project reference
                            continue;

                        ProjectInstance dependencyProject = Pom.GetProjectByName(dependency);
                        if (dependencyProject != null)
                            rootProject.MergeWithDependencyProject(dependencyProject);
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
                        dependencies[i] = dependencies[i] + prj.Extension;
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