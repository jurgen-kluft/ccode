﻿using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageContent
    {
        class ContentItem
        {
            private ContentItem()
            {
                Required = false;
            }

            public string Src { get; private set; }
            public string Dst { get; private set; }
            public string Platform { get; private set; }
            public bool Required { get; private set; }

            public static ContentItem Read(XmlNode node)
            {
                ContentItem item = new ContentItem();

                item.Platform = Attribute.Get("Platform", node, "*");
                item.Required = Boolean.Parse(Attribute.Get("Required", node, "false"));

                item.Src = Attribute.Get("Src", node, null);
                if (item.Src != null)
                {
                    item.Dst = Attribute.Get("Dst", node, null);
                    if (item.Dst != null)
                    {
                        return item;
                    }
                }
                return null;
            }
        }

        private Dictionary<string, List<ContentItem>> mContent;

        public PackageContent()
        {
            mContent = new Dictionary<string, List<ContentItem>>();
        }


        public bool Collect(string name, string platform, PackageVars vars, string rootDir, out Dictionary<string, string> outFiles)
        {
            outFiles = new Dictionary<string, string>();

            List<string> platforms = new List<string>();
            platforms.Add("*");
            platforms.Add(platform);

            vars.Add("Name", name);
            vars.Add("Platform", platform);
            vars.Add("ToolSet", vars.GetToolSet(platform));

            foreach (string p in platforms)
            {
                List<ContentItem> content;
                if (mContent.TryGetValue(p, out content))
                {
                    foreach (ContentItem item in content)
                    {
                        string src = rootDir + item.Src;
                        src = vars.ReplaceVars(src);
                        string dst = item.Dst;
                        dst = vars.ReplaceVars(dst);

                        int m = outFiles.Count;
                        Glob(src, dst, outFiles);
                        int n = outFiles.Count - m;

                        if (n == 0 && item.Required)
                        {
                            Loggy.Error(String.Format("PackageContent::Collect, error; required file {0} does not exist", src));
                        }
                    }
                }
            }
            return true;
        }

        private static void Glob(string src, string dst, Dictionary<string, string> files)
        {
            List<string> globbedFiles = PathUtil.getFiles(src);

            int r = src.IndexOf("**");
            string reldir = r >= 0 ? src.Substring(0, src.IndexOf("**")) : string.Empty;

            foreach (string src_filename in globbedFiles)
            {
                string dst_filename;
                if (r >= 0)
                    dst_filename = dst + src_filename.Substring(reldir.Length);
                else
                    dst_filename = dst + Path.GetFileName(src_filename);

                if (!files.ContainsKey(src_filename))
                    files.Add(src_filename, Path.GetDirectoryName(dst_filename));
            }
        }

        public bool Read(XmlNode node, PackageVars vars)
        {
            if (node.Name == "Content")
            {
                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        ContentItem item = ContentItem.Read(child);
                        if (item != null)
                        {
                            List<ContentItem> items;
                            if (!mContent.TryGetValue(item.Platform, out items))
                            {
                                items = new List<ContentItem>();
                                mContent.Add(item.Platform, items);
                            }
                            items.Add(item);
                        }
                    }
                }
                return true;
            }
            return false;
        }
    }
}