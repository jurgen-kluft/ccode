using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XPom
    {
        public string Name { get; set; }
        public XGroup Group { get; set; }

        public string IncludePath { get; set; }
        public string LibraryPath { get; set; }
        public string LibraryDep { get; set; }

        public List<XAttribute> DirectoryStructure { get; set; }

        public List<XDependency> Dependencies { get; set; }
        public List<XProject> Projects { get; set; }
        public XVersions Versions { get; set; }
        public Dictionary<string, XDependencyTree> DependencyTree { get; set; }

        public XPom()
        {
            Name = "Unknown";
            Group = new XGroup("com.virtuos.tnt");

            DirectoryStructure = new List<XAttribute>();
            Dependencies = new List<XDependency>();
            Projects = new List<XProject>();
            Versions = new XVersions();
            DependencyTree = new Dictionary<string, XDependencyTree>();
        }

        public XProject GetProjectByCategory(string category)
        {
            foreach (XProject p in Projects)
            {
                if (String.Compare(p.Category, category, true) == 0)
                    return p;
            }
            return null;
        }

        public XPlatform GetPlatformByCategory(string platform, string category)
        {
            foreach (XProject prj in Projects)
            {
                if (String.Compare(prj.Category, category, true) == 0)
                {
                    XPlatform p;
                    if (prj.Platforms.TryGetValue(platform, out p))
                        return p;
                }
            }
            return null;
        }

        public string[] GetCategories()
        {
            List<string> categories = new List<string>();
            foreach (XProject prj in Projects)
            {
                categories.Add(prj.Category);
            }
            return categories.ToArray();
        }

        public string[] GetPlatformsForCategory(string Category)
        {
            XProject project = GetProjectByCategory(Category);
            List<string> platforms = new List<string>();
            foreach (XPlatform p in project.Platforms.Values)
                platforms.Add(p.Name);
            return platforms.ToArray();
        }

        public string[] GetConfigsForPlatformsForCategory(string Platform, string Category)
        {
            XPlatform platform = GetPlatformByCategory(Platform, Category);
            List<string> configs = new List<string>();
            foreach (XConfig c in platform.configs.Values)
                configs.Add(c.Config);
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
            foreach (XProject p in Projects)
            {
                XProject template = XGlobal.GetTemplate(p.Language);
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
                            XDependency dependency = new XDependency();
                            dependency.Read(child);
                            Dependencies.Add(dependency);
                        }
                        else if (child.Name == "Project")
                        {
                            XProject project = new XProject();
                            project.Read(child);
                            Projects.Add(project);
                        }
                        else if (child.Name == "DirectoryStructure")
                        {
                            if (child.HasChildNodes)
                            {
                                foreach (XmlNode item in child.ChildNodes)
                                {
                                    string folder = XAttribute.Get("Folder", item, string.Empty);
                                    if (!String.IsNullOrEmpty(folder))
                                        DirectoryStructure.Add(new XAttribute("Folder", folder));
                                    string file = XAttribute.Get("File", item, string.Empty);
                                    if (!String.IsNullOrEmpty(file))
                                        DirectoryStructure.Add(new XAttribute("File", file));
                                }
                            }
                        }
                        else
                        {
                            // Elements: IncludePath, LibraryPath
                            if (child.Name == "IncludePath")
                            {
                                IncludePath = XElement.sGetXmlNodeValueAsText(child);
                            }
                            else if (child.Name == "LibraryPath")
                            {
                                LibraryPath = XElement.sGetXmlNodeValueAsText(child);
                            }
                            else if (child.Name == "LibraryDep")
                            {
                                LibraryDep = XElement.sGetXmlNodeValueAsText(child);
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

            foreach (XProject p in Projects)
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
            foreach (XProject prj in Projects)
            {
                string path = prj.Location.Replace("/", "\\");
                path = path.EndsWith("\\") ? path : (path + "\\");
                string f = path + prj.Name + prj.Extension;
                projectFilenames.Add(f);
            }

            MsDevSolutionGenerator generator = new MsDevSolutionGenerator(MsDevSolutionGenerator.EVersion.VS2010, MsDevSolutionGenerator.ELanguage.CPP);
            generator.Save(filename, projectFilenames);
        }

        public bool BuildDependencies(string Platform, XPackageRepository localRepo, XPackageRepository remoteRepo)
        {
            bool result = true;

            XDependencyTree tree;
            if (!DependencyTree.TryGetValue(Platform, out tree))
            {
                tree = new XDependencyTree();
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
            XDependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                tree.Print();
        }

        public bool CheckoutDependencies(string Path, string Platform, XPackageRepository localRepo)
        {
            bool result = false;
            XDependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                result = tree.Checkout(Path, Platform, localRepo);
            return result;
        }

        public void CollectProjectInformation(string Category, string Platform, string Config)
        {
            XDependencyTree tree;
            if (DependencyTree.TryGetValue(Platform, out tree))
                tree.CollectProjectInformation(Category, Platform, Config);
        }
    }
}
