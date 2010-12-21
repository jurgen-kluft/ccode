using System;
using System.Configuration;
using System.Collections.Generic;
using Microsoft.Win32;
using MSBuild.XCode;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode.Test
{
    static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        [STAThread]
        static void Main()
        {
            Global.TemplateDir = @"D:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\";
            Global.CacheRepoDir = @"d:\SCM_PACKAGE_REPO\";
            Global.RemoteRepoDir = @"\\cnshasap2\Hg_Repo\SCM_PACKAGE_REPO\";
            Global.Initialize();

            Global.RootDir = @"I:\HgDev.Modules\xbase\";
            
            if (false)
            {
                Package package = new Package();
                package.IsRoot = true;
                package.Name = "xbase";
                package.Group = new Group("com.virtuos.tnt");
                package.RootDir = Global.RootDir;
                package.Version = null;
                package.Branch = "default";
                package.Platform = "Win32";

                Global.RemoteRepo.Update(package, new VersionRange("[,2.0]"));
                Global.CacheRepo.Update(package, new VersionRange("[,2.0]"));
                if (package.LoadFinalPom())
                {
                    string package_filename;
                    package.Create(out package_filename);

                    package.BuildDependencies("Win32", Global.CacheRepo, Global.RemoteRepo);
                    package.PrintDependencies("Win32");
                    package.SyncDependencies("Win32", Global.CacheRepo);

                    string[] categories = package.Pom.GetCategories();
                    foreach (string category in categories)
                    {
                        string[] platforms = package.Pom.GetPlatformsForCategory(category);
                        foreach (string platform in platforms)
                        {
                            string[] configs = package.Pom.GetConfigsForPlatformsForCategory(platform, category);
                            //foreach (string config in configs)
                              //  package.Pom.CollectProjectInformation(category, platform, config);
                        }
                    }
                }
            }

            // Our test project is xproject
            Global.RootDir = @"I:\HgDev.Modules\xproject\";

            PackageConstruct construct = new PackageConstruct();
            construct.Name = "xproject";
            construct.RootDir = @"i:\HgDev.Modules\";
            construct.Language = "cpp";
            construct.CacheRepoDir = Global.CacheRepoDir;
            construct.RemoteRepoDir = Global.RemoteRepoDir;
            construct.TemplateDir = Global.TemplateDir;
            construct.Execute();    //1st
            construct.RootDir = construct.RootDir + construct.Name + "\\";
            construct.Execute();    //2nd

            PackageInfo info = new PackageInfo();
            info.RootDir = Global.RootDir;
            info.Execute();

            PackageCreate create = new PackageCreate();
            create.RootDir = Global.RootDir;
            create.Platform = "Win32";
            bool result1 = create.Execute();

            PackageInstall install = new PackageInstall();
            install.RootDir = Global.RootDir;
            install.CacheRepoDir = Global.CacheRepoDir;
            install.RemoteRepoDir = Global.RemoteRepoDir;
            install.Filename = create.Filename;
            bool result3 = install.Execute();

            PackageDeploy deploy = new PackageDeploy();
            deploy.RootDir = Global.RootDir;
            deploy.CacheRepoDir = Global.CacheRepoDir;
            deploy.RemoteRepoDir = Global.RemoteRepoDir;
            deploy.Filename = create.Filename;
            bool result4 = deploy.Execute();

            PackageVerify verify = new PackageVerify();
            verify.RootDir = Global.RootDir;
            verify.Platform = "Win32";
            verify.Branch = "default";
            bool result2 = verify.Execute();

            PackageSync sync = new PackageSync();
            sync.RootDir = Global.RootDir;
            sync.Platform = "Win32";
            sync.CacheRepoDir = Global.CacheRepoDir;
            sync.RemoteRepoDir = Global.RemoteRepoDir;
            sync.Execute();

        }
    }
}
