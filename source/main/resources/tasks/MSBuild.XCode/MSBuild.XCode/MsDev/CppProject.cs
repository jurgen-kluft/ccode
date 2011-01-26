using System;
using System.IO;
using System.Xml;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode.MsDev
{
    /// <summary>
    /// What if we change the whole approach from being template oriented to being able to
    /// load, merge and save .vcxproj.
    /// The advantage is that we can 'update' .vcxprojects as well as generate them and it is 
    /// easier to make .csprojects working.
    /// What we need for this is to be able to understand the layout of a .vcxproj. So lets
    /// assume we can load it using XmlDocument, now we need a way to find an element under
    /// a certain configuration, the elements we need to concat/replace are:
    ///    1) PreprocessorDefinitions (Default=Concat)
    ///    2) AdditionalIncludeDirectories (Default=Concat)
    ///    3) AdditionalDependencies (Default=Concat)
    ///    4) AdditionalLibraryDirectories (Default=Concat)
    /// 
    /// If we use a .vcxproj as template then merging means also to be able to 'replace' an
    /// element. 
    /// 
    /// When generating the .vcxproj from the template the main file should exist but empty.
    /// The template may contain replaceable patterns, using these:
    /// 1) ${Name}
    /// 2) ${GUID}
    /// 
    /// </summary>
    public class CppProject : BaseProject, IProject
    {
        private readonly static string[] mMergeItems = new string[]
        {
            "IncludePath",
            "LibraryPath",
            "ExecutablePath",
            "PreprocessorDefinitions", 
            "AdditionalDependencies", 
            "AdditionalLibraryDirectories", 
            "AdditionalIncludeDirectories" 
        };
        
        private readonly static string[] mContentItems = new string[]
        {
            "ClCompile", 
            "ClInclude", 
            "None" 
        };

        public CppProject()
        {
            mXmlDocMain = new XmlDocument();
        }

        public CppProject(XmlNodeList nodes)
        {
            mXmlDocMain = new XmlDocument();
            foreach(XmlNode node in nodes)
                CopyTo(mXmlDocMain, null, node);
        }

        public string Extension { get { return ".vcxproj"; } }

        public void RemovePlatform(string platform)
        {
            XmlDocument result = (XmlDocument)mXmlDocMain.Clone();
            Merge(result, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    return !HasCondition(node, platform, string.Empty);
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, true);

            mXmlDocMain = result;
        }

        public void RemoveConfigForPlatform(string config, string platform)
        {
            XmlDocument result = (XmlDocument)mXmlDocMain.Clone();
            Merge(result, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    return !HasCondition(node, platform, config);
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, true);

            mXmlDocMain = result;
        }

        public void RemoveAllBut(Dictionary<string, StringItems> platformConfigs)
        {
            XmlDocument result = (XmlDocument)mXmlDocMain.Clone();
            string platform, config;
            Merge(result, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (GetPlatformConfig(node, out platform, out config))
                    {
                        StringItems items;
                        if (platformConfigs.TryGetValue(platform, out items))
                        {
                            return (items.Contains(config));
                        }
                        return false;
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, true);

            mXmlDocMain = result;
        }

        public void RemoveAllPlatformsBut(string platformToKeep)
        {
            XmlDocument result = (XmlDocument)mXmlDocMain.Clone();
            string platform, config;
            Merge(result, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (GetPlatformConfig(node, out platform, out config))
                    {
                        if (String.Compare(platform, platformToKeep, true) == 0)
                        {
                            return true;
                        }
                        return false;
                    }
                    if (node.Name == "ItemGroup" && node.HasChildNodes)
                    {
                        if (IsOneOf(node.ChildNodes[0].Name, mContentItems))
                            return false;
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, true);

            mXmlDocMain = result;
        }

        public bool FilterItems(string[] to_remove, string[] to_keep)
        {
            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (IsOneOf(main.ParentNode.Name, mMergeItems))
                    {
                        StringItems items = new StringItems();
                        items.Add(main.Value, true);
                        items.Filter(to_remove, to_keep);
                        main.Value = items.ToString();
                    }
                }, false);

            return true;
        }
        
        public bool ExpandGlobs(string rootdir, string reldir)
        {
            List<XmlNode> removals = new List<XmlNode>();
            List<XmlNode> globs = new List<XmlNode>();

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (IsOneOf(node.Name, mContentItems))
                    {
                        foreach (XmlAttribute a in node.Attributes)
                        {
                            if (a.Name == "Include")
                            {
                                if (a.Value.Contains('*'))
                                {
                                    globs.Add(node);
                                }
                                else if (String.IsNullOrEmpty(a.Value))
                                {
                                    removals.Add(node);
                                }
                            }
                        }

                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, false);

            // Removal
            foreach (XmlNode node in removals)
            {
                XmlNode parent = node.ParentNode;
                parent.RemoveChild(node);
                if (parent.ChildNodes.Count == 0)
                {
                    XmlNode grandparent = parent.ParentNode;
                    grandparent.RemoveChild(parent);
                }
            }

            // Now do the file globbing
            HashSet<string> allGlobbedFiles = new HashSet<string>();
            foreach (XmlNode node in globs)
            {
                XmlNode parent = node.ParentNode;
                parent.RemoveChild(node);

                string glob = node.Attributes[0].Value;
                List<string> globbedFiles = PathUtil.getFiles(rootdir + glob);
                foreach (string filename in globbedFiles)
                {
                    if (!allGlobbedFiles.Contains(filename))
                    {
                        allGlobbedFiles.Add(filename);

                        XmlNode newNode = node.CloneNode(false);
                        string filedir = PathUtil.RelativePathTo(reldir, Path.GetDirectoryName(filename));
                        if (!String.IsNullOrEmpty(filedir) && !filedir.EndsWith("\\"))
                            filedir += "\\";

                        newNode.Attributes[0].Value = filedir + Path.GetFileName(filename);
                        parent.AppendChild(newNode);
                    }
                }
            }

            return true;
        }

        private bool GetPlatformConfig(XmlNode node, out string platform, out string config)
        {
            if (node.Attributes != null)
            {
                foreach (XmlAttribute a in node.Attributes)
                {
                    if (a.Name == "Condition")
                    {
                        string[] parts = StringTools.Between(a.Value, '\'', '\'');
                        if (parts.Length == 2)
                        {
                            config = StringTools.LeftOf(parts[1], '|');
                            platform = StringTools.RightOf(parts[1], '|');
                            if (!String.IsNullOrEmpty(config) && !String.IsNullOrEmpty(platform))
                                return true;
                        }
                        break;
                    }
                    else if (a.Name == "Include")
                    {
                        config = StringTools.LeftOf(a.Value, '|');
                        platform = StringTools.RightOf(a.Value, '|');
                        if (!String.IsNullOrEmpty(config) && !String.IsNullOrEmpty(platform))
                            return true;

                        break;
                    }
                }
            }
            config = null;
            platform = null;
            return false;
        }
        private bool HasCondition(XmlNode node, string platform, string config)
        {
            if (node.Attributes != null)
            {
                foreach (XmlAttribute a in node.Attributes)
                {
                    if (IsOneOf(a.Name, new string[] { "Condition", "Include" }))
                        return (a.Value.Contains(String.Format("{0}|{1}", config, platform)));
                }
            }
            return true;
        }
        
        public string[] GetPlatforms()
        {
            HashSet<string> platforms = new HashSet<string>();
            string platform, config;

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (GetPlatformConfig(node, out platform, out config))
                    {
                        if (!platforms.Contains(platform))
                            platforms.Add(platform);
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, false);
            return platforms.ToArray();
        }

        public string[] GetPlatformConfigs(string platform)
        {
            HashSet<string> configs = new HashSet<string>();
            string _platform, config;

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (GetPlatformConfig(node, out _platform, out config))
                    {
                        if (String.Compare(platform, _platform, true) == 0)
                        {
                            if (!configs.Contains(config))
                                configs.Add(config);
                        }
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                },false);
            return configs.ToArray();
        }
        
        public bool Save(string filename)
        {
            if (!Directory.Exists(Path.GetDirectoryName(filename)))
                Directory.CreateDirectory(Path.GetDirectoryName(filename));

            mXmlDocMain.Save(filename);
            return SaveFilter(filename +  ".filters");
        }

        private void WriteFilterIncludes(TextFile textFile, List<string> files, HashSet<string> directoryMap)
        {
            if (files.Count > 0)
            {
                textFile.WriteLine(1, "<ItemGroup>");
                foreach (string current_file in files)
                {
                    string path_to_file = Path.GetDirectoryName(current_file.Replace("..\\", ""));
                    string[] folders = path_to_file.Split(new char[] { '\\' }, StringSplitOptions.RemoveEmptyEntries);
                    string folderPath = string.Empty;
                    foreach (string folder in folders)
                    {
                        if (String.IsNullOrEmpty(folderPath))
                            folderPath = folder;
                        else
                            folderPath = folderPath + "\\" + folder;

                        if (!String.IsNullOrEmpty(folderPath) && !directoryMap.Contains(folderPath))
                        {
                            directoryMap.Add(folderPath);
                            textFile.WriteLine(2, "<Filter Include=\"{0}\">", folderPath);
                            textFile.WriteLine(3, "<UniqueIdentifier>{{{0}}}</UniqueIdentifier>", Guid.NewGuid().ToString());
                            textFile.WriteLine(2, "</Filter>");
                        }
                    }
                }
                textFile.WriteLine(1, "</ItemGroup>");
            }
        }

        private void WriteFileFilterBlock(TextFile textFile, List<string> files, string group_type)
        {
            if (files.Count > 0)
            {
                textFile.WriteLine(1, "<ItemGroup>");
                foreach (string current_file in files)
                {
                    string path_to_file = Path.GetDirectoryName(current_file.Replace("..\\", ""));
                    if (!String.IsNullOrEmpty(path_to_file))
                    {
                        textFile.WriteLine(2, "<{0} Include=\"{1}\">", group_type, current_file);
                        textFile.WriteLine(3, "<Filter>{0}</Filter>", path_to_file);
                        textFile.WriteLine(2, "</{0}>", group_type);
                    }
                    else
                    {
                        textFile.WriteLine(2, "<{0} Include=\"{1}\" />", group_type, current_file);
                    }
                }
                textFile.WriteLine(1, "</ItemGroup>");
            }
        }

        private bool SaveFilter(string filename)
        {
            // Find the <ItemGroup><ClCompile...
            List<string> clAll = new List<string>();
            List<string> clCompile = new List<string>();
            List<string> clInclude = new List<string>();
            List<string> clNone = new List<string>();
            List<string> clResourceCompile = new List<string>();

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    string includeAttrValue = Attribute.Get("Include", node, string.Empty);
                    if (!String.IsNullOrEmpty(includeAttrValue))
                    {
                        if (String.Compare(node.Name, "ClCompile", true) == 0)
                        {
                            clAll.Add(includeAttrValue);
                            clCompile.Add(includeAttrValue);
                        }
                        else if (String.Compare(node.Name, "ClInclude", true) == 0)
                        {
                            clAll.Add(includeAttrValue);
                            clInclude.Add(includeAttrValue);
                        }
                        else if (String.Compare(node.Name, "None", true) == 0)
                        {
                            clAll.Add(includeAttrValue);
                            clNone.Add(includeAttrValue);
                        }
                        else if (String.Compare(node.Name, "ClResourceCompile", true) == 0)
                        {
                            clAll.Add(includeAttrValue);
                            clResourceCompile.Add(includeAttrValue);
                        }
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                }, false);

            string tool_version_and_xmlns = "ToolsVersion=\"4.0\" xmlns=\"http://schemas.microsoft.com/developer/msbuild/2003\"";
            string xml_version_and_encoding = "<?xml version=\"1.0\" encoding=\"utf-8\"?>";

            TextFile textFile = new TextFile();
            textFile.Open(filename);
            textFile.WriteLine(xml_version_and_encoding);
            textFile.WriteLine("<Project " + tool_version_and_xmlns + ">");
            {
                HashSet<string> directoryMap = new HashSet<string>();
                WriteFilterIncludes(textFile, clAll, directoryMap);
                WriteFileFilterBlock(textFile, clInclude, "ClInclude");
                WriteFileFilterBlock(textFile, clCompile, "ClCompile");
                WriteFileFilterBlock(textFile, clNone, "None");
                WriteFileFilterBlock(textFile, clResourceCompile, "ResourceCompile");
            }
            textFile.WriteLine("</Project>");
            textFile.Close();

            return true;
        }

        public void MergeDependencyProject(IProject project)
        {
            Merge(project, false, false, false);
        }

        public bool Construct(IProject template)
        {
            MsDev.CppProject finalProject = new MsDev.CppProject();
            finalProject.Xml = template.Xml;
            finalProject.Merge(this, true, true, true);
            mXmlDocMain = finalProject.Xml;
            return true;
        }

        private bool Merge(IProject project, bool replaceValues, bool addMissingPlatformConfigurations, bool mergeContentItems)
        {
            string[] platforms = GetPlatforms();
            Dictionary<string, string[]> platformConfigs = new Dictionary<string, string[]>();
            foreach (string platform in platforms)
            {
                string[] configs = GetPlatformConfigs(platform);
                if (configs != null && configs.Length > 0)
                    platformConfigs.Add(platform, configs);
            }

            string _platform, _config;

            Merge(mXmlDocMain, project.Xml,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (!addMissingPlatformConfigurations)
                    {
                        // Do not merge in Configurations which do not exist in this main project
                        if (GetPlatformConfig(node, out _platform, out _config))
                        {
                            string[] configs;
                            if (platformConfigs.TryGetValue(_platform, out configs))
                            {
                                foreach (string config in configs)
                                {
                                    if (String.Compare(config, _config, true) == 0)
                                        return true;
                                }
                            }
                            return false;
                        }
                    }

                    if (!mergeContentItems)
                    {
                        if (node.Name == "ItemGroup" && node.HasChildNodes)
                        {
                            XmlNode child = node.ChildNodes[0];
                            if (IsOneOf(child.Name, mContentItems))
                            {
                                return false;
                            }
                        }
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (IsOneOf(main.ParentNode.Name, mMergeItems))
                    {
                        StringItems items = new StringItems();
                        items.Add(other.Value, true);
                        items.Add(main.Value, true);
                        items.Filter(new string[] { "%()" }, new string[0]);
                        main.Value = items.ToString();
                    }
                    else
                    {
                        if (replaceValues)
                            main.Value = other.Value;
                    }
                }, false);

            return true;
        }
    }
}
