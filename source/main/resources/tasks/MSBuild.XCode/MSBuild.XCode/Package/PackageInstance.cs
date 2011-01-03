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
        private string mLocalURL = string.Empty;

        private string mBranch;
        private string mPlatform;

        private PomInstance mPom;

        public Group Group { get { return mResource.Group; } }
        public string Name { get { return mResource.Name; } }
        public string Branch { get { return mBranch; } }
        public string Platform { get { return mPlatform; } }

        public bool IsValid { get { return mResource.IsValid; } }

        public IPackageFilename Filename { get; set; }
        public ComparableVersion Version { get; set; }

        public bool RemoteExists { get { return !String.IsNullOrEmpty(RemoteURL); } }
        public bool CacheExists { get { return !String.IsNullOrEmpty(CacheURL); } }
        public bool LocalExists { get { return !String.IsNullOrEmpty(LocalURL); } }
        public bool TargetExists { get { return !String.IsNullOrEmpty(TargetURL); } }
        public bool RootExists { get { return !String.IsNullOrEmpty(RootURL); } }

        public string RemoteURL { get; set; }
        public string CacheURL { get; set; }
        public string LocalURL { get { return mLocalURL; } }
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
            instance.mLocalURL = rootURL + "target\\";
            instance.Filename = filename;
            instance.Version = filename.Version;
            return instance;
        }

        public static PackageInstance FromFilename(IPackageFilename filename)
        {
            PackageResource resource = PackageResource.From(filename.Name, string.Empty);
            PackageInstance instance = resource.CreateInstance();
            instance.Version = new ComparableVersion(filename.Version);
            instance.mBranch = filename.Branch;
            instance.mPlatform = filename.Platform;
            return instance;
        }

        public void SetPlatform(string platform)
        {
            mPlatform = platform;
            Version = Pom.Versions.GetForPlatform(Platform);
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
                PackageResource resource = PackageResource.LoadFromPackage(CacheURL, Filename);
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
                case ELocation.Local: mLocalURL = url; break;
                case ELocation.Target: mTargetURL = url; break;
                case ELocation.Root: mRootURL = url; break;
            }
        }
        
        public void SetURL(ELocation location, string url, string filename)
        {
            SetURL(location, url);
            Filename = new PackageFilename(Path.GetFileNameWithoutExtension(filename));
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

        public bool Info()
        {
            return Pom.Info();
        }

        public bool Extract()
        {
            if (CacheExists && TargetExists)    // && File.Exists(CacheURL)), should exist since the PackageRepository assigned it
            {
                if (File.Exists(CacheURL + Filename))
                {
                    ZipFile zip = new ZipFile(CacheURL + Filename);
                    zip.ExtractAll(TargetURL, ExtractExistingFileAction.OverwriteSilently);

                    DateTime lastWriteTime = File.GetLastWriteTime(CacheURL + Filename);
                    FileInfo fi = new FileInfo(TargetURL + Path.GetFileNameWithoutExtension(Filename.ToString()) + ".t");
                    if (fi.Exists)
                    {
                        fi.LastWriteTime = lastWriteTime;
                    }
                    else
                    {
                        fi.Create().Close();
                        fi.LastWriteTime = lastWriteTime;
                    }
                    return true;
                }
            }
            return false;
        }

        public bool VerifyBeforeExtract()
        {
            if (CacheExists && TargetExists)
            {
                // Verify 'Extracted' package
                if (Verify())
                    return true;

                return Extract();
            }
            return false;
        }

        private void Glob(string src, string dst, List<KeyValuePair<string,string>> files)
        {
            src = src.Replace("${Name}", Name);
            src = src.Replace("${Platform}", Platform);

            List<string> globbedFiles = PathUtil.getFiles(src);

            int r = src.IndexOf("**");
            string reldir = r >= 0 ? src.Substring(0, src.IndexOf("**")) : string.Empty;

            foreach (string src_filename in globbedFiles)
            {
                string dst_filename;
                if (r >= 0)
                    dst_filename = dst + src_filename.Substring(reldir.Length);
                else
                    dst_filename = dst + Path.GetFileName(src_filename);

                files.Add(new KeyValuePair<string, string>(src_filename, Path.GetDirectoryName(dst_filename)));
            }
        }

        public bool Create(string branch, string platform, out IPackageFilename outFilename)
        {
            bool success = false;

            mBranch = branch;
            mPlatform = platform;
            Version = Pom.Versions.GetForPlatformWithBranch(Platform, Branch);

            /// Delete the .sfv file
            string sfv_filename = Name + ".md5";
            string dir = RootURL + "target\\outdir\\";
            if (File.Exists(dir + sfv_filename))
                File.Delete(dir + sfv_filename);

            /// 1) Create zip file
            /// 2) For every file create an MD5 and gather them into a sfv file
            /// 3) Remove root from every source file
            /// 4) Set the work directory
            /// 5) Add files to zip
            /// 6) Close
            /// 

            List<KeyValuePair<string, string>> content;
            if (!Pom.Content.TryGetValue(Platform, out content))
            {
                if (!Pom.Content.TryGetValue("*", out content))
                {
                    outFilename = new PackageFilename();
                    return false;
                }
            }
            List<KeyValuePair<string, string>> files = new List<KeyValuePair<string,string>>();
            foreach (KeyValuePair<string, string> pair in content)
            {
                Glob(RootURL + pair.Key, pair.Value, files);
            }
            
            // Is pom.xml included?
            bool includesPomXml = false;
            foreach (KeyValuePair<string, string> pair in files)
            {
                if (String.Compare(Path.GetFileName(pair.Key), "pom.xml", true) == 0)
                {
                    includesPomXml = true;
                    break;
                }
            }
            if (!includesPomXml)
            {
                Loggy.Add(String.Format("Error: Package::Create, package must include pom.xml!"));
                outFilename = new PackageFilename();
                return false;
            }

            MD5CryptoServiceProvider md5_provider = new MD5CryptoServiceProvider();
            Dictionary<string, byte[]> md5Dictionary = new Dictionary<string, byte[]>();
            foreach (KeyValuePair<string, string> pair in files)
            {
                string src_filename = pair.Key;

                FileStream fs = new FileStream(src_filename, FileMode.Open, FileAccess.Read);
                byte[] md5 = md5_provider.ComputeHash(fs);
                fs.Close();

                string dst_filename;
                if (String.IsNullOrEmpty(pair.Value))
                    dst_filename = Path.GetFileName(src_filename);
                else
                    dst_filename = pair.Value.EndWith('\\') + Path.GetFileName(src_filename);

                if (!md5Dictionary.ContainsKey(dst_filename))
                    md5Dictionary.Add(dst_filename, md5);
            }

            if (!Directory.Exists(dir))
                Directory.CreateDirectory(dir);

            using (FileStream wfs = new FileStream(dir + sfv_filename, FileMode.Create))
            {
                StreamWriter writer = new StreamWriter(wfs);
                writer.WriteLine("; Generated by MSBuild.XCode");
                foreach (KeyValuePair<string, byte[]> k in md5Dictionary)
                {
                    writer.WriteLine("{0} *{1}", k.Key, StringTools.MD5ToString(k.Value));
                }
                writer.Close();
                wfs.Close();

                files.Add(new KeyValuePair<string, string>(dir + sfv_filename, Path.GetDirectoryName(sfv_filename)));
            }

            // Add VCS Information file to the package
            if (File.Exists(RootURL + "vcs.info"))
                files.Add(new KeyValuePair<string, string>(RootURL + "vcs.info", ""));

            DependencyTree dependencyTree;
            if (DependencyTree.TryGetValue(Platform, out dependencyTree))
            {
                dependencyTree.SaveInfo(new FileDirectoryPath.FilePathAbsolute(RootURL + "dependencies.info"));
                files.Add(new KeyValuePair<string, string>(RootURL + "dependencies.info", ""));
            }

            ComparableVersion version = Pom.Versions.GetForPlatformWithBranch(Platform, Branch);
            outFilename = new PackageFilename(Name, version, Branch, Platform);
            outFilename.DateTime = DateTime.Now;

            if (File.Exists(RootURL + "target\\" + outFilename.ToString()))
            {
                try { File.Delete(RootURL + "target\\" + Filename.ToString()); }
                catch (Exception) { }
            }

            using (ZipFile zip = new ZipFile(RootURL + "target\\" + outFilename.ToString()))
            {
                foreach (KeyValuePair<string, string> p in files)
                    zip.AddFile(p.Key, p.Value);

                zip.Save();
                mLocalURL = RootURL + "target\\" + Filename.ToString();
                Filename = new PackageFilename(outFilename);
                success = true;
            }
            return success;
        }

        public bool Verify()
        {
            bool ok = false;

            if (!TargetExists || !CacheExists)
                return ok;

            string md5_file = Name + ".MD5";
            if (File.Exists(TargetURL + md5_file))
            {
                DateTime packageLastWriteTime = File.GetLastWriteTime(CacheURL);
                FileInfo fi = new FileInfo(TargetURL + Path.GetFileNameWithoutExtension(CacheURL) + ".t");
                bool markerFileExists = fi.Exists;
                bool markerFileUpToDate = markerFileExists && (fi.LastWriteTime == packageLastWriteTime);
                if (markerFileExists && markerFileUpToDate)
                {
                    MD5CryptoServiceProvider md5_provider = new MD5CryptoServiceProvider();

                    // Load MD5 file
                    ok = true;
                    string[] lines = File.ReadAllLines(TargetURL + md5_file);

                    // MD5 is relative to its own location
                    foreach (string entry in lines)
                    {
                        if (entry.Trim().StartsWith(";"))
                            continue;

                        // Get the MD5 and Filename
                        int s = entry.IndexOf('*');
                        if (s == -1)
                        {
                            ok = false;
                            break;
                        }
                        string old_md5 = entry.Substring(s + 1).Trim();
                        string filename = TargetURL + entry.Substring(0, s).Trim();

                        if (File.Exists(filename))
                        {
                            string new_md5 = string.Empty;
                            using (FileStream rfs = new FileStream(filename, FileMode.Open, FileAccess.Read))
                            {
                                byte[] new_md5_raw = md5_provider.ComputeHash(rfs);
                                new_md5 = StringTools.MD5ToString(new_md5_raw);
                                rfs.Close();
                            }

                            if (String.Compare(old_md5, new_md5) != 0)
                            {
                                ok = false;
                                break;
                            }
                        }
                        else
                        {
                            // File doesn't exist anymore
                            ok = false;
                            break;
                        }
                    }
                }
                return ok;
            }
            return ok;
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

            MsDev2010.Cpp.XCode.Solution solution = new MsDev2010.Cpp.XCode.Solution(MsDev2010.Cpp.XCode.Solution.EVersion.VS2010, MsDev2010.Cpp.XCode.Solution.ELanguage.CPP);
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
                result = tree.Sync();
            return result;
        }
    }
}
