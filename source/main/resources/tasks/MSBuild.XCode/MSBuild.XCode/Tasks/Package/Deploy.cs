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
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            RootDir = RootDir.EndWith('\\');

            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            Mercurial.Repository hg_repo = new Mercurial.Repository(RootDir);
            Mercurial.StatusCommand hg_status = new Mercurial.StatusCommand(Mercurial.FileStatusIncludes.MARM);
            IEnumerable<Mercurial.FileStatus> repo_status = hg_repo.Status(hg_status);
            if (!repo_status.IsEmpty())
            {
                Loggy.Add("Not everything checked-in!");
                return false;
            }

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
            bool ok = package.Deploy();
            return ok;
        }
    }
}
