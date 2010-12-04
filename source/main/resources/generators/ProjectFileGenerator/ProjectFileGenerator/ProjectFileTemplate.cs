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

        private string[] mProjectPlatforms;
        private string[] mProjectConfigs;

        private string[] mGroups = new string[] {"Configuration","ImportGroup","OutDir","IntDir",
                                                    "TargetName","IgnoreImportLibrary","GenerateManifest",
                                                    "LinkIncremental","ClCompile","Link","ResourceCompile",
                                                    "Lib" 
        };
        private XProject mTemplate;
        private XProject mProject;

        void InternalLoad(string template_filename, string project_filename)
        {
            mTemplate = new XProject();
            mTemplate.Initialize(mPlatforms, mConfigs, mGroups);

            XmlDocument _template = new XmlDocument();
            _template.Load(template_filename);
            Parse(_template.FirstChild, mTemplate);

            mProject = new XProject();
            mProject.Initialize(mPlatforms, mConfigs, mGroups);

            XmlDocument _project = new XmlDocument();
            _project.Load(project_filename);
            Parse(_project.FirstChild, mProject);

            // Now merge the template and the project
            // By default every element is 'Replace'
            // Some elements however are set to 'Concat'
            // Default separator is ';'
            Merge(mProject, mTemplate);
        }

        string ConvertElementToString(XElement e)
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

        void ConvertElementsToLines(List<XElement> elements, List<string> lines)
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

        List<string> InternalGetGroupElementsFor(string platform, string config, string group)
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

        class XAttribute
        {
            public XAttribute(string name, string value)
            {
                Name = name;
                Value = value;
            }

            public string Name { get; set; }
            public string Value { get; set; }

            public XAttribute Copy()
            {
                XAttribute a = new XAttribute(Name, Value);
                return a;
            }
        }


        class XElement
        {
            public XElement(string name, List<XElement> elements, List<XAttribute> attributes)
            {
                Name = name;
                Attributes = attributes;
                Elements = elements;
            }

            public string Name { get; set; }
            public List<XAttribute> Attributes { get; set; }
            public string Value { get; set; }
            public bool Concat { get; set; }
            public string Seperator { get; set; }
            public List<XElement> Elements { get; set; }

            public void Init()
            {
                Name = string.Empty;
                Attributes = new List<XAttribute>();
                Value = string.Empty;
                Concat = false;
                Seperator = ";";
                Elements = new List<XElement>();
            }

            public XElement Copy()
            {
                XElement c = new XElement(Name, new List<XElement>(), new List<XAttribute>());
                c.Value = Value;
                foreach (XAttribute a in Attributes)
                    c.Attributes.Add(a.Copy());
                foreach (XElement e in Elements)
                    c.Elements.Add(e.Copy());
                return c;
            }
        }


        private static void Merge(Dictionary<string, List<XElement>> main, Dictionary<string, List<XElement>> template)
        {
            foreach (KeyValuePair<string, List<XElement>> template_group in template)
            {
                if (main.ContainsKey(template_group.Key))
                {
                    // Merge
                    List<XElement> mainElementsList;
                    main.TryGetValue(template_group.Key, out mainElementsList);

                    Dictionary<string, XElement> mainElementsDict = new Dictionary<string, XElement>();
                    foreach (XElement e in mainElementsList)
                    {
                        mainElementsDict.Add(e.Name, e);
                    }

                    foreach (XElement e in template_group.Value)
                    {
                        if (mainElementsDict.ContainsKey(e.Name))
                        {
                            // Merge element if concatenation of the values is required
                            if (e.Concat)
                            {
                                XElement this_e;
                                mainElementsDict.TryGetValue(e.Name, out this_e);
                                this_e.Value = this_e.Value + e.Seperator + e.Value;
                            }
                        }
                        else
                        {
                            // Add element
                            mainElementsList.Add(e.Copy());
                        }
                    }

                }
                else
                {
                    // Clone
                    List<XElement> elements = new List<XElement>();
                    main.Add(template_group.Key, elements);
                    foreach (XElement e in template_group.Value)
                        elements.Add(e.Copy());
                }
            }
        }

        private void Merge(XPlatform main, XPlatform template)
        {
            Merge(main.groups, template.groups);
            foreach (KeyValuePair<string, XConfig> p in template.configs)
            {
                XConfig x;
                main.configs.TryGetValue(p.Key, out x);
                Merge(x.groups, main.groups);
                Merge(x.groups, p.Value.groups);
            }
        }

        private void Merge(XProject main, XProject template)
        {
            Merge(main.groups, template.groups);
            foreach (KeyValuePair<string, XPlatform> p in template.platforms)
            {
                XPlatform x;
                main.platforms.TryGetValue(p.Key, out x);
                Merge(x.groups, main.groups);
                Merge(x, p.Value);
            }
        }

        class XProject
        {
            protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();
            protected Dictionary<string, XPlatform> mPlatforms = new Dictionary<string,XPlatform>();

            public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
            public Dictionary<string, XPlatform> platforms { get { return mPlatforms; } }

            public void Initialize(string[] platforms, string[] configs, string[] groups)
            {
                foreach (string g in groups)
                {
                    mGroups.Add(g, new List<XElement>());
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
            protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string,List<XElement>>();
            protected Dictionary<string, XConfig> mConfigs = new Dictionary<string, XConfig>();

            public string Name { get; set; }

            public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
            public Dictionary<string, XConfig> configs { get { return mConfigs; } }

            public void Initialize(string p, string[] configs, string[] groups)
            {
                Name = p;
                foreach (string g in groups)
                {
                    mGroups.Add(g, new List<XElement>());
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
            protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string,List<XElement>>();

            public string Name { get; set; }
            public string Config { get; set; }
            public string Platform { get; set; }

            public Dictionary<string, List<XElement>> groups { get { return mGroups; } }

            public void Initialize(string p, string c, string[] groups)
            {
                Name = c + "|" + p;
                Platform = p;
                Config = c;

                foreach (string g in groups)
                {
                    mGroups.Add(g, new List<XElement>());
                }
            }
        }

        void Parse(XmlNode node, XElement parent)
        {
            if (node.Attributes != null && node.Attributes.Count > 0)
            {
                foreach (XmlAttribute a in node.Attributes)
                {
                    if (a.Name == "Concat")
                    {
                        parent.Concat = String.Compare(a.Value, "true", true) == 0 ? true : false;
                    }
                    else if (a.Name == "Seperator")
                    {
                        parent.Seperator = a.Value;
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

                Parse(child, e);
            }
        }

        void Parse(XmlNode node, XConfig cfg)
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
                        Parse(child, e);
                        break;
                    }
                }
            }
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
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        plm.groups.TryGetValue(g, out elements);
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
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        prj.groups.TryGetValue(g, out elements);
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