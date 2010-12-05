using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;

namespace MSBuild.XCode
{
    public class XConfig
    {
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();

        public string Name { get; set; }
        public string Config { get; set; }
        public string Platform { get; set; }

        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }

        public void Initialize(string p, string c, string[] groups)
        {
            Name = c + "|" + p;
            Platform = p;
            Config = c;

            foreach (string g in groups)
            {
                mGroups.Add(g, new List<XElement>());
            }
        }
    }
}