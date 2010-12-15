using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XPackage
    {
        public string Name { get; set; }
        public XGroup Group { get; set; }

        public string IncludePath { get; set; }
        public string LibraryPath { get; set; }

        public List<XAttribute> DirectoryStructure { get; set; }

        public List<XDependency> Dependencies { get; set; }
        public List<XProject> Projects { get; set; }
        public List<XProject> Templates { get; set; }
        public XVersions Versions { get; set; }
        public XDependencyTree DependencyTree { get; set; }

        public XPackage()
        {
            Name = "Unknown";
            Group = new XGroup("com.virtuos.tnt");

            DirectoryStructure = new List<XAttribute>();
            Dependencies = new List<XDependency>();
            Projects = new List<XProject>();
            Templates = new List<XProject>();
            Versions = new XVersions();
            DependencyTree = new XDependencyTree();
        }

        public void Load(string filename)
        {
            XmlDocument xmlDoc = new XmlDocument();
            xmlDoc.Load(filename);
            Read(xmlDoc.FirstChild);

            foreach (XProject p in Projects)
            {
                XProject template = null;
                foreach (XProject t in Templates)
                {
                    if (t.Language == p.Language)
                    {
                        template = t;
                        break;
                    }
                }
                if (template != null)
                    XProjectMerge.Merge(p, template);
            }
        }

        public void LoadXml(string xml)
        {
            XmlDocument xmlDoc = new XmlDocument();
            xmlDoc.LoadXml(xml);
            Read(xmlDoc.FirstChild);

            foreach (XProject p in Projects)
            {
                XProject template = null;
                foreach (XProject t in Templates)
                {
                    if (t.Language == p.Language)
                    {
                        template = t;
                        break;
                    }
                }
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
            bool result = DependencyTree.Build(Platform, localRepo, remoteRepo);
            return result;
        }

        public void PrintDependencies()
        {
            DependencyTree.Print();
        }

        public bool CheckoutDependencies(string Path, string Platform, XPackageRepository localRepo)
        {
            bool result = DependencyTree.Checkout(Path, Platform, localRepo);
            return result;
        }
    }
}
