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
        Target,     ///< Target package, an 'Extracted' package
        Root,       ///< Root package
    }

    public class PackageInstance
    {
        private PackageResource mResource;

        private string mRootURL = string.Empty;
        private string mTargetURL = string.Empty;

        private string mBranch;
        private string mPlatform;

        private PomInstance mPom;

        public Group Group { get { return mResource.Group; } }
        public string Name { get { return mResource.Name; } }
        public string Branch { get { return mBranch; } set { mBranch = value; } }
        public string Platform { get { return mPlatform; } set { mPlatform = value; } }

        public bool IsValid { get { return mResource.IsValid; } }

        public string RemoteSignature { get; set; }
        public string CacheSignature { get; set; }
        public string TargetSignature { get; set; }
        public string LocalSignature { get; set; }

        public IPackageFilename RemoteFilename { get; set; }
        public IPackageFilename CacheFilename { get; set; }
        public IPackageFilename TargetFilename { get; set; }
        public IPackageFilename LocalFilename { get; set; }

        public ComparableVersion RemoteVersion { get; set; }
        public ComparableVersion CacheVersion { get; set; }
        public ComparableVersion TargetVersion { get; set; }
        public ComparableVersion LocalVersion { get; set; }
        public ComparableVersion RootVersion { get; set; }

        public bool RemoteExists { get { return !String.IsNullOrEmpty(RemoteURL); } }
        public bool CacheExists { get { return !String.IsNullOrEmpty(CacheURL); } }
        public bool LocalExists { get { return !String.IsNullOrEmpty(LocalURL); } }
        public bool TargetExists { get { return !String.IsNullOrEmpty(TargetURL); } }
        public bool RootExists { get { return !String.IsNullOrEmpty(RootURL); } }

        public string RemoteURL { get; set; }
        public string CacheURL { get; set; }
        public string LocalURL { get; set; }
        public string TargetURL { get { return mTargetURL; } }
        public string RootURL { get { return mRootURL; } }

        public bool IsRootPackage { get { return RootExists; } }

        public PomInstance Pom { get { return mPom; } }
        public List<DependencyResource> Dependencies { get { return Pom.Dependencies; } }

        public Dictionary<string, DependencyTree> DependencyTree { get; set; }

        private PackageInstance()
        {

        }

        internal PackageInstance(PackageResource resource)
        {
            mResource = resource;
            DependencyTree = new Dictionary<string, DependencyTree>();
        }
        internal PackageInstance(PackageResource resource, PomInstance pom)
        {
            mResource = resource;
            mPom = pom;
            DependencyTree = new Dictionary<string, DependencyTree>();
        }

        public static PackageInstance From(string name, string group, string branch, string platform)
        {
            PackageResource resource = PackageResource.From(name, group);
            PackageInstance instance = resource.CreateInstance();
            instance.mBranch = branch;
            instance.mPlatform = platform;
            return instance;
        }

        public static PackageInstance LoadFromRoot(string dir)
        {
            PackageResource resource = PackageResource.LoadFromFile(dir);
            PackageInstance instance = resource.CreateInstance();
            instance.mRootURL = dir;
            return instance;
        }

        public static PackageInstance LoadFromTarget(string dir)
        {
            PackageResource resource = PackageResource.LoadFromFile(dir);
            PackageInstance instance = resource.CreateInstance();
            instance.mTargetURL = dir;
            return instance;
        }

        public static PackageInstance LoadFromLocal(string rootURL, IPackageFilename filename)
        {
            PackageResource resource = PackageResource.LoadFromPackage(rootURL + "target\\", filename);
            PackageInstance instance = resource.CreateInstance();
            instance.mRootURL = rootURL;
            instance.LocalURL = rootURL + "target\\";
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
            if (RootExists)
            {
                PackageResource resource = PackageResource.LoadFromFile(RootURL);
                mPom = resource.CreatePomInstance();
                mResource = resource;
                return true;
            }
            else if (TargetExists)
            {
                PackageResource resource = PackageResource.LoadFromFile(TargetURL);
                mPom = resource.CreatePomInstance();
                mResource = resource;
                return true;
            }
            else if (CacheExists)
            {
                PackageResource resource = PackageResource.LoadFromPackage(CacheURL, CacheFilename);
                mPom = resource.CreatePomInstance();
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
                case ELocation.Target: mTargetURL = url; break;
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
                case ELocation.Target: TargetFilename = filename; break;
                case ELocation.Root: break;
            }
        }

        public void SetSignature(ELocation location, string signature)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteSignature = signature; break;
                case ELocation.Cache: CacheSignature = signature; break;
                case ELocation.Local: LocalSignature = signature; break;
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
                case ELocation.Target: url = TargetURL; break;
                case ELocation.Root: url = RootURL; break;
            }
            return url;
        }

        public IPackageFilename GetFilename(ELocation location)
        {
            IPackageFilename filename = null;
            switch (location)
            {
                case ELocation.Remote: filename = RemoteFilename; break;
                case ELocation.Cache: filename = CacheFilename; break;
                case ELocation.Local: filename = LocalFilename; break;
                case ELocation.Target: filename = TargetFilename; break;
                case ELocation.Root: break;
            }
            return filename;
        }

        public string GetSignature(ELocation location)
        {
            string signature = string.Empty;
            switch (location)
            {
                case ELocation.Remote: signature = RemoteSignature; break;
                case ELocation.Cache: signature = CacheSignature; break;
                case ELocation.Local: signature = LocalSignature; break;
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
                case ELocation.Target: version = TargetVersion; break;
                case ELocation.Root: version = RootVersion; break;
            }
            return version;
        }

        public bool Info()
        {
            return Pom.Info();
        }

        
        

        public bool Install()
        {
            bool success = Global.CacheRepo.Add(this, ELocation.Local);
            return success;
        }

        public bool Deploy()
        {
            bool success = false;
            if (Global.CacheRepo.Update(this))
            {
                success = Global.RemoteRepo.Add(this, ELocation.Cache);
            }
            return success;
        }

        public bool BuildAllDependencies()
        {
            foreach (string platform in Pom.Platforms)
            {
                if (!BuildDependencies(platform))
                    return false;
            }
            return true;
        }

        public bool PrintAllDependencies()
        {
            foreach (string p in Pom.Platforms)
            {
                Loggy.Add(String.Format("Dependencies for platform : {0}", p));
                Loggy.Indent += 1;
                // Has valid dependency tree ?
                PrintDependencies(p);
                Loggy.Indent -= 1;
            }
            return true;
        }

        public bool SyncAllDependencies()
        {
            foreach (string platform in Pom.Platforms)
            {
                // Has valid dependency tree ?
                if (!SyncDependencies(platform))
                    return false;
            }
            return true;
        }

        public void GenerateProjects()
        {
            // Generating project files is a bit complex in that it has to merge project definitions on a per platform basis.
            // Every platform is considered to have its own package (zip -> pom.xml) containing the project settings for that platform.
            // For every platform we have to merge in only the conditional xml elements into the final project file.
            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                // Every platform has a dependency tree and every dependency package for that platform has filtered their
                // project to only keep their platform specific xml elements.
                foreach (KeyValuePair<string, DependencyTree> pair in DependencyTree)
                {
                    List<PackageInstance> allDependencyPackages = pair.Value.GetAllDependencyPackages();

                    // Merge in all Projects of those dependency packages which are already filtered on the platform
                    foreach (PackageInstance dependencyPackage in allDependencyPackages)
                    {
                        ProjectInstance dependencyProject = dependencyPackage.Pom.GetProjectByGroup(rootProject.Group);
                        if (dependencyProject != null && !dependencyProject.IsPrivate)
                            rootProject.MergeWithDependencyProject(dependencyProject);
                    }
                }
            }

            // And the root package UnitTest project generally merges with Main, since UnitTest will use Main.
            // Although the user could specify more project and different internal dependencies.
            foreach (ProjectInstance rootProject in Pom.Projects)
            {
                if (!String.IsNullOrEmpty(rootProject.DependsOn))
                {
                    string[] projectNames = rootProject.DependsOn.Split(new char[] { ';' }, StringSplitOptions.RemoveEmptyEntries);
                    foreach (string name in projectNames)
                    {
                        ProjectInstance dependencyProject = Pom.GetProjectByName(name);
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

        public void GenerateSolution()
        {
            List<string> projectFilenames = new List<string>();
            foreach (ProjectInstance prj in Pom.Projects)
            {
                string path = prj.Location.Replace("/", "\\");
                path = path.EndWith('\\');
                path = path + prj.Name + prj.Extension;
                projectFilenames.Add(path);
            }

            CppSolution solution = new CppSolution(CppSolution.EVersion.VS2010, CppSolution.ELanguage.CPP);
            string solutionFilename = RootURL + Name + ".sln";
            solution.Save(solutionFilename, projectFilenames);
        }

        public bool BuildDependencies(string Platform)
        {
            bool result = true;

            DependencyTree tree;
            if (!DependencyTree.TryGetValue(Platform, out tree))
            {
                tree = new DependencyTree();
                tree.Package = this;
                tree.Platform = Platform;
                tree.Dependencies = new List<DependencyInstance>();
                foreach (DependencyResource resource in mResource.Dependencies)
                    tree.Dependencies.Add(new DependencyInstance(Platform, resource));
                DependencyTree.Add(Platform, tree);
                result = tree.Build();
            }

            return result;
        }

        public void PrintDependencies(string Platform)
        {
            DependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                tree.Print();
        }

        public bool SyncDependencies(string Platform)
        {
            bool result = false;
            DependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                result = false; // tree.Sync();
            return result;
        }
    }
}