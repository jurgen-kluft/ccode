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
        protected MsDev2010.Cpp.XCode.Project mMsDevProject;
        protected MsDev2010.Cpp.XCode.Project mMsDevProjectFull;

        protected Dictionary<string, Element> mElements = new Dictionary<string, Element>();
        protected Dictionary<string, StringItems> mConfigs = new Dictionary<string, StringItems>();
        protected Dictionary<string, string> mTypes = new Dictionary<string, string>();

        public Dictionary<string, Element> Elements { get { return mElements; } }
        public Dictionary<string, StringItems> Configs { get { return mConfigs; } }

        public string Name { get; set; }
        public string Category { get; set; }
        public string Language { get; set; }
        public string Location { get; set; }
        public string Extension
        {
            get
            {
                if (IsCs)
                    return ".csproj";
                return ".vcxproj";
            }
        }

        public bool IsCpp { get { return (String.Compare(Language, "C++", true) == 0 || String.Compare(Language, "CPP", true) == 0); } }
        public bool IsCs { get { return (String.Compare(Language, "C#", true) == 0 || String.Compare(Language, "CS", true) == 0); } }

        public Project()
        {
            Name = "Unknown";
            Category = "Main";
            Language = "cpp";   /// "cs"
            Location = @"source\main\cpp";
        }

        public void Initialize()
        {
            Name = "Unknown";
            Category = "Main";
            Language = "cpp";
            Location = @"source\main\cpp";

            mMsDevProject = new MsDev2010.Cpp.XCode.Project();
            mMsDevProjectFull = new MsDev2010.Cpp.XCode.Project();

            mElements = new Dictionary<string, Element>();
        }

        public void Info()
        {
            Loggy.Add(String.Format("Project                    : {0}", Name));
            Loggy.Add(String.Format("Category                   : {0}", Category));
            Loggy.Add(String.Format("Language                   : {0}", Language));
            Loggy.Add(String.Format("Location                   : {0}", Location));
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
                {
                    continue;
                }
                else if (String.Compare(child.Name, "ProjectFile", true) == 0)
                {
                    mMsDevProject = new MsDev2010.Cpp.XCode.Project(child.ChildNodes);
                }
                else
                {
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

            // Now extract the platforms and configs from the ProjectFile
            string[] platforms = mMsDevProject.GetPlatforms();
            foreach (string platform in platforms)
            {
                string[] configs = mMsDevProject.GetPlatformConfigs(platform);
                mConfigs.Add(platform, new StringItems(configs));
            }
        }

        public void ConstructFullMsDevProject()
        {
            if (IsCpp)
            {
                mMsDevProjectFull = new MsDev2010.Cpp.XCode.Project();
                mMsDevProjectFull.Merge(Global.CppTemplateProject);
                mMsDevProjectFull.Merge(mMsDevProject);
                mMsDevProjectFull.RemoveAllBut(mConfigs);
            }
            else if (IsCs)
            {
                mMsDevProjectFull = new MsDev2010.Cpp.XCode.Project();
                mMsDevProjectFull.Merge(Global.CsTemplateProject);
                mMsDevProjectFull.Merge(mMsDevProject);
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

        public void Save(string rootdir, string filename)
        {
            string reldir = rootdir + Path.GetDirectoryName(filename);
            reldir = reldir.EndWith('\\');

            mMsDevProjectFull.ExpandGlobs(rootdir, reldir);
            mMsDevProjectFull.Save(rootdir + filename);
        }
    }
}