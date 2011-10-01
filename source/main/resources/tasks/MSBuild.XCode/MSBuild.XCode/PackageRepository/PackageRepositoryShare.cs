﻿using System;
using System.IO;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    /// The Share repository manages the dependencies of all root packages and the purpose of this
    /// package repository is to reduce the duplication that would otherwise exist if we use the
    /// Target repository to extract the content of dependency packages. Now these packages are
    /// able to be shared between root packages, hence the name 'Share'.
    /// 
    /// It stores (extracts) them in a folder in the working directory:
    /// - PACKAGE_REPO
    /// 
    /// It stores the packages like this:
    ///     group\package name\full filename\{content}
    ///     
    /// Example:
    ///     com.virtuos.sdk\sdk_3ds\sdk_3ds+1.0.0.2011.1.12.13.51.28+default+N3DS
    ///     com.virtuos.tnt\xbase\xbase+1.0.0.2011.1.7.20.38.54+default+Win32
    /// 
    /// The change to the pom.xml is:
    /// - Packages should use $(PACKAGE_NAME_TargetDir) for the location of their package
    ///   e.g.: $(xbase_TargetDir), like $(xbase_TargetDir)source\main\include
    /// 
    /// The change to Visual Studio Project files is that we need to supply it with the
    /// value for $(PACKAGE_NAME_TargetDir) and this is platform dependent since every
    /// package is split on a platform basis.
    /// 
    /// We can do this by using one or more .props files which contain constructions like:
    /// 
    ///      <PropertyGroup Condition="'$(Platform)'=='Win32'" Label="TargetDirs">
    ///        <xbase_TargetDir>$(SolutionDir)..\PACKAGE_REPO\com.virtuos.tnt\xbase\xbase+1.0.0.2011.1.7.20.38.54+default+Win32\</xbase_TargetDir>
    ///      </PropertyGroup>
    /// 
    /// Generating the .vcxproj project files should include constructions taht import the above .props files.
    /// 
    /// </summary>
    public class PackageRepositoryShare : IPackageRepository
    {
        public PackageRepositoryShare(string repoURL)
        {
            RepoURL = repoURL.EndWith('\\');
            Layout = new LayoutShare();
            Location = ELocation.Share;
            Valid = true;
        }

        public bool Valid { get; private set; }
        public string RepoURL { get; set; }
        public ELocation Location { get; private set; }
        private ILayout Layout { get; set; }

        public bool Query(PackageState package)
        {
            // See if this package is in the share folder and valid
            // The package has to have Target properties!
            // This is a passive repository, it will not try to find the
            // best version in its repository.
            if (!package.CacheExists)
                return false;

            IPackageFilename pf = package.GetFilename(ELocation.Cache);
            string packageURL = RepoURL + package.Group + "\\" + package.Name + "\\" + pf.FilenameWithoutExtension + "\\";
            if (Directory.Exists(packageURL))
            {
                if (File.Exists(packageURL + "pom.xml"))
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
                            package.ShareURL = packageURL;
                            package.ShareFilename = new PackageFilename(Path.GetFileNameWithoutExtension(latest_t_filename));
                            package.ShareVersion = package.CacheFilename.Version;
                            package.ShareSignature = latest_datetime;
                            return true;
                        }
                    }
                }
            }
            return false;
        }

        public bool Query(PackageState package, VersionRange versionRange)
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

        public bool Download(PackageState package, string to_filename)
        {
            return false;
        }

        public bool Link(PackageState package, out string filename)
        {
            filename = string.Empty;
            return false;
        }

        public bool Submit(PackageState package, IPackageRepository from)
        {
            // Cannot add from Target! should normally be added from Cache
            if (from.Location == ELocation.Target)
                return false;

            string packageFilenameInCache;
            if (!from.Link(package, out packageFilenameInCache))
                return false;

            if (File.Exists(packageFilenameInCache))
            {
                PackageFilename pf = new PackageFilename(packageFilenameInCache);
                string shareURL = RepoURL + package.Group + "\\" + package.Name + "\\" + pf.FilenameWithoutExtension + "\\";
                {
                    PackageZipper zip = PackageZipper.Open(packageFilenameInCache, FileAccess.Read);
                    DateTime now = DateTime.Now;
                    string destExtractDir = Path.GetPathRoot(RepoURL) + "temp\\" + String.Format("tmp.{0}.{1}.{2}.{3}.{4}.{5}\\", now.Year, now.Month, now.Day, now.Hour, now.Minute, now.Second);
                    Directory.CreateDirectory(destExtractDir);
                    zip.ExtractTo(destExtractDir);
                    zip.Close();
                    zip = null;

                    // Moving a directory only works when the destination doesn't exist
                    // So make sure the destination directory is not there
                    DirectoryInfo shareURLDirInfo;
                    if (Directory.Exists(shareURL))
                    {
                        shareURLDirInfo = new DirectoryInfo(shareURL);
                        shareURLDirInfo.Remove();
                    }

                    Directory.CreateDirectory(shareURL);
                    Directory.Delete(shareURL, false);

                    Directory.Move(destExtractDir, shareURL);

                    // Set this directory and all its children (files & directories) to readonly
                    shareURLDirInfo= new DirectoryInfo(shareURL);
                    shareURLDirInfo.Readonly();
                }

                string current_t_filename = pf.FilenameWithoutExtension + ".t";
                string[] t_filenames = Directory.GetFiles(shareURL, "*.t", SearchOption.TopDirectoryOnly);

                // Delete all old .t files?
                foreach (string t_filename in t_filenames)
                {
                    if (String.Compare(Path.GetFileNameWithoutExtension(t_filename), current_t_filename, true) != 0)
                    {
                        try { File.Delete(t_filename); }
                        catch (IOException) { }
                    }
                }

                DateTime lastWriteTime = package.GetSignature(from.Location);
                FileInfo fi = new FileInfo(shareURL + current_t_filename);
                if (fi.Exists)
                {
                    fi.LastWriteTime = lastWriteTime;
                }
                else
                {
                    fi.Create().Close();
                    fi.LastWriteTime = lastWriteTime;
                }

                package.SetURL(Location, shareURL);
                package.SetFilename(Location, pf);
                package.SetVersion(Location, new ComparableVersion(pf.Version));
                package.SetSignature(Location, lastWriteTime);

                return true;
            }
            return false;
        }
    }
}
