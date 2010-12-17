using System;
using System.Xml;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class Platform
    {
        protected Dictionary<string, List<Element>> mGroups = new Dictionary<string, List<Element>>();
        protected Dictionary<string, Config> mConfigs = new Dictionary<string, Config>();

        public string Name { get; set; }

        public Dictionary<string, List<Element>> groups { get { return mGroups; } }
        public Dictionary<string, Config> configs { get { return mConfigs; } }

        public void Initialize(string p)
        {
            Name = p;
        }

        public void Read(XmlNode node)
        {
            if (!node.HasChildNodes)
                return;

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                bool do_continue = false;

                if (String.Compare(child.Name, "Config", true) == 0)
                {
                    string c = Attribute.Get("Name", child, "None");

                    Config config;
                    if (!configs.TryGetValue(c, out config))
                    {
                        config = new Config();
                        config.Initialize(Name, c);
                        configs.Add(c, config);
                    }
                    if (child.HasChildNodes)
                        config.Read(child);
                    do_continue = true;
                }

                if (do_continue)
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
                        do_continue = true;
                        break;
                    }
                }

                if (do_continue)
                    continue;

                // It is an element
                Element element = new Element(child.Name, new List<Element>(), new List<Attribute>());
                {
                    if (child.HasChildNodes && child.FirstChild.NodeType == XmlNodeType.Text)
                        element.Value = child.FirstChild.Value;

                    if (child.Attributes != null)
                    {
                        foreach (XmlAttribute a in child.Attributes)
                        {
                            element.Attributes.Add(new Attribute(a.Name, a.Value));
                        }
                    }
                }
            }
        }
    }

}
