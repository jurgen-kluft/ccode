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
            RepoDir = targetDir.EndWith('\\');
            Layout = new LayoutTarget();
            Location = ELocation.Target;
        }

        public string RepoDir { get; set; }
        public ELocation Location { get; set; }
        public ILayout Layout { get; set; }

        public bool Update(PackageInstance package)
        {
            // See if this package is in the target folder and valid
            string packageURL = Layout.PackageRootDir(RepoDir, package.Group.ToString(), package.Name, package.Platform);
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

        public bool Update(PackageInstance package, VersionRange versionRange)
        {
            // See if this package is in the target folder and valid for the version range
            if (Update(package))
            {
                if (versionRange.IsInRange(package.GetVersion(Location)))
                {
                    return true;
                }
            }
            return false;
        }

        public bool Add(PackageInstance package, ELocation from)
        {
            // Target normally gets packages from Share

            if (package.HasURL(from))
            {
                string targetURL = Layout.PackageVersionDir(RepoDir, package.Group.ToString(), package.Name, package.Platform, package.Branch, package.GetVersion(from));
                if (!Directory.Exists(targetURL))
                    Directory.CreateDirectory(targetURL);

                // The package itself is extracted in the Share Repo

                string current_t_filename = Path.GetFileNameWithoutExtension(package.GetFilename(from).ToString()) + ".t";
                string[] t_filenames = Directory.GetFiles(targetURL, String.Format("*{0}.t", package.Platform), SearchOption.TopDirectoryOnly);
                // Delete all old .t files?
                foreach (string t_filename in t_filenames)
                {
                    if (String.Compare(Path.GetFileNameWithoutExtension(t_filename), current_t_filename, true) != 0)
                    {
                        try { File.Delete(t_filename); } catch(IOException) { }
                    }
                }

                DateTime lastWriteTime = package.GetSignature(from);
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
                package.SetFilename(Location, new PackageFilename(package.GetFilename(from)));
                package.SetVersion(Location, package.GetVersion(from));
                package.SetSignature(Location, lastWriteTime);

                GenerateProps(package);

                return true;
            }
            return false;
        }

        private bool GenerateProps(PackageInstance package)
        {
            // <?xml version="1.0" encoding="utf-8"?>
            // <Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
            //   <PropertyGroup Condition="'$(Platform)'=='Win32'" Label="TargetDirs">
            //     <xunittest_TargetDir>$(SolutionDir)..\PACKAGE_REPO\com.virtuos.tnt\xunittest\xunittest+1.0.1.2011.1.7.20.40.44+default+Win32\</xunittest_TargetDir>
            //   </PropertyGroup>
            // </Project>
            // 

            string shareURL = package.ShareURL;

            string propsFilePath = Layout.PackageRootDir(RepoDir, package.Group.ToString(), package.Name, package.Platform) + package.Name + "." + package.Platform + ".props";
            using (FileStream stream = new FileStream(propsFilePath, FileMode.Create, FileAccess.Write))
            {
                using (StreamWriter writer = new StreamWriter(stream))
                {
                    writer.WriteLine("<?xml version=\"1.0\" encoding=\"utf-8\"?>");
                    writer.WriteLine("<Project ToolsVersion=\"4.0\" xmlns=\"http://schemas.microsoft.com/developer/msbuild/2003\">");
                    writer.WriteLine("  <PropertyGroup Condition=\"'$(Platform)'=='{0}'\" Label=\"TargetDirs\">", package.Platform);
                    writer.WriteLine("    <{0}_TargetDir>{1}</{0}_TargetDir>", package.Name, shareURL);
                    writer.WriteLine("  </PropertyGroup>");
                    writer.WriteLine("</Project>");
                    writer.Close();
                }
                stream.Close();
            }

            return true;
        }
    }
}
