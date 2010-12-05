using System;
using System.Collections.Generic;

namespace MSBuild.Cod
{
    public class XPlatform
    {
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();
        protected Dictionary<string, XConfig> mConfigs = new Dictionary<string, XConfig>();

        public string Name { get; set; }

        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
        public Dictionary<string, XConfig> configs { get { return mConfigs; } }

        public void Initialize(string p, string[] groups)
        {
            Name = p;
            foreach (string g in groups)
            {
                mGroups.Add(g, new List<XElement>());
            }
        }
    }

}
