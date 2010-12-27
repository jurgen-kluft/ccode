using System;
using System.Text;
using System.Xml;

namespace MSBuild.XCode
{
    public class Attribute
    {
        public Attribute(string name, string value)
        {
            Name = name;
            Value = value;
        }

        public string Name { get; set; }
        public string Value { get; set; }

        public Attribute Copy()
        {
            Attribute a = new Attribute(Name, Value);
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