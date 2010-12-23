﻿using System;
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

        public bool FilterItems(string[] to_remove, string[] to_keep)
        {
            Merge(mXmlDocMain, mXmlDocMain,
                delegate(XmlNode node)
                {
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (main.ParentNode.Name == "PreprocessorDefinitions" ||
                        main.ParentNode.Name == "AdditionalDependencies" ||
                        main.ParentNode.Name == "AdditionalLibraryDirectories" ||
                        main.ParentNode.Name == "AdditionalIncludeDirectories")
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
                    if (node.Name == "ClCompile" || node.Name == "ClInclude" || node.Name == "None")
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

            // Now do the globbing
            foreach (XmlNode node in globs)
            {
                XmlNode parent = node.ParentNode;
                parent.RemoveChild(node);

                string glob = node.Attributes[0].Value;
                int index = glob.IndexOf('*');

                foreach (string filename in Globber.Glob(rootdir + glob))
                {
                    XmlNode newNode = node.CloneNode(false);
                    string filedir = PathUtil.RelativePathTo(reldir, Path.GetDirectoryName(filename));
                    if (!String.IsNullOrEmpty(filedir) && !filedir.EndsWith("\\"))
                        filedir += "\\";

                    newNode.Attributes[0].Value = filedir + Path.GetFileName(filename);
                    parent.AppendChild(newNode);
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
                    int begin = -1;
                    int end = -1;

                    if (a.Name == "Condition")
                    {
                        int cursor = a.Value.IndexOf("==");
                        if (cursor >= 0)
                        {
                            cursor += 2;
                            if ((cursor + 1) < a.Value.Length && a.Value[cursor] == '\'')
                                cursor += 1;
                            else
                                cursor = -1;
                        }
                        begin = cursor;
                        end = begin>=0 ? a.Value.IndexOf("'", begin) : begin;
                    }
                    else if (a.Name == "Include")
                    {
                        begin = 0;
                        end = a.Value.Length;
                    }

                    if (begin >= 0 && end > begin)
                    {
                        string configplatform = a.Value.Substring(begin, end - begin);
                        string[] items = configplatform.Split(new char[] { '|' }, StringSplitOptions.RemoveEmptyEntries);
                        if (items.Length == 2)
                        {
                            config = items[0];
                            platform = items[1];
                            return true;
                        }
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
                    if (a.Name == "Condition" || a.Name == "Include")
                    {
                        if (a.Value.Contains(String.Format("{0}|{1}", config, platform)))
                        {
                            return true;
                        }
                        else
                        {
                            return false;
                        }
                    }
                }
            }
            return true;
        }

        private string GetItem(string platform, string config, string itemName)
        {
            StringItems concat = new StringItems();

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(XmlNode node)
                {
                    return HasCondition(node, platform, config);
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (main.ParentNode.Name == itemName)
                    {
                        concat.Add(main.Value, true);
                        concat.Add(other.Value, true);
                    }
                });
            return concat.Get();
        }

        public bool GetPreprocessorDefinitions(string platform, string config, out string defines)
        {
            defines = GetItem(platform, config, "PreprocessorDefinitions");
            return true;
        }
        public bool GetAdditionalIncludeDirectories(string platform, string config, out string includeDirectories)
        {
            includeDirectories = GetItem(platform, config, "AdditionalIncludeDirectories");
            return true;
        }
        public bool GetAdditionalLibraryDirectories(string platform, string config, out string libraryDirectories)
        {
            libraryDirectories = GetItem(platform, config, "AdditionalLibraryDirectories");
            return true;
        }
        public bool GetAdditionalDependencies(string platform, string config, out string libraryDependencies)
        {
            libraryDependencies = GetItem(platform, config, "AdditionalDependencies");
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

        public bool SetItem(string platform, string config, string itemName, string itemValue)
        {
            StringItems concat = new StringItems();
            concat.Add(itemValue, true);

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(XmlNode node)
                {
                    return HasCondition(node, platform, config);
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (main.ParentNode.Name == itemName)
                    {
                        concat.Add(main.Value, true);
                        concat.Add(other.Value, true);
                        main.Value = concat.Get();
                    }
                });
            return true;
        }
        public bool SetPreprocessorDefinitions(string platform, string config, string defines)
        {
            return SetItem(platform, config, "PreprocessorDefinitions", defines);
        }
        public bool SetAdditionalIncludeDirectories(string platform, string config, string includeDirectories)
        {
            return SetItem(platform, config, "AdditionalIncludeDirectories", includeDirectories);
        }
        public bool SetAdditionalLibraryDirectories(string platform, string config, string libraryDirectories)
        {
            return SetItem(platform, config, "AdditionalLibraryDirectories", libraryDirectories);
        }
        public bool SetAdditionalDependencies(string platform, string config, string libraryDependencies)
        {
            return SetItem(platform, config, "AdditionalDependencies", libraryDependencies);
        }

        public bool Save(string filename)
        {
            mXmlDocMain.Save(filename);
            return true;
        }

        public bool Merge(Project project)
        {
            Merge(mXmlDocMain, project.mXmlDocMain,
                delegate(XmlNode node)
                {
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    // Merge:
                    // - PreprocessorDefinitions
                    // - AdditionalIncludeDirectories
                    // - AdditionalLibraryDirectories
                    // - AdditionalDependencies
                    if (main.ParentNode.Name == "PreprocessorDefinitions" ||
                        main.ParentNode.Name == "AdditionalIncludeDirectories" ||
                        main.ParentNode.Name == "AdditionalLibraryDirectories" ||
                        main.ParentNode.Name == "AdditionalDependencies")
                    {
                        StringItems items = new StringItems();
                        items.Add(other.Value, true);
                        items.Add(main.Value, true);
                        main.Value = items.ToString();
                    }
                    else
                    {
                        // Replace
                        main.Value = other.Value;
                    }
                });
            return false;
        }

        private bool HasSameAttributes(XmlNode a, XmlNode b)
        {
            if (a.Attributes == null && b.Attributes == null)
                return true;
            if (a.Attributes != null && b.Attributes != null)
            {
                bool the_same = true;
                int na = 0;
                foreach (XmlAttribute aa in a.Attributes)
                {
                    if (aa.Name == "Concat")
                        continue;
                    ++na;
                }
                int nb = 0;
                foreach (XmlAttribute ab in b.Attributes)
                {
                    if (ab.Name == "Concat")
                        continue;
                    ++nb;
                }

                the_same = (na == nb);
                if (the_same)
                {
                    foreach (XmlAttribute aa in a.Attributes)
                    {
                        if (aa.Name == "Concat")
                            continue;

                        bool found = false;
                        foreach (XmlAttribute ab in b.Attributes)
                        {
                            if (ab.Name == "Concat")
                                continue;

                            if (ab.Name == aa.Name)
                            {
                                if (ab.Value == aa.Value)
                                {
                                    found = true;
                                    break;
                                }
                            }
                        }
                        if (!found)
                        {
                            the_same = false;
                            break;
                        }
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
