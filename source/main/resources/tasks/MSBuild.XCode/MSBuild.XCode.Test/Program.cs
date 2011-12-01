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

        [STAThread]
        static void Main()
        {
            //MainCpp();
            MainCs();
        }

        static void MainCs()
        {
            Loggy.ToConsole = true;

            string name = "xprojectB";

            // home (hp laptop)
            //RemoteRepoDir = @"db::server=127.0.0.1;port=3306;database=xcode;uid=root;password=p1|fs::D:\PACKAGE_REPO_TEST\";
            
            // work
            RemoteRepoDir = @"db::server=cnshasap2;port=3307;database=xcode_cpp;uid=developer;password=Fastercode189|storage::\\cnshasap2\Hg_Repo\PACKAGE_REPO\.storage\";
            
            CacheRepoDir = @"k:\Dev.Cs.Packages\PACKAGE_REPO\";
            RootDir = @"k:\Dev.Cs.Packages\" + name + "\\";
            XCodeRepoDir = @"k:\Dev.Cs.Packages\PACKAGE_REPO\com\virtuos\xcode\publish\";
            TemplateDir = XCodeRepoDir + @"templates\";

            PackageInstance.TemplateDir = TemplateDir;
            if (!PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir))
                return;

            if (Update("1.1.0.4"))
            {
                string platform = "x86";
                Construct(name, platform, "Cs");
                Create(name, platform, "Cs");
                Install(name, platform, "Cs");
                Deploy(name, platform, "Cs");
            }
        }

        
        static void MainCpp()
        {
            Loggy.ToConsole = true;

            string name = "xbase";

            // home (hp laptop)
            //RemoteRepoDir = @"db::server=127.0.0.1;port=3306;database=xcode;uid=root;password=p1|fs::D:\PACKAGE_REPO_TEST\";
            
            // work
            RemoteRepoDir = @"db::server=cnshasap2;port=3307;database=xcode_cpp;uid=developer;password=Fastercode189|storage::\\cnshasap2\Hg_Repo\PACKAGE_REPO\.storage\";
            
            CacheRepoDir = @"k:\Dev.C++.Packages\PACKAGE_REPO\";
            RootDir = @"k:\Dev.C++.Packages\" + name + "\\";
            XCodeRepoDir = @"k:\Dev.C++.Packages\PACKAGE_REPO\com\virtuos\xcode\publish\";
            TemplateDir = XCodeRepoDir + @"templates\";

            PackageInstance.TemplateDir = TemplateDir;
            if (!PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir))
                return;

            if (Update("1.1.0.3"))
            {
                string platform = "*";
                Construct(name, platform, "C++");
                Create(name, platform, "C++");
                Install(name, platform, "C++");
                Deploy(name, platform, "C++");
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

        public static void Create(string name, string platform, string language)
        {
            if (platform == "*")
                platform = "Win32";

            PackageCreate cmd = new PackageCreate();
            cmd.RemoteRepoDir = RemoteRepoDir;
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.RootDir = "k:\\Dev." + language + ".Packages\\" + name + "\\";
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
            cmd.RootDir = "k:\\Dev." + language + ".Packages\\" + name + "\\";
            cmd.Platform = platform;
            cmd.Execute();
        }

        public static void Deploy(string name, string platform, string language)
        {
            if (platform == "*")
                platform = "Win32";

            PackageDeploy cmd = new PackageDeploy();
            cmd.RemoteRepoDir = RemoteRepoDir;
            cmd.CacheRepoDir = CacheRepoDir;
            cmd.RootDir = "k:\\Dev." + language + ".Packages\\" + name + "\\";
            cmd.Platform = platform;
            cmd.Execute();
        }

        public static void Construct(string name, string platform, string language)
        {
            PackageConstruct construct = new PackageConstruct();
            construct.Name = name;
            construct.RootDir = "k:\\Dev." + language + ".Packages\\";
            construct.CacheRepoDir = CacheRepoDir;
            construct.RemoteRepoDir = RemoteRepoDir;
            construct.TemplateDir = TemplateDir;
            construct.Language = "C++";
            if (language == "Cs")
                construct.Language = "C#";
            construct.RootDir = construct.RootDir + construct.Name + "\\";
            construct.Platform = platform;
            construct.Action = "vs2010";
            construct.Execute();
        }
    }
}