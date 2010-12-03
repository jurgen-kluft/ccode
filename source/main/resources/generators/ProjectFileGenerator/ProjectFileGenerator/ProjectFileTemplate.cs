using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace ProjectFileGenerator
{
    public partial class ProjectFileTemplate
    {
        private string[] mPlatforms;
        private string[] mConfigs;
        private string[] mGroups = new string[] {"Configuration","ImportGroup","OutDir","IntDir",
                                                    "TargetName","IgnoreImportLibrary","GenerateManifest",
                                                    "LinkIncremental","ClCompile","Link","ResourceCompile",
                                                    "Lib" 
        };

        void InternalLoad(string filename)
        {
            mTemplateProject = new XProject();
            mTemplateProject.Initialize(mPlatforms, mConfigs, mGroups);

            XmlDocument _doc = new XmlDocument();
            _doc.Load(filename);
            Parse(_doc.FirstChild, mTemplateProject);
        }

        List<string> InternalGetGroupElementsFor(string platform, string config, string group)
        {
            List<string> lines = new List<string>();
            
            XPlatform xplatform;
            if (mTemplateProject.platforms.TryGetValue(platform, out xplatform))
            {
                XConfig xconfig;
                if (xplatform.configs.TryGetValue(config, out xconfig))
                {
                    List<XElement> elements;
                    if (xconfig.template.TryGetValue(group, out elements))
                    {
                        // Build the lines
                        if (group == "Import")
                        {

                        }

                    }
                }
            }

            return lines;
        }
        
        class XElement
        {
            public XElement(string name, List<XElement> elements)
            {
                Name = name;
                Elements = elements;
            }

            public string Name { get; set; }
            public string Value { get; set; }
            public List<XElement> Elements { get; set; }

            public void Init()
            {
                Elements = new List<XElement>();
            }
        }

        class XProject
        {
            protected Dictionary<string, List<XElement>> mTemplate = new Dictionary<string, List<XElement>>();
            protected Dictionary<string, XPlatform> mPlatforms = new Dictionary<string,XPlatform>();

            public Dictionary<string, List<XElement>> template { get { return mTemplate; } }
            public Dictionary<string, XPlatform> platforms { get { return mPlatforms; } }

            public void Initialize(string[] platforms, string[] configs, string[] groups)
            {
                foreach (string g in groups)
                {
                    mTemplate.Add(g, new List<XElement>());
                }

                foreach (string p in platforms)
                {
                    XPlatform platform = new XPlatform();
                    platform.Initialize(p, configs, groups);
                    mPlatforms.Add(p, platform);
                }
            }
        }
        class XPlatform
        {
            protected Dictionary<string, List<XElement>> mTemplate = new Dictionary<string,List<XElement>>();
            protected Dictionary<string, XConfig> mConfigs = new Dictionary<string, XConfig>();

            public string Name { get; set; }

            public Dictionary<string, List<XElement>> template { get { return mTemplate; } }
            public Dictionary<string, XConfig> configs { get { return mConfigs; } }

            public void Initialize(string p, string[] configs, string[] groups)
            {
                Name = p;
                foreach (string g in groups)
                {
                    mTemplate.Add(g, new List<XElement>());
                }

                foreach (string c in configs)
                {
                    XConfig config = new XConfig();
                    config.Initialize(p, c, groups);
                    mConfigs.Add(c, config);
                }
            }
        }

        class XConfig
        {
            protected Dictionary<string, List<XElement>> mTemplate = new Dictionary<string,List<XElement>>();

            public string Name { get; set; }
            public string Config { get; set; }
            public string Platform { get; set; }

            public Dictionary<string, List<XElement>> template { get { return mTemplate; } }

            public void Initialize(string p, string c, string[] groups)
            {
                Name = c + "|" + p;
                Platform = p;
                Config = c;

                foreach (string g in groups)
                {
                    mTemplate.Add(g, new List<XElement>());
                }
            }
        }

        private XProject mTemplateProject;

        void Parse(XmlNode node, XElement parent)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;
                if (child.NodeType == XmlNodeType.Text)
                {
                    parent.Value = child.Value;
                    continue;
                }

                XElement e = new XElement(child.Name, new List<XElement>());
                parent.Elements.Add(e);

                Parse(child, e);
            }
        }

        void Parse(XmlNode node, XConfig cfg)
        {
            if (node.NodeType == XmlNodeType.Comment)
                return;

        }

        void Parse(XmlNode node, XPlatform plm)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool do_continue = false;
                foreach (string g in mGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>());
                        List<XElement> elements;
                        plm.template.TryGetValue(g, out elements);
                        elements.Add(e);
                        Parse(child, e);
                        do_continue = true;
                        break;
                    }
                }
                if (do_continue)
                    continue;

                foreach (string c in mConfigs)
                {
                    if (String.Compare(child.Name, c, true) == 0)
                    {
                        XConfig config;
                        plm.configs.TryGetValue(c, out config);
                        Parse(child.FirstChild, config);
                        do_continue = true;
                        break;
                    }
                }
                if (do_continue)
                    continue;
            }
        }

        void Parse(XmlNode node, XProject prj)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool do_continue = false;
                foreach (string g in mGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>());
                        List<XElement> elements;
                        prj.template.TryGetValue(g, out elements);
                        elements.Add(e);
                        Parse(child, e);
                        do_continue = true;
                        break;
                    }
                }
                if (do_continue)
                    continue;

                foreach (string p in mPlatforms)
                {
                    if (String.Compare(child.Name, p, true) == 0)
                    {
                        XPlatform platform;
                        prj.platforms.TryGetValue(p, out platform);
                        if (child.HasChildNodes)
                            Parse(child.FirstChild, platform);
                        do_continue = true;
                        break;
                    }
                }
                if (do_continue)
                    continue;
            }
        }


    }
}