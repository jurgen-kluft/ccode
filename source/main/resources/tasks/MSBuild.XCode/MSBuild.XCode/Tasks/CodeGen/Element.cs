using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace MSBuild.XCode
{
    public class Element
    {
        private Element()
        {

        }

        public Element(string name, List<Element> elements, List<Attribute> attributes)
        {
            Name = name;
            Attributes = attributes;
            Elements = elements;
            Value = string.Empty;
            Concat = false;
            Separator = ";";
            IsGroup = false;
        }

        public string Name { get; set; }
        public List<Attribute> Attributes { get; set; }
        public string Value { get; set; }
        public bool Concat { get; set; }
        public string Separator { get; set; }
        public bool IsGroup { get; set; }
        public List<Element> Elements { get; set; }

        public Element Copy()
        {
            Element c = new Element(Name, new List<Element>(), new List<Attribute>());
            c.Value = Value;
            c.Concat = Concat;
            c.Separator = Separator;
            foreach (Attribute a in Attributes)
                c.Attributes.Add(a.Copy());
            c.IsGroup = IsGroup;
            foreach (Element e in Elements)
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
                        Attributes.Add(new Attribute(a.Name, a.Value));
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

                Element e = new Element(child.Name, new List<Element>(), new List<Attribute>());
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
                foreach (Attribute a in this.Attributes)
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