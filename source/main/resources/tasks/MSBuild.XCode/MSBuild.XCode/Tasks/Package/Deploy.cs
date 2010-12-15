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
    ///	Will copy a new package release to the remote-package-repository. 
    /// </summary>
    public class PackageDeploy : Task
    {
        [Required]
        public string Path { get; set; }
        [Required]
        public string RepoPath { get; set; }

        [Required]
        public string Platform { get; set; }
        [Required]
        public string Branch { get; set; }
        [Required]
        public string Version { get; set; }

        public override bool Execute()
        {
            if (!Path.EndsWith("\\"))
                Path = Path + "\\";

            XPackage package = new XPackage();
            package.Load(Path + "package.xml");

            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            // - Strip (Year).(Month).(Day).(Minute).(Second) from version of filename
            // - Commit version to remote package repository

            // - Commit version to local package repository
            XPackageRepository repo = new XPackageRepository(RepoPath);
            bool ok = repo.Commit(package.Group.Full, Path, package.Name, Branch, Platform, new XVersion(Version));

            return ok;
        }

    }
}
