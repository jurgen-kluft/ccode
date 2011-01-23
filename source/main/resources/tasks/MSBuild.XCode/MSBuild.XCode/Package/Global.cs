using System;
using System.IO;
using System.Xml;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public partial class PackageInstance
    {
        public static bool IsInitialized { get; set; }

        public static string CacheRepoDir { get; set; }
        public static string RemoteRepoDir { get; set; }
        public static string TemplateDir { get; set; }
        public static string RootDir { get; set; }

        public static IPackageRepository RemoteRepo { get; set; }
        public static IPackageRepository CacheRepo { get; set; }
        public static IPackageRepository ShareRepo { get; set; }

        public static MsDev.CppProject CppTemplateProject { get; set; }
        public static MsDev.CsProject CsTemplateProject { get; set; }

        static PackageInstance()
        {
            IsInitialized = false;
        }

        public static bool Initialize()
        {
            if (IsInitialized)
                return true;

            Loggy.ToConsole = true;
            Loggy.TaskLogger = null;
            Loggy.Indentor = "\t";

            if (!String.IsNullOrEmpty(CacheRepoDir))
            {
                if (!Directory.Exists(CacheRepoDir))
                {
                    Loggy.Error(String.Format("Error: Initialization of Global failed since cache repo {0} doesn't exist", CacheRepoDir));
                    return false;
                }
            }
            if (!String.IsNullOrEmpty(RemoteRepoDir))
            {
                if (!Directory.Exists(RemoteRepoDir))
                {
                    Loggy.Error(String.Format("Error: Initialization of Global failed since remote repo {0} doesn't exist", RemoteRepoDir));
                    return false;
                }
            }
            if (!String.IsNullOrEmpty(TemplateDir))
            {
                if (!Directory.Exists(TemplateDir))
                {
                    Loggy.Error(String.Format("Error: Initialization of Global failed since template dir {0} doesn't exist", TemplateDir));
                    return false;
                }
            }

            RemoteRepo = new PackageRepositoryFileSystem(RemoteRepoDir, ELocation.Remote);
            CacheRepo = new PackageRepositoryFileSystem(CacheRepoDir, ELocation.Cache);
            ShareRepo = new PackageRepositoryShare(CacheRepoDir + ".share\\");

            if (!String.IsNullOrEmpty(TemplateDir))
            {
                // For C++
                CppTemplateProject = new MsDev.CppProject();
                if (!CppTemplateProject.Load(TemplateDir + "main" + CppTemplateProject.Extension))
                {
                    Loggy.Error(String.Format("Error: Initialization of Global failed in due to failure in loading {0}", TemplateDir + "main" + CppTemplateProject.Extension));
                    return false;
                }

                // For C#
                CsTemplateProject = new MsDev.CsProject();
                if (!CsTemplateProject.Load(TemplateDir + "main" + CsTemplateProject.Extension))
                {
                    Loggy.Error(String.Format("Error: Initialization of Global failed in due to failure in loading {0}", TemplateDir + "main" + CsTemplateProject.Extension));
                    return false;
                }
            }

            IsInitialized = true;
            return true;
        }

    }
}
