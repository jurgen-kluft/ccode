using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class Pom
    {
        public string Name { get; set; }
        public Group Group { get; set; }

        public EType Type { get; set; }
        public bool Main { get { return Type == EType.Main; } set { if (value) Type = EType.Main; else Type = EType.Dependency; } }

        public List<Attribute> DirectoryStructure { get; set; }

        public Dictionary<string, List<KeyValuePair<string,string>>> Content { get; set; }
        public List<Dependency> Dependencies { get; set; }
        public List<Project> Projects { get; set; }
        public List<string> Platforms { get; set; }
        public Versions Versions { get; set; }
        public Dictionary<string, DependencyTree> DependencyTree { get; set; }

        public enum EType
        {
            Main,
            Dependency,
        }

        public Pom(EType type)
        {
            Name = "Unknown";
            Group = new Group("com.virtuos.tnt");
            Type = type;

            DirectoryStructure = new List<Attribute>();
            Content = new Dictionary<string, List<KeyValuePair<string, string>>>();
            Dependencies = new List<Dependency>();
            Projects = new List<Project>();
            Platforms = new List<string>();
            Versions = new Versions();
            DependencyTree = new Dictionary<string, DependencyTree>();
        }

        public bool Info()
        {
            Loggy.Add(String.Format("Name                       : {0}", Name));
            Loggy.Add(String.Format("Group                      : {0}", Group.ToString()));
            Versions.Info();
            {
                Loggy.Indent += 1;
                foreach (Project p in Projects)
                {
                    Loggy.Add(String.Format("----------------------------"));
                    p.Info();
                }

                Loggy.Add(String.Format("----------------------------"));
                foreach (Dependency d in Dependencies)
                    d.Info();

                Loggy.Indent -= 1;
            }
            return true;
        }

        public Project GetProjectByCategory(string category)
        {
            foreach (Project p in Projects)
            {
                if (String.Compare(p.Category, category, true) == 0)
                    return p;
            }
            return null;
        }

        public string[] GetCategories()
        {
            List<string> categories = new List<string>();
            foreach (Project prj in Projects)
            {
                categories.Add(prj.Category);
            }
            return categories.ToArray();
        }

        public string[] GetPlatformsForCategory(string Category)
        {
            Project project = GetProjectByCategory(Category);
            if (project != null)
                return project.GetPlatforms();
            return new string[0];
        }

        public string[] GetConfigsForPlatformsForCategory(string Platform, string Category)
        {
            Project project = GetProjectByCategory(Category);
            if (project!=null)
                return project.GetConfigsForPlatform(Platform);
            return new string[0];
        }

        public void Load(string filename)
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

        public void PostLoad()
        {
            foreach (Project prj in Projects)
            {
                prj.ConstructFullMsDevProject();
            }
        }

        public void Read(XmlNode node)
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
                            Group.Full = a.Value;
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
                        else if (child.Name == "Content")
                        {
                            if (child.HasChildNodes)
                            {
                                foreach (XmlNode item in child.ChildNodes)
                                {
                                    string platform = Attribute.Get("Platform", item, "*");
                                    string src = Attribute.Get("Src", item, string.Empty);
                                    if (!String.IsNullOrEmpty(src))
                                    {
                                        string dst = Attribute.Get("Dst", item, string.Empty);
                                        if (!String.IsNullOrEmpty(dst))
                                        {
                                            List<KeyValuePair<string,string>> items;
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
                            Dependency dependency = new Dependency();
                            dependency.Read(child);
                            Dependencies.Add(dependency);
                        }
                        else if (child.Name == "Project")
                        {
                            Project project = new Project();
                            project.Main = Main;
                            project.Read(child);
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
                                        DirectoryStructure.Add(new Attribute("Folder", folder));
                                    string file = Attribute.Get("File", item, string.Empty);
                                    if (!String.IsNullOrEmpty(file))
                                        DirectoryStructure.Add(new Attribute("File", file));
                                }
                            }
                        }
                    }
                }
            }

            HashSet<string> all_platforms = new HashSet<string>();
            string[] categories = GetCategories();
            foreach (string category in categories)
            {
                string[] platforms = GetPlatformsForCategory(category);
                foreach (string platform in platforms)
                {
                    if (!all_platforms.Contains(platform))
                        all_platforms.Add(platform);
                }
            }
            foreach (string platform in all_platforms)
                Platforms.Add(platform);
        }

        public void GenerateProjects(string root)
        {
            root = root.EndWith('\\');

            // Here merge with all Projects of the dependency packages, but only:
            ///    1) PreprocessorDefinitions
            ///    2) AdditionalIncludeDirectories
            ///    3) AdditionalDependencies
            ///    4) AdditionalLibraryDirectories
            /// Also, missing nodes should not be added!

            foreach (Project p in Projects)
            {
                string path = p.Location.Replace("/", "\\");
                path = path.EndWith('\\');
                string filename = path + p.Name + p.Extension;
                p.Save(root, filename);
            }
        }

        public void GenerateSolution(string root)
        {
            root = root.EndWith('\\');

            string filename = root + Name + ".sln";

            List<string> projectFilenames = new List<string>();
            foreach (Project prj in Projects)
            {
                string path = prj.Location.Replace("/", "\\");
                path = path.EndWith('\\');
                string f = path + prj.Name + prj.Extension;
                projectFilenames.Add(f);
            }

            MsDev2010.Cpp.XCode.Solution solution = new MsDev2010.Cpp.XCode.Solution(MsDev2010.Cpp.XCode.Solution.EVersion.VS2010, MsDev2010.Cpp.XCode.Solution.ELanguage.CPP);
            solution.Save(filename, projectFilenames);
        }

        public bool BuildDependencies(string Platform)
        {
            bool result = true;

            DependencyTree tree;
            if (!DependencyTree.TryGetValue(Platform, out tree))
            {
                tree = new DependencyTree();
                tree.Name = Name;
                tree.Dependencies = Dependencies;
                tree.Package = this;
                DependencyTree.Add(Platform, tree);

                tree.Version = Versions.GetForPlatform(Platform);
                result = tree.Build(Platform);
            }

            return result;
        }

        public void PrintDependencies(string Platform)
        {
            DependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                tree.Print();
        }

        public bool SyncDependencies(string Platform)
        {
            bool result = false;
            DependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                result = tree.Sync(Platform);
            return result;
        }

    }
}
