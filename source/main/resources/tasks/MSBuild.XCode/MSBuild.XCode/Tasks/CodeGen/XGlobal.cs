using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode
{
    public static class XGlobal
    {
        public static string LocalRepoDir { get; set; }
        public static string RemoteRepoDir { get; set; }
        public static string TemplateDir { get; set; }

        public static XPackageRepository LocalRepo { get; set; }
        public static XPackageRepository RemoteRepo { get; set; }

        public static List<XProject> TemplateProjects { get; set; }

        public static void Initialize()
        {
            LocalRepo = new XPackageRepository(LocalRepoDir);
            RemoteRepo = new XPackageRepository(RemoteRepoDir);

            TemplateProjects = new List<XProject>();

            if (!String.IsNullOrEmpty(TemplateDir))
            {
                // For C++
                XProject CppProjectTemplate = new XProject();
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

        public static XProject GetTemplate(string language)
        {
            XProject template = null;
            foreach (XProject t in TemplateProjects)
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
