using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace MSBuild.XCode
{
    public class XProjectReader
    {
        private string[] mGroups = new string[] {"Configuration","ImportGroup","OutDir","IntDir",
                                                    "TargetName","IgnoreImportLibrary","GenerateManifest",
                                                    "LinkIncremental","ClCompile","Link","ResourceCompile",
                                                    "Lib" 
        };

        public XProject Read(string filename)
        {
            XProject prj = new XProject();
            prj.Initialize(mGroups);

            XmlDocument _project = new XmlDocument();
            _project.Load(filename);
            Read(_project.FirstChild, prj);

            return prj;
        }

        private void Read(XmlNode node, XElement parent)
        {
            if (node.Attributes != null && node.Attributes.Count > 0)
            {
                foreach (XmlAttribute a in node.Attributes)
                {
                    if (a.Name == "Concat")
                    {
                        parent.Concat = String.Compare(a.Value, "true", true) == 0 ? true : false;
                    }
                    else if (a.Name == "Separator")
                    {
                        parent.Separator = a.Value;
                    }
                    else
                    {
                        // A real attribute
                        parent.Attributes.Add(new XAttribute(a.Name, a.Value));
                    }
                }
            }

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;
                if (child.NodeType == XmlNodeType.Text)
                {
                    parent.Value = child.Value;
                    continue;
                }

                XElement e = new XElement(child.Name, new List<XElement>(), new List<XAttribute>());
                parent.Elements.Add(e);

                Read(child, e);
            }
        }

        private void Read(XmlNode node, XConfig cfg)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                foreach (string g in mGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        cfg.groups.TryGetValue(g, out elements);
                        elements.Add(e);
                        Read(child, e);
                        break;
                    }
                }
            }
        }

        private void Read(XmlNode node, XPlatform plm)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool do_continue = false;

                if (String.Compare(child.Name, "Config", true) == 0)
                {
                    string c = "None";
                    if (child.Attributes != null)
                    {
                        foreach (XmlAttribute a in child.Attributes)
                        {
                            if (a.Name == "Name")
                                c = a.Value;
                        }
                    }
                    XConfig config;
                    plm.configs.TryGetValue(c, out config);
                    if (child.HasChildNodes)
                        Read(child.FirstChild, config);
                    do_continue = true;
                }

                if (do_continue)
                    continue;

                foreach (string g in mGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        plm.groups.TryGetValue(g, out elements);
                        elements.Add(e);
                        Read(child, e);
                        break;
                    }
                }
            }
        }

        private void Read(XmlNode node, XProject prj)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool do_continue = false;

                if (String.Compare(child.Name, "Platform", true) == 0)
                {
                    string p = "None";
                    if (child.Attributes != null)
                    {
                        foreach (XmlAttribute a in child.Attributes)
                        {
                            if (a.Name == "Name")
                                p = a.Value;
                        }
                    }

                    XPlatform platform;
                    if (!prj.platforms.TryGetValue(p, out platform))
                    {
                        platform = new XPlatform();
                        prj.platforms.Add(p, platform);
                    }
                    if (child.HasChildNodes)
                        Read(child.FirstChild, platform);
                    do_continue = true;
                }

                if (do_continue)
                    continue;

                foreach (string g in mGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        prj.groups.TryGetValue(g, out elements);
                        elements.Add(e);
                        Read(child, e);
                        do_continue = true;
                        break;
                    }
                }

                if (do_continue)
                    continue;

                // It is an element
                XElement element = new XElement(child.Name, new List<XElement>(), new List<XAttribute>());
                {
                    if (child.HasChildNodes && child.FirstChild.NodeType == XmlNodeType.Text)
                        element.Value = child.FirstChild.Value;

                    if (child.Attributes != null)
                    {
                        foreach (XmlAttribute a in child.Attributes)
                        {
                            element.Attributes.Add(new XAttribute(a.Name, a.Value));
                        }
                    }
                }
            }
        }
    }
}