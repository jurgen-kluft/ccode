using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode
{
    public static class Global
    {
        public static string CacheRepoDir { get; set; }
        public static string RemoteRepoDir { get; set; }
        public static string TemplateDir { get; set; }
        public static string RootDir { get; set; }

        public static PackageRepository CacheRepo { get; set; }
        public static PackageRepository RemoteRepo { get; set; }

        public static List<Project> TemplateProjects { get; set; }

        public static void Initialize()
        {
            RemoteRepo = new PackageRepository(RemoteRepoDir, ELocation.Remote);
            CacheRepo = new PackageRepository(CacheRepoDir, ELocation.Cache);

            TemplateProjects = new List<Project>();

            if (!String.IsNullOrEmpty(TemplateDir))
            {
                // For C++
                Project CppProjectTemplate = new Project();
                CppProjectTemplate.Language = "cpp";
                CppProjectTemplate.Load(TemplateDir + "vcxproj.xml.template");
                TemplateProjects.Add(CppProjectTemplate);

                // For C#
                //XProject CsProjectTemplate = new XProject();
                //CsProjectTemplate.Language = "cs";
                //CsProjectTemplate.Load(TemplateDir + "csproj.xml.template");
                //TemplateProjects.Add(CsProjectTemplate);
            }
        }

        public static Project GetTemplate(string language)
        {
            Project template = null;
            foreach (Project t in TemplateProjects)
            {
                if (String.Compare("C++", language, true) == 0 || String.Compare("CPP", language, true) == 0)
                {
                    if (String.Compare("C++", t.Language, true) == 0 || String.Compare("CPP", t.Language, true) == 0)
                    {
                        return t;
                    }
                }
            }
            return template;
        }
    }
}
