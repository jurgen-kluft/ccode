using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class Pom
    {
        public string Name { get; set; }
        public Group Group { get; set; }

        public string IncludePath { get; set; }
        public string LibraryPath { get; set; }
        public string LibraryDep { get; set; }

        public List<Attribute> DirectoryStructure { get; set; }

        public List<Dependency> Dependencies { get; set; }
        public List<Project> Projects { get; set; }
        public Versions Versions { get; set; }
        public Dictionary<string, DependencyTree> DependencyTree { get; set; }

        public Pom()
        {
            Name = "Unknown";
            Group = new Group("com.virtuos.tnt");

            DirectoryStructure = new List<Attribute>();
            Dependencies = new List<Dependency>();
            Projects = new List<Project>();
            Versions = new Versions();
            DependencyTree = new Dictionary<string, DependencyTree>();
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

        public Platform GetPlatformByCategory(string platform, string category)
        {
            foreach (Project prj in Projects)
            {
                if (String.Compare(prj.Category, category, true) == 0)
                {
                    Platform p;
                    if (prj.Platforms.TryGetValue(platform, out p))
                        return p;
                }
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
            List<string> platforms = new List<string>();
            foreach (Platform p in project.Platforms.Values)
                platforms.Add(p.Name);
            return platforms.ToArray();
        }

        public string[] GetConfigsForPlatformsForCategory(string Platform, string Category)
        {
            Platform platform = GetPlatformByCategory(Platform, Category);
            List<string> configs = new List<string>();
            foreach (Config c in platform.configs.Values)
                configs.Add(c.Configuration);
            return configs.ToArray();
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
            foreach (Project p in Projects)
            {
                Project template = Global.GetTemplate(p.Language);
                if (template != null)
                    XProjectMerge.Merge(p, template);
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
                        else if (child.Name == "Dependency")
                        {
                            Dependency dependency = new Dependency();
                            dependency.Read(child);
                            Dependencies.Add(dependency);
                        }
                        else if (child.Name == "Project")
                        {
                            Project project = new Project();
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
                        else
                        {
                            // Elements: IncludePath, LibraryPath
                            if (child.Name == "IncludePath")
                            {
                                IncludePath = Element.sGetXmlNodeValueAsText(child);
                            }
                            else if (child.Name == "LibraryPath")
                            {
                                LibraryPath = Element.sGetXmlNodeValueAsText(child);
                            }
                            else if (child.Name == "LibraryDep")
                            {
                                LibraryDep = Element.sGetXmlNodeValueAsText(child);
                            }
                        }
                    }
                }
            }
        }

        public void GenerateProjects(string root)
        {
            if (!root.EndsWith("\\"))
                root = root + "\\";

            foreach (Project p in Projects)
            {
                MsDevProjectFileGenerator generator = new MsDevProjectFileGenerator(p.Name, p.UUID, MsDevProjectFileGenerator.EVersion.VS2010, MsDevProjectFileGenerator.ELanguage.CPP, p);
                string path = p.Location.Replace("/", "\\");
                path = path.EndsWith("\\") ? path : (path + "\\");
                string filename = root + path + p.Name + p.Extension;
                generator.Save(filename);
            }
        }

        public void GenerateSolution(string root)
        {
            if (!root.EndsWith("\\"))
                root = root + "\\";

            string filename = root + Name + ".sln";

            List<string> projectFilenames = new List<string>();
            foreach (Project prj in Projects)
            {
                string path = prj.Location.Replace("/", "\\");
                path = path.EndsWith("\\") ? path : (path + "\\");
                string f = path + prj.Name + prj.Extension;
                projectFilenames.Add(f);
            }

            MsDevSolutionGenerator generator = new MsDevSolutionGenerator(MsDevSolutionGenerator.EVersion.VS2010, MsDevSolutionGenerator.ELanguage.CPP);
            generator.Save(filename, projectFilenames);
        }

        public bool BuildDependencies(string Platform, PackageRepository localRepo, PackageRepository remoteRepo)
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

        public bool SyncDependencies(string Platform, PackageRepository localRepo)
        {
            bool result = false;
            DependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                result = tree.Sync(Platform, localRepo);
            return result;
        }

        public void CollectProjectInformation(string Category, string Platform, string Config)
        {
            DependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                tree.CollectProjectInformation(Category, Platform, Config);
        }
    }
}
