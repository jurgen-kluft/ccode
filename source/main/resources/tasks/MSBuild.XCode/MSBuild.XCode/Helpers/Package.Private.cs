using System;
using System.Xml;
using System.IO;
using System.Collections.Generic;
using System.Text;

namespace MSBuild.Cod.Helpers
{
    public partial class Package
    {
        private string[] mGroups = new string[]
        {
            "Dependency",
            "Project",
        };

        class XDependency
        {
            public string Name { get; set; }
            public bool IsPackage { get; set; }
            List<XElement> Elements { get; set; }
        }

        class XPackage
        {
            List<XElement> Elements { get; set; }
        }

        private List<XProject> mProjects;


        private void _Load(string filename)
        {
            XmlDocument _package = new XmlDocument();
            _package.Load(filename);

        }
    }
}
