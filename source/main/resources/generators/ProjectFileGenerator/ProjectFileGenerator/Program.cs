using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace ProjectFileGenerator
{
    class Program
    {
        static void Main(string[] args)
        {
            string[] all_platforms = new string[] { "Win32", "NintendoDS", "NintendoWII", "Nintendo3DS", "SonyPSP", "SonyPS2", "SonyPS3", "Microsoft360" };
            string[] all_configs = new string[] { "Debug", "Release", "Profile", "Final" };

            string[] project_platforms = new string[] { "Win32", "SonyPS3" };
            string[] project_configs = new string[] { "Debug", "Release" };

            ProjectFileTemplate p = new ProjectFileTemplate(all_platforms, all_configs, project_platforms, project_configs);
            p.Load(@"d:\Dev\HgDev.Modules\xmaven\source\main\resources\generators\ProjectFileGenerator\ProjectFileGenerator\vcxproj_template.xml",
                @"d:\Dev\HgDev.Modules\xmaven\source\main\resources\generators\ProjectFileGenerator\ProjectFileGenerator\vcxproj_project.xml");

            ProjectFileGenerator g = new ProjectFileGenerator("xbase", Guid.NewGuid().ToString(), ProjectFileGenerator.EVersion.VS2010,
                ProjectFileGenerator.ELanguage.CPP, project_platforms, project_configs, p);
            g.Save("D:\\xbase.vcxproj");
        }
    }
}
