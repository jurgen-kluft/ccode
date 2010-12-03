using System;
using System.Configuration;
using System.Collections.Generic;
using Microsoft.Win32;
using msbuild.xmaven;

namespace xmaven
{
    static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        [STAThread]
        static void Main()
        {
            PackageSync sync = new PackageSync();
            sync.Name = "xbase";
            sync.Path = @"i:\HgDev.Modules\xbase\";
            sync.Dep = "dep.props";
            sync.Execute();

            PackageCreate create = new PackageCreate();
            create.Path = @"i:\HgDev.Modules\xbase\target\";
            create.Name = "xbase";
            create.ZipFilename = @"i:\HgDev.Modules\xbase\target\xbase_1.0.2010.11.default.959a52b10784_Win32.zip";
            bool result1 = create.Execute();

            PackageVerify verify = new PackageVerify();
            verify.Name = "xbase";
            verify.Path = @"i:\HgDev.Modules\xbase\target\";
            bool result2 = verify.Execute();

            PackageInstall install = new PackageInstall();
            install.RepoPath = @"D:\SCM_PACKAGE_REPO\com\virtuos\tnt\xbase\";
            install.LatestPath = @"latest\";
            install.OldLatest = "xbase_*default*_Win32*.latest";
            install.SourceFilename = "xbase_1.0.2010.11.default.959a52b10784_Win32.zip";
            install.SourcePath = @"i:\HgDev.Modules\xbase\target\";
            install.VersionPath = @"2010\11\";
            bool result3 = install.Execute();

            PackageDeploy deploy = new PackageDeploy();
            deploy.RepoPath = @"\\cnshasap2\Hg_Repo\SCM_PACKAGE_REPO\com\virtuos\tnt\xbase\";
            deploy.LatestPath = @"latest\";
            deploy.OldLatest = "xbase_*default*_Win32*.latest";
            deploy.SourceFilename = "xbase_1.0.2010.11.default.959a52b10784_Win32.zip";
            deploy.SourcePath = @"i:\HgDev.Modules\xbase\target\";
            deploy.VersionPath = @"2010\11\";
            bool result4 = deploy.Execute();

        }
    }
}
