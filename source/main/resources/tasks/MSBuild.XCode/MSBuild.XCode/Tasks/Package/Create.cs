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
    public class PackageCreate : Task
    {
        public string RootDir { get; set; }
        public string Platform { get; set; }

        [Output]
        public string Version{ get; set; }
        [Output]
        public string Filename { get; set; }

        public override bool Execute()
        {
            bool success = false;

            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            RootDir = RootDir.EndWith('\\');

            Environment.CurrentDirectory = RootDir;

            Global.TemplateDir = string.Empty;
            Global.CacheRepoDir = Global.CacheRepoDir;
            Global.RemoteRepoDir = Global.RemoteRepoDir;
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
            package.Version = package.Pom.Versions.GetForPlatformWithBranch(Platform, branch);
            package.Branch = branch;
            package.Platform = Platform;

            string filename;
            if (package.Create(out filename))
            {
                Filename = filename;
                success = true;
            }

            return success;
        }
    }
}
