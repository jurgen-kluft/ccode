using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageVars
    {
        private Dictionary<string, string> mVars;

        public PackageVars()
        {
            mVars = new Dictionary<string, string>();
        }

        public string ReplaceVars(string str)
        {
            foreach (KeyValuePair<string, string> var in mVars)
                str = str.Replace(String.Format("${{{0}}}", var.Key), var.Value);
            return str;
        }

        public void Add(string name, string value)
        {
            if (String.IsNullOrEmpty(value))
                return;

            if (!mVars.ContainsKey(name))
                mVars.Add(name, value);
        }

        public bool Read(XmlNode node)
        {
            if (node.Name == "Variables")
            {
                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        string text = Element.GetText(child);
                        Add(child.Name, text);
                    }
                } return true;
            }
            return false;
        }
    }
}