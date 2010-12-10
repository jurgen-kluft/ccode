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
            {
                XPackageRepository repo = new XPackageRepository(@"d:\SCM_PACKAGE_REPO\");
                repo.CheckoutVersion("com.virtuos.tnt", @"i:\temp\", "xbase", "default", "Win32", new XVersionRange("[,2.0]"));
            }

            {
                XVersionRange xrange1 = new XVersionRange("(,1.0],[1.2,)");
                bool in_range = xrange1.IsInRange(new XVersion("1.1.2"));
                in_range = xrange1.IsInRange(new XVersion("0.9"));

                XVersion xversion1 = new XVersion("1.2.23.0");
                string[] version_components1 = xversion1.ToStrings();
                XVersion xversion2 = new XVersion("1.0.0.0");
            }

            {
                XProject template = new XProject();
                template.Load(@"D:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\vcxproj.xml.template");

                XPackage xpack = new XPackage();
                xpack.Templates.Add(template);
                xpack.Load(@"D:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\package.xml.template");
            }

            PackageConstruct construct = new PackageConstruct();
            construct.Name = "xproject";
            construct.Path = @"i:\HgDev.Modules\";
            construct.Language = "cpp";
            construct.TemplatePath = @"d:\SCM_PACKAGE_REPO\com\virtuos\xcode\templates\";
            construct.Execute();
            construct.Path = construct.Path + construct.Name + "\\";
            construct.Execute();
            return;

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
