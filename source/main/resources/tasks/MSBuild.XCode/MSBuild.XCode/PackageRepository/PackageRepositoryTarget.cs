using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using System.Security.Cryptography;
using Ionic.Zip;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    /// The target repository manages the dependencies of the root package.
    /// It stores (extracts) them to (relative to the root directory):
    /// - Target\PackageName\Platform
    /// 
    /// Note: We do need a way to identify the version of the extracted package.
    ///       This is done by using the filename of the marker file (.t).
    ///       
    /// From the root pom.xml we can collect its dependencies. From there we
    /// can load the pom.xml of those packages from the target. If one of them
    /// doesn't exist we stop and move to using the cache repository. Under
    /// normal circumstances we are able to load all dependency packages from
    /// the target folder without ever using the cache and remote repository.
    /// However we do need to query the cache and remote for a more up-to-date
    /// version. We can however have the cache and remote give us a signature
    /// that can tell us when a package has been updated by a install/deploy cmd.
    /// When the signature has changed we can move to the next phase and ask for 
    /// a better version of that package.
    /// </summary>
    public class PackageRepositoryTarget : IPackageRepository
    {
        public PackageRepositoryTarget(string targetDir)
        {
            RepoURL = targetDir.EndWith('\\');
            Layout = new LayoutTarget();
            Location = ELocation.Target;
            Valid = true;
        }

        public bool Valid { get; private set; }
        public string RepoURL { get; set; }
        public ELocation Location { get; private set; }
        private ILayout Layout { get; set; }

        public bool Query(Package package)
        {
            // See if this package is in the target folder and valid
            string packageURL = Layout.PackageRootDir(RepoURL, package.Group, package.Name, package.Platform);
            if (Directory.Exists(packageURL))
            {
                // A .t file needs to exist as well as a .props
                if (File.Exists(packageURL + package.Name + "." + package.Platform + ".props"))
                {
                    string[] t_filenames = Directory.GetFiles(packageURL, String.Format("*{0}.t", package.Platform), SearchOption.TopDirectoryOnly);
                    if (t_filenames.Length > 0)
                    {
                        // Find the one with the latest LastWriteTime
                        DateTime latest_datetime = DateTime.MinValue;
                        string latest_t_filename = string.Empty;
                        foreach (string t_filename in t_filenames)
                        {
                            DateTime datetime = File.GetLastWriteTime(t_filename);
                            if (datetime > latest_datetime)
                            {
                                latest_datetime = datetime;
                                latest_t_filename = t_filename;
                            }
                        }
                        // Extract the version from the filename
                        if (!String.IsNullOrEmpty(latest_t_filename))
                        {
                            package.TargetURL = packageURL;
                            package.TargetFilename = new PackageFilename(Path.GetFileNameWithoutExtension(latest_t_filename));
                            package.TargetVersion = package.TargetFilename.Version;
                            package.TargetSignature = latest_datetime;
                            return true;
                        }
                    }
                }
            }
            return false;
        }

        public bool Query(Package package, VersionRange versionRange)
        {
            // See if this package is in the target folder and valid for the version range
            if (Query(package))
            {
                if (versionRange.IsInRange(package.GetVersion(Location)))
                {
                    return true;
                }
            }
            return false;
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
            // Target normally gets packages from Share
            if (from.Location != ELocation.Share)
                return false;

            if (package.HasURL(from.Location))
            {
                string targetURL = Layout.PackageVersionDir(RepoURL, package.Group, package.Name, package.Platform, package.Branch, package.GetVersion(from.Location));
                if (!Directory.Exists(targetURL))
                    Directory.CreateDirectory(targetURL);

                // The package itself is extracted in the Share Repo

                string current_t_filename = Path.GetFileNameWithoutExtension(package.GetFilename(from.Location).ToString()) + ".t";
                string[] t_filenames = Directory.GetFiles(targetURL, String.Format("*{0}.t", package.Platform), SearchOption.TopDirectoryOnly);
                // Delete all old .t files?
                foreach (string t_filename in t_filenames)
                {
                    if (String.Compare(Path.GetFileNameWithoutExtension(t_filename), current_t_filename, true) != 0)
                    {
                        try { File.Delete(t_filename); } catch(IOException) { }
                    }
                }

                DateTime lastWriteTime = package.GetSignature(from.Location);
                FileInfo fi = new FileInfo(targetURL + current_t_filename);
                if (fi.Exists)
                {
                    fi.LastWriteTime = lastWriteTime;
                }
                else
                {
                    fi.Create().Close();
                    fi.LastWriteTime = lastWriteTime;
                }

                package.SetURL(Location, targetURL);
                package.SetFilename(Location, new PackageFilename(package.GetFilename(from.Location)));
                package.SetVersion(Location, package.GetVersion(from.Location));
                package.SetSignature(Location, lastWriteTime);

                //GenerateProps(package);

                return true;
            }
            return false;
        }

    }
}
