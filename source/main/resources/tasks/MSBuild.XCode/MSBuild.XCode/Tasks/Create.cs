using System;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
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
        public string IDE { get; set; }         ///< Eclipse, Visual Studio, XCODE
        public string ToolSet { get; set; }     ///< GCC, Visual Studio (v90, v100, v110, v120)
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

            IDE = !String.IsNullOrEmpty(IDE) ? IDE.ToLower() : "vs2012";
            ToolSet = !String.IsNullOrEmpty(ToolSet) ? ToolSet.ToLower() : "v110";

            if (!PackageInstance.IsInitialized)
            {
                PackageInstance.TemplateDir = string.Empty;
                if (!PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir))
                {
                    return false;
                }
            }

            PackageVars vars = new PackageVars();
            vars.Add("Platform", Platform);
            vars.Add("IDE", IDE);
            vars.Add("ToolSet", ToolSet);

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);
            package.SetPlatform(Platform);

            if (package.IsValid)
            {
                PackageState p = package.Package;
                
                // - Increment the build ?
                p.IncrementVersion = IncrementBuild;
                Loggy.Info(String.Format("Increment build number : {0}", IncrementBuild ? "ON" : "OFF"));

                ComparableVersion rootVersion = package.Pom.Versions.GetForPlatform(Platform);

                // - Create
                bool created = PackageInstance.RepoActor.Create(p, package.Pom.Content, vars, rootVersion);
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
