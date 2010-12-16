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

        public class Settings  // For every platform
        {
            // TargetConfig
            private Dictionary<string, HashSet<string>> mIncludeDirs;
            private Dictionary<string, HashSet<string>> mLibraryDirs;
            private Dictionary<string, HashSet<string>> mLibraryDeps;

            public Settings()
            {
                mIncludeDirs = new Dictionary<string, HashSet<string>>();
                mLibraryDirs = new Dictionary<string, HashSet<string>>();
                mLibraryDeps = new Dictionary<string, HashSet<string>>();
            }

            private void Add(string config, string value, bool concat, string seperator, Dictionary<string, HashSet<string>> items)
            {
                HashSet<string> content;
                if (items.TryGetValue(config, out content))
                {
                    if (!concat)
                        content.Clear();
                }
                else
                {
                    content = new HashSet<string>();
                    items.Add(config, content);
                }

                if (!String.IsNullOrEmpty(value))
                {
                    string[] values = value.Split(new string[] { seperator }, StringSplitOptions.RemoveEmptyEntries);
                    foreach (string v in values)
                    {
                        if (v.StartsWith("#"))  // Skip any variable
                            continue;

                        if (!content.Contains(v))
                            content.Add(v);
                    }
                }
            }

            private string Get(string config, Dictionary<string, HashSet<string>> items)
            {
                HashSet<string> content;
                string str = string.Empty;
                if (items.TryGetValue(config, out content))
                {
                    string seperator = string.Empty;
                    foreach (string s in content)
                    {
                        str = str + seperator + s;
                        seperator = ";";
                    }
                }
                return str;
            }

            public void AddIncludeDir(string config, string value, bool concat, string seperator)
            {
                Add(config, value, concat, seperator, mIncludeDirs);
            }
            public void AddLibraryDir(string config, string value, bool concat, string seperator)
            {
                Add(config, value, concat, seperator, mLibraryDirs);
            }
            public void AddLibraryDep(string config, string value, bool concat, string seperator)
            {
                Add(config, value, concat, seperator, mLibraryDeps);
            }
            public string GetIncludeDir(string config)
            {
                return Get(config, mIncludeDirs);
            }
            public string GetLibraryDir(string config)
            {
                return Get(config, mLibraryDirs);
            }
            public string GetLibraryDep(string config)
            {
                return Get(config, mLibraryDeps);
            }
        }

        protected Dictionary<string, XElement> mElements = new Dictionary<string, XElement>();
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();
        protected Dictionary<string, XConfig> mConfigs = new Dictionary<string, XConfig>();
        protected Dictionary<string, string> mTypes = new Dictionary<string, string>();
        protected Dictionary<string, XPlatform> mPlatforms = new Dictionary<string, XPlatform>();
        protected Dictionary<string, Settings> mSettings = new Dictionary<string, Settings>();

        public Dictionary<string, XElement> elements { get { return mElements; } }
        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
        public Dictionary<string, XConfig> configs { get { return mConfigs; } }
        public Dictionary<string, string> types { get { return mTypes; } }
        public Dictionary<string, XPlatform> Platforms { get { return mPlatforms; } }

        public string Name { get; set; }
        public string Category { get; set; }
        public string Language { get; set; }
        public string Location { get; set; }
        public string UUID { get; set; }
        public string Extension
        {
            get
            {
                if (String.Compare(Language,"C#",true)==0 || String.Compare(Language, "cs", true)==0)
                    return ".csproj";
                return ".vcxproj";
            }
        }

        public XProject()
        {
            Name = "Unknown";
            Category = "Main";
            Language = "cpp";   /// "cs"
            Location = @"source\main\cpp";
            UUID = string.Empty;
        }

        public void Initialize()
        {
            Name = "Unknown";
            Category = "Main";
            Language = "cpp";
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
            Category = "Main";
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
            this.Category = XAttribute.Get("Category", node, "Main");
            this.Language = XAttribute.Get("Language", node, "cpp");
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
                    if (!this.Platforms.TryGetValue(p, out platform))
                    {
                        platform = new XPlatform();
                        platform.Initialize(p);
                        this.Platforms.Add(p, platform);
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

        public void AddIncludeDir(string platform, string targetconfig, string value, bool concat, string seperator)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
            {
                settings = new Settings();
                mSettings.Add(platform, settings);
            }
            settings.AddIncludeDir(targetconfig, value, concat, seperator);
        }
        public void AddLibraryDir(string platform, string targetconfig, string value, bool concat, string seperator)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
            {
                settings = new Settings();
                mSettings.Add(platform, settings);
            }
            settings.AddLibraryDir(targetconfig, value, concat, seperator);
        }
        public void AddLibraryDep(string platform, string targetconfig, string value, bool concat, string seperator)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
            {
                settings = new Settings();
                mSettings.Add(platform, settings);
            }
            settings.AddLibraryDep(targetconfig, value, concat, seperator);
        }

        public string GetIncludeDir(string platform, string config)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
                return string.Empty;
            return settings.GetIncludeDir(config);
        }
        public string GetLibraryDir(string platform, string config)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
                return string.Empty;
            return settings.GetLibraryDir(config);
        }
        public string GetLibraryDep(string platform, string config)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
                return string.Empty;
            return settings.GetLibraryDep(config);
        }
    }
}