using System;
using System.Text;
using System.Xml;

namespace MSBuild.XCode
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

        public static string Get(string attrName, XmlNode node, string _default)
        {
            string v = _default;
            if (node.Attributes != null)
            {
                foreach (XmlAttribute a in node.Attributes)
                {
                    if (a.Name == attrName)
                    {
                        v = a.Value;
                        break;
                    }
                }
            }
            return v;
        }
    }
}