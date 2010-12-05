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

        private string ConvertElementToString(XElement e)
        {
            string str;
            if (e.Attributes.Count == 0)
            {
                if (String.IsNullOrEmpty(e.Value))
                    str = String.Format("<{0} />", e.Name);
                else
                    str = String.Format("<{0}>{1}</{0}>", e.Name, e.Value);
            }
            else
            {
                string attributes = string.Empty;
                foreach (XAttribute a in e.Attributes)
                {
                    string attribute = a.Name + "=\"" + a.Value + "\"";
                    if (String.IsNullOrEmpty(attributes))
                        attributes = attribute;
                    else
                        attributes = attributes + " " + attribute;
                }
                if (String.IsNullOrEmpty(e.Value))
                    str = String.Format("<{0} {2} />", e.Name, e.Value, attributes);
                else
                    str = String.Format("<{0} {2}>{1}</{0}>", e.Name, e.Value, attributes);
            }
            return str;
        }

        private void ConvertElementsToLines(List<XElement> elements, List<string> lines)
        {
            // Build the lines
            // If contains #(Configuration) and/or #(Platform) then iterate
            foreach (XElement e in elements)
            {
                string line = ConvertElementToString(e);
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
            if (mProject.platforms.TryGetValue(platform, out xplatform))
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