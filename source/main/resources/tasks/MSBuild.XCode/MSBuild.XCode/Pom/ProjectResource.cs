using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class ProjectResource
    {
        protected CppProject mMsDevProject;

        protected Dictionary<string, StringItems> mConfigs = new Dictionary<string, StringItems>();

        public Dictionary<string, StringItems> Configs { get { return mConfigs; } }

        public string Name { get; set; }
        public string Scope { get; set; }       ///< Public / Private
        public string Group { get; set; }
        public string Language { get; set; }
        public string Location { get; set; }
        public string DependsOn { get; set; }

        internal CppProject MsDevProject { get { return mMsDevProject; } }

        public bool IsCpp { get { return (String.Compare(Language, "C++", true) == 0 || String.Compare(Language, "CPP", true) == 0); } }
        public bool IsCs { get { return (String.Compare(Language, "C#", true) == 0 || String.Compare(Language, "CS", true) == 0); } }

        public bool IsPrivate { get { return String.Compare(Scope, "Private") == 0; } }

        public ProjectResource()
        {
            Name = "Unknown";
            Scope = "Public";
            Group = "Main";
            Language = "C++";
            Location = @"source\main\cpp";
            DependsOn = "";

            mMsDevProject = new CppProject();
        }

        public void Info()
        {
            Loggy.Add(String.Format("Project                    : {0}", Name));
            Loggy.Add(String.Format("Category                   : {0}", Group));
            Loggy.Add(String.Format("Language                   : {0}", Language));
            Loggy.Add(String.Format("Location                   : {0}", Location));
        }

        public void Read(XmlNode node)
        {
            this.Name = Attribute.Get("Name", node, "Unknown");
            this.Group = Attribute.Get("Group", node, "Main");
            this.Language = Attribute.Get("Language", node, "C++");
            this.Location = Attribute.Get("Location", node, "source\\main\\cpp");
            this.Scope = Attribute.Get("Scope", node, "Public");
            this.DependsOn = Attribute.Get("DependsOn", node, "");

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                {
                    continue;
                }
                else if (String.Compare(child.Name, "ProjectFile", true) == 0)
                {
                    mMsDevProject = new CppProject(child.ChildNodes);
                }
                else
                {
                    // It is an unknown element
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
                return items.ToArray();
            return new string[0];
        }

        public bool HasPlatformWithConfig(string platform, string config)
        {
            StringItems items;
            if (mConfigs.TryGetValue(platform, out items))
                return items.Contains(config);
            return false;
        }
    }
}