using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyResource
    {
        private Dictionary<string, string> mPlatformBranch;
        private Dictionary<string, VersionRange> mPlatformBranchVersions;

        private string mName;
        private Group mGroup;
        private string mType;

        public DependencyResource()
        {
            mName = "?";
            mGroup = new Group("com.virtuos.tnt");
            mType = "Package";
            mPlatformBranch = new Dictionary<string, string>();
            mPlatformBranchVersions = new Dictionary<string, VersionRange>();
        }

        public string Name { get { return mName; } }
        public Group Group { get { return mGroup; } }
        public string Type { get { return mType; } }

        private string ReplaceVars(string str, Dictionary<string,string> vars)
        {
            foreach (KeyValuePair<string, string> var in vars)
                str = str.Replace(String.Format("${{{0}}}", var.Key), var.Value);
            return str;
        }

        public void ExpandVars(Dictionary<string, string> vars)
        {
            mName = ReplaceVars(mName, vars);
            mGroup.ExpandVars(vars);
            mType = ReplaceVars(mType, vars);
        }

        public void Info()
        {
            Loggy.Add(String.Format("Dependency                 : {0}", Name));
            Loggy.Add(String.Format("Group                      : {0}", Group.ToString()));
            Loggy.Add(String.Format("Type                       : {0}", Type));

            bool first = true;
            foreach (KeyValuePair<string, VersionRange> pair in mPlatformBranchVersions)
            {
                if (first)
                    Loggy.Add(String.Format("Versions[]                 : {0} = {1}", pair.Key, pair.Value.ToString()));
                else
                    Loggy.Add(String.Format("                             {0} = {1}", pair.Key, pair.Value.ToString()));
                first = false;
            }
        }

        public string GetBranch(string platform, string defaultBranch)
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

        public delegate VersionRange ReturnVersionRangeDelegate();

        public VersionRange GetVersionRange(string platform, ReturnVersionRangeDelegate returnDefaultVersionRangeDelegate)
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

        public bool IsEqual(DependencyResource dependency)
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

        public void Read(XmlNode node)
        {
            if (node.Name == "Dependency")
            {
                mName = Attribute.Get("Package", node, "Unknown");

                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        if (child.NodeType == XmlNodeType.Comment)
                            continue;

                        if (child.Name == "Group")
                        {
                            mGroup.Full = Element.GetText(child);
                        }
                        else if (child.Name == "Version")
                        {
                            string platform = Attribute.Get("Platform", child, "*").ToLower();
                            string branch = Attribute.Get("Branch", child, "default").ToLower();
                            if (branch == "*")
                                branch = "default";

                            VersionRange versionRange = new VersionRange(Element.GetText(child));

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
                            mType = Element.GetText(child);
                        }
                    }
                }
            }
        }
   }
}