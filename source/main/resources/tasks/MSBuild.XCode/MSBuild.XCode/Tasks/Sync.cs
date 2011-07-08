using System;
using System.Xml;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will sync the local-package-repository with the remote-package-repository. 
    ///	Will sync dependencies specified into the target folder
    /// </summary>
    public class PackageSync : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string Platform { get; set; }
        [Required]
        public string TemplateDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;
            RootDir = RootDir.EndWith('\\');

            PackageInstance.TemplateDir = TemplateDir;
            PackageInstance.CacheRepoDir = CacheRepoDir;
            PackageInstance.RemoteRepoDir = RemoteRepoDir;
            PackageInstance.Initialize();

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            if (package.IsValid)
            {
                package.SetPlatform(Platform);
                PackageDependencies dependencies = new PackageDependencies(package);

                if (!dependencies.BuildForPlatform(Platform))
                {
                    Loggy.Error(String.Format("Error: Failed to build dependencies in Package::Sync"));
                    return false;
                }
                dependencies.SaveInfo(Platform, new FileDirectoryPath.FilePathAbsolute(RootDir + "\\target\\" + package.Name + "\\build\\"  + Platform + "\\dependencies.info"));
            }
            else
            {
                Loggy.Error(String.Format("Error: Failed to load 'pom.xml' in Package::Sync"));
                return false;
            }

            return true;
        }
    }
}
