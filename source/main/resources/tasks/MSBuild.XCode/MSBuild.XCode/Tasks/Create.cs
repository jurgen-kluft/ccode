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
        [Required]
        public string RemoteRepoDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string Platform { get; set; }
        [Required]
        public bool IncrementBuild { get; set; }
        [Output]
        public string Filename { get; set; }

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
                Package p = package.Package;
                
                // - Increment the build ?
                p.IncrementVersion = IncrementBuild;
                Loggy.Info(String.Format("Info: Package::Create, increment build number : {0}", IncrementBuild ? "ON" : "OFF"));

                ComparableVersion rootVersion = package.Pom.Versions.GetForPlatform(Platform);

                // - Create
                bool created = PackageInstance.RepoActor.Create(p, package.Pom.Content, rootVersion);
                if (created)
                {
                    return true;
                }
                else
                {
                    Loggy.Error(String.Format("Error: Package::Create, failed to create package {0}", p.LocalURL));
                }
            }
            else
            {
                Loggy.Error(String.Format("Error: Package::Create, failed to load pom.xml"));
            }

            return false;
        }
    }
}
