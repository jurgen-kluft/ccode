using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System;
using System.IO;
using System.Collections.Generic;
using System.Security.Cryptography;
using System.Linq;
using System.Text;
using System.Runtime;
using Ionic.Zip;
using Ionic.Zlib;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class XCodeUpdate : Task
    {
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string XCodeRepoDir { get; set; }

        public override bool Execute()
        {
            bool success = false;
            Loggy.TaskLogger = Log;

            CacheRepoDir = CacheRepoDir.EndWith('\\');

            XCodeRepoDir = XCodeRepoDir.EndWith('\\');
            XCodeRepoDir = XCodeRepoDir.Replace('\\', '/');

            Environment.CurrentDirectory = CacheRepoDir;

            string dst_path = CacheRepoDir;
            string sub_path = @"com\virtuos\xcode\publish";

            if (!Directory.Exists(dst_path + sub_path))
                Directory.CreateDirectory(dst_path + sub_path);

            // First check if the repo is there, if not clone it, otherwise update it
            Mercurial.Repository hg_repo = new Mercurial.Repository(dst_path + sub_path);
            if (hg_repo.Exists)
            {
                // Update
                hg_repo.Update();
            }
            else
            {
                if (Directory.Exists(dst_path + sub_path))
                    Directory.Delete(dst_path + sub_path, true);
                Directory.CreateDirectory(dst_path + sub_path);

                // Clone
                hg_repo = new Mercurial.Repository(dst_path + sub_path);
                Mercurial.CloneCommand clone_cmd = new Mercurial.CloneCommand();
                clone_cmd.CompressedTransfer = false;
                clone_cmd.Source = XCodeRepoDir;

                hg_repo.Clone(clone_cmd);

                //Loggy.Error(String.Format("Error: Loading package failed in Package::Configs"));
            }

            return success;
        }
    }
}
