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
            RepoDir = rootDir;
            Layout = new LayoutLocal();
            Location = ELocation.Local;
        }

        public string RepoDir { get; set; }
        public ELocation Location { get; set; }
        public ILayout Layout { get; set; }

        public bool Update(PackageInstance package)
        {
            // See if there are one or more created packages in the Target\Build\ folder.
            // If so then together with the content description and doing a check if files have been
            // modified see if we need to create a new zipped package.
            // Lastly delete any old zip packages.
            string rootURL = RepoDir;
            string buildURL = rootURL + "target\\build\\";
            string[] package_filenames = Directory.GetFiles(buildURL, "*.zip", SearchOption.TopDirectoryOnly);
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
                    string version = Layout.FilenameToVersion(latest_package_filename);
                    package.LocalVersion = new ComparableVersion(version);
                    package.LocalFilename = new PackageFilename(Path.GetFileName(latest_package_filename));

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

        public bool Update(PackageInstance package, VersionRange versionRange)
        {
            // See if this package is in the target folder and valid for the version range
            if (Update(package))
            {
                return true;
            }
            return false;
        }

        public bool Add(PackageInstance package, ELocation from)
        {
            // From = Root
            // Create a new package from the root package and store in the local package repository
            bool success = false;
            if (from != ELocation.Root)
                return false;

            string branch = package.Branch;
            string platform = package.Platform;
            ComparableVersion version = package.Pom.Versions.GetForPlatformWithBranch(platform, branch);

            /// Delete the .sfv file
            string sfv_filename = package.Name + ".md5";
            string rootURL = RepoDir;
            string buildURL = rootURL + "target\\build\\";

            if (!Directory.Exists(buildURL))
                Directory.CreateDirectory(buildURL);

            if (File.Exists(buildURL + sfv_filename))
                File.Delete(buildURL + sfv_filename);

            /// 1) Create zip file
            /// 2) For every file create an MD5 and gather them into a sfv file
            /// 3) Remove root from every source file
            /// 4) Set the work directory
            /// 5) Add files to zip
            /// 6) Close
            /// 

            List<KeyValuePair<string, string>> content;
            if (!package.Pom.Content.TryGetValue(platform, out content))
            {
                if (!package.Pom.Content.TryGetValue("*", out content))
                {
                    package.LocalFilename = new PackageFilename();
                    return false;
                }
            }
            List<KeyValuePair<string, string>> files = new List<KeyValuePair<string,string>>();
            foreach (KeyValuePair<string, string> pair in content)
            {
                string src = rootURL + pair.Key;
                src = src.Replace("${Name}", package.Name);
                src = src.Replace("${Platform}", platform);

                Glob(src, pair.Value, files);
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
                Loggy.Add(String.Format("Error: PackageRepositoryLocal::Add, package must include pom.xml!"));
                package.LocalFilename = new PackageFilename();
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

            using (FileStream wfs = new FileStream(buildURL + sfv_filename, FileMode.Create))
            {
                StreamWriter writer = new StreamWriter(wfs);
                writer.WriteLine("; Generated by MSBuild.XCode");
                foreach (KeyValuePair<string, byte[]> k in md5Dictionary)
                {
                    writer.WriteLine("{0} *{1}", k.Key, StringTools.MD5ToString(k.Value));
                }
                writer.Close();
                wfs.Close();

                files.Add(new KeyValuePair<string, string>(buildURL + sfv_filename, Path.GetDirectoryName(sfv_filename)));
            }

            // Add VCS Information file to the package
            if (File.Exists(buildURL + "vcs.info"))
                files.Add(new KeyValuePair<string, string>(rootURL + "vcs.info", ""));
            // Add Dependency Information file to the package
            if (File.Exists((rootURL + "dependencies.info")))
                files.Add(new KeyValuePair<string, string>(rootURL + "dependencies.info", ""));

            package.LocalFilename = new PackageFilename(package.Name, version, branch, platform);
            package.LocalFilename.DateTime = DateTime.Now;

            if (File.Exists(buildURL + package.LocalFilename.ToString()))
            {
                try { File.Delete(buildURL + package.LocalFilename.ToString()); }
                catch (Exception) { }
            }

            using (ZipFile zip = new ZipFile(buildURL + package.LocalFilename.ToString()))
            {
                foreach (KeyValuePair<string, string> p in files)
                    zip.AddFile(p.Key, p.Value);

                zip.Save();
                package.LocalURL = buildURL;
                success = true;
            }
            return success;
        }


        private static void Glob(string src, string dst, List<KeyValuePair<string, string>> files)
        {
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
    }
}
