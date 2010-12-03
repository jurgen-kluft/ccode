using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace ProjectFileGenerator
{
    public partial class ProjectFileGenerator
    {
        private string mRootDir = string.Empty;

        private EVersion mVersion = EVersion.VS2010;
        private ELanguage mLanguage = ELanguage.CS;

        private string[] mPlatforms;
        private string[] mConfigs;

        private string mProjectName;
        private string mProjectGuid;

        private ProjectFileTemplate mTemplate;

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

            int pi = 0;
            int cursor = 0;
            while (p != null && pi < p.Length)
            {
                cursor = text.IndexOf("%s", cursor);
                if (cursor >= 0)
                {
                    text = text.Substring(0, cursor) + p[pi] + text.Substring(cursor + 2);
                    ++pi;
                    cursor = cursor + p[pi].Length;
                }
                break;
            }
            mWriter.WriteLine(text);
        }

        private void EmitGroupElementsFor(int indent, string platform, string config, string group)
        {
            List<string> lines = mTemplate.GetGroupElementsFor(platform, config, group);
            foreach (string line in lines)
                _p(indent, line);
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
            _p(1, "PropertyGroup>");
        }

        public void _SaveConfigTypeBlock()
        {
            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    _p(1, "<PropertyGroup " + if_config_and_platform() + " Label=\"Configuration\">", Helpers.esc(cfginfo.name));
                    EmitGroupElementsFor(2,p,c,"Configuration");
                    _p(1, "</PropertyGroup>");
                }
            }
        }
        public void _SaveImportProps()
        {
            foreach (string c in mConfigs)
            {
                foreach (string p in mPlatforms)
                {
                    _p(1, "<ImportGroup " + if_config_and_platform() + " Label=\"PropertySheets\">", Helpers.esc(cfginfo.name));
                    EmitGroupElementsFor(2, p, c, "ImportGroup");
                    _p(1, "</ImportGroup>");
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
                    EmitGroup(2, p, c, "OutDir");
                    EmitGroup(2, p, c, "IntDir");
                    EmitGroup(2, p, c, "TargetName");
                    EmitGroup(2, p, c, "IgnoreImportLibrary");
                    EmitGroup(2, p, c, "LinkIncremental");
                    EmitGroup(2, p, c, "GenerateManifest");
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
                    _p(1, "<ItemDefinitionGroup " + if_config_and_platform() + ">", Helpers.esc(cfginfo.name));
                    _p(2, "<ClCompile>")
                    clcompile(cfg);
                    resource_compile(cfg);
                    item_def_lib(cfg);
                    item_link(cfg);
                    event_hooks(cfg);
                    _p(1, "</ItemDefinitionGroup>");
                }
            }
        }

        public void _Save(string filename)
        {
            FileStream wfs = new FileStream(filename, FileMode.Create, FileAccess.Write);
            mWriter = new StreamWriter(wfs);

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
            item_definitions();
            vcxproj_files();

            _p(1, "<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.targets\" />");
            _p(1, "<ImportGroup Label=\"ExtensionTargets\">");
            _p(1, "</ImportGroup>");

            _p(0, "</Project>");
        }
    }
}
