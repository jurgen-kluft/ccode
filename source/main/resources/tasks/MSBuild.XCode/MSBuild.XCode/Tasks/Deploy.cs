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
        public string IDE { get; set; }         ///< Eclipse, Visual Studio, XCODE
        public string ToolSet { get; set; }     ///< GCC, Visual Studio (v90, v100, v110, v120)

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
                PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir);
            }

            PackageVars vars = new PackageVars();
            vars.Add("IDE", IDE);
            vars.Add(Platform + "ToolSet", ToolSet);
            vars.SetToolSet(Platform, ToolSet, true);
            PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);
            package.SetPlatform(Platform);

            if (package.IsValid)
            {
                PackageState p = package.Package;

                // - Submit version to remote package repository
                if (PackageInstance.RepoActor.Deploy(p))
                {
                    Loggy.Info(String.Format("Package '{0}' for platform '{1}' on branch '{2}' has been deployed with version {3}", p.Name, Platform, p.Branch, p.RemoteVersion.ToString()));
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