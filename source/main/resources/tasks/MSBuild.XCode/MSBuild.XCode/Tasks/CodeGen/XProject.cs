using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;

namespace MSBuild.XCode
{
    public class XProject
    {
        public static readonly string[] AllGroups = new string[] {"Configuration","ImportGroup","OutDir","IntDir",
                                                                   "TargetName","IgnoreImportLibrary","GenerateManifest",
                                                                        "LinkIncremental","ClCompile","Link","ResourceCompile",
                                                                            "Lib" 
        };

        protected Dictionary<string, XElement> mElements = new Dictionary<string, XElement>();
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();
        protected Dictionary<string, XConfig> mConfigs = new Dictionary<string, XConfig>();
        protected Dictionary<string, string> mTypes = new Dictionary<string, string>();
        protected Dictionary<string, XPlatform> mPlatforms = new Dictionary<string, XPlatform>();

        public Dictionary<string, XElement> elements { get { return mElements; } }
        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
        public Dictionary<string, XConfig> configs { get { return mConfigs; } }
        public Dictionary<string, string> types { get { return mTypes; } }
        public Dictionary<string, XPlatform> platforms { get { return mPlatforms; } }

        public string Name { get; set; }
        public string Language { get; set; }
        public string Location { get; set; }
        public string UUID { get; set; }
        public string Extension
        {
            get
            {
                if (Language == "C#")
                    return ".csproj";
                return ".vcxproj";
            }
        }

        public XProject()
        {
            Name = "Unknown";
            Language = "C++";
            Location = @"source\main\cpp";
            UUID = string.Empty;
        }

        public void Initialize()
        {
            Name = "Unknown";
            Language = "C++";
            Location = @"source\main\cpp";
            UUID = Guid.NewGuid().ToString();

            mElements = new Dictionary<string, XElement>();

            foreach (string g in AllGroups)
            {
                mGroups.Add(g, new List<XElement>());
            }
        }

        public void Load(string filename)
        {
            Name = string.Empty;
            Language = string.Empty;
            Location = string.Empty;

            XmlDocument xmlDoc = new XmlDocument();
            xmlDoc.Load(filename);
            Read(xmlDoc.FirstChild);
        }

        public void Read(XmlNode node)
        {
            Initialize();

            this.Name = XAttribute.Get("Name", node, "Unknown");
            this.Language = XAttribute.Get("Language", node, "C++");
            this.Location = XAttribute.Get("Location", node, "source\\main\\cpp");

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool do_continue = false;

                if (String.Compare(child.Name, "Platform", true) == 0)
                {
                    string p = XAttribute.Get("Name", child, "Unknown");

                    XPlatform platform;
                    if (!this.platforms.TryGetValue(p, out platform))
                    {
                        platform = new XPlatform();
                        platform.Initialize(p);
                        this.platforms.Add(p, platform);
                    }
                    platform.Read(child);
                    do_continue = true;
                }

                if (do_continue)
                    continue;

                if (String.Compare(child.Name, "Config", true) == 0)
                {
                    string c = XAttribute.Get("Name", child, "Unknown");

                    XConfig config;
                    if (!this.configs.TryGetValue(c, out config))
                    {
                        config = new XConfig();
                        config.Initialize("Any", c);
                        this.configs.Add(c, config);
                    }
                    config.Read(child);
                    do_continue = true;
                }

                if (do_continue)
                    continue;

                if (String.Compare(child.Name, "Type", true) == 0)
                {
                    string t = XAttribute.Get("Name", child, "Unknown");

                    string type;
                    if (!this.types.TryGetValue(t, out type))
                    {
                        this.types.Add(t, t);
                    }

                    do_continue = true;
                }

                if (do_continue)
                    continue;

                foreach (string g in XProject.AllGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        this.groups.TryGetValue(g, out elements);
                        elements.Add(e);
                        e.Read(child);
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