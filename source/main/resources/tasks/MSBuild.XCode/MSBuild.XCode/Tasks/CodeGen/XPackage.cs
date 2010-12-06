using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XPackage
    {
        public string Name { get; set; }
        public XVersion Version { get; set; }
        public XGroup Group { get; set; }

        public string IncludePath { get; set; }
        public string LibraryPath { get; set; }

        public List<XDependency> Dependencies { get; set; }
        public List<XProject> Projects { get; set; }
        public List<XProject> Templates { get; set; }

        public XPackage()
        {
            Name = "Unknown";
            Group = new XGroup("com.virtuos.tnt");
            Version = new XVersion("1.0");

            Dependencies = new List<XDependency>();
            Projects = new List<XProject>();
            Templates = new List<XProject>();
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
                        else if (a.Name == "Version")
                        {
                            Version.ParseVersion(a.Value);
                        }
                        else if (a.Name == "Group")
                        {
                            Group.Group = a.Value;
                        }
                    }
                }

                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        if (child.Name == "Dependency")
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

        public void GenerateProjects()
        {
            foreach (XProject p in Projects)
            {
                MsDevProjectFileGenerator generator = new MsDevProjectFileGenerator(p.Name, p.UUID, MsDevProjectFileGenerator.EVersion.VS2010, MsDevProjectFileGenerator.ELanguage.CPP, p);
                string path = p.Location.Replace("/", "\\");
                path = path.EndsWith("\\") ? path : (path + "\\");
                string filename = path + p.Name + p.Extension;
                generator.Save(filename);
            }
        }

        public void GenerateSolution(string path)
        {
            MsDevSolutionGenerator generator = new MsDevSolutionGenerator(MsDevSolutionGenerator.EVersion.VS2010, MsDevSolutionGenerator.ELanguage.CPP);
            string filename = path.EndsWith("\\") ? (path + Name + ".sln") : (path + "\\" + Name + ".sln");

            List<string> projectFilenames = new List<string>();
            foreach (XProject prj in Projects)
            {
                string p = prj.Location.Replace("/", "\\");
                p = p.EndsWith("\\") ? p : (p + "\\");
                string f = p + prj.Name + prj.Extension;
                projectFilenames.Add(f);
            }

            generator.Save(filename, projectFilenames);
        }
    }
}
