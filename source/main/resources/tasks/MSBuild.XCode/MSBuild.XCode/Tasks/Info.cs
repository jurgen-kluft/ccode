using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageInfo : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            bool success = false;
            Loggy.TaskLogger = Log;

            RootDir = RootDir.EndWith('\\');
            CacheRepoDir = CacheRepoDir.EndWith('\\');
            RemoteRepoDir = RemoteRepoDir.EndWith('\\');

            Environment.CurrentDirectory = RootDir;

            PackageInstance.TemplateDir = string.Empty;
            PackageInstance.Initialize("VS2012", RemoteRepoDir, CacheRepoDir, RootDir);

            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            Mercurial.Repository hg_repo = new Mercurial.Repository(RootDir);
            string branch = hg_repo.Branch();

            PackageVars vars = new PackageVars();
            PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);
            if (package.IsValid)
            {
                PackageDependencies dependencies = new PackageDependencies(package);
                if (dependencies.BuildForAllPlatforms())
                {
                    success = package.Info();
                }
                else
                {
                    success = false;
                }
            }
            return success;
        }
    }
}
