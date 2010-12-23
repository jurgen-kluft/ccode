using System;
using System.IO;
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
            Global.TemplateDir = @"\\cnshasap2\Hg_Repo\PACKAGE_REPO\com\virtuos\xcode\data\templates\";
            Global.CacheRepoDir = @"d:\PACKAGE_REPO\";
            Global.RemoteRepoDir = @"\\cnshasap2\Hg_Repo\PACKAGE_REPO\";
            Global.Initialize();
           
            // Our test project is xproject
            Global.RootDir = @"I:\HgDev.Modules\xproject\";

            PackageConstruct construct = new PackageConstruct();
            construct.Name = "xproject";
            construct.RootDir = @"i:\HgDev.Modules\";
            construct.Language = "cpp";
            construct.CacheRepoDir = Global.CacheRepoDir;
            construct.RemoteRepoDir = Global.RemoteRepoDir;
            construct.TemplateDir = Global.TemplateDir;
            construct.Action = "init";
            construct.Execute();
            construct.RootDir = construct.RootDir + construct.Name + "\\";
            construct.Action = "dir";
            construct.Execute();
            construct.Action = "vs2010";
            construct.Execute();

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
