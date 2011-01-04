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

        public void OnlyKeepPlatformSpecifics(string platform)
        {
            foreach (Project prj in Projects)
            {
                prj.OnlyKeepPlatformSpecifics(platform);
            }
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
                                    string src = Attribute.Get("Src", item, null);
                                    if (src != null)
                                    {
                                        string dst = Attribute.Get("Dst", item, null);
                                        if (dst != null)
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
                            Project project = new Project(Main);
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
            // Generating project files is a bit complex in that it has to merge project definitions on a per platform basis.
            // Every platform is considered to have its own package (zip -> pom.xml) containing the project settings for that platform.
            // For every platform we have to merge in only the conditional xml elements into the final project file.
            foreach (Project rootProject in Projects)
            {
                // Every platform has a dependency tree and every dependency package for that platform has filtered their
                // project to only keep their platform specific xml elements.
                foreach (KeyValuePair<string, DependencyTree> pair in DependencyTree)
                {
                    List<Package> allDependencyPackages = pair.Value.GetAllDependencyPackages();

                    // Merge in all Projects of those dependency packages which are already filtered on the platform
                    foreach (Package dependencyPackage in allDependencyPackages)
                    {
                        Project dependencyProject = dependencyPackage.Pom.GetProjectByCategory(rootProject.Category);
                        if (dependencyProject!=null && !dependencyProject.IsPrivate)
                            rootProject.MergeWithDependencyProject(dependencyProject);
                    }
                }
            }

            root = root.EndWith('\\');
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

            List<string> projectFilenames = new List<string>();
            foreach (Project prj in Projects)
            {
                string path = prj.Location.Replace("/", "\\");
                path = path.EndWith('\\');
                path = path + prj.Name + prj.Extension;
                projectFilenames.Add(path);
            }

            MsDev2010.Cpp.XCode.Solution solution = new MsDev2010.Cpp.XCode.Solution(MsDev2010.Cpp.XCode.Solution.EVersion.VS2010, MsDev2010.Cpp.XCode.Solution.ELanguage.CPP);
            string solutionFilename = root + Name + ".sln";
            solution.Save(solutionFilename, projectFilenames);
        }

        public bool BuildDependencies(string Platform)
        {
            bool result = true;

            DependencyTree tree;
            if (!DependencyTree.TryGetValue(Platform, out tree))
            {
                tree = new DependencyTree();
                tree.Package = this;
                tree.Platform = Platform;
                tree.Dependencies = Dependencies;
                DependencyTree.Add(Platform, tree);
                result = tree.Build();
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
                result = tree.Sync();
            return result;
        }

    }
}
