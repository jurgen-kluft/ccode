using System;
using System.IO;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;
using Ionic.Zip;

namespace MSBuild.XCode
{
    public class PackageArchive
    {
        public static string RetrieveFileAsText(string package_filename, string filename)
        {
            string text = null;
            if (File.Exists(package_filename))
            {
                ZipFile zip = new ZipFile(package_filename);
                if (zip.Entries.Count > 0)
                {
                    ZipEntry entry = zip[filename];
                    if (entry != null)
                    {
                        using (MemoryStream stream = new MemoryStream())
                        {
                            entry.Extract(stream);
                            stream.Position = 0;
                            using (StreamReader reader = new StreamReader(stream))
                            {
                                text = reader.ReadToEnd();
                                reader.Close();
                                stream.Close();
                            }
                        }
                    }
                }
                zip.Dispose();
            }
            return text;
        }

        public static bool RetrieveDependencies(string package_filename, out List<KeyValuePair<string, Int64>> dependencies)
        {
            dependencies = new List<KeyValuePair<string, Int64>>();

            string text = RetrieveFileAsText(package_filename, "dependencies.info");
            if (text == null)
                return false;

            try
            {
                StringReader sr = new StringReader(text);
                // Skip the first line which contains information
                // about the root package
                string line = sr.ReadLine();
                while (true)
                {
                    line = sr.ReadLine();
                    if (String.IsNullOrEmpty(line))
                        break;

                    string[] parts = line.Split(new char[] { ',' }, StringSplitOptions.RemoveEmptyEntries);

                    string name = parts[0].Trim();
                    string version_str = parts[1].Replace("version=", "").Trim();
                    string[] version_items = version_str.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
                    version_str = String.Format("{0}.{1}.{2}", version_items[0], version_items[1], version_items[2]);
                    ComparableVersion version_cp = new ComparableVersion(version_str);
                    Int64 version_int = version_cp.ToInt();
                    dependencies.Add(new KeyValuePair<string, Int64>(name, version_int));
                }
                return true;
            }
            catch (Exception e)
            {
                Loggy.Error(String.Format("Exception: {0}", e.ToString())); 
                return false;
            }
        }

        public static bool Create(Package package, PackageContent content, string rootURL)
        {
            try
            {
                ComparableVersion version = package.CreateVersion;
                string branch = package.Branch;
                string platform = package.Platform;

                /// Delete the SFV file
                string sfv_filename = package.Name + ".md5";
                string buildURL = rootURL + "target\\" + package.Name + "\\build\\" + platform + "\\";

                if (!Directory.Exists(buildURL))
                    Directory.CreateDirectory(buildURL);

                if (!File.Exists(rootURL + "pom.xml"))
                {
                    Loggy.Error(String.Format("Error: PackageRepositoryLocal::Add, package root has no pom.xml!"));
                    package.LocalFilename = new PackageFilename();
                    return false;
                }
                if (!File.Exists(buildURL + "dependencies.info"))
                {
                    Loggy.Error(String.Format("Error: PackageRepositoryLocal::Add, package must include dependencies.info!"));
                    package.LocalFilename = new PackageFilename();
                    return false;
                }
                if (!File.Exists(buildURL + "vcs.info"))
                {
                    Loggy.Error(String.Format("Error: PackageRepositoryLocal::Add, package must include vcs.info!"));
                    package.LocalFilename = new PackageFilename();
                    return false;
                }

                Dictionary<string, string> files;

                if (!content.Collect(package.Name, platform, rootURL, out files))
                {
                    package.LocalFilename = new PackageFilename();
                    return false;
                }

                // Add VCS Information file to the package
                files.Add(buildURL + "vcs.info", "");
                files.Add(buildURL + "dependencies.info", "");

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
                    Loggy.Error(String.Format("Error: PackageRepositoryLocal::Add, package must include pom.xml!"));
                    package.LocalFilename = new PackageFilename();
                    return false;
                }

                PackageSfvFile newSfvFile = PackageSfvFile.New(new List<string>(files.Keys));

                if (package.LocalFilename != null)
                {
                    string old_package_filename = package.LocalFilename.ToString();
                    PackageSfvFile oldSfvFile = PackageSfvFile.LoadFromText("", RetrieveFileAsText(buildURL + old_package_filename, sfv_filename));
                    PackageSfvFile convertedNewSfvFile = newSfvFile.Rooted(files);

                    // If the content of the new package is the same as the package that we
                    // created before than it makes no sense to build a new one.
                    if (PackageSfvFile.AreEqual(oldSfvFile, convertedNewSfvFile) && File.Exists(buildURL + old_package_filename))
                    {
                        // Actually we can just rename the current zip file
                        package.LocalFilename = new PackageFilename(package.Name, version, branch, platform);
                        package.LocalFilename.DateTime = DateTime.Now;
                        package.LocalVersion = version;
                        package.LocalSignature = package.LocalFilename.DateTime;

                        // Rename
                        File.Move(buildURL + old_package_filename, buildURL + package.LocalFilename.ToString());
                        return true;
                    }
                }

                // Delete the old md5 file
                if (File.Exists(buildURL + sfv_filename))
                    File.Delete(buildURL + sfv_filename);

                if (!newSfvFile.Save(buildURL, package.Name, files))
                {
                    Loggy.Error(String.Format("Error: PackageRepositoryLocal::Add, failed to save sfv file!"));
                    return false;
                }
                
                if (!newSfvFile.Save(buildURL, package.Name + ".source"))
                {
                    Loggy.Error(String.Format("Error: PackageRepositoryLocal::Add, failed to save sfv source file!"));
                    return false;
                }

                // Add SFV file to the package
                if (File.Exists(buildURL + sfv_filename))
                    files.Add(buildURL + sfv_filename, "");

                // Construct the full filename including name, version, date-time, branch and platform
                package.LocalFilename = new PackageFilename(package.Name, version, branch, platform);
                package.LocalFilename.DateTime = DateTime.Now;
                package.LocalVersion = version;
                package.LocalSignature = package.LocalFilename.DateTime;

                if (File.Exists(buildURL + package.LocalFilename.ToString()))
                {
                    try { File.Delete(buildURL + package.LocalFilename.ToString()); }
                    catch (Exception) { }
                }

                string zipPath = buildURL + package.LocalFilename.ToString();
                using (ZipFile zip = new ZipFile(zipPath))
                {
                    foreach (KeyValuePair<string, string> p in files)
                        zip.AddFile(p.Key, p.Value);

                    ZipSaveProgress progress = new ZipSaveProgress(files.Count);
                    Console.Write("Creating Package {0} for platform {1}: ", package.Name, package.Platform);
                    zip.SaveProgress += progress.EventHandler;
                    zip.Save();
                    Console.WriteLine("Done");
                    File.SetLastWriteTime(zipPath, package.LocalSignature);
                    package.LocalURL = buildURL;
                    return true;
                }
            }
            catch (Exception e)
            {
                Loggy.Error(String.Format("Exception: {0}", e.ToString())); 
            }

            return false;
        }

        public class ZipSaveProgress
        {
            private int mNumCharsDisplayed = 0;

            private int mNumEntries;
            public ZipSaveProgress(int numEntries)
            {
                mNumEntries = numEntries;
            }

            public void EventHandler(object sender, SaveProgressEventArgs e)
            {
                int numChars = ((100 * e.EntriesSaved) / mNumEntries);
                while (mNumCharsDisplayed < numChars)
                {
                    Console.Write("{0, 3}%", numChars);
                    Console.CursorLeft = Console.CursorLeft - 4;

                    ++mNumCharsDisplayed;
                }
            }
        }


    }
}