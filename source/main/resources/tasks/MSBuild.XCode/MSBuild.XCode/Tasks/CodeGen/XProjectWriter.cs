using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace MSBuild.XCode
{
    public class XProjectWriter
    {
        private XProject mProject;
        private string[] mProjectPlatforms;
        private string[] mProjectConfigs;

        public XProjectWriter(XProject project, string[] platforms, string[] configs)
        {
            mProject = project;
            mProjectPlatforms = platforms;
            mProjectConfigs = configs;
        }

        public void Write(string filename, XProject project, string[] platforms, string[] configs)
        {
            mProject = project;
            mProjectPlatforms = platforms;
            mProjectConfigs = configs;
        }

        private string GetAdditionalIncludeDirectories()
        {
            // Dependency:
            //   If Type == "Package"
            //     $(SolutionDir)target\$(package_name)\$(platform)\include
            //   Else
            //     $(SolutionDir)target\$(package_name)\$(platform)\source\main\include
            return string.Empty;
        }

        private string GetAdditionalLibraryDirectories()
        {
            // Dependency:
            //  $(SolutionDir)target\$(package_name)\$(platform)\libs
            return string.Empty;
        }

        private string GetAdditionalLibraryDependencies()
        {
            // Dependency:
            //  $(SolutionDir)target\$(package_name)\$(platform)\libs\$(package_name)_Dev_Debug_Win32.lib
            return string.Empty;
        }

        private void ConvertElementsToLines(List<XElement> elements, List<string> lines)
        {
            // Build the lines
            // If contains #(Configuration) and/or #(Platform) then iterate
            foreach (XElement e in elements)
            {
                string line = e.ToString();
                bool iterator_platform = line.Contains("#(Platform)");
                bool iterator_config = line.Contains("#(Configuration)");
                if (iterator_platform && iterator_config)
                {
                    foreach (string p in mProjectPlatforms)
                    {
                        string l1 = line.Replace("#(Platform)", p);
                        foreach (string c in mProjectConfigs)
                        {
                            string l2 = l1.Replace("#(Configuration)", c);
                            lines.Add(l2);
                        }
                    }
                }
                else if (iterator_platform)
                {
                    foreach (string p in mProjectPlatforms)
                    {
                        string l1 = line.Replace("#(Platform)", p);
                        lines.Add(l1);
                    }
                }
                else if (iterator_config)
                {
                    foreach (string c in mProjectConfigs)
                    {
                        string l1 = line.Replace("#(Configuration)", c);
                        lines.Add(l1);
                    }
                }
                else
                {
                    lines.Add(line);
                }
            }
        }

        public List<string> GetGroupElementsFor(string platform, string config, string group)
        {
            List<string> lines = new List<string>();

            XPlatform xplatform;
            if (mProject.Platforms.TryGetValue(platform, out xplatform))
            {
                XConfig xconfig;
                if (xplatform.configs.TryGetValue(config, out xconfig))
                {
                    List<XElement> elements;
                    if (xconfig.groups.TryGetValue(group, out elements))
                    {
                        if (elements.Count == 1 && elements[0].Name == group)
                            ConvertElementsToLines(elements[0].Elements, lines);
                        else
                            ConvertElementsToLines(elements, lines);
                    }
                }
            }
            return lines;
        }
    }
}