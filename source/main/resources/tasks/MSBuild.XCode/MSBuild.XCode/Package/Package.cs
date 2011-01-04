using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using System.Security.Cryptography;
using Ionic.Zip;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public enum ELocation 
    {
        Remote,     ///< Remote package repository
        Cache,      ///< Cache package repository (on local machine)
        Local        ///< Local package, a 'Created' package
    }

    public class Package
    {
        ///< if IsRoot=True then 
        ///     LocalURL = 'Created' package ready for 'Install' or 'Deploy'
        ///     Pom      = Loaded from Cache
        ///  else 
        ///     LocalURL = Rootdir
        ///     Pom      = Loaded from RootDir
        public bool IsRoot { get; set; }
        
        public string RootDir { get; set; }
        public Group Group { get; set; }
        public string Name { get; set; }
        public string Branch { get; set; }
        public ComparableVersion Version { get; set; }
        public string Platform { get; set; }

        public bool RemoteExists { get { return !String.IsNullOrEmpty(RemoteURL); } }
        public bool CacheExists { get { return !String.IsNullOrEmpty(CacheURL); } }
        public bool LocalExists { get { return !String.IsNullOrEmpty(LocalURL); }  }

        public string RemoteURL { get; set; }
        public string CacheURL { get; set; }
        public string LocalURL { get; set; }    

        public Pom Pom { get; set; }
        public bool HasPom { get { return Pom != null; } }
        public bool IsFinalPom { get; set; }

        public void SetPropertiesFromFilename(string filename)
        {
            filename = Path.GetFileNameWithoutExtension(filename);

            string[] parts = filename.Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
            if (parts.Length == 4)
            {
                Name = parts[0];
                Version = new ComparableVersion(parts[1]);
                Branch = parts[2];
                Platform = parts[3];
            }
        }

        public void SetURL(ELocation location, string url)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteURL = url; break;
                case ELocation.Cache: CacheURL = url; break;
                case ELocation.Local: LocalURL = url; break;
            }
        }

        public string GetURL(ELocation location)
        {
            string url = string.Empty;
            switch (location)
            {
                case ELocation.Remote: url = RemoteURL; break;
                case ELocation.Cache: url = CacheURL; break;
                case ELocation.Local: url = LocalURL; break;
            }
            return url;
        }

        public bool Info()
        {
            if (HasPom)
                return Pom.Info();
            return false;
        }

        public bool Extract()
        {
            if (IsRoot)
                return false;

            if (CacheExists)    // && File.Exists(CacheURL)), should exist since the PackageRepository assigned it
            {
                ZipFile zip = new ZipFile(CacheURL);
                string path = RootDir + "target\\" + Name + "\\" + Platform + "\\";
                zip.ExtractAll(path, ExtractExistingFileAction.OverwriteSilently);

                DateTime lastWriteTime = File.GetLastWriteTime(CacheURL);
                FileInfo fi = new FileInfo(RootDir + "target\\" + Path.GetFileNameWithoutExtension(CacheURL) + ".t");
                if (fi.Exists)
                {
                    fi.LastWriteTime = lastWriteTime;
                }
                else
                {
                    fi.Create().Close();
                    fi.LastWriteTime = lastWriteTime;
                }
                return true;
            }
            return false;
        }

        public bool VerifyBeforeExtract()
        {
            if (IsRoot)
                return false;

            if (CacheExists)
            {
                // Verify 'Extracted' package
                if (Verify())
                    return true;

                return Extract();
            }
            return false;
        }

        private void Glob(string src, string dst, List<KeyValuePair<string,string>> files)
        {
            src = src.Replace("${Name}", Name);
            src = src.Replace("${Platform}", Platform);

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

        public bool Create(out string Filename)
        {
            bool success = false;

            /// Delete the .sfv file
            string sfv_filename = Name + ".md5";
            string dir = RootDir + "target\\" + Name + "\\" + Platform + "\\";
            if (File.Exists(dir + sfv_filename))
                File.Delete(dir + sfv_filename);

            /// 1) Create zip file
            /// 2) For every file create an MD5 and gather them into a sfv file
            /// 3) Remove root from every source file
            /// 4) Set the work directory
            /// 5) Add files to zip
            /// 6) Close
            /// 

            List<KeyValuePair<string, string>> content;
            if (!Pom.Content.TryGetValue(Platform, out content))
            {
                if (!Pom.Content.TryGetValue("*", out content))
                {
                    Filename = string.Empty;
                    return false;
                }
            }
            List<KeyValuePair<string, string>> files = new List<KeyValuePair<string,string>>();
            foreach (KeyValuePair<string, string> pair in content)
            {
                Glob(RootDir + pair.Key, pair.Value, files);
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
                Loggy.Add(String.Format("Error: Package::Create, package must include pom.xml!"));
                Filename = string.Empty;
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

            if (!Directory.Exists(dir))
                Directory.CreateDirectory(dir);

            using (FileStream wfs = new FileStream(dir + sfv_filename, FileMode.Create))
            {
                StreamWriter writer = new StreamWriter(wfs);
                writer.WriteLine("; Generated by MSBuild.XCode");
                foreach (KeyValuePair<string, byte[]> k in md5Dictionary)
                {
                    writer.WriteLine("{0} *{1}", k.Key, StringTools.MD5ToString(k.Value));
                }
                writer.Close();
                wfs.Close();

                files.Add(new KeyValuePair<string, string>(dir + sfv_filename, Path.GetDirectoryName(sfv_filename)));
            }

            // Add VCS Information file to the package
            if (File.Exists(RootDir + "vcs.info"))
                files.Add(new KeyValuePair<string, string>(RootDir + "vcs.info", ""));

            // Versioning:
            // - Get Version
            // - Build = (DateTime.Year).(DateTime.Month).(DateTime.Day).(DateTime.Hour).(DateTime.Minute).(DateTime.Second)

            // @TODO - 
            // - VCS: Get revision information and write it into a file and add that file to the source file list to include it into the zip package
            // - Write a 'marker' in the 

            ComparableVersion version = Pom.Versions.GetForPlatformWithBranch(Platform, Branch);

            DateTime t = DateTime.Now;
            string versionStr = version.ToString() + String.Format(".{0:yyyy.M.d.H.m.s}", t);

            Filename = Name + "+" + versionStr + "+" + Branch + "+" + Platform + ".zip";
            if (File.Exists(RootDir + "target\\" + Filename))
            {
                try { File.Delete(RootDir + "target\\" + Filename); }
                catch (Exception) { }
            }

            using (ZipFile zip = new ZipFile(RootDir + "target\\" + Filename))
            {
                foreach (KeyValuePair<string, string> p in files)
                    zip.AddFile(p.Key, p.Value);

                zip.Save();
                success = true;
            }
            return success;
        }

        public bool Verify()
        {
            bool ok = false;

            if (IsRoot)
                return ok;

            string targetDir = RootDir + "target\\";
            string subDir = Name + "\\" + Platform + "\\";
            string md5_file = subDir + Name + ".MD5";
            if (File.Exists(targetDir + md5_file))
            {
                DateTime packageLastWriteTime = File.GetLastWriteTime(CacheURL);
                FileInfo fi = new FileInfo(targetDir + Path.GetFileNameWithoutExtension(CacheURL) + ".t");
                bool markerFileExists = fi.Exists;
                bool markerFileUpToDate = markerFileExists && (fi.LastWriteTime == packageLastWriteTime);
                if (markerFileExists && markerFileUpToDate)
                {
                    MD5CryptoServiceProvider md5_provider = new MD5CryptoServiceProvider();

                    // Load MD5 file
                    ok = true;
                    string[] lines = File.ReadAllLines(targetDir + md5_file);

                    // MD5 is relative to its own location
                    foreach (string entry in lines)
                    {
                        if (entry.Trim().StartsWith(";"))
                            continue;

                        // Get the MD5 and Filename
                        int s = entry.IndexOf('*');
                        if (s == -1)
                        {
                            ok = false;
                            break;
                        }
                        string old_md5 = entry.Substring(s + 1).Trim();
                        string filename = targetDir + subDir + entry.Substring(0, s).Trim();

                        if (File.Exists(filename))
                        {
                            string new_md5 = string.Empty;
                            using (FileStream rfs = new FileStream(filename, FileMode.Open, FileAccess.Read))
                            {
                                byte[] new_md5_raw = md5_provider.ComputeHash(rfs);
                                new_md5 = StringTools.MD5ToString(new_md5_raw);
                                rfs.Close();
                            }

                            if (String.Compare(old_md5, new_md5) != 0)
                            {
                                ok = false;
                                break;
                            }
                        }
                        else
                        {
                            // File doesn't exist anymore
                            ok = false;
                            break;
                        }
                    }
                }
                return ok;
            }
            return ok;
        }

        public bool Install()
        {
            bool success = Global.CacheRepo.Add(this, ELocation.Local);
            return success;
        }

        public bool Deploy()
        {
            bool success = false;
            if (Global.CacheRepo.Update(this))
            {
                success = Global.RemoteRepo.Add(this, ELocation.Cache);
            }
            return success;
        }

        public bool LoadPom()
        {
            Pom = null;
            IsFinalPom = false;

            if (IsRoot)
            {
                if (!String.IsNullOrEmpty(RootDir) && Directory.Exists(RootDir))
                {
                    string pomPath = RootDir + "pom.xml";
                    if (File.Exists(pomPath))
                    {
                        Pom = new Pom(Pom.EType.Main);
                        Pom.Load(pomPath);
                        return true;
                    }
                }
            }
            else if (CacheExists && File.Exists(CacheURL))
            {
                ZipFile zip = new ZipFile(CacheURL);
                if (zip.Entries.Count > 0)
                {
                    ZipEntry entry = zip["pom.xml"];
                    if (entry != null)
                    {
                        using (MemoryStream stream = new MemoryStream())
                        {
                            entry.Extract(stream);
                            stream.Position = 0;
                            using (StreamReader reader = new StreamReader(stream))
                            {
                                string xml = reader.ReadToEnd();
                                reader.Close();
                                stream.Close();
                                Pom = new Pom(Pom.EType.Dependency);
                                Pom.LoadXml(xml);
                                return true;
                            }
                        }
                    }
                }
            }
            return false;
        }

        public bool LoadFinalPom()
        {
            if (LoadPom())
            {
                Pom.PostLoad();
                IsFinalPom = true;
                return true;
            }
            return false;
        }

        public bool BuildDependencies(string Platform)
        {
            if (HasPom && IsRoot)
            {
                return Pom.BuildDependencies(Platform);
            }
            return false;
        }

        public bool BuildAllDependencies()
        {
            if (HasPom && IsRoot)
            {
                foreach (string platform in Pom.Platforms)
                {
                    if (!Pom.BuildDependencies(platform))
                        return false;
                }
                return true;
            }
            return false;
        }

        public bool PrintDependencies(string Platform)
        {
            if (HasPom && IsRoot)
            {
                // Has valid dependency tree ?
                Pom.PrintDependencies(Platform);
                return true;
            }
            return false;
        }

        public bool PrintAllDependencies()
        {
            if (HasPom && IsRoot)
            {
                foreach (string p in Pom.Platforms)
                {
                    Loggy.Add(String.Format("Dependencies for platform : {0}", p));
                    Loggy.Indent += 1;
                    // Has valid dependency tree ?
                    Pom.PrintDependencies(p);
                    Loggy.Indent -= 1;
                }
                return true;
            }
            return false;
        }
 
        public bool SyncDependencies(string Platform)
        {
            if (HasPom && IsRoot)
            {
                // Has valid dependency tree ?
                return Pom.SyncDependencies(Platform);
            }
            return false;
        }

        public bool SyncAllDependencies()
        {
            if (HasPom && IsRoot)
            {
                foreach (string platform in Pom.Platforms)
                {
                    // Has valid dependency tree ?
                    if (!Pom.SyncDependencies(platform))
                        return false;
                }
                return true;
            }
            return false;
        }

        public bool GenerateProjects()
        {
            if (!IsRoot || !HasPom)
                return false;
            Pom.GenerateProjects(RootDir);
            return true;
        }

        public bool GenerateSolution()
        {
            if (!IsRoot || !HasPom)
                return false;
            Pom.GenerateSolution(RootDir);
            return true;
        }

    }
}
