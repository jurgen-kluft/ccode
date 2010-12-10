using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public interface XPackageRepositoryLayout
    {
        string RepoPath { get; set; }

        string CheckoutVersion(string group, string package_path, string package_name, string branch, string platform, XVersionRange versionRange);
        string CommitVersion(string group, string package_path, string package_name, string branch, string platform, XVersion version);

        void SyncTo(string group, string package_name, string branch, string platform, XVersionRange range, XPackageRepositoryLayout to);
    }

    public class XPackageRepositoryLayoutDefault : XPackageRepositoryLayout
    {
        public static class Layout
        {
            public static string VersionToPath(XVersion version)
            {
                string path = string.Empty;
                string[] components = version.ToStrings();
                // Keep it to X.Y.Z
                for (int i = 0; i < components.Length && i < 3; ++i)
                {
                    path = path + components[0] + "\\";
                }
                return path;
            }
            
            public static string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, XVersion version)
            {
                return String.Format("{0}+{1}+{2}+{3}", package_name, version.ToString(), branch, platform);
            }

            public static string VersionToFilename(string package_name, string branch, string platform, XVersion version)
            {
                return VersionToFilenameWithoutExtension(package_name, branch, platform, version) + ".zip";
            }

            public static string FilenameToVersion(string filename)
            {
                string[] parts = filename.Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
                return parts[1];
            }

            public static string FullRootPath(string repoPath, string group, string package_name)
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

            public static string FullVersionPath(string repoPath, string group, string package_name, XVersion version)
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
                string fullPath = repoPath + groupPath + package_name + "\\version\\" + VersionToPath(version);
                return fullPath;
            }
        }

        public XPackageRepositoryLayoutDefault(string repoPath)
        {
            RepoPath = repoPath;
        }

        public string RepoPath { get; set; }

        private void CheckoutVersion(string group, string package_path, string package_name, string branch, string platform, XVersion version)
        {
            string src_path = Layout.FullVersionPath(RepoPath, group, package_name, version);
            string src_filename = Layout.VersionToFilename(package_name, branch, platform, version);

            if (File.Exists(src_path + src_filename) && Directory.Exists(package_path))
            {
                if (!File.Exists(package_path + src_filename))
                {
                    File.Copy(src_path + src_filename, package_path + src_filename);
                }
            }
        }

        public string CheckoutVersion(string group, string package_path, string package_name, string branch, string platform, XVersionRange versionRange)
        {
            XVersion version = FindBestVersion(group, package_name, branch, platform, versionRange);
            if (version != null)
                CheckoutVersion(group, package_path, package_name, branch, platform, version);
            return string.Empty;
        }

        public string CommitVersion(string group, string package_path, string package_name, string branch, string platform, XVersion version)
        {
            string package_filename = Layout.VersionToFilename(package_name, branch, platform, version);
            if (File.Exists(package_path))
            {
                string dest_path = Layout.FullVersionPath(RepoPath, group, package_name, version);
                if (!Directory.Exists(dest_path))
                {
                    Directory.CreateDirectory(dest_path);
                }
                File.Copy(package_path, dest_path + package_filename, true);
                DirtyVersionCache(group, package_name);
                return package_filename;
            }
            return string.Empty;
        }


        /// Database file should contain a table like this:
        ///     Version                 |    Branch    |     Platform     |     Change-Set ID
        ///     1.0.0.0.10.12.10.17.20  |    default   |       Win32      |   AAAAAAAAAAAAAAAAAAAAA


        public void SyncTo(string group, string package_name, string branch, string platform, XVersionRange versionRange, XPackageRepositoryLayout to)
        {
            // Sync the best version from this repository to another
            XVersion version = FindBestVersion(group, package_name, branch, platform, versionRange);
            if (version != null)
            {
                string dir = Layout.FullVersionPath(RepoPath, group, package_name, version);
                string path = dir + Layout.VersionToFilename(package_name, branch, platform, version);
                to.CommitVersion(group, path, package_name, branch, platform, version);
            }
        }

        public void UpdateVersionCache(string group, string package_name)
        {
            string path = Layout.FullRootPath(RepoPath, group, package_name);
            if (!Directory.Exists(path))
                return;
            if (File.Exists(path + "version.cache.writelock"))
            {
                long ticks = File.GetLastWriteTime(path + "version.cache.writelock").Ticks;
                long current_ticks = DateTime.Now.Ticks;
                TimeSpan timespan = new TimeSpan(current_ticks - ticks);
                TimeSpan timeout = new TimeSpan(0, 5, 0);
                bool _return = true;
                if (timespan > timeout)
                {
                    try { _return = false;  File.Delete(path + "version.cache.writelock"); }
                    catch (SystemException) { _return = true; }
                }
                if (_return)
                    return;
            }

            using (FileStream stream = new FileStream(path + "version.cache.writelock", FileMode.Create, FileAccess.Write, FileShare.None))
            {
                using (StreamWriter writer = new StreamWriter(stream))
                {
                    string[] dirtyMarkerFiles = Directory.GetFiles(path, "*.dirty", SearchOption.TopDirectoryOnly);
                    foreach (string dirty in dirtyMarkerFiles)
                    {
                        try { File.Delete(dirty); }
                        catch (SystemException) { }
                    }
                    string[] packages = Directory.GetFiles(path + "version\\", "*.zip", SearchOption.AllDirectories);
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

                    try { File.Copy(path + "version.cache.writelock", path + "version.cache", true); }
                    catch (SystemException) { }
                    try { File.Delete(path + "version.cache.writelock"); }
                    catch (SystemException) { }
                }
            }
        }

        private string[] LoadVersionCache(string group, string package_name)
        {
            string path = Layout.FullRootPath(RepoPath, group, package_name);
            string[] versions = new string[0];

            int retry = 0;
            bool exception = true;
            while (exception && retry < 5)
            {
                try
                {
                    exception = false;
                    versions = File.ReadAllLines(path + "version.cache");
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
            string path = Layout.FullRootPath(RepoPath, group, package_name);
            if (!File.Exists(path + Environment.MachineName + ".dirty"))
            {
                FileStream s = File.Create(path + Environment.MachineName + ".dirty");
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

        private XVersion FindBestVersion(string group, string package_name, string branch, string platform, XVersionRange versionRange)
        {
            string[] versions = LoadVersionCache(group, package_name);
            if (versions.Length > 1)
            {
                if (versionRange.Kind == XVersionRange.EKind.UniqueVersion)
                {
                    List<XVersion> all = Construct(versions);
                    XVersion version = LowerBound(all, versionRange.From, versionRange.IncludeFrom);
                    // This might not be a match, but we have to return the best possible match
                    return version;
                }
                else
                {
                    XVersion highest = new XVersion(versions[versions.Length - 1]);

                    // High probability that the highest version will match
                    if (versionRange.IsInRange(highest))
                        return new XVersion(highest);

                    if (versionRange.Kind == XVersionRange.EKind.VersionToUnbound)
                    {
                        // lowest ---------------------------------------------------------- highest
                        // xxxxxxxxxxxxxxxx from >-----------------------------------------------------------
                        // highest is not in range so there will be no matching version
                        return null;
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
                            return LowerBound(all, versionRange.From, versionRange.IncludeFrom);
                        }
                        else if (versionRange.Kind == XVersionRange.EKind.UnboundToVersion)
                        {
                            // highest failed and this can only be in the following case:
                            //     lowest ---------------------------------------------------------- highest
                            //     ----------------------< to xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
                            // Find a version in the UnboundToVersion, where Version is 'to'
                            List<XVersion> all = Construct(versions);
                            return LowerBound(all, versionRange.To, versionRange.IncludeTo);
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
                                    return version;
                            }
                        }
                    }
                }
            }
            else if (versions.Length == 1)
            {
                XVersion version = new XVersion(versions[0]);
                if (versionRange.IsInRange(version))
                    return version;
            }

            return null;
        }
    }

    public static class MyListExtensions
    {
        public static int IndexOfUsingBinarySearch<T>(this List<T> sortedCollection, T value) where T : IComparable<T>
        {
            if (sortedCollection == null)
                return -1;

            int begin = 0;
            int end = sortedCollection.Count - 1;
            int index = 0;
            while (end >= begin)
            {
                index = (begin + end) / 2;
                T val = sortedCollection[index];
                int compare = val.CompareTo(value);
                if (compare == 0)
                    return index;
                if (compare > 0)
                    end = index - 1;
                else
                    begin = index + 1;
            }

            return ~index;  // Not found, return bitwise complement of the index.
        }

        public static int LowerBound<T>(this List<T> sortedCollection, T value, bool lessOrEqual) where T : IComparable<T>
        {
            int index = IndexOfUsingBinarySearch(sortedCollection, value);
            if (index < 0)
                index = ~index;

            if (lessOrEqual)
            {
                while (index > 0 && value.CompareTo(sortedCollection[index-1]) == -1)
                    --index;
                while (index < sortedCollection.Count && value.CompareTo(sortedCollection[index]) != -1)
                    ++index;
            }
            else
            {
                while (index > 0 && value.CompareTo(sortedCollection[index-1]) != 1)
                    --index;
                while (index < sortedCollection.Count && value.CompareTo(sortedCollection[index]) == 1)
                    ++index;
            }

            return index;
        }
    }
}
