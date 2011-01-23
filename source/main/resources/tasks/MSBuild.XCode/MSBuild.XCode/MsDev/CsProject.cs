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
    /// C# projects are much simpler than C++ projects.
    /// We do not have to merge anything, the only action
    /// that has to be done is to add references like:
    /// 
    ///     <Reference Include="#Ionic.Zip">
    ///       <HintPath>$(packageName_TargetDir)lib\Ionic.Zip.dll</HintPath>
    ///     </Reference>
    /// 
    /// No need for include files, preprocessor definitions
    /// and all that.
    /// 
    /// </summary>
    public class CsProject : BaseProject, IProject
    {
        private readonly static string[] mContentItems = new string[]
        {
            "Reference",
            "Compile", 
            "EmbeddedResource", 
            "BootstrapperPackage", 
            "None" 
        };

        public CsProject()
        {
            mXmlDocMain = new XmlDocument();
        }

        public CsProject(XmlNodeList nodes)
        {
            mXmlDocMain = new XmlDocument();
            foreach(XmlNode node in nodes)
                CopyTo(mXmlDocMain, null, node);
        }

        public string Extension { get { return ".csproj"; } }

        public void RemovePlatform(string platform)
        {
            XmlDocument result = new XmlDocument();
            mAllowRemoval = true;
            Merge(result, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
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
                delegate(bool isMainNode, XmlNode node)
                {
                    return !HasCondition(node, platform, config);
                },
                delegate(XmlNode main, XmlNode other)
                {
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
        }


        public void RemoveAllPlatformsBut(string platformToKeep)
        {
            XmlDocument result = (XmlDocument)mXmlDocMain.Clone();
            mAllowRemoval = true;
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
                        string childNodeName = node.ChildNodes[0].Name;
                        if (IsOneOf(childNodeName, mContentItems))
                        {
                            // References should not be stripped
                            if (String.Compare(childNodeName, "Reference", true) == 0)
                                return true;

                            return false;
                        }
                    }
                    return true;
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
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
        }

        /// if action == "Compile" and filename.endswith(".cs")
        ///    if (filename.endswith(".designer.cs"))
        ///        basename = filename.replace(".designer.cs", ".cs")
        ///        if (files.has(basename))
        ///            return ["dependency", basename]
        ///        endif
        ///        basename = basename.replace(".cs", ".resx")
        ///        if (files.has(basename))
        ///            return ["AutoGen", basename]
        ///        endif
        ///    else
        ///        basename = filename.replace(".cs", ".designer.cs")
        ///        if (files.has(basename))
        ///            return "SubTypeForm"
        ///        endif
        ///    endif
        /// endif
        ///
        ///
        /// if action == "EmbeddedResource" and filename.endswith(".resx")
        ///    basename = filename.replace(".resx", ".cs")
        ///    if (files.has(basename))
        ///        testname = basename.replace(".cs", ".designer.cs")
        ///        if (files.has(testname))
        ///            return ["DesignerType", basename]
        ///        else
        ///            return ["Dependency", testname]
        ///        endif
        ///    else
        ///        testname = basename.replace(".cs", ".designer.cs")
        ///        if (files.has(testname))
        ///            return "AutoGenerated"
        ///        endif
        ///    endif
        /// endif
        /// 
        /// return "None"
        /// 

        private KeyValuePair<string, string> GetElements(HashSet<string> files, string action, string filename)
        {
            if (action == "Compile" && filename.EndsWith(".cs"))
            {
                if (filename.EndsWith(".designer.cs"))
                {
                    string basename = filename.Replace(".designer.cs", ".cs");
                    if (files.Contains(basename))
                        return new KeyValuePair<string, string>("Dependency", basename);

                    basename = basename.Replace(".cs", ".resx");
                    if (files.Contains(basename))
                        return new KeyValuePair<string, string>("AutoGen", basename);
                }
                else
                {
                    string basename = filename.Replace(".cs", ".designer.cs");
                    if (files.Contains(basename))
                        return new KeyValuePair<string, string>("SubTypeForm", string.Empty);
                }
            }
            else if (action == "EmbeddedResource" && filename.EndsWith(".resx"))
            {
                string basename = filename.Replace(".resx", ".cs");
                if (files.Contains(basename))
                {
                    string testname = basename.Replace(".cs", ".designer.cs");
                    if (files.Contains(testname))
                        return new KeyValuePair<string, string>("DesignerType", basename);
                    else
                        return new KeyValuePair<string, string>("Dependency", testname);
                }
                else
                {
                    string testname = basename.Replace(".cs", ".designer.cs");
                    if (files.Contains(testname))
                        return new KeyValuePair<string, string>("AutoGenerated", string.Empty);
                }
            }
            else if (action == "Content")
            {
                return new KeyValuePair<string, string>("CopyNewest", string.Empty);
            }

            return new KeyValuePair<string, string>("None", string.Empty);
        }

        class Comparer<T> : IEqualityComparer<T>
        {
            private readonly Func<T, T, bool> _comparer;
            public Comparer(Func<T, T, bool> comparer)
            {
                if (comparer == null) 
                    throw new ArgumentNullException("comparer");
                _comparer = comparer;
            } 
            public bool Equals(T x, T y) 
            {
                return _comparer(x, y);
            }
            public int GetHashCode(T obj)
            {
                return obj.ToString().ToLower().GetHashCode();
            }
        }

        public bool ExpandGlobs(string rootdir, string reldir)
        {
            List<XmlNode> removals = new List<XmlNode>();
            List<XmlNode> globs = new List<XmlNode>();

            Merge(mXmlDocMain, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (node.Name == "Compile" || node.Name == "EmbeddedResource" || node.Name == "Content" || node.Name == "None")
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
            // First collect all files
            HashSet<string> allFiles = new HashSet<string>(new Comparer<string>((x, y) => String.Compare(x, y, true) == 0));
            List<HashSet<string>> filesPerNode = new List<HashSet<string>>();
            foreach (XmlNode node in globs)
            {
                string glob = node.Attributes[0].Value;
                int index = glob.IndexOf('*');

                HashSet<string> files = new HashSet<string>(new Comparer<string>((x, y) => String.Compare(x, y, true) == 0));
                List<string> globbedFiles = PathUtil.getFiles(rootdir + glob);
                foreach (string filename in globbedFiles)
                {
                    XmlNode newNode = node.CloneNode(false);
                    string filedir = PathUtil.RelativePathTo(reldir, Path.GetDirectoryName(filename));
                    if (!String.IsNullOrEmpty(filedir) && !filedir.EndsWith("\\"))
                        filedir += "\\";

                    string file = filedir + Path.GetFileName(filename);
                    if (!allFiles.Contains(file))
                    {
                        allFiles.Add(file);
                        if (!files.Contains(file))
                            files.Add(file);
                    }
                }
                filesPerNode.Add(files);
            }

            // Second update the xml nodes
            foreach (var nf in globs.Zip(filesPerNode, (n, f) => new { Node = n, Files = f }))
            //foreach (XmlNode node in globs)
            {
                XmlNode parent = nf.Node.ParentNode;
                parent.RemoveChild(nf.Node);

                foreach (string filename in nf.Files)
                {
                    XmlNode newNode = nf.Node.CloneNode(false);

                    KeyValuePair<string, string> element = GetElements(allFiles, nf.Node.Name, filename);
                    /// if element.Key == "None" then
                    /// 	_p('    <%s Include="%s" />', action, fname)
                    /// else
                    /// 	_p('    <%s Include="%s">', action, fname)
                    /// 	if element.Key == "AutoGen" then
                    /// 		_p('      <AutoGen>True</AutoGen>')
                    /// 	elseif element.Key == "AutoGenerated" then
                    /// 		_p('      <SubType>Designer</SubType>')
                    /// 		_p('      <Generator>ResXFileCodeGenerator</Generator>')
                    /// 		_p('      <LastGenOutput>%s.Designer.cs</LastGenOutput>', premake.esc(path.getbasename(fcfg.name)))
                    /// 	elseif element.Key == "SubTypeDesigner" then
                    /// 		_p('      <SubType>Designer</SubType>')
                    /// 	elseif element.Key == "SubTypeForm" then
                    /// 		_p('      <SubType>Form</SubType>')
                    /// 	elseif element.Key == "PreserveNewest" then
                    /// 		_p('      <CopyToOutputDirectory>PreserveNewest</CopyToOutputDirectory>')
                    /// 	end
                    /// 	if (!String.IsNullOrEmpty(element.Value))
                    /// 		_p('      <DependentUpon>%s</DependentUpon>', path.translate(premake.esc(dependency), "\\"))
                    /// 	end
                    /// 	_p('    </%s>', action)
                    /// end

                    newNode.Attributes[0].Value = filename;
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
                            if (a.Value[cursor] == ' ')
                                cursor++;

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
                });
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
                });
            return configs.ToArray();
        }

        public bool FilterItems(string[] to_remove, string[] to_keep)
        {
            mAllowRemoval = true;
            XmlDocument result = (XmlDocument)mXmlDocMain.Clone();
            Merge(result, mXmlDocMain,
                delegate(bool isMainNode, XmlNode node)
                {
                    if (node.Name == "Reference")
                    {
                        string include = Attribute.Get("Include", node, string.Empty);
                        foreach (string keeper in to_keep)
                        {
                            if (include.StartsWith(keeper))
                            {
                                Attribute.Set("Include", node, include.Substring(keeper.Length));
                                return true;
                            }
                        }
                        foreach (string remover in to_remove)
                        {
                            if (include.StartsWith(remover))
                            {
                                return false;
                            }
                        }
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                });

            mXmlDocMain = result;
            mAllowRemoval = false;
            return true;
        }

        public bool Save(string filename)
        {
            mXmlDocMain.Save(filename);
            return true;
        }

        public void Copy(XmlDocument doc)
        {
            mXmlDocMain = (XmlDocument)doc.Clone();
        }

        public bool Merge(IProject project)
        {
            Merge(mXmlDocMain, project.Xml,
                delegate(bool isMainNode, XmlNode node)
                {
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    main.Value = other.Value;
                });

            return true;
        }

        public bool Construct(IProject template)
        {
            MsDev.CsProject finalProject = new MsDev.CsProject();
            finalProject.Xml = template.Xml;
            finalProject.Merge(this);
            mXmlDocMain = finalProject.Xml;
            return true;
        }

        public void MergeDependencyProject(IProject project)
        {
            Merge(mXmlDocMain, project.Xml,
                delegate(bool isMainNode, XmlNode node)
                {
                    /// Only merge dependency project elements like:
                    ///     <Reference Include="#Ionic.Zip">
                    ///       <HintPath>$(packageName_TargetDir)lib\Ionic.Zip.dll</HintPath>
                    ///     </Reference>
                    if (!isMainNode)
                    {
                        if (node.Name == "Reference" || node.Name == "ImportGroup" || node.Name == "Import")
                        {
                            return true;
                        }
                        return false;
                    }
                    return true;
                },
                delegate(XmlNode main, XmlNode other)
                {
                    
                });
        }

    }
}
