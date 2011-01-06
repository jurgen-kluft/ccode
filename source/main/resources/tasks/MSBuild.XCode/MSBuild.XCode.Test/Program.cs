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
            /// MsDev2010.Cs.XCode.Project p1 = new MsDev2010.Cs.XCode.Project();
            /// p1.Load(@"d:\Dev\HgDev.Modules\xcode\source\main\resources\templates\main.csproj");
            /// MsDev2010.Cs.XCode.Project p2 = new MsDev2010.Cs.XCode.Project();
            /// p2.Load(@"d:\temp\test_cs_project\package.csproj");
            /// 
            /// p2.Merge(p1);
            /// p2.ExpandGlobs(@"d:\temp\test_cs_project\", @"d:\temp\test_cs_project\");
            /// 
            Global.TemplateDir = @"d:\REMOTE_PACKAGE_REPO\com\virtuos\xcode\publish\templates\";
            Global.CacheRepoDir = @"d:\PACKAGE_REPO\";
            Global.RemoteRepoDir = @"d:\REMOTE_PACKAGE_REPO\";
            Global.Initialize();
           
            // Our test project is xproject
            Global.RootDir = @"I:\Packages\xunittest\";

            Mercurial.Repository hg_repo = new Mercurial.Repository(Global.RootDir);
            if (!hg_repo.Exists)
            {
                Loggy.Add(String.Format("Error: Package::Create failed since there is no Hg (Mercurial) repository!"));
                return;
            }

            PackageConfigs configs = new PackageConfigs();
            configs.RootDir = Global.RootDir;
            configs.Platform = "Win32";
            configs.Category = "UnitTest";
            configs.TemplateDir = Global.TemplateDir;
            configs.Execute();
            
            string createdPackageFilename = @"xunittest+1.0.1.2010.12.29.16.38.53+default+Win32.zip";
            if (false)
            {
                PackageCreate create = new PackageCreate();
                create.RootDir = Global.RootDir;
                create.Platform = "Win32";
                bool result1 = create.Execute();
                createdPackageFilename = create.Filename;
            }

            Construct("xstring");

            PackageInstall install = new PackageInstall();
            install.RootDir = Global.RootDir;
            install.CacheRepoDir = Global.CacheRepoDir;
            install.RemoteRepoDir = Global.RemoteRepoDir;
            install.Filename = createdPackageFilename;
            bool result3 = install.Execute();

            PackageDeploy deploy = new PackageDeploy();
            deploy.RootDir = Global.RootDir;
            deploy.CacheRepoDir = Global.CacheRepoDir;
            deploy.RemoteRepoDir = Global.RemoteRepoDir;
            deploy.Filename = createdPackageFilename;
            bool result4 = deploy.Execute();

            PackageSync sync = new PackageSync();
            sync.RootDir = Global.RootDir;
            sync.Platform = "Win32";
            sync.CacheRepoDir = Global.CacheRepoDir;
            sync.RemoteRepoDir = Global.RemoteRepoDir;
            sync.Execute();

            PackageInfo info = new PackageInfo();
            info.RootDir = Global.RootDir;
            info.Execute();

            Global.RootDir = @"I:\Packages\xstring\";

            PackageVerify verify = new PackageVerify();
            verify.RootDir = Global.RootDir;
            verify.Name = "xbase";
            verify.Platform = "Win32";
            bool result2 = verify.Execute();
        }

        public static void Construct(string name)
        {
            PackageConstruct construct = new PackageConstruct();
            construct.Name = name;
            construct.RootDir = @"i:\Packages\";
            construct.CacheRepoDir = Global.CacheRepoDir;
            construct.RemoteRepoDir = Global.RemoteRepoDir;
            construct.TemplateDir = Global.TemplateDir;
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
