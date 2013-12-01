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
        private Dictionary<string, List<string>> mPerPlatformToolSets;
        private Dictionary<string, string> mVars;

        public PackageVars()
        {
            Clear();
            Init();
        }

        public void Clear()
        {
            mPerPlatformToolSets = new Dictionary<string, List<string>>();
            mVars = new Dictionary<string, string>();
        }

        private void Init()
        {
            SetToolSet("Win32", "v100", false);
            SetToolSet("Win32", "v110", true);

            SetToolSet("x64", "v100", false);
            SetToolSet("x64", "v110", true);

            SetToolSet("PS3", "GCC", false);
            SetToolSet("PS3", "SNC", true);

            SetToolSet("PS3SPU", "SNC", true);


            //@TODO: Add other platforms
        }

        public string ReplaceVars(string str)
        {
            foreach (KeyValuePair<string, string> var in mVars)
            {
                string occurence = String.Format("${{{0}}}", var.Key);
                while (str.Contains(occurence))
                    str = str.Replace(occurence, var.Value);
            }
            return str;
        }

        public string ReplaceVars(string platform, string str)
        {
            string toolset = GetToolSet(platform);
            if (!String.IsNullOrEmpty(toolset))
            {
                string occurence = String.Format("${{{0}}}", "ToolSet");
                while (str.Contains(occurence))
                    str = str.Replace(occurence, toolset);
            }

            foreach (KeyValuePair<string, string> var in mVars)
            {
                string occurence = String.Format("${{{0}}}", var.Key);
                while (str.Contains(occurence))
                    str = str.Replace(occurence, var.Value);
            }

            return str;
        }
        public void Add(string name, string value)
        {
            if (String.IsNullOrEmpty(value))
                return;

            // If we already have this key, remove it and add the incoming key-value
            if (mVars.ContainsKey(name))
                mVars.Remove(name);
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


        public string GetToolSet(string platform)
        {
            string toolset = string.Empty;
            List<string> toolsets;
            if (mPerPlatformToolSets.TryGetValue(platform, out toolsets))
            {
                // Always take the top one from the list
                if (toolsets.Count > 0)
                    toolset = toolsets[0];
            }
            return toolset;
        }

        public void SetToolSet(string platform, string toolset, bool setfirst)
        {
            if (String.IsNullOrEmpty(toolset))
                return;

            platform = platform.Replace(" ", "");
            platform = platform.Replace("_", "");

            Add(platform + "ToolSet", toolset);

            if (!mPerPlatformToolSets.ContainsKey(platform))
                mPerPlatformToolSets.Add(platform, new List<string>());

            List<string> toolsets;
            mPerPlatformToolSets.TryGetValue(platform, out toolsets);
            if (toolsets.Contains(toolset))
                toolsets.Remove(toolset);
            if (setfirst)
                toolsets.Insert(0, toolset);
            else
                toolsets.Add(toolset);
        }
    }
}