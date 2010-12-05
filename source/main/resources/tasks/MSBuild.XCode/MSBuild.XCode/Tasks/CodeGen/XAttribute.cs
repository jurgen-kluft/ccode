using System;
using System.Text;
using System.Xml;

namespace MSBuild.Cod
{
    public class XAttribute
    {
        public XAttribute(string name, string value)
        {
            Name = name;
            Value = value;
        }

        public string Name { get; set; }
        public string Value { get; set; }

        public XAttribute Copy()
        {
            XAttribute a = new XAttribute(Name, Value);
            return a;
        }
    }
}