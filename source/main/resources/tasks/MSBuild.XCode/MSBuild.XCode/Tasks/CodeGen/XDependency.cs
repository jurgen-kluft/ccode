using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XDependency
    {
        public XDependency()
        {
            Group = new XGroup("com.virtuos.tnt");
            Version = new XVersion("1.0");
            Branch = "default";
            Type = "Package";
        }

        public XGroup Group { get; set; }
        public XVersion Version { get; set; }
        public string Branch { get; set; }
        public string Type { get; set; }

        public void Read(XmlNode node)
        {
            if (node.Name == "Dependency")
            {
                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        if (child.NodeType == XmlNodeType.Comment)
                            continue;

                        if (child.Name == "Group")
                        {
                            Group.Group = XElement.sGetXmlNodeValueAsText(child);
                        }
                        else if (child.Name == "Version")
                        {
                            Version.ParseVersion(XElement.sGetXmlNodeValueAsText(child));
                        }
                        else if (child.Name == "Branch")
                        {
                            Branch = XElement.sGetXmlNodeValueAsText(child);
                        }
                        else if (child.Name == "Type")
                        {
                            Type = XElement.sGetXmlNodeValueAsText(child);
                        }
                    }
                }
            }
        }

        public void Sync(string remote_repo, string local_repo, string path, string[] platforms)
        {

        }
   }
}