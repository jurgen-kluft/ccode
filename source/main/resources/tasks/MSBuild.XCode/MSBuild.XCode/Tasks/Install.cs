using System;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will 
    ///	   Create
    ///	   Install
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

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;

            RootDir = RootDir.EndWith('\\');
            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            if (!PackageInstance.IsInitialized)
            {
                PackageInstance.TemplateDir = string.Empty;
                if (!PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir))
                {
                    return false;
                }
            }

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            package.SetPlatform(Platform);

            if (package.IsValid)
            {
                PackageState p = package.Package;

                // - Commit version to cache package repository
                if (PackageInstance.RepoActor.Install(p))
                {
                    Loggy.Info(String.Format("Package '{0}' for platform '{1}' on branch '{2}' has been installed with version {3}", p.Name, Platform, p.Branch, p.CacheVersion.ToString()));
                    return true;
                }
                else
                {
                    Loggy.Error(String.Format("Package::Install, failed to add package {0} to cache package repository at {1}", p.LocalURL, CacheRepoDir));
                }
            }
            else
            {
                Loggy.Error(String.Format("Package::Install, failed to load pom.xml"));
            }

            return false;
        }
    }
}
