using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Configuration;
using System.Collections.Generic;
using Microsoft.Win32;
using MSBuild.XCode;
using MSBuild.XCode.Helpers;
using FileDirectoryPath;

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
            PackageInstance.TemplateDir = @"d:\REMOTE_PACKAGE_REPO\com\virtuos\xcode\publish\templates\";
            PackageInstance.CacheRepoDir = @"d:\PACKAGE_REPO\";
            PackageInstance.RemoteRepoDir = @"d:\REMOTE_PACKAGE_REPO\";
            PackageInstance.Initialize();
           
            // Our test project is xproject
            PackageInstance.RootDir = @"I:\Packages\xstl\";

            Mercurial.Repository hg_repo = new Mercurial.Repository(PackageInstance.RootDir);
            if (!hg_repo.Exists)
            {
                Loggy.Error(String.Format("Error: Package::Create failed since there is no Hg (SVN) repository!"));
                return;
            }

            Construct("xstl");

            PackageConfigs configs = new PackageConfigs();
            configs.RootDir = PackageInstance.RootDir;
            configs.Platform = "Win32";
            configs.ProjectGroup = "UnitTest";
            configs.TemplateDir = PackageInstance.TemplateDir;
            configs.Execute();

            string createdPackageFilename;
            if (true)
            {
                PackageCreate create = new PackageCreate();
                create.RootDir = PackageInstance.RootDir;
                create.Platform = "Win32";
                bool result1 = create.Execute();
                createdPackageFilename = create.Filename;
            }

            PackageInstall install = new PackageInstall();
            install.RootDir = PackageInstance.RootDir;
            install.CacheRepoDir = PackageInstance.CacheRepoDir;
            install.RemoteRepoDir = PackageInstance.RemoteRepoDir;
            install.Filename = createdPackageFilename;
            bool result3 = install.Execute();

            PackageDeploy deploy = new PackageDeploy();
            deploy.RootDir = PackageInstance.RootDir;
            deploy.CacheRepoDir = PackageInstance.CacheRepoDir;
            deploy.RemoteRepoDir = PackageInstance.RemoteRepoDir;
            deploy.Filename = createdPackageFilename;
            bool result4 = deploy.Execute();

            PackageSync sync = new PackageSync();
            sync.RootDir = PackageInstance.RootDir;
            sync.Platform = "Win32";
            sync.CacheRepoDir = PackageInstance.CacheRepoDir;
            sync.RemoteRepoDir = PackageInstance.RemoteRepoDir;
            sync.Execute();

            PackageInfo info = new PackageInfo();
            info.RootDir = PackageInstance.RootDir;
            info.CacheRepoDir = PackageInstance.CacheRepoDir;
            info.RemoteRepoDir = PackageInstance.RemoteRepoDir;
            info.Execute();

            PackageInstance.RootDir = @"I:\Packages\xstring\";

            PackageVerify verify = new PackageVerify();
            verify.RootDir = PackageInstance.RootDir;
            verify.Name = "xbase";
            verify.Platform = "Win32";
            bool result2 = verify.Execute();
        }

        public static void Construct(string name)
        {
            PackageConstruct construct = new PackageConstruct();
            construct.Name = name;
            construct.RootDir = @"i:\Packages\";
            construct.CacheRepoDir = PackageInstance.CacheRepoDir;
            construct.RemoteRepoDir = PackageInstance.RemoteRepoDir;
            construct.TemplateDir = PackageInstance.TemplateDir;
            //construct.Action = "init";
            //construct.Execute();
            construct.RootDir = construct.RootDir + construct.Name + "\\";
            //construct.Action = "dir";
            //construct.Execute();
            construct.Action = "vs2010";
            construct.Execute();
        }
    }
}
