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
            XProject template = new XProject();
            template.Load(@"D:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\vcxproj.xml.template");

            XPackage xpack = new XPackage();
            xpack.Templates.Add(template);
            xpack.Load(@"D:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\package.xml.template");

            PackageSync sync = new PackageSync();
            sync.Path = @"i:\HgDev.Modules\xbase\";
            sync.LocalRepoPath = @"D:\SCM_PACKAGE_REPO\com\virtuos\tnt\xbase\";
            sync.RemoteRepoPath = @"\\cnshasap2\Hg_Repo\SCM_PACKAGE_REPO\com\virtuos\tnt\xbase\";
            sync.Execute();

            PackageCreate create = new PackageCreate();
            create.Path = @"i:\HgDev.Modules\xbase\";
            create.Platform = "Win32";
            bool result1 = create.Execute();

            PackageVerify verify = new PackageVerify();
            verify.Path = @"i:\HgDev.Modules\xbase\";
            bool result2 = verify.Execute();

            PackageInstall install = new PackageInstall();
            install.Path = @"i:\HgDev.Modules\xbase\";
            install.RepoPath = @"D:\SCM_PACKAGE_REPO\com\virtuos\tnt\xbase\";
            bool result3 = install.Execute();

            PackageDeploy deploy = new PackageDeploy();
            deploy.Path = @"i:\HgDev.Modules\xbase\";
            deploy.RepoPath = @"\\cnshasap2\Hg_Repo\SCM_PACKAGE_REPO\com\virtuos\tnt\xbase\";
            bool result4 = deploy.Execute();

        }
    }
}
