using System;
using System.IO;
using System.Xml;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MsDev2010.Cpp.XCode
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
    public class Project
    {
        private bool mAllowRemoval;
        private XmlDocument mXmlDocMain;

        /// <summary>
        /// Copy a node from a source XmlDocument to a target XmlDocument
        /// </summary>
        /// <param name="domTarget">The XmlDocument to which we want to copy</param>
        /// <param name="node">The node we want to copy</param>
        private XmlNode CopyTo(XmlDocument xmlDoc, XmlNode xmlDocNode, XmlNode nodeToCopy)
        {
            XmlNode copy = xmlDoc.ImportNode(nodeToCopy, true);
            if (xmlDocNode != null)
                xmlDocNode.AppendChild(copy);
            else
                xmlDoc.AppendChild(copy);
            return copy;
        }

        public Project()
        {
            mXmlDocMain = new XmlDocument();
        }

        public Project(XmlNodeList nodes)
        {
            mXmlDocMain = new XmlDocument();
            foreach(XmlNode node in nodes)
                CopyTo(mXmlDocMain, null, node);
        }

        public bool Load(string filename)
        {
            if (File.Exists(filename))
            {
                mXmlDocMain = new XmlDocument();
                mXmlDocMain.Load(filename);
                return true;
            }
            return false;
        }

        public void RemovePlatform(string platform)
        {
            XmlDocument result = new XmlDocument();
            mAllowRemoval = true;
            Merge(result, mXmlDocMain,
                delegate(XmlNode node)
                {
                    return !HasCondition(node, platform, string.Empty);
                },
                delegate(XmlNode main, XmlNode other)
                {
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
        }

        public void RemoveConfigForPlatform(string config, string platform)
        {
            XmlDocument result = new XmlDocument();
            mAllowRemoval = true; 
            Merge(result, mXmlDocMain,
                delegate(XmlNode node)
                {
                    return !HasCondition(node, platform, config);
                },
                delegate(XmlNode main, XmlNode other)
                {
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
        }

        public void RemoveAllBut(Dictionary<string, StringItems> platformConfigs)
        {
            XmlDocument result = new XmlDocument();
            mAllowRemoval = true;
            string platform, config;
            Merge(result, mXmlDocMain,
                delegate(XmlNode node)
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
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
        }

        public void RemoveAllPlatformsBut(string platformToKeep)
        {
            XmlDocument result = new XmlDocument();
            mAllowRemoval = true;
            string platform, config;
            Merge(result, mXmlDocMain,
                delegate(XmlNode node)
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
                        if (IsOneOf(node.ChildNodes[0].Name, new string[] { "ClCompile", "ClInclude", "None" }))
                            return false;
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
        }

        public bool FilterItems(string[] to_remove, string[] to_keep)
        {
            Merge(mXmlDocMain, mXmlDocMain,
                delegate(XmlNode node)
                {
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (IsOneOf(main.ParentNode.Name, new string[] { "PreprocessorDefinitions", "AdditionalDependencies", "AdditionalLibraryDirectories", "AdditionalIncludeDirectories" }))
                    {
                        StringItems items = new StringItems();
                        items.Add(main.Value, true);
                        items.Filter(to_remove, to_keep);
                        main.Value = items.ToString();
                    }
                });

            return true;
        }

        public bool ExpandGlobs(string rootdir, string reldir)
        {
            List<XmlNode> removals = new List<XmlNode>();
            List<XmlNode> globs = new List<XmlNode>();

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(XmlNode node)
                {
                    if (IsOneOf(node.Name, new string[] { "ClCompile", "ClInclude", "None" }))
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
                });

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
                delegate(XmlNode node)
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
                });
            return platforms.ToArray();
        }

        public string[] GetPlatformConfigs(string platform)
        {
            HashSet<string> configs = new HashSet<string>();
            string _platform, config;

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(XmlNode node)
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
                });
            return configs.ToArray();
        }
        
        public bool Save(string filename)
        {
            mXmlDocMain.Save(filename);
            return true;
        }

        private static bool IsOneOf(string str, string[] strs)
        {
            if (String.IsNullOrEmpty(str))
                return false;

            foreach (string s in strs)
            {
                if ((str.Length == s.Length) && (str[0] == s[0]))
                {
                    if (String.Compare(str, s, true) == 0)
                        return true;
                }
            }
            return false;
        }

        public bool Merge(Project project, bool replace, bool content)
        {
            Merge(mXmlDocMain, project.mXmlDocMain,
                delegate(XmlNode node)
                {
                    // If (content == False)
                    //    do not merge these:
                    //          <ItemGroup>
                    //            <ClCompile Include="source.cpp">
                    //          </ItemGroup>
                    //          <ItemGroup>
                    //            <ClInclude Include="source.h">
                    //          </ItemGroup>
                    //          <ItemGroup>
                    //            <None Include="source.png">
                    //          </ItemGroup>
                    // Endif
                    if (!content && node.Name == "ItemGroup" && node.HasChildNodes)
                    {
                        XmlNode child = node.ChildNodes[0];
                        if (IsOneOf(child.Name, new string[] { "ClCompile", "ClInclude", "None" }))
                        {
                            return false;
                        }
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (IsOneOf(main.ParentNode.Name, new string[] { "PreprocessorDefinitions", "AdditionalIncludeDirectories", "AdditionalLibraryDirectories", "AdditionalDependencies" }))
                    {
                        StringItems items = new StringItems();
                        items.Add(other.Value, true);
                        items.Add(main.Value, true);
                        items.Filter(new string[] { "%()" }, new string[0]);
                        main.Value = items.ToString();
                    }
                    else
                    {
                        if (replace)
                            main.Value = other.Value;
                    }
                });
            return true;
        }

        private bool HasSameAttributes(XmlNode a, XmlNode b)
        {
            if (a.Attributes == null && b.Attributes == null)
                return true;
            if (a.Attributes != null && b.Attributes != null)
            {
                bool the_same = (a.Attributes.Count == b.Attributes.Count);
                if (the_same)
                {
                    foreach (XmlAttribute aa in a.Attributes)
                    {
                        the_same = false;
                        foreach (XmlAttribute ab in b.Attributes)
                        {
                            if ((ab.Name == aa.Name) && (ab.Value == aa.Value))
                            {
                                the_same = true;
                                break;
                            }
                        }
                        if (!the_same)
                            break;
                    }
                }
                return the_same;
            }
            return false;
        }

        private XmlNode FindNode(XmlNode nodeToFind, XmlNodeList children)
        {
            // vcxproj has multiple <ItemGroup> nodes with different content, we
            // need to make sure we pick the right one

            XmlNode foundNode = null;
            foreach (XmlNode child in children)
            {
                // First, match by name
                if (child.Name == nodeToFind.Name)
                {
                    // Now see if the attributes match
                    if (HasSameAttributes(nodeToFind, child))
                    {
                        if (!nodeToFind.HasChildNodes && !child.HasChildNodes)
                        {
                            foundNode = child;
                            break;
                        }
                        else if (nodeToFind.HasChildNodes && child.HasChildNodes)
                        {
                            if (nodeToFind.Name == "ItemGroup" && nodeToFind.Attributes.Count==0)
                            {
                                if (nodeToFind.ChildNodes[0].Name == child.ChildNodes[0].Name)
                                {
                                    foundNode = child;
                                    break;
                                }
                            }
                            else
                            {
                                foundNode = child;
                                break;
                            }
                        }
                        
                    }
                }
            }
            return foundNode;
        }

        public delegate void NodeMergeDelegate(XmlNode main, XmlNode other);
        public delegate bool NodeConditionDelegate(XmlNode node);

        private void LockStep(XmlDocument mainXmlDoc, XmlDocument otherXmlDoc, Stack<XmlNode> mainPath, Stack<XmlNode> otherPath, NodeConditionDelegate nodeConditionDelegate, NodeMergeDelegate nodeMergeDelegate)
        {
            XmlNode mainNode = mainPath.Peek();
            XmlNode otherNode = otherPath.Peek();

            if (mainNode.NodeType == XmlNodeType.Comment)
            {
            }
            else if (mainNode.NodeType == XmlNodeType.Text)
            {
                nodeMergeDelegate(mainNode, otherNode);
            }
            else
            {
                foreach (XmlNode otherChildNode in otherNode.ChildNodes)
                {
                    XmlNode mainChildNode = FindNode(otherChildNode, mainNode.ChildNodes);
                    if (mainChildNode == null)
                    {
                        if (nodeConditionDelegate(otherChildNode))
                        {
                            mainChildNode = CopyTo(mainXmlDoc, mainNode, otherChildNode);

                            mainPath.Push(mainChildNode);
                            otherPath.Push(otherChildNode);
                            LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate);
                        }
                    }
                    else
                    {
                        if (nodeConditionDelegate(mainChildNode))
                        {
                            mainPath.Push(mainChildNode);
                            otherPath.Push(otherChildNode);
                            LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate);
                        }
                        else if (mAllowRemoval)
                        {
                            // Removal
                            mainNode.RemoveChild(mainChildNode);
                        }
                    }
                }
            }
        }

        private void Merge(XmlDocument mainXmlDoc, XmlDocument otherXmlDoc, NodeConditionDelegate nodeConditionDelegate, NodeMergeDelegate nodeMergeDelegate)
        {
            // Lock-Step Merge the xml tree
            // 1) When encountering a node which does not exist in the main doc, insert it
            Stack<XmlNode> mainPath = new Stack<XmlNode>();
            Stack<XmlNode> otherPath = new Stack<XmlNode>();
            foreach (XmlNode otherChildNode in otherXmlDoc)
            {
                XmlNode mainChildNode = FindNode(otherChildNode, mainXmlDoc.ChildNodes);
                if (mainChildNode == null)
                {
                    if (nodeConditionDelegate(otherChildNode))
                    {
                        mainChildNode = CopyTo(mainXmlDoc, null, otherChildNode);

                        mainPath.Push(mainChildNode);
                        otherPath.Push(otherChildNode);
                        LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate);
                    }
                }
                else
                {
                    if (nodeConditionDelegate(mainChildNode))
                    {
                        mainPath.Push(mainChildNode);
                        otherPath.Push(otherChildNode);
                        LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate);
                    }
                    else if (mAllowRemoval)
                    {
                        // Removal
                        mainXmlDoc.RemoveChild(mainChildNode);
                    }
                }
            }
        }
    }
}
