using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class ProjectProperties
    {
        public class Item
        {
            private string mDependencyType;
            private string mPlatform;
            private List<string> mContent;

            public Item(string depType, string platform, List<string> content)
            {
                mDependencyType = depType;
                mPlatform = platform;
                mContent = content;
            }

            public string DependencyType { get { return mDependencyType; } }

            public void ExpandVars(PackageVars vars)
            {
                for (int i = 0; i < mContent.Count; ++i)
                {
                    string str = mContent[i];
                    str = vars.ReplaceVars(str);
                    mContent[i] = str;
                }
            }

            public void write(string propsFilePath, string location)
            {
                // <?xml version="1.0" encoding="utf-8"?>
                // <Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
                //     mXmlDoc
                // </Project>
                //
                string path = Path.GetDirectoryName(propsFilePath);
                if (!Directory.Exists(path))
                    Directory.CreateDirectory(path);

                using (FileStream stream = new FileStream(propsFilePath, FileMode.Create, FileAccess.Write))
                {
                    using (StreamWriter writer = new StreamWriter(stream))
                    {
                        writer.WriteLine("<?xml version=\"1.0\" encoding=\"utf-8\"?>");
                        writer.WriteLine("<Project ToolsVersion=\"4.0\" xmlns=\"http://schemas.microsoft.com/developer/msbuild/2003\">");
                        foreach (string l in mContent)
                        {
                            string ml = l.Replace("${Location}", location);
                            writer.WriteLine("    {0}", ml);
                        }
                        writer.WriteLine("</Project>");
                        writer.Close();
                    }
                    stream.Close();
                }
            }
        }

        private Dictionary<string, Dictionary<string, Item>> mProperties;

        public ProjectProperties()
        {
            mProperties = new Dictionary<string, Dictionary<string, Item>>();
        }

        private Item GetItemFor(string platform, string dependencyType)
        {
            Dictionary<string, Item> items;
            if (mProperties.TryGetValue(platform, out items))
            {
                Item existingItem;
                if (items.TryGetValue(dependencyType.ToLower(), out existingItem))
                {
                    return existingItem;
                }
            }
            else if (mProperties.TryGetValue("*", out items))
            {
                Item existingItem;
                if (items.TryGetValue(dependencyType.ToLower(), out existingItem))
                {
                    return existingItem;
                }
            }
            return null;
        }

        public void ExpandVars(PackageVars vars)
        {
            // Replace on all items that we hold
            foreach (KeyValuePair<string, Dictionary<string, Item>> items in mProperties)
            {
                foreach (KeyValuePair<string, Item> item in items.Value)
                {
                    item.Value.ExpandVars(vars);
                }
            }
        }

        public bool Write(string propsFilePath, string platform, string dependencyType, string location)
        {
            Item item = GetItemFor(platform, dependencyType);
            if (item != null)
            {
                item.write(propsFilePath, location);
                return true;
            }
            return false;
        }

        public bool WriteAll(string propsFilePath, string[] platforms, string dependencyType, string location)
        {
            foreach(string platform in platforms)
            {
                if (Write(propsFilePath, platform, dependencyType, location) == false)
                    return false;
            }
            return true;
        }

        public void SetDefault(string name)
        {
            List<string> rootContent = new List<string>();
            rootContent.Add("<PropertyGroup Label=\"${Name}_TargetDirs\">");
            rootContent.Add("    <${Name}_RootDir>$(SolutionDir)</${Name}_RootDir>");
            rootContent.Add("    <${Name}_TargetDir>$(SolutionDir)</${Name}_TargetDir>");
            rootContent.Add("    <${Name}_LibraryDir>$(SolutionDir)target\\${Name}\\outdir\\${Name}_$(Configuration)_$(Platform)\\</${Name}_LibraryDir>");
            rootContent.Add("    <${Name}_IncludeDir>$(SolutionDir)source\\main\\include\\</${Name}_IncludeDir>");
            rootContent.Add("</PropertyGroup>");

            List<string> packageContent = new List<string>();
            packageContent.Add("<PropertyGroup Label=\"${Name}_TargetDirs\">");
            packageContent.Add("    <${Name}_RootDir>${Location}</${Name}_RootDir>");
            packageContent.Add("    <${Name}_TargetDir>${Location}</${Name}_TargetDir>");
            packageContent.Add("    <${Name}_LibraryDir>${Location}libs\\</${Name}_LibraryDir>");
            packageContent.Add("    <${Name}_IncludeDir>${Location}source\\main\\include\\</${Name}_IncludeDir>");
            packageContent.Add("</PropertyGroup>");

            List<string> sourceContent = new List<string>();
            sourceContent.Add("<PropertyGroup Label=\"${Name}_TargetDirs\">");
            sourceContent.Add("    <${Name}_RootDir>$(SolutionDir)target\\${Name}\\</${Name}_RootDir>");
            sourceContent.Add("    <${Name}_TargetDir>$(SolutionDir)</${Name}_TargetDir>");
            sourceContent.Add("    <${Name}_LibraryDir>$(SolutionDir)target\\${Name}\\target\\${Name}\\outdir\\${Name}_$(Configuration)_$(Platform)\\</${Name}_LibraryDir>");
            sourceContent.Add("    <${Name}_IncludeDir>$(SolutionDir)target\\${Name}\\source\\main\\include\\</${Name}_IncludeDir>");
            sourceContent.Add("</PropertyGroup>");

            Dictionary<string, Item> items = new Dictionary<string, Item>();
            Item item1 = new Item("Root", "*", rootContent);
            items.Add(item1.DependencyType.ToLower(), item1);
            Item item2 = new Item("Package", "*", packageContent);
            items.Add(item2.DependencyType.ToLower(), item2);
            Item item3 = new Item("Source", "*", sourceContent);
            items.Add(item3.DependencyType.ToLower(), item3);

            mProperties.Clear();
            mProperties.Add("*", items);
        }

        public bool Read(XmlNode node)
        {
            foreach (XmlNode child in node.ChildNodes)
            {
                if (child.Name == "Properties")
                {
                    string dependencyType = Attribute.Get("DependencyType", child, "Package");
                    string platform = Attribute.Get("Platform", child, "*");
                    string content = child.InnerXml;
                    string[] contentLines = content.Split(new string[] { Environment.NewLine }, StringSplitOptions.RemoveEmptyEntries);

                    Item item = new Item(dependencyType, platform, new List<string>(contentLines));

                    Dictionary<string, Item> items;
                    if (mProperties.TryGetValue(platform, out items))
                    {
                        Item existingItem;
                        if (items.TryGetValue(dependencyType.ToLower(), out existingItem))
                        {
                            // Already exists!
                        }
                        else
                        {
                            items.Add(dependencyType.ToLower(), item);
                        }
                    }
                    else
                    {
                        items = new Dictionary<string, Item>();
                        items.Add(dependencyType.ToLower(), item);
                        mProperties.Add(platform, items);
                    }
                }
            }
            return true;
        }
    }
}