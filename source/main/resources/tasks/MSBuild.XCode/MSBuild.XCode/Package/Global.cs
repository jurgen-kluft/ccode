using System;
using System.IO;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public partial class PackageInstance
    {
        public static bool IsInitialized { get; set; }

        public static string TemplateDir { get; set; }
        public static string RootDir { get; set; }

        public static PackageRepositoryActor RepoActor { get; set; }

        public static MsDev.CppProject CppTemplateProject { get; set; }
        public static MsDev.CsProject CsTemplateProject { get; set; }

        static PackageInstance()
        {
            IsInitialized = false;
        }

        public static bool Initialize(string RemoteRepoURL, string CacheRepoURL, string RootURL)
        {
            if (IsInitialized)
                return true;

            Loggy.Indentor = "\t";

            RepoActor = new PackageRepositoryActor();
            if (!RepoActor.Initialize(RemoteRepoURL, CacheRepoURL, RootURL))
            {
                Loggy.Error(String.Format("Error: Initialization of Repository Actor failed", TemplateDir));
                return false;
            }

            if (!String.IsNullOrEmpty(TemplateDir))
            {
                if (!Directory.Exists(TemplateDir))
                {
                    Loggy.Error(String.Format("Error: Initialization of Global failed since template dir {0} doesn't exist", TemplateDir));
                    return false;
                }
            }

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
