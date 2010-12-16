using System;
using System.Configuration;
using System.Collections.Generic;
using Microsoft.Win32;
using MSBuild.XCode;

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
            XGlobal.TemplateDir = @"D:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\";
            XGlobal.LocalRepoDir = @"d:\SCM_PACKAGE_REPO\";
            XGlobal.RemoteRepoDir = @"\\cnshasap2\Hg_Repo\SCM_PACKAGE_REPO\";

            XGlobal.Initialize();

            {
                XPackage package = new XPackage();
                package.Name = "xbase";
                package.Group = new XGroup("com.virtuos.tnt");
                package.Path = string.Empty;
                package.Local = true;
                package.Version = null;
                package.Branch = "default";
                package.Platform = "Win32";

                XGlobal.LocalRepo.Checkout(package, new XVersionRange("[,2.0]"));
                if (package.LoadPom())
                {
                    package.Pom.BuildDependencies("Win32", XGlobal.LocalRepo, XGlobal.RemoteRepo);
                    package.Pom.PrintDependencies("Win32");
                    package.Pom.CheckoutDependencies(@"i:\temp\target\", "Win32", XGlobal.LocalRepo);

                    string[] categories = package.Pom.GetCategories();
                    string[] platforms = package.Pom.GetPlatformsForCategory("Main");
                    string[] configs = package.Pom.GetConfigsForPlatformsForCategory("Win32", "Main");
                    foreach (string category in categories)
                    {
                        foreach (string platform in platforms)
                            foreach (string config in configs)
                                package.Pom.CollectProjectInformation(category, platform, config);
                    }
                }
            }

            PackageConstruct construct = new PackageConstruct();
            construct.Name = "xproject";
            construct.RootDir = @"i:\HgDev.Modules\";
            construct.Language = "cpp";
            construct.TemplateDir = XGlobal.LocalRepoDir + @"com\virtuos\xcode\templates\";
            construct.Execute();    //1st
            construct.RootDir = construct.RootDir + construct.Name + "\\";
            construct.Execute();    //2nd

            PackageCreate create = new PackageCreate();
            create.RootDir = @"i:\HgDev.Modules\xbase\";
            create.Platform = "Win32";
            create.Branch = "default";
            bool result1 = create.Execute();

            PackageVerify verify = new PackageVerify();
            verify.RootDir = @"i:\HgDev.Modules\xbase\";
            verify.Platform = "Win32";
            verify.Branch = "default";
            bool result2 = verify.Execute();

            PackageSync sync = new PackageSync();
            sync.RootDir = @"i:\HgDev.Modules\xbase\";
            sync.Platform = "Win32";
            sync.LocalRepoDir = XGlobal.LocalRepoDir;
            sync.RemoteRepoDir = XGlobal.RemoteRepoDir;
            sync.Execute();

            PackageInstall install = new PackageInstall();
            install.RootDir = @"i:\HgDev.Modules\xbase\";
            install.LocalRepoDir = XGlobal.LocalRepoDir;
            install.RemoteRepoDir = XGlobal.RemoteRepoDir;
            bool result3 = install.Execute();

            PackageDeploy deploy = new PackageDeploy();
            deploy.RootDir = @"i:\HgDev.Modules\xbase\";
            deploy.LocalRepoDir = XGlobal.LocalRepoDir;
            deploy.RemoteRepoDir = XGlobal.RemoteRepoDir;
            bool result4 = deploy.Execute();

        }
    }
}
