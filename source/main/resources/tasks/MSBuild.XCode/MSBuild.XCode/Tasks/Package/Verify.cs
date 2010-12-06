using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using System.Security.Cryptography;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will copy a new package release to the local-package-repository. 
    ///	Also updates 'latest'.
    /// </summary>
    public class PackageVerify : Task
    {
        public string Path { get; set; }

        public override bool Execute()
        {
            bool ok = false;

            if (!Path.EndsWith("\\"))
                Path = Path + "\\";

            Environment.CurrentDirectory = Path;

            XPackage package = new XPackage();
            package.Load(Path + "package.xml");

            string md5_file = Path + package.Name + ".MD5";
            if (File.Exists(md5_file))
            {
                MD5CryptoServiceProvider md5_provider = new MD5CryptoServiceProvider();

                // Load MD5 file
                ok = true;
                string[] lines = File.ReadAllLines(md5_file);
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
                    string old_md5 = entry.Substring(s+1).Trim();
                    string filename = entry.Substring(0, s).Trim();

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
                return ok;
            }

            return ok;
        }
    }
}
