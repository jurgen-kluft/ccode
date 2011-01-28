using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PomResource
    {
        private string mName;
        private Group mGroup;

        public string Name { get { return mName; } private set { mName = value; } }
        public Group Group { get { return mGroup; } }

        public bool IsValid { get { return !String.IsNullOrEmpty(Name); } }

        public List<Attribute> DirectoryStructure { get; set; }

        public Dictionary<string, List<KeyValuePair<string,string>>> Content { get; set; }
        public Dictionary<string, string> Vars { get; set; }
        public List<DependencyResource> Dependencies { get; set; }
        public List<ProjectResource> Projects { get; set; }
        public List<string> Platforms { get; set; }
        public Versions Versions { get; set; }

        public PomResource()
        {
            Name = string.Empty;
            mGroup = new Group(string.Empty);

            DirectoryStructure = new List<Attribute>();
            Content = new Dictionary<string, List<KeyValuePair<string, string>>>();
            Vars = new Dictionary<string, string>();
            Dependencies = new List<DependencyResource>();
            Projects = new List<ProjectResource>();
            Platforms = new List<string>();
            Versions = new Versions();
        }

        public static PomResource From(string name, string group)
        {
            PomResource resource = new PomResource();
            resource.Name = name;
            resource.mGroup = new Group(group);
            return resource;
        }

        public bool Info()
        {
            Loggy.Info(String.Format("Name                       : {0}", Name));

            Versions.Info();
            {
                Loggy.Indent += 1;
                foreach (ProjectResource p in Projects)
                {
                    Loggy.Info(String.Format("----------------------------"));
                    p.Info();
                }

                Loggy.Info(String.Format("----------------------------"));
                foreach (DependencyResource d in Dependencies)
                    d.Info();

                Loggy.Indent -= 1;
            }
            return true;
        }

        public ProjectResource GetProjectByName(string name)
        {
            foreach (ProjectResource p in Projects)
            {
                if (String.Compare(p.Name, name, true) == 0)
                    return p;
            }
            return null;
        }

        public void LoadFile(string filename)
        {
            XmlDocument xmlDoc = new XmlDocument();
            xmlDoc.Load(filename);
            Read(xmlDoc.FirstChild);
        }

        public void LoadXml(string xml)
        {
            XmlDocument xmlDoc = new XmlDocument();
            xmlDoc.LoadXml(xml);
            Read(xmlDoc.FirstChild);
        }

        private string ReplaceVars(string str, Dictionary<string, string> vars)
        {
            foreach (KeyValuePair<string, string> var in vars)
                str = str.Replace(String.Format("${{{0}}}", var.Key), var.Value);
            return str;
        }

        private void Read(XmlNode node)
        {
            if (node.Name == "Package")
            {
                if (node.Attributes != null)
                {
                    foreach (XmlAttribute a in node.Attributes)
                    {
                        if (a.Name == "Name")
                        {
                            Name = a.Value;
                        }
                        else if (a.Name == "Group")
                        {
                            mGroup = new Group(a.Value);
                        }
                    }
                }

                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        if (child.Name == "Versions")
                        {
                            Versions.Read(child);
                        }
                        else if (child.Name == "Variables")
                        {
                            if (child.HasChildNodes)
                            {
                                foreach (XmlNode item in child.ChildNodes)
                                {
                                    string text;
                                    if (!Vars.TryGetValue(item.Name, out text))
                                    {
                                        text = Element.GetText(item);
                                        Vars.Add(item.Name, text);
                                    }
                                }
                            }
                        }
                        else if (child.Name == "Content")
                        {
                            if (child.HasChildNodes)
                            {
                                foreach (XmlNode item in child.ChildNodes)
                                {
                                    string platform = Attribute.Get("Platform", item, "*");
                                    string src = Attribute.Get("Src", item, null);
                                    if (src != null)
                                    {
                                        ReplaceVars(src, Vars);
                                        string dst = Attribute.Get("Dst", item, null);
                                        if (dst != null)
                                        {
                                            ReplaceVars(dst, Vars);
                                            List<KeyValuePair<string, string>> items;
                                            if (!Content.TryGetValue(platform, out items))
                                            {
                                                items = new List<KeyValuePair<string, string>>();
                                                Content.Add(platform, items);
                                            }
                                            items.Add(new KeyValuePair<string, string>(src, dst));
                                        }
                                    }
                                }
                            }
                        }
                        else if (child.Name == "Dependency")
                        {
                            DependencyResource dependency = new DependencyResource();
                            dependency.Read(child);
                            Dependencies.Add(dependency);
                        }
                        else if (child.Name == "Project")
                        {
                            ProjectResource project = new ProjectResource();
                            project.Read(child, Vars);
                            Projects.Add(project);
                        }
                        else if (child.Name == "DirectoryStructure")
                        {
                            if (child.HasChildNodes)
                            {
                                foreach (XmlNode item in child.ChildNodes)
                                {
                                    string folder = Attribute.Get("Folder", item, string.Empty);
                                    if (!String.IsNullOrEmpty(folder))
                                    {
                                        folder = folder.Replace("${Name}", Name);
                                        DirectoryStructure.Add(new Attribute("Folder", folder));
                                    }
                                    string file = Attribute.Get("File", item, string.Empty);
                                    if (!String.IsNullOrEmpty(file))
                                    {
                                        file = file.Replace("${Name}", Name);
                                        DirectoryStructure.Add(new Attribute("File", file));
                                    }
                                }
                            }
                        }
                    }
                }
            }

            Name = ReplaceVars(Name, Vars);
            Group.ExpandVars(Vars);

            foreach (DependencyResource dependencyResource in Dependencies)
                dependencyResource.ExpandVars(Vars);

            HashSet<string> all_platforms = new HashSet<string>();
            foreach (ProjectResource project in Projects)
            {
                string[] platforms = project.GetPlatforms();
                foreach (string platform in platforms)
                {
                    if (!all_platforms.Contains(platform))
                        all_platforms.Add(platform);
                }
            }
            foreach (string platform in all_platforms)
                Platforms.Add(platform);
        }
    }
}
