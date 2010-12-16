using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace MSBuild.XCode
{
    public class XElement
    {
        private XElement()
        {

        }

        public XElement(string name, List<XElement> elements, List<XAttribute> attributes)
        {
            Name = name;
            Attributes = attributes;
            Elements = elements;
            Value = string.Empty;
            Concat = false;
            Separator = ";";
        }

        public string Name { get; set; }
        public List<XAttribute> Attributes { get; set; }
        public string Value { get; set; }
        public bool Concat { get; set; }
        public string Separator { get; set; }
        public List<XElement> Elements { get; set; }

        public XElement Copy()
        {
            XElement c = new XElement(Name, new List<XElement>(), new List<XAttribute>());
            c.Value = Value;
            c.Concat = Concat;
            c.Separator = Separator;
            foreach (XAttribute a in Attributes)
                c.Attributes.Add(a.Copy());
            foreach (XElement e in Elements)
                c.Elements.Add(e.Copy());
            return c;
        }

        public void Read(XmlNode node)
        {
            if (node.Attributes != null && node.Attributes.Count > 0)
            {
                foreach (XmlAttribute a in node.Attributes)
                {
                    if (a.Name == "Concat")
                    {
                        Concat = String.Compare(a.Value, "true", true) == 0 ? true : false;
                    }
                    else if (a.Name == "Separator")
                    {
                        Separator = a.Value;
                    }
                    else
                    {
                        // A real attribute
                        Attributes.Add(new XAttribute(a.Name, a.Value));
                    }
                }
            }

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;
                if (child.NodeType == XmlNodeType.Text)
                {
                    Value = child.Value;
                    continue;
                }

                XElement e = new XElement(child.Name, new List<XElement>(), new List<XAttribute>());
                Elements.Add(e);
                e.Read(child);
            }
        }

        public static string sGetXmlNodeValueAsText(XmlNode node)
        {
            if (!String.IsNullOrEmpty(node.Value))
            {
                return node.Value;
            }
            else if (node.HasChildNodes && node.FirstChild.NodeType == XmlNodeType.Text)
            {
                return node.FirstChild.Value;
            }
            return string.Empty;
        }

        public override string ToString()
        {
            string str;
            if (this.Attributes.Count == 0)
            {
                if (String.IsNullOrEmpty(this.Value))
                    str = String.Format("<{0} />", this.Name);
                else
                    str = String.Format("<{0}>{1}</{0}>", this.Name, this.Value);
            }
            else
            {
                string attributes = string.Empty;
                foreach (XAttribute a in this.Attributes)
                {
                    string attribute = a.Name + "=\"" + a.Value + "\"";
                    if (String.IsNullOrEmpty(attributes))
                        attributes = attribute;
                    else
                        attributes = attributes + " " + attribute;
                }
                if (String.IsNullOrEmpty(this.Value))
                    str = String.Format("<{0} {2} />", this.Name, this.Value, attributes);
                else
                    str = String.Format("<{0} {2}>{1}</{0}>", this.Name, this.Value, attributes);
            }
            return str;
        }
    }

}