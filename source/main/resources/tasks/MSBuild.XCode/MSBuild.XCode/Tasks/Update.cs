using System;
using System.IO;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System.Reflection;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class XCodeUpdate : Task
    {
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string XCodeRepoDir { get; set; }
        [Required]
        public string XCodeVersion { get; set; }

        public override bool Execute()
        {
            bool success = false;
            Loggy.TaskLogger = Log;

            CacheRepoDir = CacheRepoDir.EndWith('\\');

            XCodeRepoDir = XCodeRepoDir.EndWith('\\');
            XCodeRepoDir = XCodeRepoDir.Replace('\\', '/');

            ComparableVersion xcode_assembly_version = new ComparableVersion(Assembly.GetExecutingAssembly().GetName().Version.ToString());
            ComparableVersion xcode_msbuild_version = new ComparableVersion(XCodeVersion);
            
            if (xcode_assembly_version > xcode_msbuild_version)
            {
                Loggy.Error(String.Format("Error: Msbuild files with xCode version {0} is incompatible with xCode version {1}", xcode_msbuild_version.ToString(), xcode_assembly_version.ToString()));
                return false;
            }

            Loggy.Info(String.Format("XCode assembly version: {0}", xcode_assembly_version.ToString()));
            Loggy.Info(String.Format("XCode msbuild version: {0}", xcode_msbuild_version.ToString()));

            Environment.CurrentDirectory = CacheRepoDir;

            string dst_path = CacheRepoDir;
            string sub_path = @"com\virtuos\xcode\publish";

            if (!Directory.Exists(dst_path + sub_path))
                Directory.CreateDirectory(dst_path + sub_path);

            // First check if the repo is there, if not clone it, otherwise update it
            Mercurial.Repository hg_repo = new Mercurial.Repository(dst_path + sub_path);
            if (hg_repo.Exists)
            {
                try
                {
                    hg_repo.Pull();
                    hg_repo.Update();
                    success = true;
                }
                catch (System.Exception)
                {
                    success = false;
                    Loggy.Error(String.Format("Error: Updating xCode publish repository at {0} failed", dst_path + sub_path));
                }
            }
            else
            {
                try
                {
                    if (Directory.Exists(dst_path + sub_path))
                        Directory.Delete(dst_path + sub_path, true);
                    Directory.CreateDirectory(dst_path + sub_path);

                    hg_repo = new Mercurial.Repository(dst_path + sub_path);
                    Mercurial.CloneCommand clone_cmd = new Mercurial.CloneCommand();
                    clone_cmd.CompressedTransfer = false;
                    clone_cmd.Source = XCodeRepoDir;

                    hg_repo.Clone(clone_cmd);

                    success = true;
                }
                catch (System.Exception)
                {
                    success = false;
                    Loggy.Error(String.Format("Error: Cloning XCode publish repository at {0} to {1} failed", XCodeRepoDir, dst_path + sub_path));
                }
            }

            return success;
        }
    }
}
