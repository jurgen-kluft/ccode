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
        private Project mProject;
        private string[] mProjectPlatforms;
        private string[] mProjectConfigs;

        public XProjectWriter(Project project, string[] platforms, string[] configs)
        {
            mProject = project;
            mProjectPlatforms = platforms;
            mProjectConfigs = configs;
        }

        public Project Project { get { return mProject; } }

        public void Write(string filename, Project project, string[] platforms, string[] configs)
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

        private void ConvertElementsToLines(List<Element> elements, List<string> lines, string platform, string config)
        {
            // Build the lines
            // If contains #(Configuration) and/or #(Platform) then iterate
            foreach (Element e in elements)
            {
                string line = e.ToString();
                bool iterator_platform = line.Contains("#(Platform)");
                bool iterator_config = line.Contains("#(Configuration)");
                if (iterator_platform && iterator_config)
                {
                    //foreach (string platform in mProjectPlatforms)
                    {
                        string l1 = line.Replace("#(Platform)", platform);
                        //foreach (string config in mProjectConfigs)
                        {
                            string l2 = l1.Replace("#(Configuration)", config);
                            lines.Add(l2);
                        }
                    }
                }
                else if (iterator_platform)
                {
                    //foreach (string platform in mProjectPlatforms)
                    {
                        string l1 = line.Replace("#(Platform)", platform);
                        lines.Add(l1);
                    }
                }
                else if (iterator_config)
                {
                    //foreach (string config in mProjectConfigs)
                    {
                        string l1 = line.Replace("#(Configuration)", config);
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

            Platform xplatform;
            if (mProject.Platforms.TryGetValue(platform, out xplatform))
            {
                Config xconfig;
                if (xplatform.configs.TryGetValue(config, out xconfig))
                {
                    List<Element> elements;
                    if (xconfig.groups.TryGetValue(group, out elements))
                    {
                        if (elements.Count > 0)
                        {
                            if (elements[0].IsGroup)
                                ConvertElementsToLines(elements[0].Elements, lines, platform, config);
                            else
                                ConvertElementsToLines(elements, lines, platform, config);
                        }
                    }
                }
            }
            return lines;
        }
       
    }
}