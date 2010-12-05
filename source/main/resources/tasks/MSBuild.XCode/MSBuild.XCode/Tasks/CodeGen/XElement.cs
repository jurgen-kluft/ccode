using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace MSBuild.Cod
{
    public class XElement
    {
        public XElement(string name, List<XElement> elements, List<XAttribute> attributes)
        {
            Name = name;
            Attributes = attributes;
            Elements = elements;
        }

        public string Name { get; set; }
        public List<XAttribute> Attributes { get; set; }
        public string Value { get; set; }
        public bool Concat { get; set; }
        public string Separator { get; set; }
        public List<XElement> Elements { get; set; }

        public void Init()
        {
            Name = string.Empty;
            Attributes = new List<XAttribute>();
            Value = string.Empty;
            Concat = false;
            Separator = ";";
            Elements = new List<XElement>();
        }

        public XElement Copy()
        {
            XElement c = new XElement(Name, new List<XElement>(), new List<XAttribute>());
            c.Value = Value;
            foreach (XAttribute a in Attributes)
                c.Attributes.Add(a.Copy());
            foreach (XElement e in Elements)
                c.Elements.Add(e.Copy());
            return c;
        }
    }

}