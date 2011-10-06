using System;
using System.IO;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    /// The local package repository manages the root package.
    /// It stores (creates) them to (relative to the root directory):
    /// - Target\Build\
    /// 
    /// Here we can also add the mechanism that will identify if there
    /// is any need to rebuild the package, maybe the last build package
    /// is already up-to-date so there will be no need to zip a new one.
    /// 
    /// </summary>
    public class PackageRepositoryLocal : IPackageRepository
    {
        public PackageRepositoryLocal(string rootDir)
        {
            RepoURL = rootDir.EndWith('\\');
            Layout = new LayoutLocal();
            Location = ELocation.Local;
            Valid = true;
        }

        public bool Valid { get; private set; }
        public string RepoURL { get; set; }
        public ELocation Location { get; private set; }
        public ILayout Layout { get; set; }

        public bool Query(PackageState package)
        {
            // See if there are one or more created packages in the Target\Build\ folder.
            // If so then together with the content description and doing a check if files have been
            // modified see if we need to create a new zipped package.
            // If pom.xml is not modified and all content of the previous package are identical
            // Lastly delete any old zip packages.
            string rootURL = RepoURL;
            string buildURL = String.Format("{0}target\\{1}\\build\\{2}\\", rootURL, package.Name, package.Platform);
            if (!Directory.Exists(buildURL))
                return false;

            string[] package_filenames = Directory.GetFiles(buildURL, String.Format("*{0}+{1}.zip", package.Branch, package.Platform), SearchOption.TopDirectoryOnly);
            if (package_filenames.Length > 0)
            {
                // Find the one with the latest LastWriteTime
                DateTime latest_datetime = DateTime.MinValue;
                string latest_package_filename = string.Empty;
                foreach (string package_filename in package_filenames)
                {
                    DateTime datetime = File.GetLastWriteTime(package_filename);
                    if (datetime > latest_datetime)
                    {
                        latest_datetime = datetime;
                        latest_package_filename = package_filename;
                    }
                }
                // Delete old .zip packages from the build folder
                if (!String.IsNullOrEmpty(latest_package_filename))
                {
                    string version = Layout.FilenameToVersion(Path.GetFileNameWithoutExtension(latest_package_filename));
                    package.LocalFilename = new PackageFilename(Path.GetFileNameWithoutExtension(latest_package_filename));
                    package.Platform = package.LocalFilename.Platform;
                    package.Branch = package.LocalFilename.Branch;
                    package.LocalVersion = new ComparableVersion(version);
                    package.LocalURL = buildURL;
                    package.LocalSignature = File.GetLastWriteTime(latest_package_filename);

                    // Delete the old .zip files
                    foreach (string package_filename in package_filenames)
                    {
                        if (String.Compare(package_filename, latest_package_filename, true) != 0)
                        {
                            try { File.Delete(package_filename); } catch (IOException) { }
                        }
                    }

                    return true;
                }
            }

            return false;
        }

        public bool Query(PackageState package, VersionRange versionRange)
        {
            // See if this package is in the target folder and valid for the version range
            if (Query(package))
            {
                return true;
            }
            return false;
        }

        public bool Link(PackageState package, out string filename)
        {
            string package_f = package.GetURL(Location) + package.GetFilename(Location);
            if (File.Exists(package_f))
            {
                filename = package_f;
                return true;
            }
            filename = string.Empty;
            return false;
        }

        public bool Download(PackageState package, string to_filename)
        {
            string src_path = package.GetURL(Location) + package.GetFilename(Location);
            if (File.Exists(src_path))
            {
                AsyncUnbufferedCopy xcopy = new AsyncUnbufferedCopy();
                try
                {
                    xcopy.ProgressFormatStr = "Copy package, progress: {0}%";
                    xcopy.AsyncCopyFileUnbuffered(src_path, to_filename, true, false, false, 1 * 1024 * 1024, true);
                }
                catch (Exception)
                {
                    return false;
                }
                finally
                {
                    xcopy = null;
                }
                return true;
            }
            return false;
        }

        public bool Submit(PackageState package, IPackageRepository from)
        {
            string package_f = package.GetURL(Location) + package.GetFilename(Location);
            return (File.Exists(package_f));
        }

    }
}
