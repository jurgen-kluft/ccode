using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class Project
    {
        public static readonly List<string> AllGroups = new List<string>
        {
            "Configuration",
            "ImportGroup",
            "OutDir",
            "IntDir",
            "TargetName",
            "IgnoreImportLibrary",
            "GenerateManifest",
            "LinkIncremental",
            "ClCompile",
            "Link",
            "ResourceCompile",
            "Lib" 
        };
        public static readonly List<bool> IsGroup = new List<bool>
        {
            true,    /// "Configuration",
            true,    /// "ImportGroup",
            false,   /// "OutDir",
            false,   /// "IntDir",
            false,   /// "TargetName",
            false,   /// "IgnoreImportLibrary",
            false,   /// "GenerateManifest",
            false,   /// "LinkIncremental",
            true,    /// "ClCompile",
            true,    /// "Link",
            true,    /// "ResourceCompile",
            true,    /// "Lib" 
        };

        public class Settings  // For every platform
        {
            // TargetConfig
            private Dictionary<string, HashSet<string>> mPreprocessorDefinitions;
            private Dictionary<string, HashSet<string>> mIncludeDirs;
            private Dictionary<string, HashSet<string>> mLibraryDirs;
            private Dictionary<string, HashSet<string>> mLibraryDeps;

            public Settings()
            {
                mPreprocessorDefinitions = new Dictionary<string, HashSet<string>>();
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
                    bool skip = false;
                    foreach (string v in values)
                    {
                        if (v.StartsWith("#"))  // Skip any variable
                        {
                            skip = true;
                            break;
                        }
                    }
                    if (!skip)
                    {
                        foreach (string v in values)
                        {
                            if (v.StartsWith("#"))  // Skip any variable
                                continue;

                            if (!content.Contains(v))
                                content.Add(v);
                        }
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

            public void AddPreprocessorDefinitions(string config, string value, bool concat, string seperator)
            {
                Add(config, value, concat, seperator, mPreprocessorDefinitions);
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
            public string GetPreprocessorDefinitions(string config)
            {
                return Get(config, mPreprocessorDefinitions);
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

        protected Dictionary<string, Element> mElements = new Dictionary<string, Element>();
        protected Dictionary<string, List<Element>> mGroups = new Dictionary<string, List<Element>>();
        protected Dictionary<string, StringItems> mConfigs = new Dictionary<string, StringItems>();
        protected Dictionary<string, string> mTypes = new Dictionary<string, string>();
        protected Dictionary<string, Settings> mSettings = new Dictionary<string, Settings>();

        public Dictionary<string, Element> elements { get { return mElements; } }
        public Dictionary<string, List<Element>> groups { get { return mGroups; } }
        public Dictionary<string, StringItems> configs { get { return mConfigs; } }
        public Dictionary<string, string> types { get { return mTypes; } }

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

        public Project()
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

            mElements = new Dictionary<string, Element>();

            foreach (string g in AllGroups)
            {
                mGroups.Add(g, new List<Element>());
            }
        }

        public void Info()
        {
            Logger.Add(String.Format("Project                    : {0}", Name));
            Logger.Add(String.Format("Category                   : {0}", Category));
            Logger.Add(String.Format("Language                   : {0}", Language));
            Logger.Add(String.Format("Location                   : {0}", Location));
            Logger.Add(String.Format("UUID                       : {0}", UUID));
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

            this.Name = Attribute.Get("Name", node, "Unknown");
            this.Category = Attribute.Get("Category", node, "Main");
            this.Language = Attribute.Get("Language", node, "cpp");
            this.Location = Attribute.Get("Location", node, "source\\main\\cpp");

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool _continue = false;

                if (String.Compare(child.Name, "Configuration", true) == 0)
                {
                    StringItems platforms = new StringItems();
                    platforms.Add(Attribute.Get("Platform", child, string.Empty), true);
                    StringItems configs2 = new StringItems();
                    configs2.Add(Attribute.Get("Config", child, string.Empty), true);

                    foreach(string platform in platforms.ToArray())
                    {
                        StringItems items;
                        if (!this.configs.TryGetValue(platform, out items))
                        {
                            items = new StringItems();
                            this.configs.Add(platform, items);
                        }
                        items.Add(configs2);
                     }
                    _continue = true;
                }

                if (_continue)
                    continue;
                
                if (String.Compare(child.Name, "UUID", true) == 0)
                {
                    UUID = Element.sGetXmlNodeValueAsText(child);
                    _continue = true;
                }

                if (_continue)
                    continue;

                int index = Project.AllGroups.IndexOf(child.Name);
                if (index >= 0)
                {
                    string g = Project.AllGroups[index];
                    Element e = new Element(g, new List<Element>(), new List<Attribute>());
                    e.IsGroup = Project.IsGroup[index];
                    List<Element> elements;
                    this.groups.TryGetValue(g, out elements);
                    elements.Add(e);
                    e.Read(child);
                    _continue = true;
                }

                if (_continue)
                    continue;

                // It is an element
                Element element = new Element(child.Name, new List<Element>(), new List<Attribute>());
                {
                    if (child.HasChildNodes && child.FirstChild.NodeType == XmlNodeType.Text)
                        element.Value = child.FirstChild.Value;

                    if (child.Attributes != null)
                    {
                        foreach (XmlAttribute a in child.Attributes)
                        {
                            element.Attributes.Add(new Attribute(a.Name, a.Value));
                        }
                    }
                }
            }
        }

        public string[] GetPlatforms()
        {
            string[] platforms = new string[mConfigs.Keys.Count];
            mConfigs.Keys.CopyTo(platforms, 0);
            return platforms;
        }

        public string[] GetConfigsForPlatform(string platform)
        {
            StringItems items;
            if (mConfigs.TryGetValue(platform, out items))
            {
                return items.ToArray();
            }
            return new string[0];
        }

        public bool HasPlatformWithConfig(string platform, string config)
        {
            StringItems items;
            if (mConfigs.TryGetValue(platform, out items))
            {
                return items.Contains(config);
            }
            return false;
        }

        public void AddPreprocessorDefinitions(string platform, string targetconfig, string value, bool concat, string seperator)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
            {
                settings = new Settings();
                mSettings.Add(platform, settings);
            }
            settings.AddPreprocessorDefinitions(targetconfig, value, concat, seperator);
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
        public string GetPreprocessorDefinitions(string platform, string config)
        {
            Settings settings;
            if (!mSettings.TryGetValue(platform, out settings))
                return string.Empty;
            return settings.GetPreprocessorDefinitions(config);
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