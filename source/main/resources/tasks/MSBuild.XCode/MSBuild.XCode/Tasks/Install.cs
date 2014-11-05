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
        public string IDE { get; set; }         ///< Eclipse, Visual Studio, XCODE
        public string ToolSet { get; set; }     ///< GCC, Visual Studio (v90, v100, v110, v120)
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

            IDE = !String.IsNullOrEmpty(IDE) ? IDE.ToLower() : "vs2012";
            ToolSet = !String.IsNullOrEmpty(ToolSet) ? ToolSet.ToLower() : "v110";

            if (!PackageInstance.IsInitialized)
            {
                PackageInstance.TemplateDir = string.Empty;
                if (!PackageInstance.Initialize(IDE, RemoteRepoDir, CacheRepoDir, RootDir))
                {
                    return false;
                }
            }

            PackageVars vars = new PackageVars();
            vars.Add("IDE", IDE);
            vars.Add("Platform", Platform);
            vars.SetToolSet(Platform, ToolSet, true);

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);

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
