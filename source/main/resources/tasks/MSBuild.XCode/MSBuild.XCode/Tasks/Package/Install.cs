using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will copy a new package release to the local-package-repository. 
    /// </summary>
    public class PackageInstall : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string LocalRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        [Required]
        public string Platform { get; set; }
        [Required]
        public string Branch { get; set; }
        [Required]
        public string Version { get; set; }
        [Required]
        public string Filename { get; set; }

        public override bool Execute()
        {
            if (!RootDir.EndsWith("\\"))
                RootDir = RootDir + "\\";

            XPom pom = new XPom();
            pom.Load(RootDir + "pom.xml");

            XGlobal.TemplateDir = string.Empty;
            XGlobal.LocalRepoDir = LocalRepoDir;
            XGlobal.RemoteRepoDir = RemoteRepoDir;
            XGlobal.Initialize();

            XPackage package = new XPackage();
            package.Name = pom.Name;
            package.Group = pom.Group;
            package.Path = RootDir + "target\\" + Filename;
            package.Target = true;
            package.Version = pom.Versions.GetForPlatformWithBranch(Platform, Branch);
            package.Branch = Branch;
            package.Platform = Platform;

            // - Commit version to local package repository
            bool ok = XGlobal.LocalRepo.Checkin(package);

            return ok;
        }
    }
}
