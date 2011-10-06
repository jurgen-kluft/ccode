using System;
using System.IO;
using System.Collections.Generic;
using System.Security.Cryptography;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageSfvFile
    {
        private MD5CryptoServiceProvider mProvider;
        private string mEmptyMD5;
        private Dictionary<string, string> mFileHashes;

        private PackageSfvFile()
        {
            mProvider = new MD5CryptoServiceProvider();
            mEmptyMD5 = "DEAD";
            mFileHashes = new Dictionary<string, string>();
        }

        public static bool AreEqual(PackageSfvFile a, PackageSfvFile b)
        {
            if (a.mFileHashes.Count != b.mFileHashes.Count)
                return false;

            foreach (KeyValuePair<string, string> k in a.mFileHashes)
            {
                string val;
                if (b.mFileHashes.TryGetValue(k.Key, out val))
                {
                    if (String.Compare(k.Value, val, true) != 0)
                        return false;
                }
                else
                {
                    return false;
                }
            }
            return true;
        }

        public static PackageSfvFile LoadFromText(string base_path, string sfv_text)
        {
            PackageSfvFile sfvFile = new PackageSfvFile();

            StringReader reader = new StringReader(sfv_text);
            while (true)
            {
                string entry = reader.ReadLine();
                if (String.IsNullOrEmpty(entry))
                    break;

                if (entry.Trim().StartsWith(";"))   /// Skip comments
                    continue;

                // Get the MD5 and Filename
                int s = entry.IndexOf('*');
                if (s >= 0)
                {
                    string entry_md5 = entry.Substring(s + 1).Trim();
                    string entry_filename = base_path + entry.Substring(0, s).Trim();
                    sfvFile.mFileHashes.Add(entry_filename, entry_md5);
                }
            }
            reader.Close();

            return sfvFile;
        }

        public static PackageSfvFile Load(string URL, string filename, string base_path)
        {
            PackageSfvFile sfvFile;

            string md5_file = filename + ".md5";
            if (File.Exists(URL + md5_file))
            {
                // Load SFV file
                string sfv_text = File.ReadAllText(URL + md5_file);
                sfvFile = LoadFromText(base_path, sfv_text);
            }
            else
            {
                sfvFile = new PackageSfvFile();
            }

            return sfvFile;
        }

        public static PackageSfvFile New(List<string> files)
        {
            Loggy.RestoreConsoleCursor();

            int _origCol, _origRow;
            _origCol = Console.CursorLeft;
            _origRow = Console.CursorTop;

            Loggy.Info("Building Package: 0%");

            float progress_m = files.Count;
            float progress_i = 1.0f;
            PackageSfvFile sfvFile = new PackageSfvFile();
            foreach (string src_filename in files)
            {
                float progress = progress_i * 100.0f / progress_m;
                Console.SetCursorPosition(_origCol, _origRow);
                Console.Write("Building Package: {0}%  ", (int)System.Math.Min(progress, 100.0f));
                sfvFile.Add(src_filename);
                progress_i++;
            }
            return sfvFile;
        }

        public bool Add(string file)
        {
            if (File.Exists(file))
            {
                using (FileStream fs = new FileStream(file, FileMode.Open, FileAccess.Read))
                {
                    byte[] md5 = mProvider.ComputeHash(fs);
                    string md5_str = StringTools.MD5ToString(md5);
                    fs.Close();

                    if (!mFileHashes.ContainsKey(file))
                        mFileHashes.Add(file, md5_str);

                    return true;
                }
            }
            else
            {
                if (!mFileHashes.ContainsKey(file))
                    mFileHashes.Add(file, mEmptyMD5);
            }
            return false;
        }

        public void Unroot(string root)
        {
            Dictionary<string, string> fileHashes = new Dictionary<string, string>();
            foreach (KeyValuePair<string, string> k in mFileHashes)
            {
                string new_filename = k.Key.Replace(root, "");
                fileHashes.Add(new_filename, k.Value);
            }
            mFileHashes = fileHashes;
        }

        public void Root(Dictionary<string, string> rootPerFile)
        {
            Dictionary<string, string> fileHashes = new Dictionary<string, string>();
            foreach (KeyValuePair<string, string> k in mFileHashes)
            {
                string root;
                if (rootPerFile.TryGetValue(k.Key, out root))
                {
                    string dst_filename = k.Key;
                    if (String.IsNullOrEmpty(root))
                        dst_filename = Path.GetFileName(dst_filename);
                    else
                        dst_filename = root.EndWith('\\') + Path.GetFileName(dst_filename);

                    fileHashes.Add(dst_filename, k.Value);
                }
            }
            mFileHashes = fileHashes;
        }

        public PackageSfvFile Rooted(Dictionary<string, string> rootPerFile)
        {
            PackageSfvFile n = new PackageSfvFile();
            n.mFileHashes = new Dictionary<string, string>();
            foreach (KeyValuePair<string, string> k in mFileHashes)
                n.mFileHashes.Add(k.Key, k.Value);
            n.Root(rootPerFile);
            return n;
        }

        public void Root(string root)
        {
            Dictionary<string, string> fileHashes = new Dictionary<string, string>();
            foreach (KeyValuePair<string, string> k in mFileHashes)
                fileHashes.Add(root + k.Key, k.Value);
            mFileHashes = fileHashes;
        }

        public bool Save(string URL, string filename)
        {
            using (FileStream wfs = new FileStream(URL + filename + ".md5", FileMode.Create))
            {
                StreamWriter writer = new StreamWriter(wfs);
                foreach (KeyValuePair<string, string> k in mFileHashes)
                {
                    writer.WriteLine("{0} *{1}", k.Key, k.Value);
                }
                writer.Close();
                wfs.Close();
                return true;
            }
        }

        /// <summary>
        /// Save the .md5 file and use a dictionary to map source file to a new directory
        /// </summary>
        /// <param name="URL">The directory of where to store the .md5 file</param>
        /// <param name="filename">The name of the file, will be appended with the .md5 extension</param>
        /// <param name="filenameDstMap">The source file path to destination path map</param>
        /// <returns></returns>
        public bool Save(string URL, string filename, Dictionary<string, string> rootPerFile)
        {
            PackageSfvFile s = Rooted(rootPerFile);
            return s.Save(URL, filename);
        }

        public bool Verify(string URL)
        {
            bool ok = true;
            foreach (KeyValuePair<string,string> pair in mFileHashes)
            {
                string filename = pair.Key;
                if (File.Exists(URL + filename))
                {
                    string old_md5 = pair.Value;
                    string new_md5 = string.Empty;
                    using (FileStream rfs = new FileStream(URL + filename, FileMode.Open, FileAccess.Read))
                    {
                        byte[] new_md5_raw = mProvider.ComputeHash(rfs);
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
                    if (pair.Value != mEmptyMD5)
                    {
                        ok = false;
                        break;
                    }
                }
            }
            return ok;
        }
    }
}