using System;
using System.IO;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageArchive
    {
        public static string RetrieveFileAsText(string package_filename, string filename)
        {
            string text = null;
            if (File.Exists(package_filename))
            {
                PackageZipper pz = PackageZipper.Open(package_filename, FileAccess.Read);
                PackageZipper.ZipFileEntry e;
                if (pz.FindEntryByFilename(filename, out e))
                {
                    using (MemoryStream stream = new MemoryStream())
                    {
                        if (pz.ExtractFile(e, stream))
                        {
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
                pz.Close();
            }
            return text;
        }

        public static bool RetrieveDependencies(string package_filename, out List<KeyValuePair<Package, Int64>> dependencies)
        {
            dependencies = new List<KeyValuePair<Package,Int64>>();

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
                    for (int i = 0; i < parts.Length; ++i)
                        parts[i] = parts[i].Trim();

                    Int64 v = 0;
                    Package p = new Package();
                    foreach (string part in parts)
                    {
                        if (part.StartsWith("name="))
                        {
                            p.Name = part.Split('=')[1].Trim();
                        }
                        else if (part.StartsWith("branch="))
                        {
                            p.Branch = part.Split('=')[1].Trim();
                        }
                        else if (part.StartsWith("group="))
                        {
                            p.Group = part.Split('=')[1].Trim();
                        }
                        else if (part.StartsWith("platform="))
                        {
                            p.Platform = part.Split('=')[1].Trim();
                        }
                        else if (part.StartsWith("language="))
                        {
                            p.Language = part.Split('=')[1].Trim();
                        }
                        else if (part.StartsWith("version="))
                        {
                            string version_str = part.Split('=')[1].Trim();
                            string[] version_items = version_str.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
                            version_str = String.Format("{0}.{1}.{2}", version_items[0], version_items[1], version_items[2]);
                            ComparableVersion version_cp = new ComparableVersion(version_str);
                            v = version_cp.ToInt();
                        }
                    }
                    dependencies.Add(new KeyValuePair<Package, Int64>(p, v));
                }
                return true;
            }
            catch (Exception e)
            {
                Loggy.Error(String.Format("Exception: {0}", e.ToString())); 
                return false;
            }
        }

        public static bool Create(PackageState package, PackageContent content, string rootURL)
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
                using (PackageZipper zip = PackageZipper.Create(zipPath, string.Empty))
                {
                    Loggy.RestoreConsoleCursor();
                    string progressFormatStr = String.Format("Creating Package {0} for platform {1}: ", package.Name, package.Platform) + "{0}%";
                   
                    int cl, ct;
                    cl = Console.CursorLeft;
                    ct = Console.CursorTop;
                    int max = files.Count;
                    int cnt = 1;
                    // Reserve a line in the log
                    Loggy.Info(String.Format(progressFormatStr, (cnt * 100) / max));

                    foreach (KeyValuePair<string, string> p in files)
                    {
                        Console.SetCursorPosition(cl, ct);
                        Console.Write(progressFormatStr, (cnt * 100) / max);
                        string src_filepath = p.Key;
                        string zip_filepath = String.IsNullOrEmpty(p.Value) ? (Path.GetFileName(src_filepath)) : (p.Value.EndWith('\\') + Path.GetFileName(src_filepath));
                        zip.AddFile(src_filepath, zip_filepath);
                        ++cnt;
                    }

                    zip.Close();
                    Loggy.Info("Done");
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
    }
}