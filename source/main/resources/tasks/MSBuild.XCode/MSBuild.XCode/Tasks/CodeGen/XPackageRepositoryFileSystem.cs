using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public interface ILayout
    {
        string VersionToDir(XVersion version);
        string VersionToFilename(string package_name, string branch, string platform, XVersion version);
        string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, XVersion version);
        string FilenameToVersion(string filename);
        string PackageRootDir(string repoPath, string group, string package_name);
        string PackageVersionDir(string repoPath, string group, string package_name, XVersion version);
    }

    public class XLayoutDefault : ILayout
    {
        public string VersionToDir(XVersion version)
        {
            string path = string.Empty;
            string[] components = version.ToStrings(3);
            // Keep it to X.Y.Z
            for (int i = 0; i < components.Length && i < 3; ++i)
                path = path + components[i] + "\\";
            return path;
        }

        public string VersionToFilename(string package_name, string branch, string platform, XVersion version)
        {
            return VersionToFilenameWithoutExtension(package_name, branch, platform, version) + ".zip";
        }
        
        public string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, XVersion version)
        {
            return String.Format("{0}+{1}+{2}+{3}", package_name, version.ToString(), branch, platform);
        }

        public string FilenameToVersion(string filename)
        {
            string[] parts = filename.Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
            return parts[1];
        }

        public string PackageRootDir(string repoPath, string group, string package_name)
        {
            // Path = group[] \ group[] ... \ package_name \ version.cache
            string[] splitted_group = group.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
            string groupPath = string.Empty;
            foreach (string g in splitted_group)
            {
                if (String.IsNullOrEmpty(groupPath))
                    groupPath = g + "\\";
                else
                    groupPath = groupPath + g + "\\";
            }
            string fullPath = repoPath + groupPath + package_name + "\\";
            return fullPath;
        }

        public string PackageVersionDir(string repoPath, string group, string package_name, XVersion version)
        {
            // Path = group[] \ group[] ... \ package_name \ version.cache
            string[] splitted_group = group.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
            string groupPath = string.Empty;
            foreach (string g in splitted_group)
            {
                if (String.IsNullOrEmpty(groupPath))
                    groupPath = g + "\\";
                else
                    groupPath = groupPath + g + "\\";
            }
            string fullPath = repoPath + groupPath + package_name + "\\version\\" + VersionToDir(version);
            return fullPath;
        }
    }


    public interface IPackageRepository
    {
        string RepoDir { get; set; }
        ILayout Layout { get; set; }

        bool Checkout(XPackage package, XVersionRange versionRange);
        bool Checkout(XPackage package);
        bool Checkin(XPackage package);
    }

    public class XPackageRepositoryFileSystem : IPackageRepository
    {
        public XPackageRepositoryFileSystem(string repoDir)
        {
            RepoDir = repoDir;
            Layout = new XLayoutDefault();
        }

        public string RepoDir { get; set; }
        public ILayout Layout { get; set; }

        public bool Checkout(XPackage package)
        {
            string src_dir = Layout.PackageVersionDir(RepoDir, package.Group.ToString(), package.Name, package.Version);
            string src_filename = Layout.VersionToFilename(package.Name, package.Branch, package.Platform, package.Version);
            string src_path = src_dir + src_filename;
            if (File.Exists(src_path))
            {
                package.Path = src_path;
                package.Local = true;
                return true;
            }
            return false;
        }

        public bool Checkout(XPackage package, XVersionRange versionRange)
        {
            if (FindBestVersion(package, versionRange))
            {
                return Checkout(package);
            }
            return false;
        }

        public bool Checkin(XPackage package)
        {
            string package_filename = Layout.VersionToFilename(package.Name, package.Branch, package.Platform, package.Version);
            if (File.Exists(package.Path))
            {
                string dest_dir = Layout.PackageVersionDir(RepoDir, package.Group.ToString(), package.Name, package.Version);
                if (!Directory.Exists(dest_dir))
                {
                    Directory.CreateDirectory(dest_dir);
                }
                if (!File.Exists(dest_dir + package_filename))
                {
                    File.Copy(package.Path, dest_dir + package_filename, true);
                    DirtyVersionCache(package.Group.ToString(), package.Name);
                }
                return true;
            }
            return false;
        }

        public string[] RetrieveVersionsFor(string group, string package_name, string branch, string platform)
        {
            UpdateVersionCache(group, package_name, branch, platform);
            return LoadVersionCache(group, package_name, branch, platform);
        }

        // 
        // OPTIMIZATION, THIS CAN BE DONE IN MANY DIFFERENT WAYS
        // 
        public void UpdateVersionCache(string group, string package_name, string branch, string platform)
        {
            string root_dir = Layout.PackageRootDir(RepoDir, group, package_name);
            if (!Directory.Exists(root_dir))
                return;

            string filename = String.Format("versions.{0}.{1}.cache", branch, platform);

            if (File.Exists(root_dir + filename + ".writelock"))
            {
                long ticks = File.GetLastWriteTime(root_dir + filename + ".writelock").Ticks;
                long current_ticks = DateTime.Now.Ticks;
                TimeSpan timespan = new TimeSpan(current_ticks - ticks);
                TimeSpan timeout = new TimeSpan(0, 5, 0);
                bool _return = true;
                if (timespan > timeout)
                {
                    try { _return = false;  File.Delete(root_dir + filename + ".writelock"); }
                    catch (SystemException) { _return = true; }
                }
                if (_return)
                    return;
            }

            using (FileStream stream = new FileStream(root_dir + filename + ".writelock", FileMode.Create, FileAccess.Write, FileShare.None))
            {
                using (StreamWriter writer = new StreamWriter(stream))
                {
                    string[] dirtyMarkerFiles = Directory.GetFiles(root_dir, "*.dirty", SearchOption.TopDirectoryOnly);
                    foreach (string dirty in dirtyMarkerFiles)
                    {
                        try { File.Delete(dirty); }
                        catch (SystemException) { }
                    }
                    string[] packages = Directory.GetFiles(root_dir + "version\\", "*.zip", SearchOption.AllDirectories);
                    SortedDictionary<XVersion, bool> sortedVersions = new SortedDictionary<XVersion, bool>();
                    foreach (string package in packages)
                    {
                        string[] c = Path.GetFileNameWithoutExtension(package).Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
                        if (c.Length == 4)
                        {
                            XVersion version = new XVersion(c[1]);
                            if (!sortedVersions.ContainsKey(version))
                                sortedVersions.Add(version, true);
                        }
                    }

                    foreach (XVersion v in sortedVersions.Keys)
                        writer.WriteLine(v.ToString());
                    writer.Close();
                    stream.Close();

                    try { File.Copy(root_dir + filename + ".writelock", root_dir + filename, true); }
                    catch (SystemException) { }
                    try { File.Delete(root_dir + filename + ".writelock"); }
                    catch (SystemException) { }
                }
            }
        }

        private string[] LoadVersionCache(string group, string package_name, string branch, string platform)
        {
            string root_dir = Layout.PackageRootDir(RepoDir, group, package_name);
            string[] versions = new string[0];

            string filename = String.Format("versions.{0}.{1}.cache", branch, platform);

            int retry = 0;
            bool exception = true;
            while (exception && retry < 5)
            {
                try
                {
                    exception = false;
                    versions = File.ReadAllLines(root_dir + filename);
                }
                catch (SystemException)
                {
                    exception = true;
                    System.Threading.Thread.Sleep(1 * 1000);
                    retry++;
                }
            }
            return versions;
        }

        private void DirtyVersionCache(string group, string package_name)
        {
            string root_dir = Layout.PackageRootDir(RepoDir, group, package_name);
            if (!File.Exists(root_dir + Environment.MachineName + ".dirty"))
            {
                FileStream s = File.Create(root_dir + Environment.MachineName + ".dirty");
                s.Close();
            }
        }

        private XVersion LowerBound(List<XVersion> all, XVersion value, bool lessOrEqual)
        {
            // LowerBound will return the index of the version that is higher than the lower bound
            int i = all.LowerBound(value, lessOrEqual) - 1;
            if (i >= 0 && i<all.Count)
                return all[i];

            return null;
        }

        private List<XVersion> Construct(string[] versions)
        {
            List<XVersion> all = new List<XVersion>();
            foreach (string v in versions)
                all.Add(new XVersion(v));
            return all;
        }

        private bool FindBestVersion(XPackage package, XVersionRange versionRange)
        {
            string[] versions = RetrieveVersionsFor(package.Group.ToString(), package.Name, package.Branch, package.Platform);
            if (versions.Length > 1)
            {
                if (versionRange.Kind == XVersionRange.EKind.UniqueVersion)
                {
                    List<XVersion> all = Construct(versions);
                    package.Version = LowerBound(all, versionRange.From, versionRange.IncludeFrom);
                    // This might not be a match, but we have to return the best possible match
                    return true;
                }
                else
                {
                    XVersion highest = new XVersion(versions[versions.Length - 1]);

                    // High probability that the highest version will match
                    if (versionRange.IsInRange(highest))
                    {
                        package.Version = new XVersion(highest);
                        return true;
                    }

                    if (versionRange.Kind == XVersionRange.EKind.VersionToUnbound)
                    {
                        // lowest ---------------------------------------------------------- highest
                        // xxxxxxxxxxxxxxxx from >-----------------------------------------------------------
                        // highest is not in range so there will be no matching version
                        return false;
                    }
                    else
                    {
                        // Search for a matching version
                        if (versionRange.Kind == XVersionRange.EKind.UnboundToVersionOrVersionToUnbound)
                        {
                            // highest failed and this can only be in the following case:
                            //      lowest ------------------------------------- highest
                            //      ---------------------< from xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx to >--------------------------
                            // Find a version in the UnboundToVersion, where Version is 'from'
                            List<XVersion> all = Construct(versions);
                            XVersion version = LowerBound(all, versionRange.From, versionRange.IncludeFrom);
                            package.Version = version;
                            return true;
                        }
                        else if (versionRange.Kind == XVersionRange.EKind.UnboundToVersion)
                        {
                            // highest failed and this can only be in the following case:
                            //     lowest ---------------------------------------------------------- highest
                            //     ----------------------< to xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
                            // Find a version in the UnboundToVersion, where Version is 'to'
                            List<XVersion> all = Construct(versions);
                            XVersion version = LowerBound(all, versionRange.To, versionRange.IncludeTo);
                            package.Version = version;
                            return true;
                        }
                        else if (versionRange.Kind == XVersionRange.EKind.VersionToVersion)
                        {
                            // highest failed and this can only be in the following two cases:
                            // 1)
                            //     lowest ---------------------------------------------------------- highest
                            //     xxxxxxxxxxxxxxxx from >---------------------------------< to xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
                            // 2)
                            //     lowest --------highest
                            //     xxxxxxxxxxxxxxxxxxxxxxxxxxxxx from >---------------------------------< to xxxxxxxxxxxxxxxxx
                            // Only case 1 might give as a version
                            if (!highest.LessThan(versionRange.From, !versionRange.IncludeFrom))
                            {
                                List<XVersion> all = Construct(versions);
                                XVersion version = LowerBound(all, versionRange.To, versionRange.IncludeTo);
                                if (version != null && versionRange.IsInRange(version))
                                {
                                    package.Version = version;
                                    return true;
                                }
                            }
                        }
                    }
                }
            }
            else if (versions.Length == 1)
            {
                XVersion version = new XVersion(versions[0]);
                if (versionRange.IsInRange(version))
                {
                    package.Version = version;
                    return true;
                }
            }
            return false;
        }
    }

}
