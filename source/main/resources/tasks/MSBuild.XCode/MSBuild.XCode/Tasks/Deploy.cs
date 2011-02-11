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
        public string Platform { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;

            RootDir = RootDir.EndWith('\\');
            
            if (!PackageInstance.IsInitialized)
            {
                PackageInstance.TemplateDir = string.Empty;
                PackageInstance.CacheRepoDir = CacheRepoDir;
                PackageInstance.RemoteRepoDir = RemoteRepoDir;
                PackageInstance.Initialize();
            }

            Mercurial.Repository hg_repo = new Mercurial.Repository(RootDir);
            Mercurial.Changeset hg_changeset = hg_repo.GetChangeSet();

            bool ok = false;
            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            if (package.IsValid)
            {
                package.Branch = hg_changeset.Branch;
                package.Platform = Platform;

                PackageRepositoryLocal localPackageRepo = new PackageRepositoryLocal(RootDir);
                if (localPackageRepo.Update(package))
                {
                    // Tag the mercurial repository
                    if (hg_repo.Exists)
                    {
                        ComparableVersion packageVersion = package.Pom.Versions.GetForPlatformWithBranch(package.Platform, package.Branch);

                        // Strip the date-time from the version, only tag with the package version + platform
                        hg_repo.Tag(packageVersion.ToString() + "_" + package.Platform);
                    }

                    // - Commit version to remote package repository from local
                    ok = PackageInstance.RemoteRepo.Add(package, localPackageRepo.Location);
                }
            }
            return ok;
        }
    }
}
