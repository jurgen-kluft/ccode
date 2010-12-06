using System;
using System.Xml;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XPlatform
    {
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();
        protected Dictionary<string, XConfig> mConfigs = new Dictionary<string, XConfig>();

        public string Name { get; set; }

        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
        public Dictionary<string, XConfig> configs { get { return mConfigs; } }

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
                    string c = XAttribute.Get("Name", child, "None");

                    XConfig config;
                    if (!configs.TryGetValue(c, out config))
                    {
                        config = new XConfig();
                        config.Initialize(Name, c);
                        configs.Add(c, config);
                    }
                    if (child.HasChildNodes)
                        config.Read(child);
                    do_continue = true;
                }

                if (do_continue)
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
                        do_continue = true;
                        break;
                    }
                }

                if (do_continue)
                    continue;

                // It is an element
                XElement element = new XElement(child.Name, new List<XElement>(), new List<XAttribute>());
                {
                    if (child.HasChildNodes && child.FirstChild.NodeType == XmlNodeType.Text)
                        element.Value = child.FirstChild.Value;

                    if (child.Attributes != null)
                    {
                        foreach (XmlAttribute a in child.Attributes)
                        {
                            element.Attributes.Add(new XAttribute(a.Name, a.Value));
                        }
                    }
                }
            }
        }
    }

}
