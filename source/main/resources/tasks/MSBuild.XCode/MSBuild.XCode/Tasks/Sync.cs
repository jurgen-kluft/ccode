using System;
using System.Collections.Generic;
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
        public string IDE { get; set; }         ///< Eclipse, Visual Studio, XCODE
        public string ToolSet { get; set; }     ///< GCC, Visual Studio (v90, v100, v110, v120)

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
            if (!PackageInstance.Initialize(IDE, RemoteRepoDir, CacheRepoDir, RootDir))
            {
                Loggy.Error(String.Format("Error: Failed to initialize in Package::Sync"));
                return false;
            }

            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            IDE = !String.IsNullOrEmpty(IDE) ? IDE.ToLower() : "vs2012";
            ToolSet = !String.IsNullOrEmpty(ToolSet) ? ToolSet.ToLower() : "v110";

            PackageVars vars = new PackageVars();
            vars.Add("Platform", Platform);
            vars.Add("IDE", IDE);
            vars.Add("ToolSet", ToolSet);
            vars.SetToolSet(Platform, ToolSet, true);

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);

            if (package.IsValid)
            {
                PackageDependencies dependencies = new PackageDependencies(package);
                if (!dependencies.BuildForPlatform(Platform))
                {
                    Loggy.Error(String.Format("Error: Failed to build dependencies in Package::Sync"));
                    return false;
                }
                List<string> platforms = new List<string>();
                platforms.Add(Platform);
                dependencies.SaveInfoForPlatforms(platforms, vars);
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
