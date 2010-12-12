using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XDependency
    {
        private Dictionary<string, Version> mPlatformVersions;
        private Dictionary<string, Version> mPlatformBranchVersions;

        public XDependency()
        {
            Group = new XGroup("com.virtuos.tnt");
            Type = "Package";
            mPlatformVersions = new Dictionary<string, Version>();
            mPlatformBranchVersions = new Dictionary<string, Version>();
        }

        public class Version
        {
            public string Platform { get; set; }
            public string Branch { get; set; }
            public XVersionRange VersionRange { get; set; }
        }

        public string Name { get; set; }
        public XGroup Group { get; set; }
        public string Type { get; set; }

        public XVersionRange GetVersionRange(string platform, string branch)
        {
            if (String.IsNullOrEmpty(branch))
            {
                Version v;
                if (mPlatformVersions.TryGetValue(platform.ToLower(), out v))
                {
                    return v.VersionRange;
                }
                if (mPlatformVersions.TryGetValue("All".ToLower(), out v))
                {
                    return v.VersionRange;
                }
            }
            else
            {
                Version v;
                string platformBranch = (platform + "|" + branch).ToLower();
                if (mPlatformBranchVersions.TryGetValue(platformBranch, out v))
                {
                    return v.VersionRange;
                }
                platformBranch = ("All" + "|" + branch).ToLower();
                if (mPlatformBranchVersions.TryGetValue(platformBranch, out v))
                {
                    return v.VersionRange;
                }
            }

            // By default return x >= 1.0
            return new XVersionRange("[1.0,)");
        }

        public void Read(XmlNode node)
        {
            if (node.Name == "Dependency")
            {
                Name = XAttribute.Get("Package", node, "Unknown");

                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        if (child.NodeType == XmlNodeType.Comment)
                            continue;

                        if (child.Name == "Group")
                        {
                            Group.Group = XElement.sGetXmlNodeValueAsText(child);
                        }
                        else if (child.Name == "Version")
                        {
                            Version v = new Version();
                            v.Platform = XAttribute.Get("Platform", child, "All");
                            v.Branch = XAttribute.Get("Branch", child, "default");
                            v.VersionRange = new XVersionRange(XElement.sGetXmlNodeValueAsText(child));

                            string platform = v.Platform.ToLower();
                            if (mPlatformVersions.ContainsKey(platform))
                                mPlatformVersions.Remove(platform);
                            mPlatformVersions.Add(platform, v);

                            string platformBranch = (v.Platform + "|" + v.Branch).ToLower();
                            if (mPlatformBranchVersions.ContainsKey(platformBranch))
                                mPlatformBranchVersions.Remove(platformBranch);
                            mPlatformBranchVersions.Add(platformBranch, v);
                        }
                        else if (child.Name == "Type")
                        {
                            Type = XElement.sGetXmlNodeValueAsText(child);
                        }
                    }
                }
            }
        }

        public void Sync(string remote_repo, string local_repo, string path, string[] platforms)
        {

        }
   }
}