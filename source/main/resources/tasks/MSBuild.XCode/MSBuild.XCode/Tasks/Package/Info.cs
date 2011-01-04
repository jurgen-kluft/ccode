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
    public class PackageInfo : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            bool success = false;
            Loggy.TaskLogger = Log;

            RootDir = RootDir.EndWith('\\');
            CacheRepoDir = CacheRepoDir.EndWith('\\');
            RemoteRepoDir = RemoteRepoDir.EndWith('\\');

            Environment.CurrentDirectory = RootDir;

            Global.TemplateDir = string.Empty;
            Global.CacheRepoDir = CacheRepoDir;
            Global.RemoteRepoDir = RemoteRepoDir;
            Global.Initialize();

            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            Mercurial.Repository hg_repo = new Mercurial.Repository(RootDir);
            string branch = hg_repo.Branch();

            Package package = new Package();
            package.IsRoot = true;
            package.RootDir = RootDir;
            package.LoadPom();

            package.Name = package.Pom.Name;
            package.Group = package.Pom.Group;
            package.Version = null;
            package.Branch = branch;

            success = package.Info();

            return success;
        }
    }
}
