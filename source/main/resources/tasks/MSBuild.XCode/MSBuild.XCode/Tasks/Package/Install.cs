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
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        [Required]
        public string Filename { get; set; }

        public override bool Execute()
        {
            RootDir = RootDir.EndWith('\\');

            Global.TemplateDir = string.Empty;
            Global.CacheRepoDir = CacheRepoDir;
            Global.RemoteRepoDir = RemoteRepoDir;
            Global.Initialize();

            Package package = new Package();
            package.IsRoot = true;
            package.RootDir = RootDir;
            package.LoadPom();
            package.SetPropertiesFromFilename(Filename);
            package.Name = package.Pom.Name;
            package.Group = package.Pom.Group;
            package.LocalURL = RootDir + "target\\" + Filename;

            // - Commit version to local package repository
            bool ok = package.Install();
            return ok;
        }
    }
}
