using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;

namespace msbuild.xmaven
{
    public class XProject
    {
        protected Dictionary<string, XElement> mElements = new Dictionary<string, XElement>();
        protected Dictionary<string, List<XElement>> mGroups = new Dictionary<string, List<XElement>>();
        protected Dictionary<string, XPlatform> mPlatforms = new Dictionary<string, XPlatform>();

        public Dictionary<string, XElement> elements { get { return mElements; } }
        public Dictionary<string, List<XElement>> groups { get { return mGroups; } }
        public Dictionary<string, XPlatform> platforms { get { return mPlatforms; } }

        public void Initialize(string[] groups)
        {
            mElements = new Dictionary<string, XElement>();

            foreach (string g in groups)
            {
                mGroups.Add(g, new List<XElement>());
            }
        }
    }

}