using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    ///
    /// A collection of versions, where a version falls under a platform-branch
    /// 
    public class Versions
    {
        private Dictionary<String, ComparableVersion> mPlatformBranchSpecificVersions;

        private string BuildTag(string platform, string branch)
        {
            if (String.IsNullOrEmpty(platform) || String.Compare(platform, "all", true)==0)
                platform = "*";
            else
                platform = platform.ToLower();

            if (String.IsNullOrEmpty(branch) || String.Compare(branch, "default", true)==0)
                branch = "*";
            else
                branch = branch.ToLower();

            return platform + "|" + branch;
        }

        public Versions()
        {
            mPlatformBranchSpecificVersions = new Dictionary<string, ComparableVersion>();
        }

        public void Info()
        {
            bool first = true;
            foreach (KeyValuePair<string, ComparableVersion> pair in mPlatformBranchSpecificVersions)
            {
                if (first)
                    Loggy.Add(String.Format("Versions[]                 : {0}={1}", pair.Key, pair.Value.ToString()));
                else
                    Loggy.Add(String.Format("                             {0}={1}", pair.Key, pair.Value.ToString()));
                first = false;
            }
        }

        public void Add(string platform, ComparableVersion item)
        {
            Add(platform, null, item);
        }

        public void Add(string platform, string branch, ComparableVersion item)
        {
            if (!Contains(platform, branch))
            {
                mPlatformBranchSpecificVersions.Add(BuildTag(platform, branch), item);
            }
        }

        public void Clear()
        {
            mPlatformBranchSpecificVersions.Clear();
        }

        public bool Contains(string platform)
        {
            return mPlatformBranchSpecificVersions.ContainsKey(BuildTag(platform, null));
        }
        public bool Contains(string platform, string branch)
        {
            return mPlatformBranchSpecificVersions.ContainsKey(BuildTag(platform, branch));
        }

        public ComparableVersion GetForPlatform(string platform)
        {
            return GetForPlatformWithBranch(platform, null);
        }

        public ComparableVersion GetForPlatformWithBranch(string platform, string branch)
        {
            ComparableVersion version;
            string tag = BuildTag(platform, branch);
            if (mPlatformBranchSpecificVersions.TryGetValue(tag, out version))
            {
                return version;
            }
            else if (!tag.StartsWith("*|"))
            {
                tag = BuildTag(null, branch);
                if (mPlatformBranchSpecificVersions.TryGetValue(tag, out version))
                    return version;

                tag = BuildTag(null, null);
                if (mPlatformBranchSpecificVersions.TryGetValue(tag, out version))
                    return version;
            }
            return null;
        }

        public int Count
        {
            get
            {
                return mPlatformBranchSpecificVersions.Keys.Count;
            }
        }

        public void Read(XmlNode node)
        {
            if (!node.HasChildNodes)
                return;

            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.NodeType == XmlNodeType.Comment)
                    continue;

                string v = Element.sGetXmlNodeValueAsText(child);
                v = (String.IsNullOrEmpty(v)) ? "1.0.0" : v;
                string platform = Attribute.Get("Platform", child, "*");
                string branch = Attribute.Get("Branch", child, "*");
                Add(platform, branch, new ComparableVersion(v));
            }
        }
    }

}