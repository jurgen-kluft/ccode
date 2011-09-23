using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;
using xpackage_repo;

namespace MSBuild.XCode
{
    public class PackageRepositoryRemoteFs : IPackageRepository
    {
        private string mDatabaseURL;
        private string mStorageURL;

        public PackageRepositoryRemoteFs(string repoURL, ELocation location)
        {
            RepoURL = repoURL.EndWith('\\');
            Layout = new LayoutDefault();
            Location = location;
            Valid = true;
        }

        public bool Valid { get; private set; }
        public string RepoURL { get; set; }
        public ELocation Location { get; private set; }
        private ILayout Layout { get; set; }

        public bool Query(Package package)
        {
            return Query(package, new VersionRange("[1.0,)"));
        }

        public bool Query(Package package, VersionRange versionRange)
        {
            return (FindBestVersion(package, versionRange));
        }

        public bool Download(Package package, string to_filename)
        {
            return false;
        }

        public bool Link(Package package, out string filename)
        {
            string package_d = Layout.PackageVersionDir(RepoURL, package.Group, package.Name, package.Platform, package.Branch, package.GetVersion(Location));
            string package_f = package.GetFilename(Location).ToString();
            if (File.Exists(package_d + package_f))
            {
                filename = package_d + package_f;
                return true;
            }
            filename = string.Empty;
            return false;
        }

        public bool Submit(Package package, IPackageRepository from)
        {
            string package_d = Layout.PackageVersionDir(RepoURL, package.Group, package.Name, package.Platform, package.Branch, package.GetVersion(from.Location));
            if (!Directory.Exists(package_d))
                Directory.CreateDirectory(package_d);

            string package_f = Layout.VersionToFilename(package.Name, package.Branch, package.Platform, package.GetVersion(from.Location));

            if (!from.Download(package, package_d + package_f))
                return false;

            package.SetURL(Location, package_d);
            package.SetFilename(Location, new PackageFilename(package_f));
            package.SetVersion(Location, package.GetVersion(from.Location));
            package.SetSignature(Location, package.GetSignature(from.Location));
            return true;
        }

        private PackageFilename FindBest(PackageFilename[] versions, VersionRange versionRange)
        {
            if (versions == null || versions.Length == 0)
                return null;

            PackageFilename best = versions[versions.Length - 1];

            // High probability that the highest version will match
            if (versionRange.IsInRange(best.Version))
                return best;

            // The list is sorted
            best = null;
            foreach (PackageFilename v in versions)
            {
                if (versionRange.IsInRange(v.Version))
                {
                    best = v;
                }
            }
            return best;
        }

        private PackageFilename[] RetrieveVersionsFor(string group, string package_name, string branch, string platform)
        {
            string root_dir = Layout.PackageRootDir(RepoURL, group, package_name, platform);
            if (Directory.Exists(root_dir))
            {
                string[] filenames = Directory.GetFiles(root_dir + "version\\", String.Format("*+{0}+{1}.zip", branch, platform), SearchOption.AllDirectories);
                SortedDictionary<string, PackageFilename> sortedVersions = new SortedDictionary<string, PackageFilename>();
                foreach (string file in filenames)
                {
                    PackageFilename packageFilename = new PackageFilename(file);
                    string full_version = packageFilename.VersionAndDateTimeComparable;
                    if (!sortedVersions.ContainsKey(full_version))
                        sortedVersions.Add(full_version, packageFilename);
                }
                List<PackageFilename> versions = new List<PackageFilename>();
                foreach (KeyValuePair<string, PackageFilename> v in sortedVersions)
                    versions.Add(v.Value);
                return versions.ToArray();
            }
            return new PackageFilename[0];
        }

        private bool FindBestVersion(Package package, VersionRange versionRange)
        {
            PackageFilename[] versions = RetrieveVersionsFor(package.Group, package.Name, package.Branch, package.Platform);
            PackageFilename best = FindBest(versions, versionRange);

            if (best == null)
                return false;

            package.SetVersion(Location, new ComparableVersion(best.Version));
            package.SetFilename(Location, best);
            package.SetURL(Location, RepoURL);

            string package_f = best.Filename;
            string package_d = Layout.PackageVersionDir(RepoURL, package.Group, package.Name, package.Platform, package.Branch, best.Version);

            package.SetSignature(Location, File.GetLastWriteTime(package_d + package_f));
            return true;
        }
    }
}
