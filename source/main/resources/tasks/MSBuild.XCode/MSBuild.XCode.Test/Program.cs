using System;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode.Test
{
    static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        private static string RemoteRepoDir;
        private static string CacheRepoDir;
        private static string RootDir;
        private static string XCodeRepoDir;
		private static string TemplateDir;
		private static string MsDev;
		private static string MsDevToolSet;

        private static string Name;
        private static string BaseDirCpp;
        private static string BaseDirCs;

        [STAThread]
        static void Main()
        {
            MainCpp();
            //MainCs();
        }

        static void MainCs()
        {
            Loggy.ToConsole = true;

            Name = "xprojectB";
            BaseDirCpp = "";
            BaseDirCs = @"J:\Dev.Cs.Packages\";
			MsDev = "VS2013";
			MsDevToolSet = "vs120";

            // home (hp laptop)
            //RemoteRepoDir = @"db::server=127.0.0.1;port=3306;database=xcode;uid=root;password=p1|fs::D:\PACKAGE_REPO_TEST\";
            
            // work
            RemoteRepoDir = @"db::server=cnshasap2;port=3307;database=xcode_cpp;uid=developer;password=Fastercode189|storage::\\cnshasap2\Hg_Repo\PACKAGE_REPO\.storage\";
            
            CacheRepoDir = @"PACKAGE_REPO\";
            RootDir = BaseDirCs + Name + "\\";
            XCodeRepoDir = BaseDirCs + @"PACKAGE_REPO\com\virtuos\xcode\publish\";
            TemplateDir = XCodeRepoDir + @"templates\";

            PackageInstance.TemplateDir = TemplateDir;
			if (!PackageInstance.Initialize(MsDev, RemoteRepoDir, CacheRepoDir, RootDir))
                return;

            if (Update("1.1.0.4"))
            {
                string platform = "x86";
                Construct(Name, platform, "Cs");
                Create(Name, platform, "Cs");
                Install(Name, platform, "Cs");
                Deploy(Name, platform, "Cs");
            }
        }

        
        static void MainCpp()
        {
            Loggy.ToConsole = true;

            string name = "xbase";
            BaseDirCpp = @"J:\Dev.C++.Packages.Bitbucket\";
            BaseDirCs = "";
			MsDev = "VS2013";
			MsDevToolSet = "vs120";

            // home (hp laptop)
            //RemoteRepoDir = @"db::server=127.0.0.1;port=3306;database=xcode;uid=root;password=p1|fs::D:\PACKAGE_REPO_TEST\";
            
            // work
            // RemoteRepoDir = @"db::server=cnshasap2;port=3307;database=xcode_cpp;uid=developer;password=Fastercode189|storage::\\cnshasap2\Hg_Repo\PACKAGE_REPO\.storage\";
            RemoteRepoDir = @"fs::" + BaseDirCpp + @"REMOTE_PACKAGE_REPO\";
            
            CacheRepoDir = BaseDirCpp + @"PACKAGE_REPO\";
            RootDir = BaseDirCpp + name + "\\";
            XCodeRepoDir = BaseDirCpp + @"PACKAGE_REPO\com\virtuos\xcode\publish\";
            TemplateDir = XCodeRepoDir + @"templates\";

            PackageInstance.TemplateDir = TemplateDir;
			if (!PackageInstance.Initialize(MsDev, RemoteRepoDir, CacheRepoDir, RootDir))
                return;

            if (Update("1.1.0.4"))
            {
                string platform = "x64";
                Construct(name, platform, "C++");
                //Create(name, platform, "C++");
                //Sync(name, platform, "C++");
                //Install(name, platform, "C++");
                //Deploy(name, platform, "C++");
            }
        }

        public static bool Update(string version)
        {
            XCodeUpdate cmd = new XCodeUpdate();
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.XCodeRepoDir = XCodeRepoDir;
            cmd.XCodeVersion = version;
            return cmd.Execute();
        }

        public static void Sync(string name, string platform, string language)
        {
            if (platform == "*")
                platform = "Win32";

            PackageSync cmd = new PackageSync();
            cmd.RemoteRepoDir = RemoteRepoDir;
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.RootDir = (language == "Cs" ? BaseDirCs : BaseDirCpp) + name + "\\";
            cmd.Platform = platform;
            cmd.Execute();
        }

        public static void Create(string name, string platform, string language)
        {
            if (platform == "*")
                platform = "Win32";

            PackageCreate cmd = new PackageCreate();
            cmd.RemoteRepoDir = RemoteRepoDir;
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.RootDir = (language == "Cs" ? BaseDirCs : BaseDirCpp) + name + "\\";
            cmd.Platform = platform;
            cmd.IncrementBuild = true;
            cmd.Execute();
        }

        public static void Install(string name, string platform, string language)
        {
            if (platform == "*")
                platform = "Win32";

            PackageInstall cmd = new PackageInstall();
            cmd.RemoteRepoDir = RemoteRepoDir;
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.RootDir = (language == "Cs" ? BaseDirCs : BaseDirCpp) + name + "\\";
            cmd.Platform = platform;
			cmd.IDE = MsDev;
			cmd.ToolSet = MsDevToolSet;
            cmd.Execute();
        }

        public static void Deploy(string name, string platform, string language)
        {
            if (platform == "*")
                platform = "Win32";

            PackageDeploy cmd = new PackageDeploy();
            cmd.RemoteRepoDir = RemoteRepoDir;
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.RootDir = (language == "Cs" ? BaseDirCs : BaseDirCpp) + name + "\\";
            cmd.Platform = platform;
            cmd.Execute();
        }

        public static void Construct(string name, string platform, string language)
        {
            PackageConstruct construct = new PackageConstruct();
            construct.Name = name;
            construct.RootDir = language == "Cs" ? BaseDirCs : BaseDirCpp;
            construct.CacheRepoDir = CacheRepoDir;
            construct.RemoteRepoDir = RemoteRepoDir;
            construct.TemplateDir = TemplateDir;
            construct.Language = "C++";
            if (language == "Cs")
                construct.Language = "C#";
            construct.RootDir = construct.RootDir + construct.Name + "\\";
            construct.Platform = platform;
			construct.ToolSet = MsDevToolSet;
			construct.IDE = MsDev;
            construct.Action = "genprj";
            construct.Execute();
        }
    }
}