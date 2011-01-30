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
        public string Platform { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        [Required]
        public string Filename { get; set; }

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;
            RootDir = RootDir.EndWith('\\');

            PackageInstance.TemplateDir = string.Empty;
            PackageInstance.CacheRepoDir = CacheRepoDir;
            PackageInstance.RemoteRepoDir = RemoteRepoDir;
            PackageInstance.Initialize();

            bool ok = false;

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            if (package.IsValid)
            {
                package.Platform = Platform;

                PackageRepositoryLocal localPackageRepo = new PackageRepositoryLocal(RootDir);
                if (localPackageRepo.Update(package))
                {
                    // - Commit version to cache package repository
                    ok = PackageInstance.CacheRepo.Add(package, localPackageRepo.Location);
                }
            }
            
            if (!ok)
                Loggy.Error(String.Format("Error: Package::Install, failed to add {0} to {1}", Filename, CacheRepoDir));

            return ok;
        }
    }
}
