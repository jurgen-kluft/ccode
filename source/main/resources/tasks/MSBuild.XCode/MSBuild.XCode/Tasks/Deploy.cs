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
    ///	Will 
    ///	   Create
    ///	   Install
    ///	   Deploy 
    ///	
    /// </summary>
    public class PackageDeploy : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }
        [Required]
        public string Platform { get; set; }

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;

            RootDir = RootDir.EndWith('\\');
            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            if (!PackageInstance.IsInitialized)
            {
                PackageInstance.TemplateDir = string.Empty;
                PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir);
            }

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            package.SetPlatform(Platform);

            if (package.IsValid)
            {
                Package p = package.Package;

                // - Submit version to remote package repository
                if (PackageInstance.RepoActor.Deploy(p))
                {
                    Loggy.Info(String.Format("Info: Package {0} for Platform {1} on Branch {2} has been deployed with version {3}", p.Name, Platform, p.Branch, p.RemoteVersion.ToString()));
                    return true;
                }
                else
                {
                    Loggy.Error(String.Format("Error: Package::Deploy, failed to deploy package"));
                }
            }
            else
            {
                Loggy.Error(String.Format("Error: Package::Deploy, failed to load pom.xml"));
            }

            return false;
        }
    }
}