using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageRepositoryFileSystem : IPackageRepository
    {
        public PackageRepositoryFileSystem(string repoDir, ELocation location)
        {
            RepoDir = repoDir;
            Layout = new LayoutDefault();
            Location = location;
        }

        public string RepoDir { get; set; }
        public ELocation Location { get; set; }
        public ILayout Layout { get; set; }

        private bool GetSignatureFor(PackageInstance package)
        {
            string package_dir = Layout.PackageRootDir(RepoDir, package.Group.ToString(), package.Name, package.Platform);

            string signatureFilename = ".signature";
            if (!File.Exists(package_dir + signatureFilename))
                File.Create(package_dir + signatureFilename).Close();

            DateTime last_write_time = File.GetLastWriteTime(package_dir + signatureFilename);
            package.SetSignature(Location, last_write_time);

            return true;
        }
        private void UpdateSignatureOf(PackageInstance package)
        {
            string package_dir = Layout.PackageRootDir(RepoDir, package.Group.ToString(), package.Name, package.Platform);

            string signatureFilename = ".signature";
            if (!File.Exists(package_dir + signatureFilename))
                File.Create(package_dir + signatureFilename).Close();

            DateTime last_write_time = DateTime.Now;
            File.SetLastWriteTime(package_dir + signatureFilename, last_write_time);
            package.SetSignature(Location, last_write_time);
        }

        public bool Update(PackageInstance package)
        {
            string src_dir = Layout.PackageVersionDir(RepoDir, package.Group.ToString(), package.Name, package.Platform, package.GetVersion(Location));
            string src_filename = Layout.VersionToFilename(package.Name, package.Branch, package.Platform, package.GetVersion(Location));
            string src_path = src_dir + src_filename;
            if (File.Exists(src_path))
            {
                package.SetURL(Location, src_dir);
                package.SetFilename(Location, new PackageFilename(src_filename));
                return GetSignatureFor(package);
            }
            return false;
        }

        public bool Update(PackageInstance package, VersionRange versionRange)
        {
            if (FindBestVersion(package, versionRange))
            {
                GetSignatureFor(package); 
                return Update(package);
            }
            return false;
        }

        public bool Add(PackageInstance package, ELocation from)
        {
            if (File.Exists(package.GetURL(from)))
            {
                string dest_dir = Layout.PackageVersionDir(RepoDir, package.Group.ToString(), package.Name, package.Platform, package.GetVersion(from));
                if (!Directory.Exists(dest_dir))
                {
                    Directory.CreateDirectory(dest_dir);
                }
                string package_filename = Layout.VersionToFilename(package.Name, package.Branch, package.Platform, package.GetVersion(from));
                if (!File.Exists(dest_dir + package_filename))
                {
                    File.Copy(package.GetURL(from), dest_dir + package_filename, true);
                    DirtyVersionCache(package.Group.ToString(), package.Name, package.Platform);
                }

                package.SetURL(Location, dest_dir);
                package.SetFilename(Location, new PackageFilename(package_filename));
                package.SetVersion(Location, package.GetVersion(from));

                UpdateSignatureOf(package);

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
            string root_dir = Layout.PackageRootDir(RepoDir, group, package_name, platform);
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
                    string[] filenames = Directory.GetFiles(root_dir + "version\\", "*.zip", SearchOption.AllDirectories);
                    SortedDictionary<ComparableVersion, bool> sortedVersions = new SortedDictionary<ComparableVersion, bool>();
                    foreach (string file in filenames)
                    {
                        PackageFilename packageFilename = new PackageFilename(file);
                        if (String.Compare(packageFilename.Platform, platform, true) == 0)
                        {
                            ComparableVersion version = packageFilename.Version;
                            if (!sortedVersions.ContainsKey(version))
                                sortedVersions.Add(version, true);
                        }
                    }

                    foreach (ComparableVersion v in sortedVersions.Keys)
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
            string root_dir = Layout.PackageRootDir(RepoDir, group, package_name, platform);
            string[] versions = new string[0];
            if (!Directory.Exists(root_dir))
                return versions;
            string filename = String.Format("versions.{0}.{1}.cache", branch, platform);
            if (!File.Exists(root_dir + filename))
                return versions;

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

        private void DirtyVersionCache(string group, string package_name, string platform)
        {
            string root_dir = Layout.PackageRootDir(RepoDir, group, package_name, platform);
            if (!File.Exists(root_dir + Environment.MachineName + ".dirty"))
            {
                FileStream s = File.Create(root_dir + Environment.MachineName + ".dirty");
                s.Close();
            }
        }

        private ComparableVersion LowerBound(List<ComparableVersion> all, ComparableVersion value, bool lessOrEqual)
        {
            // LowerBound will return the index of the version that is higher than the lower bound
            int i = all.LowerBound(value, lessOrEqual) - 1;
            if (i >= 0 && i<all.Count)
                return all[i];

            return null;
        }

        private List<ComparableVersion> Construct(string[] versions)
        {
            List<ComparableVersion> all = new List<ComparableVersion>();
            foreach (string v in versions)
                all.Add(new ComparableVersion(v));
            return all;
        }

        private bool FindBestVersion(PackageInstance package, VersionRange versionRange)
        {
            string[] versions = RetrieveVersionsFor(package.Group.ToString(), package.Name, package.Branch, package.Platform);
            if (versions.Length > 1)
            {
                if (versionRange.Kind == VersionRange.EKind.UniqueVersion)
                {
                    List<ComparableVersion> all = Construct(versions);
                    package.SetVersion(Location, LowerBound(all, versionRange.From, versionRange.IncludeFrom));
                    // This might not be a match, but we have to return the best possible match
                    return true;
                }
                else
                {
                    ComparableVersion highest = new ComparableVersion(versions[versions.Length - 1]);

                    // High probability that the highest version will match
                    if (versionRange.IsInRange(highest))
                    {
                        package.SetVersion(Location, new ComparableVersion(highest));
                        return true;
                    }

                    if (versionRange.Kind == VersionRange.EKind.VersionToUnbound)
                    {
                        // lowest ---------------------------------------------------------- highest
                        // xxxxxxxxxxxxxxxx from >-----------------------------------------------------------
                        // highest is not in range so there will be no matching version
                        return false;
                    }
                    else
                    {
                        // Search for a matching version
                        if (versionRange.Kind == VersionRange.EKind.UnboundToVersionOrVersionToUnbound)
                        {
                            // highest failed and this can only be in the following case:
                            //      lowest ------------------------------------- highest
                            //      ---------------------< from xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx to >--------------------------
                            // Find a version in the UnboundToVersion, where Version is 'from'
                            List<ComparableVersion> all = Construct(versions);
                            ComparableVersion version = LowerBound(all, versionRange.From, versionRange.IncludeFrom);
                            package.SetVersion(Location, version);
                            return true;
                        }
                        else if (versionRange.Kind == VersionRange.EKind.UnboundToVersion)
                        {
                            // highest failed and this can only be in the following case:
                            //     lowest ---------------------------------------------------------- highest
                            //     ----------------------< to xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
                            // Find a version in the UnboundToVersion, where Version is 'to'
                            List<ComparableVersion> all = Construct(versions);
                            ComparableVersion version = LowerBound(all, versionRange.To, versionRange.IncludeTo);
                            package.SetVersion(Location, version);
                            return true;
                        }
                        else if (versionRange.Kind == VersionRange.EKind.VersionToVersion)
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
                                List<ComparableVersion> all = Construct(versions);
                                ComparableVersion version = LowerBound(all, versionRange.To, versionRange.IncludeTo);
                                if (version != null && versionRange.IsInRange(version))
                                {
                                    package.SetVersion(Location, version);
                                    return true;
                                }
                            }
                        }
                    }
                }
            }
            else if (versions.Length == 1)
            {
                ComparableVersion version = new ComparableVersion(versions[0]);
                if (versionRange.IsInRange(version))
                {
                    package.SetVersion(Location, version);
                    return true;
                }
            }
            return false;
        }
    }

}
