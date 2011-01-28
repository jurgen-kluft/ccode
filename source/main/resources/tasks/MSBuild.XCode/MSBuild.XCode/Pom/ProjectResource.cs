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
        protected MsDev.IProject mMsDevProject;

        protected Dictionary<string, StringItems> mConfigs = new Dictionary<string, StringItems>();

        public Dictionary<string, StringItems> Configs { get { return mConfigs; } }

        public string Name { get; set; }
        public string Scope { get; set; }       ///< Public / Private
        public string Language { get; set; }
        public string Location { get; set; }
        public string DependsOn { get; set; }

        internal MsDev.IProject MsDevProject { get { return mMsDevProject; } }

        public bool IsCpp { get { return (String.Compare(Language, "C++", true) == 0 || String.Compare(Language, "CPP", true) == 0); } }
        public bool IsCs { get { return (String.Compare(Language, "C#", true) == 0 || String.Compare(Language, "CS", true) == 0); } }

        public bool IsPrivate { get { return String.Compare(Scope, "Private") == 0; } }

        public ProjectResource()
        {
            Name = "Unknown";
            Scope = "Public";
            Language = "C++";
            Location = @"source\main\cpp";
            DependsOn = "";
        }

        public ProjectInstance CreateInstance(bool main)
        {
            ProjectInstance instance = new ProjectInstance(main, this);
            return instance;
        }

        public void Info()
        {
            Loggy.Info(String.Format("Project                    : {0}", Name));
            Loggy.Info(String.Format("Language                   : {0}", Language));
            Loggy.Info(String.Format("Location                   : {0}", Location));
        }

        private string ReplaceVars(string str, Dictionary<string, string> vars)
        {
            foreach (KeyValuePair<string, string> var in vars)
                str = str.Replace(String.Format("${{{0}}}", var.Key), var.Value);
            return str;
        }

        public void Read(XmlNode node, Dictionary<string, string> vars)
        {
            this.Name = Attribute.Get("Name", node, "Unknown");
            this.Language = Attribute.Get("Language", node, "C++");
            this.Location = Attribute.Get("Location", node, "source\\main\\cpp");
            this.Scope = Attribute.Get("Scope", node, "Public");
            this.DependsOn = Attribute.Get("DependsOn", node, "");

            this.Name = ReplaceVars(this.Name, vars);
            this.Language = ReplaceVars(this.Language, vars);
            this.Location = ReplaceVars(this.Location, vars);
            this.Scope = ReplaceVars(this.Scope, vars);
            this.DependsOn = ReplaceVars(this.DependsOn, vars);

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                {
                    continue;
                }
                else if (String.Compare(child.Name, "ProjectFile", true) == 0)
                {
                    if (IsCpp)
                        mMsDevProject = new MsDev.CppProject(child.ChildNodes);
                    else
                        mMsDevProject = new MsDev.CsProject(child.ChildNodes);
                    mMsDevProject.ExpandVars(vars);
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