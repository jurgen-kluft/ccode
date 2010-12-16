using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XDependency
    {
        private Dictionary<string, string> mPlatformBranch;
        private Dictionary<string, XVersionRange> mPlatformBranchVersions;

        public XDependency()
        {
            Group = new XGroup("com.virtuos.tnt");
            Type = "Package";
            mPlatformBranch = new Dictionary<string, string>();
            mPlatformBranchVersions = new Dictionary<string, XVersionRange>();
        }

        public string Name { get; set; }
        public XGroup Group { get; set; }
        public string Type { get; set; }

        private string GetBranch(string platform, string defaultBranch)
        {
            string branch;
            if (!mPlatformBranch.TryGetValue(platform.ToLower(), out branch))
            {
                return defaultBranch;
            }
            else
            {
                return branch;
            }
        }

        public string GetBranch(string platform)
        {
            return GetBranch(platform, "default");
        }

        private delegate XVersionRange ReturnVersionRangeDelegate();

        private XVersionRange GetVersionRange(string platform, ReturnVersionRangeDelegate returnDefaultVersionRangeDelegate)
        {
            string branch = GetBranch(platform);
            XVersionRange versionRange;
            string platformBranch = (platform.ToLower() + "|" + branch);
            if (mPlatformBranchVersions.TryGetValue(platformBranch, out versionRange))
                return versionRange;

            if (platform != "*")
            {
                platform = "*";
                branch = GetBranch(platform);
                platformBranch = (platform + "|" + branch);
                if (mPlatformBranchVersions.TryGetValue(platformBranch, out versionRange))
                    return versionRange;
            }

            // By default return x >= 1.0
            return returnDefaultVersionRangeDelegate();
        }

        public XVersionRange GetVersionRange(string platform)
        {
            return GetVersionRange(platform, delegate() { return new XVersionRange("[1.0,)"); } );
        }

        public bool IsEqual(XDependency dependency)
        {
            if (String.Compare(Name, dependency.Name, true)==0)
            {
                if (String.Compare(Group.Full, dependency.Group.Full, true) == 0)
                {
                    if (String.Compare(Type, dependency.Type, true) == 0)
                    {
                        if ((mPlatformBranch != null && dependency.mPlatformBranch != null) && mPlatformBranch.Count==dependency.mPlatformBranch.Count)
                        {
                            // Check content
                            foreach (string ap in mPlatformBranch.Keys)
                            {
                                string ab = GetBranch(ap, "a");
                                string bb = dependency.GetBranch(ap, "b");
                                if (String.Compare(ab, bb, true) != 0)
                                    return false;
                            }
                            foreach (string ap in dependency.mPlatformBranch.Keys)
                            {
                                string ab = GetBranch(ap, "a");
                                string bb = dependency.GetBranch(ap, "b");
                                if (String.Compare(ab, bb, true) != 0)
                                    return false;
                            }
                            if ((mPlatformBranchVersions != null && dependency.mPlatformBranchVersions != null) && mPlatformBranchVersions.Count == dependency.mPlatformBranchVersions.Count)
                            {
                                foreach (string ap in mPlatformBranch.Keys)
                                {
                                    XVersionRange a = GetVersionRange(ap, delegate() { return null; });
                                    XVersionRange b = dependency.GetVersionRange(ap, delegate() { return null; });
                                    if (a != b)
                                        return false;
                                }
                            }
                        }
                    }
                }
            }
            return false;
        }

        // Merge with same package dependency
        // Return True when merge resulted in an updated dependency (A change in XVersionRange)
        public bool Merge(XDependency dependency)
        {
            bool modified = false;
            if (String.Compare(Name, dependency.Name, true) != 0)
                return modified;

            // Merge the type
            if (String.Compare(Type, dependency.Type, true) != 0)
            {
                // Currently there are only 2 types, Package and Source
                if (String.Compare(Type, "Package", true) != 0)
                {
                    Type = "Package";
                    modified = true;
                }
            }

            // Merge the version range
            foreach (KeyValuePair<string,string> Platform_Branch in mPlatformBranch)
            {
                XVersionRange thisRange = GetVersionRange(Platform_Branch.Key);
                XVersionRange thatRange = dependency.GetVersionRange(Platform_Branch.Key);
                if (thisRange.Merge(thatRange))
                    modified = true;
            }
            return modified;
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
                            Group.Full = XElement.sGetXmlNodeValueAsText(child);
                        }
                        else if (child.Name == "Version")
                        {
                            string platform = XAttribute.Get("Platform", child, "*").ToLower();
                            string branch = XAttribute.Get("Branch", child, "default").ToLower();
                            if (branch == "*")
                                branch = "default";
                            XVersionRange versionRange = new XVersionRange(XElement.sGetXmlNodeValueAsText(child));

                            if (mPlatformBranch.ContainsKey(platform))
                                mPlatformBranch.Remove(platform);
                            mPlatformBranch.Add(platform, branch);

                            string platformBranch = (platform + "|" + branch);
                            if (mPlatformBranchVersions.ContainsKey(platformBranch))
                                mPlatformBranchVersions.Remove(platformBranch);
                            mPlatformBranchVersions.Add(platformBranch, versionRange);
                        }
                        else if (child.Name == "Type")
                        {
                            Type = XElement.sGetXmlNodeValueAsText(child);
                        }
                    }
                }
            }
        }
   }
}