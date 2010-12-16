using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;

namespace MSBuild.XCode
{
    public class XConfig
    {
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();

        public string Name { get; set; }
        public string Config { get; set; }
        public string Platform { get; set; }

        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }

        public void Initialize(string p, string c)
        {
            Name = c + "|" + p;
            Platform = p;
            Config = c;
        }

        public XElement FindElement(string group, string element)
        {
            List<XElement> elements;
            if (groups.TryGetValue(group, out elements))
            {
                if (elements.Count > 0)
                {
                    foreach (XElement e in elements[0].Elements)
                    {
                        if (String.Compare(e.Name, element, true) == 0)
                            return e;
                    }
                }
            }
            return null;
        }

        public void Read(XmlNode node)
        {
            if (!node.HasChildNodes)
                return;

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                foreach (string g in XProject.AllGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        XElement e = new XElement(g, new List<XElement>(), new List<XAttribute>());
                        List<XElement> elements;
                        if (!groups.TryGetValue(g, out elements))
                        {
                            elements = new List<XElement>();
                            groups.Add(g, elements);
                        }
                        elements.Add(e);
                        e.Read(child);
                        break;
                    }
                }
            }
        }

    }
}