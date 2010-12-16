using System;
using System.IO;
using System.Collections;
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
        public string RootDir { get; set; }
        [Required]
        public string Filename { get; set; }
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

            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            Mercurial.Repository hg_repo = new Mercurial.Repository(RootDir);
            Mercurial.StatusCommand hg_status = new Mercurial.StatusCommand();
            hg_status.AddArgument("-m");
            hg_status.AddArgument("-r");
            hg_status.AddArgument("-a");
            IEnumerable<Mercurial.FileStatus> repo_status = hg_repo.Status(hg_status);
            if (!repo_status.IsEmpty())
                return false;

            IEnumerable<Mercurial.Changeset> repo_outgoing = hg_repo.Outgoing();
            if (!repo_outgoing.IsEmpty())
                return false;

            XPackage package = new XPackage();
            package.Group = new XGroup(pom.Group);
            package.Name = pom.Name;
            package.Branch = Branch;
            package.Version = new XVersion(Version);
            package.Platform = Platform;

            // - Strip (Year).(Month).(Day).(Minute).(Second) from version of filename
            package.Path = RootDir + "target\\" + Filename;

            // - Commit version to remote package repository
            bool ok = XGlobal.RemoteRepo.Checkin(package);

            return ok;
        }

    }
}
