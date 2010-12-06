using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode
{
    public partial class MsDevProjectFileGenerator
    {
        private string mRootDir = string.Empty;

        private EVersion mVersion = EVersion.VS2010;
        private ELanguage mLanguage = ELanguage.CS;

        private string[] mPlatforms;
        private string[] mConfigs;

        private string mProjectName;
        private string mProjectGuid;

        private XProjectWriter mXProjectWriter;
        private StreamWriter mWriter;

        private string mCondition = "Condition=\"'$(Configuration)|$(Platform)'=='%s'\"";
        private string mToolVersionAndXmlns = "ToolsVersion=\"4.0\" xmlns=\"http://schemas.microsoft.com/developer/msbuild/2003\"";
        private string mXmlVersionAndEncoding = "<?xml version=\"1.0\" encoding=\"utf-8\"?>";

        private void _p(int tab, string text)
        {
            _p(tab, text, null);
        }

        private void _p(int tab, string text, params string[] p)
        {
            for (int i = 0; i < tab; ++i)
                mWriter.Write("\t");
            string s = _c(text, p);
            mWriter.WriteLine(s);
        }

        private string _c(string text, params string[] p)
        {
            int pi = 0;
            int cursor = 0;
            while (p != null && pi < p.Length)
            {
                cursor = text.IndexOf("%s", cursor);
                if (cursor >= 0)
                {
                    text = text.Substring(0, cursor) + p[pi] + text.Substring(cursor + 2);
                    cursor = cursor + p[pi].Length;
                    ++pi;
                }
                break;
            }
            return text;
        }

        private void _SaveGroup(int indent, string platform, string config, string group)
        {
            List<string> lines = mXProjectWriter.GetGroupElementsFor(platform, config, group);
            if (lines.Count > 0)
            {
                _p(indent, "<" + group + ">");
                {
                    foreach (string line in lines)
                        _p(indent + 1, line);
                }
                _p(indent, "</" + group + ">");
            }
        }

        private void _SaveElement(int indent, string platform, string config, string group)
        {
            List<string> lines = mXProjectWriter.GetGroupElementsFor(platform, config, group);
            foreach (string line in lines)
                _p(indent, line);
        }

        private void _SaveGroup(int indent, string begin, string end, string platform, string config, string group)
        {
            List<string> lines = mXProjectWriter.GetGroupElementsFor(platform, config, group);
            if (lines.Count > 0)
            {
                _p(indent, begin);
                foreach (string line in lines)
                    _p(indent+1, line);
                _p(indent, end);
            }
        }
        private void _SaveConfig()
        {
            _p(1, "<ItemGroup Label=\"ProjectConfigurations\">");
            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    _p(2, "<ProjectConfiguration Include=\"%s\">", c + "|" + p);
                        _p(3, "<Configuration>%s</Configuration>", c);
                        _p(3, "<Platform>%s</Platform>", p);
                    _p(2, "</ProjectConfiguration>");
                }
            }
            _p(1, "</ItemGroup>");        
        }

        public void _SaveGlobals()
        {
            _p(1, "<PropertyGroup Label=\"Globals\">");
                _p(2, "<ProjectGuid>{%s}</ProjectGuid>", mProjectGuid);
                _p(2, "<RootNamespace>%s</RootNamespace>", mProjectName);
                _p(2, "<Keyword>Win32Proj</Keyword>");
            _p(1, "</PropertyGroup>");
        }

        public void _SaveConfigTypeBlock()
        {
            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    string begin = _c("<PropertyGroup " + mCondition + " Label=\"Configuration\">", c + "|" + p);
                    string end = _c("</PropertyGroup>");
                    _SaveGroup(1, begin, end, p, c, "Configuration");
                }
            }
        }
        public void _SaveImportProps()
        {
            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    string begin = _c("<ImportGroup " + mCondition + " Label=\"PropertySheets\">", c + "|" + p);
                    string end = _c("</ImportGroup>");
                    _SaveGroup(1, begin, end, p, c, "ImportGroup");
                }
            }
        }

        public void _SaveIntermediateAndOutDirs()
        {
            _p(1, "<PropertyGroup>");
            _p(2, "<_ProjectFileVersion>10.0.30319.1</_ProjectFileVersion>");

            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    _SaveElement(2, p, c, "OutDir");
                    _SaveElement(2, p, c, "IntDir");
                    _SaveElement(2, p, c, "TargetName");
                    _SaveElement(2, p, c, "IgnoreImportLibrary");
                    _SaveElement(2, p, c, "LinkIncremental");
                    _SaveElement(2, p, c, "GenerateManifest");
                }
            }
            _p(1, "</PropertyGroup>");
        }

        public void _SaveItemDefinitions()
        {
            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    _p(1, "<ItemDefinitionGroup " + mCondition + ">", c + "|" + p);
                    {
                        _SaveGroup(2, p, c, "ClCompile");
                        _SaveGroup(2, p, c, "ResourceCompile");
                        _SaveGroup(2, p, c, "Lib");
                        _SaveGroup(2, p, c, "Link");
                        _SaveGroup(2, p, c, "PostBuildEvent");
                        _SaveGroup(2, p, c, "PreBuildEvent");
                        _SaveGroup(2, p, c, "PreLinkEvent");
                    }
                    _p(1, "</ItemDefinitionGroup>");
                }
            }
        }

        public void _Save(string filename)
        {
            using (FileStream wfs = new FileStream(filename, FileMode.Create, FileAccess.Write))
            {
                using (mWriter = new StreamWriter(wfs))
                {
                    _p(0, mXmlVersionAndEncoding);
                    _p(0, "<Project DefaultTargets=\"Build\" " + mToolVersionAndXmlns + ">");
                    _SaveConfig();
                    _SaveGlobals();

                    _p(1, "<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.Default.props\" />");

                    _SaveConfigTypeBlock();

                    _p(1, "<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.props\" />");

                    // check what this section is doing
                    _p(1, "<ImportGroup Label=\"ExtensionSettings\">");
                    _p(1, "</ImportGroup>");

                    _SaveImportProps();

                    // what type of macros are these?
                    _p(1, "<PropertyGroup Label=\"UserMacros\" />");

                    _SaveIntermediateAndOutDirs();
                    _SaveItemDefinitions();

                    // save files ?

                    _p(1, "<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.targets\" />");
                    _p(1, "<ImportGroup Label=\"ExtensionTargets\">");
                    _p(1, "</ImportGroup>");

                    _p(0, "</Project>");
                    mWriter.Close();
                    mWriter = null;
                }
                wfs.Close();
            }
        }
    }
}
