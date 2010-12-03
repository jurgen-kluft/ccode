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
            string[] platforms = new string[] { "Win32", "NintendoDS", "NintendoWII", "Nintendo3DS", "SonyPSP", "SonyPS2", "SonyPS3", "Microsoft360" };
            string[] configs = new string[] { "Debug", "Release", "Profile", "Final" };

            ProjectFileTemplate p = new ProjectFileTemplate(platforms, configs);
            p.Load(@"d:\.NET_MSBUILD\ProjectFileGenerator\ProjectFileGenerator\vcxproj_template.xml");
        }
    }
}
