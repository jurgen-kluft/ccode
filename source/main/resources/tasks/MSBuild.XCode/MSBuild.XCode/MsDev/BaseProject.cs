using System;
using System.IO;
using System.Collections.Generic;
using System.Xml;
using System.Text;

namespace MSBuild.XCode.MsDev
{
    public abstract class BaseProject
    {
        protected XmlDocument mXmlDocMain;

        public XmlDocument Xml
        {
            get
            {
                return mXmlDocMain;
            }
            set
            {
                mXmlDocMain = (XmlDocument)value.Clone();
            }
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

        public bool ExpandVars(Dictionary<string, string> vars)
        {
            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (node.Attributes != null)
                    {
                        foreach (XmlAttribute a in node.Attributes)
                        {
                            a.Value = ReplaceVars(a.Value, vars);
                        }
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    if (main.Attributes != null)
                    {
                        foreach (XmlAttribute a in main.Attributes)
                        {
                            foreach (KeyValuePair<string, string> var in vars)
                                a.Value = ReplaceVars(a.Value, vars);
                        }
                    }

                    foreach (KeyValuePair<string, string> var in vars)
                    {
                        if (!String.IsNullOrEmpty(main.Value))
                            main.Value = ReplaceVars(main.Value, vars);
                    }
                }, false);
            return true;
        }

        /// <summary>
        /// Copy a node from a source XmlDocument to a target XmlDocument
        /// </summary>
        /// <param name="domTarget">The XmlDocument to which we want to copy</param>
        /// <param name="node">The node we want to copy</param>
        protected static XmlNode CopyTo(XmlDocument xmlDoc, XmlNode xmlDocNode, XmlNode nodeToCopy)
        {
            XmlNode copy = xmlDoc.ImportNode(nodeToCopy, true);
            if (xmlDocNode != null)
                xmlDocNode.AppendChild(copy);
            else
                xmlDoc.AppendChild(copy);
            return copy;
        }

        protected string ReplaceVars(string str, Dictionary<string, string> vars)
        {
            foreach (KeyValuePair<string, string> var in vars)
                str = str.Replace(String.Format("${{{0}}}", var.Key), var.Value);
            return str;
        }

        protected static bool IsOneOf(string str, string[] strs)
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

        protected static bool HasSameAttributes(XmlNode a, XmlNode b)
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

        protected static XmlNode FindNode(XmlNode nodeToFind, XmlNodeList children)
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
                            if (nodeToFind.Name == "ItemGroup" && nodeToFind.Attributes.Count == 0)
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
                        else
                        {
                            foundNode = child;
                            break;
                        }
                    }
                }
            }
            return foundNode;
        }

        public delegate void NodeMergeDelegate(XmlNode main, XmlNode other);
        public delegate bool NodeConditionDelegate(bool isMainNode, XmlNode node);

        protected static void LockStep(XmlDocument mainXmlDoc, XmlDocument otherXmlDoc, Stack<XmlNode> mainPath, Stack<XmlNode> otherPath, NodeConditionDelegate nodeConditionDelegate, NodeMergeDelegate nodeMergeDelegate, bool allowRemoval)
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
                        if (nodeConditionDelegate(false, otherChildNode))
                        {
                            mainChildNode = CopyTo(mainXmlDoc, mainNode, otherChildNode);

                            mainPath.Push(mainChildNode);
                            otherPath.Push(otherChildNode);
                            LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate, allowRemoval);
                        }
                    }
                    else
                    {
                        if (nodeConditionDelegate(true, mainChildNode))
                        {
                            mainPath.Push(mainChildNode);
                            otherPath.Push(otherChildNode);
                            LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate, allowRemoval);
                        }
                        else if (allowRemoval)
                        {
                            // Removal
                            mainNode.RemoveChild(mainChildNode);
                        }
                    }
                }
            }
        }

        protected static void Merge(XmlDocument mainXmlDoc, XmlDocument otherXmlDoc, NodeConditionDelegate nodeConditionDelegate, NodeMergeDelegate nodeMergeDelegate, bool allowRemoval)
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
                    if (nodeConditionDelegate(false, otherChildNode))
                    {
                        mainChildNode = CopyTo(mainXmlDoc, null, otherChildNode);

                        mainPath.Push(mainChildNode);
                        otherPath.Push(otherChildNode);
                        LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate, allowRemoval);
                    }
                }
                else
                {
                    if (nodeConditionDelegate(true, mainChildNode))
                    {
                        mainPath.Push(mainChildNode);
                        otherPath.Push(otherChildNode);
                        LockStep(mainXmlDoc, otherXmlDoc, mainPath, otherPath, nodeConditionDelegate, nodeMergeDelegate, allowRemoval);
                    }
                    else if (allowRemoval)
                    {
                        // Removal
                        mainXmlDoc.RemoveChild(mainChildNode);
                    }
                }
            }
        }
    }
}
