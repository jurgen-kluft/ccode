using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;

namespace MSBuild.XCode
{
    public class Config
    {
        protected Dictionary<string, List<Element>> mGroups = new Dictionary<string, List<Element>>();

        public string Name { get; set; }
        public string Configuration { get; set; }
        public string Platform { get; set; }

        public Dictionary<string, List<Element>> groups { get { return mGroups; } }

        public void Initialize(string p, string c)
        {
            Name = c + "|" + p;
            Platform = p;
            Configuration = c;
        }

        public Element FindElement(string group, string element)
        {
            List<Element> elements;
            if (groups.TryGetValue(group, out elements))
            {
                if (elements.Count > 0)
                {
                    foreach (Element e in elements[0].Elements)
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

                foreach (string g in Project.AllGroups)
                {
                    if (String.Compare(child.Name, g, true) == 0)
                    {
                        Element e = new Element(g, new List<Element>(), new List<Attribute>());
                        List<Element> elements;
                        if (!groups.TryGetValue(g, out elements))
                        {
                            elements = new List<Element>();
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