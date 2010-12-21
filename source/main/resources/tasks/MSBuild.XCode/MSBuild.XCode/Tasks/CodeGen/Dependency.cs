using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class Dependency
    {
        private Dictionary<string, string> mPlatformBranch;
        private Dictionary<string, VersionRange> mPlatformBranchVersions;

        public Dependency()
        {
            Group = new Group("com.virtuos.tnt");
            Type = "Package";
            mPlatformBranch = new Dictionary<string, string>();
            mPlatformBranchVersions = new Dictionary<string, VersionRange>();
        }

        public string Name { get; set; }
        public Group Group { get; set; }
        public string Type { get; set; }

        public void Info()
        {
            Logger.Add(String.Format("Dependency                 : {0}", Name));
            Logger.Add(String.Format("Group                      : {0}", Group.ToString()));
            Logger.Add(String.Format("Type                       : {0}", Type));

            bool first = true;
            foreach (KeyValuePair<string, VersionRange> pair in mPlatformBranchVersions)
            {
                if (first)
                    Logger.Add(String.Format("Versions[]                 : {0} = {1}", pair.Key, pair.Value.ToString()));
                else
                    Logger.Add(String.Format("                             {0} = {1}", pair.Key, pair.Value.ToString()));
                first = false;
            }
        }

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

        private delegate VersionRange ReturnVersionRangeDelegate();

        private VersionRange GetVersionRange(string platform, ReturnVersionRangeDelegate returnDefaultVersionRangeDelegate)
        {
            string branch = GetBranch(platform);
            VersionRange versionRange;
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

        public VersionRange GetVersionRange(string platform)
        {
            return GetVersionRange(platform, delegate() { return new VersionRange("[1.0,)"); } );
        }

        public bool IsEqual(Dependency dependency)
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
                                    VersionRange a = GetVersionRange(ap, delegate() { return null; });
                                    VersionRange b = dependency.GetVersionRange(ap, delegate() { return null; });
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
        public bool Merge(Dependency dependency)
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
                VersionRange thisRange = GetVersionRange(Platform_Branch.Key);
                VersionRange thatRange = dependency.GetVersionRange(Platform_Branch.Key);
                if (thisRange.Merge(thatRange))
                    modified = true;
            }
            return modified;
        }

        public void Read(XmlNode node)
        {
            if (node.Name == "Dependency")
            {
                Name = Attribute.Get("Package", node, "Unknown");

                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        if (child.NodeType == XmlNodeType.Comment)
                            continue;

                        if (child.Name == "Group")
                        {
                            Group.Full = Element.sGetXmlNodeValueAsText(child);
                        }
                        else if (child.Name == "Version")
                        {
                            string platform = Attribute.Get("Platform", child, "*").ToLower();
                            string branch = Attribute.Get("Branch", child, "default").ToLower();
                            if (branch == "*")
                                branch = "default";
                            VersionRange versionRange = new VersionRange(Element.sGetXmlNodeValueAsText(child));

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
                            Type = Element.sGetXmlNodeValueAsText(child);
                        }
                    }
                }
            }
        }
   }
}