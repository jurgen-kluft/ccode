using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace ProjectFileGenerator
{
    public class ProjectFileVs2010
    {
        private StreamWriter mWriter;

        private string tool_version_and_xmlns = "ToolsVersion=\"4.0\" xmlns=\"http://schemas.microsoft.com/developer/msbuild/2003\"";
	    private string xml_version_and_encoding = "<?xml version=\"1.0\" encoding=\"utf-8\"?>";

        public class Project
        {
            public string name;
            public string uuid;

            public List<ConfigInfo> vstudio_configs;
            public Dictionary<string, Config> configs;

            public class ConfigInfo
            {
                public string name;
                public string buildcfg;
                public string platform;
            }

            public enum EConfigType
            {
                SharedLib,
                StaticLib,
                ConsoleApp,
                WindowedApp,
            }

            public class Config
            {
                public string name;             /// Debug|Win32
                public string buildcfg;         /// Debug/Release/Profile/Final
                public string platform;         /// Win32/NintendoDS/NintendoWII/Nintendo3DS/SonyPSP/SonyPS2/SonyPS3/Microsoft360
                public string language;         /// C / C++

                public EConfigType type;
                public List<string> flags;      /// {Optimize / OptimizeSize / OptimizeSpeed}

                public bool StaticRuntime;

                public List<string> defines;
                public List<string> includedirs;
                public List<string> links;

                public bool NoExceptions;
                public bool SEH;

                public bool NoRTTI;

                public bool NativeWChar;
                public bool NoNativeWChar;

                public bool EnableSSE;
                public bool EnableSSE2;

                public bool FloatFast;
                public bool FloatStrict;
                                              
                public bool NoPCH;
                public string pchheader;
                public string pchsource;

                public bool NoMinimalRebuild;

                public bool Symbols;
                public bool DebugInfo;          /// ProgramDatabase/EditAndContinue/OldStyle
                public bool NoEditAndContinue;

                public bool WholeProgramOptimization;

                public bool NoImportLib;

                public string OutDir;
                public string IntDir;
                public string TargetName;

                public bool NoManifest;

                public List<string> libdirs;
                public string OutputFile;

                public string TargetMachine;	/// MachineX86, MachineX64, ...
                                                /// 
                public bool ExtraWarnings;
                public string WarningLevel;
                public bool FatalWarnings;

                public bool NoFramePointer;

                public string CharSet;		/// "Unicode","MultiByte", "NotSet"
                public bool MFC;

                public string WinMain;

                public bool isoptimizedbuild()
                {
                    // Compute from the settings
                    return false;
                }
                public bool isdebugbuild()
                {
	                return false;
                }
                public bool should_link_incrementally()
                {
	                return false;
                }

            }
        }

        public static class Helpers
        {
            // Apply XML escaping to a value.
            public static string esc(string value)
            {
			    value = value.Replace("&",  "&amp;");
			    value = value.Replace("\"", "&quot;");
			    value = value.Replace("'",  "&apos;");
			    value = value.Replace("<",  "&lt;");
			    value = value.Replace(">",  "&gt;");
			    value = value.Replace("\r", "&#x0D;");
			    value = value.Replace("\n", "&#x0A;");
			    return value;
            }

            public static string concat(List<string> list, string seperator)
            {
                string str = string.Empty;
                foreach(string s in list)
                {
                    if (String.IsNullOrEmpty(str))
                        str = s;
                    else
                        str = str + seperator + s;
                }
                return str;
            }

            public static string fix_slash(string path)
            {
                return path.Replace("/", "\\");
            }

            public static string remove_relative_path(string file)
            {
                return System.IO.Path.GetFileName(file);
            }
            public static string file_path(string file)
            {
                return System.IO.Path.GetDirectoryName(file);
            }

		    public static string[] list_of_directories_in_path(string path)
            {
                path = fix_slash(path);
                if (path.IndexOf(":\\")>0)
                {
                    int start = path.IndexOf(":\\");
                    path = path.Substring(start + 2);
                }
                path = path.Trim();
                path = path.TrimEnd('\\');
                string[] list = path.Split(new char[] {'\\'}, StringSplitOptions.RemoveEmptyEntries);
                return list;
            }

	        public static string[] table_of_file_filters(Dictionary<string, List<string>> files)
            {
                Dictionary<string,string> filters = new Dictionary<string,string>();
		        foreach(KeyValuePair<string, List<string>> p in files)
                {
			        foreach(string entry in p.Value)
                    {
				        string[] result = Helpers.list_of_directories_in_path(entry);
				        foreach(string dir in result)
                        {
                            if (!filters.ContainsKey(dir))
                                filters.Add(dir, dir);
                        }
			        }
		        }
		        return filters.Values.ToArray<string>();
	        }

            public static string file_extension(string file)
            {
                return System.IO.Path.GetExtension(file);
            }
	
	
	        // also translates file paths from '/' to '\\'
            private static Dictionary<string, string> mTypes = new Dictionary<string,string>
            {
                    { "h", "ClInclude" },
                    { "hpp", "ClInclude" },
                    { "hxx", "ClInclude" },
                    { "c", "ClCompile" },
                    { "cpp", "ClCompile" },
                    { "cxx", "ClCompile" },
                    { "cc", "ClCompile" },
                    { "rc", "ResourceCompile" }
            };

            public static void sort_input_files(List<string> files, Dictionary<string, List<string>> sorted_container)
            {
                if (!sorted_container.ContainsKey("None"))
                    sorted_container.Add("None", new List<string>());

                foreach(string v in mTypes.Values)
                {
                    if (!sorted_container.ContainsKey(v))
                        sorted_container.Add(v, new List<string>());
                }

                foreach (string current_file in files)
                {
                    string translated_path = current_file.Replace("/", "\\");
                    string ext = Helpers.file_extension(translated_path);
                    if (!String.IsNullOrEmpty(ext))
                    {
                        string type = mTypes[ext];
                        if (String.IsNullOrEmpty(type))
                            type = "None";

                        List<string> f;
                        sorted_container.TryGetValue(type, out f);
                        f.Add(current_file);
                    }
                }
            }
        }

	
	    
        //
        // "Immediate If" - returns one of the two values depending on the value of expr.
        //
	    public string iif(bool expr, string trueval, string falseval)
        {
		    if (expr)
			    return trueval;
		    else
			    return falseval;
        }

        private void _p(int tab, string text)
        {
            _p(tab, text, null);
        }

        private void _p(int tab, string text, params string[] p)
        {
            for (int i=0; i<tab; ++i)
                mWriter.Write("\t");

            int pi = 0;
            int cursor=0;
            while (p!=null && pi<p.Length)
            {
                cursor = text.IndexOf("%s", cursor);
                if (cursor >= 0)
                {
                    text = text.Substring(0, cursor) + p[pi] + text.Substring(cursor+2);
                    ++pi;
                    cursor = cursor + p[pi].Length;
                }
                    break;
            }
            mWriter.WriteLine(text);
        }

	    public void config(Project prj)
        {
		    _p(1,"<ItemGroup Label=\"ProjectConfigurations\">");
		    foreach (Project.ConfigInfo cfginfo in prj.vstudio_configs)
            {
				_p(2,"<ProjectConfiguration Include=\"%s\">", Helpers.esc(cfginfo.name));
					_p(3,"<Configuration>%s</Configuration>", cfginfo.buildcfg);
					_p(3,"<Platform>%s</Platform>", cfginfo.platform);
				_p(2,"</ProjectConfiguration>");
		    }
		    _p(1,"</ItemGroup>");
	    }

        public void globals(Project prj)
        {
		    _p(1,"<PropertyGroup Label=\"Globals\">");
			    _p(2,"<ProjectGuid>{%s}</ProjectGuid>", prj.uuid);
			    _p(2,"<RootNamespace>%s</RootNamespace>", prj.name);
			    _p(2,"<Keyword>Win32Proj</Keyword>");
		    _p(1,"PropertyGroup>");
        }

	    public string config_type(Project.Config cfg)
        {
            string t = string.Empty;
            switch (cfg.type)
            {
                case Project.EConfigType.SharedLib: t = "DynamicLibrary"; break;
                case Project.EConfigType.StaticLib: t = "StaticLibrary";break;
                case Project.EConfigType.ConsoleApp: t = "Application";break;
                case Project.EConfigType.WindowedApp: t = "Application";break;
            }
            return t;
        }

	    public string if_config_and_platform()
        {
            return "Condition=\"'$(Configuration)|$(Platform)'=='%s'\"";
        }
	
	    public string optimisation(Project.Config cfg)
        {
            string result = "Disable";
		    foreach(string value in cfg.flags)
            {
			    if (value == "Optimize")
				    result = "Full";
			    else if (value == "OptimizeSize")
				    result = "MinSpace";
			    else if (value == "OptimizeSpeed")
				    result = "MaxSpeed";
            }
		    return result;
        }

		public void config_type(Project prj)
        {
		    foreach(Project.ConfigInfo cfginfo in prj.vstudio_configs)
			{
                Project.Config cfg;
                prj.configs.TryGetValue(cfginfo.buildcfg, out cfg);
			    
                _p(1,"<PropertyGroup " + if_config_and_platform() + " Label=\"Configuration\">", Helpers.esc(cfginfo.name));
				    _p(2,"<ConfigurationType>%s</ConfigurationType>", config_type(cfg));
				    _p(2,"<CharacterSet>%s</CharacterSet>", cfg.CharSet);
			
			    if (cfg.MFC)
				    _p(2,"<UseOfMfc>Dynamic</UseOfMfc>");
			
			    string use_debug = "false";
			    if (cfg.isoptimizedbuild())
				    use_debug = "true";
			    else
				    _p(2,"<WholeProgramOptimization>true</WholeProgramOptimization>");

				    _p(2,"<UseDebugLibraries>%s</UseDebugLibraries>", use_debug);
			    _p(1,"</PropertyGroup>");
            }
        }

	    public void config_type_block(Project prj)
        {
		    foreach(Project.ConfigInfo cfginfo in prj.vstudio_configs)
            {
                Project.Config cfg;
                prj.configs.TryGetValue(cfginfo.buildcfg, out cfg);

			    _p(1,"<PropertyGroup " + if_config_and_platform()  + " Label=\"Configuration\">", Helpers.esc(cfginfo.name));
				    _p(2,"<ConfigurationType>%s</ConfigurationType>", config_type(cfg));
				    _p(2,"<CharacterSet>%s</CharacterSet>", cfg.CharSet);
			
			    if (cfg.MFC)
				    _p(2,"<UseOfMfc>Dynamic</UseOfMfc>");
			
			    string use_debug = "false";
			    if (cfg.isoptimizedbuild())
				    use_debug = "true";
				
                _p(2,"<WholeProgramOptimization>%s</WholeProgramOptimization>", cfg.WholeProgramOptimization ? "true" : "false");
                _p(2,"<UseDebugLibraries>%s</UseDebugLibraries>", use_debug);

			    _p(1,"</PropertyGroup>");
            }
        }
	
	    public void import_props(Project prj)
        {
		    foreach(Project.ConfigInfo cfginfo in prj.vstudio_configs)
            {
                Project.Config cfg;
                prj.configs.TryGetValue(cfginfo.buildcfg, out cfg);

			    _p(1,"<ImportGroup " + if_config_and_platform()  + " Label=\"PropertySheets\">", Helpers.esc(cfginfo.name));
				    _p(2,"<Import Project=\"$(UserRootDir)\\Microsoft.Cpp.$(Platform).user.props\" Condition=\"exists('$(UserRootDir)\\Microsoft.Cpp.$(Platform).user.props')\" Label=\"LocalAppDataPlatform\" />");
			    _p(1,"</ImportGroup>");
             }
        }
	    public void incremental_link(Project.Config cfg, Project.ConfigInfo cfginfo)
        {
		    if (cfg.type != Project.EConfigType.StaticLib)
                _p(2, "<LinkIncremental " + if_config_and_platform() + ">%s</LinkIncremental>", Helpers.esc(cfginfo.name), cfg.should_link_incrementally() ? "true" : "false");
        }		
		
	    public void ignore_import_lib(Project.Config cfg, Project.ConfigInfo cfginfo)
        {
		    if (cfg.type != Project.EConfigType.SharedLib)
            {
                string shouldIgnore = "false";
			    if (cfg.NoImportLib) 
                    shouldIgnore = "true";
			     _p(2,"<IgnoreImportLibrary " + if_config_and_platform()  + ">%s</IgnoreImportLibrary>", Helpers.esc(cfginfo.name), shouldIgnore);
            }	
        }

        
	    public void clcompile(Project.Config cfg)
        {
		    _p(2,"<ClCompile>");
		
		    if (cfg.buildoptions.Count > 0)
			    _p(3,"<AdditionalOptions>%s %%(AdditionalOptions)</AdditionalOptions>", Helpers.concat(Helpers.esc(cfg.buildoptions), " "));
		
		    _p(3,"<Optimization>%s</Optimization>", optimisation(cfg));
		
			include_dirs(3,cfg);
			preprocessor(3,cfg);
			minimal_build(cfg);
		
		    if (!cfg.isoptimizedbuild())
            {
			    _p(3,"<BasicRuntimeChecks>EnableFastChecks</BasicRuntimeChecks>");
			    if (cfg.ExtraWarnings)
				    _p(3,"<SmallerTypeCheck>true</SmallerTypeCheck>");
            }
		    else
            {
			    _p(3,"<StringPooling>true</StringPooling>");
            }
	
			_p(3,"<RuntimeLibrary>%s</RuntimeLibrary>", runtime(cfg));
		
			_p(3,"<FunctionLevelLinking>true</FunctionLevelLinking>");
			
			precompiled_header(cfg);
		
			_p(3,"<WarningLevel>%s</WarningLevel>", cfg.WarningLevel);
			
		    if (cfg.FatalWarnings)
			    _p(3,"<TreatWarningAsError>true</TreatWarningAsError>");
	
			exceptions(cfg);
			rtti(cfg);
			wchar_t_buildin(cfg);
			sse(cfg);
			floating_point(cfg);
			debug_info(cfg);
		
		    if (cfg.NoFramePointer)
			    _p(3,"<OmitFramePointers>true</OmitFramePointers>");
			
			compile_language(cfg);

		    _p(2,"</ClCompile>");
        }

	    public void event_hooks(Project.Config cfg)	
        {
		    if (cfg.postbuildcommands.Count > 0)
            {
                _p(2,"<PostBuildEvent>");
				    _p(3,"<Command>%s</Command>", Helpers.esc(table.implode(cfg.postbuildcommands, "", "", "\r\n")));
			    _p(2,"</PostBuildEvent>");
		    }
		
		    if (cfg.prebuildcommands.Count > 0)
            {
                _p(2,"<PreBuildEvent>");
				    _p(3,"<Command>%s</Command>", Helpers.esc(table.implode(cfg.prebuildcommands, "", "", "\r\n")));
			    _p(2,"</PreBuildEvent>");
		    }
		
		    if (cfg.prelinkcommands.Count > 0)
            {
                _p(2,"<PreLinkEvent>");
				    _p(3,"<Command>%s</Command>", Helpers.esc(table.implode(cfg.prelinkcommands, "", "", "\r\n")));
			    _p(2,"</PreLinkEvent>");
		    }
        }

	    public void additional_options(int indent, Project.Config cfg)
        {
		    if (cfg.linkoptions > 0)
            {
                _p(indent, "<AdditionalOptions>%s %%(AdditionalOptions)</AdditionalOptions>", Helpers.concat(Helpers.esc(cfg.linkoptions), " "));
            }
        }
		
	    public void item_def_lib(Project.Config cfg)
        {
		    if (cfg.type == Project.EConfigType.StaticLib)
            {
			    _p(1,"<Lib>");
				    _p(2,"<OutputFile>$(OutDir)%s</OutputFile>", cfg.buildtarget.name);
                additional_options(2,cfg);
			    _p(1,"</Lib>");
            }
        }
	
	    public void link_target_machine(Project.Config cfg)
        {
		    string target = string.Empty;
		    if (cfg.platform == null || cfg.platform == "x32")
                target ="MachineX86";
		    else if (cfg.platform == "x64")
                target ="MachineX64";

		    _p(3,"<TargetMachine>%s</TargetMachine>", target);
	    }

        public void import_lib(Project.Config cfg)
        {
		    // Prevent the generation of an import library for a Windows DLL.
		    if (cfg.type == Project.EConfigType.SharedLib)
            {
			    string implibname = cfg.linktarget.fullpath;
                _p(3, "<ImportLibrary>%s</ImportLibrary>", iif(cfg.NoImportLib, cfg.IntDir + "\\" + System.IO.Path.GetFileNameWithoutExtension(implibname), implibname));
            }
        }

        public void common_link_section(Project.Config cfg)
        {
		    _p(3,"<SubSystem>%s</SubSystem>", iif(cfg.type == Project.EConfigType.ConsoleApp, "Console", "Windows"));
		
		    if (cfg.Symbols)
			    _p(3,"<GenerateDebugInformation>true</GenerateDebugInformation>");
		    else
			    _p(3,"<GenerateDebugInformation>false</GenerateDebugInformation>");

            if (cfg.isoptimizedbuild())
            {
                _p(3, "<OptimizeReferences>true</OptimizeReferences>");
                _p(3, "<EnableCOMDATFolding>true</EnableCOMDATFolding>");
            }
		
		    if (cfg.Symbols)
			    _p(3,"<ProgramDataBaseFileName>$(OutDir)%s.pdb</ProgramDataBaseFileName>", System.IO.Path.GetFileNameWithoutExtension(cfg.TargetName));
	    }
		
	    public void intermediate_and_out_dirs(Project prj)
        {
		    _p(1,"<PropertyGroup>");
			    _p(2,"<_ProjectFileVersion>10.0.30319.1</_ProjectFileVersion>");
			
    	    foreach(Project.ConfigInfo cfginfo in prj.vstudio_configs)
            {
                Project.Config cfg;
                prj.configs.TryGetValue(cfginfo.buildcfg, out cfg);

				_p(2,"<OutDir " + if_config_and_platform()  + ">%s\\</OutDir>", Helpers.esc(cfginfo.name), Helpers.esc(cfg.OutDir));
				_p(2,"<IntDir " + if_config_and_platform()  + ">%s\\</IntDir>", Helpers.esc(cfginfo.name), Helpers.esc(cfg.IntDir));
                _p(2, "<TargetName " + if_config_and_platform() + ">%s</TargetName>", Helpers.esc(cfginfo.name), Helpers.esc(cfg.TargetName));

				ignore_import_lib(cfg, cfginfo);
				incremental_link(cfg, cfginfo);
				if (cfg.NoManifest)
				    _p(2,"<GenerateManifest " + if_config_and_platform()  + ">false</GenerateManifest>", Helpers.esc(cfginfo.name));
            }
		    _p(1,"</PropertyGroup>");
        }	
	
	    public string runtime(Project.Config cfg)
        {
		    string runtime = string.Empty;
		    if (cfg.isdebugbuild())
			    runtime = iif(cfg.StaticRuntime, "MultiThreadedDebug", "MultiThreadedDebugDLL");
		    else
			    runtime = iif(cfg.StaticRuntime, "MultiThreaded"     , "MultiThreadedDLL");

		    return runtime;
        }
	
	    public void precompiled_header(Project.Config cfg)
        {
      	    if (!cfg.NoPCH && !String.IsNullOrEmpty(cfg.pchheader))
            {
			    _p(3,"<PrecompiledHeader>Use</PrecompiledHeader>");
			    _p(3,"<PrecompiledHeaderFile>%s</PrecompiledHeaderFile>", cfg.pchheader);
            }
		    else
            {
			    _p(3,"<PrecompiledHeader></PrecompiledHeader>");
            }
        }

	    public void preprocessor(int indent, Project.Config cfg)
        {
            if (cfg.defines.Count > 0)
                _p(indent, "<PreprocessorDefinitions>%s;%%(PreprocessorDefinitions)</PreprocessorDefinitions>", Helpers.esc(Helpers.concat(cfg.defines, ";")));
            else
                _p(indent, "<PreprocessorDefinitions></PreprocessorDefinitions>");
        }

	    public void include_dirs(int indent, Project.Config cfg)
        {
		    if (cfg.includedirs.Count > 0)
			    _p(indent, "<AdditionalIncludeDirectories>%s;%%(AdditionalIncludeDirectories)</AdditionalIncludeDirectories>", Helpers.esc(Helpers.fix_slash(Helpers.concat(cfg.includedirs, ";"))));
        }
	
	    public void resource_compile(Project.Config cfg)
        {
		    _p(2,"<ResourceCompile>");
			    preprocessor(3,cfg);
                include_dirs(3, cfg);
		    _p(2,"</ResourceCompile>");
        }
	
	    public void exceptions(Project.Config cfg)
        {
		    if (cfg.NoExceptions)
			    _p(2,"<ExceptionHandling>false</ExceptionHandling>");
		    else if (cfg.SEH)
			    _p(2,"<ExceptionHandling>Async</ExceptionHandling>");
        }
	
	    public void rtti(Project.Config cfg)
		{
            if (cfg.NoRTTI)
			    _p(3,"<RuntimeTypeInfo>false</RuntimeTypeInfo>");
        }	

	    public void wchar_t_buildin(Project.Config cfg)
        {
		    if (cfg.NativeWChar)
			    _p(3,"<TreatWChar_tAsBuiltInType>true</TreatWChar_tAsBuiltInType>");
		    else if (cfg.NoNativeWChar)
			    _p(3,"<TreatWChar_tAsBuiltInType>false</TreatWChar_tAsBuiltInType>");
        }
	
	    public void sse(Project.Config cfg)
        {
		    if (cfg.EnableSSE)
			    _p(3,"<EnableEnhancedInstructionSet>StreamingSIMDExtensions</EnableEnhancedInstructionSet>");
		    else if (cfg.EnableSSE2)
			    _p(3,"<EnableEnhancedInstructionSet>StreamingSIMDExtensions2</EnableEnhancedInstructionSet>");
        }
	
	    public void floating_point(Project.Config cfg)
        {
            if (cfg.FloatFast)
			    _p(3,"<FloatingPointModel>Fast</FloatingPointModel>");
		    else if (cfg.FloatStrict)
			    _p(3,"<FloatingPointModel>Strict</FloatingPointModel>");
        }

	    public void debug_info(Project.Config cfg)
        {
	        //
	        //	EditAndContinue /ZI
	        //	ProgramDatabase /Zi
	        //	OldStyle C7 Compatable /Z7
	        //
		    string debug_info = string.Empty;
		    if (cfg.DebugInfo)
            {
			    if (cfg.isoptimizedbuild() || cfg.NoEditAndContinue)
				    debug_info = "ProgramDatabase";
			    else if (cfg.platform != "x64")
				    debug_info = "EditAndContinue";
			    else
				    debug_info = "OldStyle";
            }		
		    _p(3,"<DebugInformationFormat>%s</DebugInformationFormat>", debug_info);
	    }
	
	    public void minimal_build(Project.Config cfg)
        {
		    _p(3,"<MinimalRebuild>%s</MinimalRebuild>", cfg.NoMinimalRebuild ? "false":"true");
        }
	
	    public void compile_language(Project.Config cfg)
        {
		    if (cfg.language == "C")
			    _p(3,"<CompileAs>CompileAsC</CompileAs>");
        }

        public void item_link(Project.Config cfg)
        {
		    _p(2,"<Link>");
		    if (cfg.type != Project.EConfigType.StaticLib)
            {
			    if (cfg.links.Count > 0)
                {
				    _p(3,"<AdditionalDependencies>%s;%%(AdditionalDependencies)</AdditionalDependencies>", Helpers.concat(cfg.links, ";"));
			    }
				_p(3,"<OutputFile>$(OutDir)%s</OutputFile>", cfg.OutputFile);

                _p(3, "<AdditionalLibraryDirectories>%s%s%%(AdditionalLibraryDirectories)</AdditionalLibraryDirectories>", Helpers.esc(Helpers.fix_slash(Helpers.concat(cfg.libdirs, ";"))), iif(cfg.libdirs.Count > 0, ";", ""));
							
			    common_link_section(cfg);
			
			    if (config_type(cfg) == "Application" && String.IsNullOrEmpty(cfg.WinMain))
				{
                    _p(3,"<EntryPointSymbol>mainCRTStartup</EntryPointSymbol>");
			    }

			    import_lib(cfg);
			
			    _p(3,"<TargetMachine>%s</TargetMachine>", cfg.TargetMachine);
			
			    additional_options(3,cfg);
            }
		    else
            {
                common_link_section(cfg);
		    }
		
		    _p(2,"</Link>");
	    }

        public void item_definitions(Project prj)
        {
		    foreach(Project.ConfigInfo cfginfo in prj.vstudio_configs)
            {
                Project.Config cfg;
                prj.configs.TryGetValue(cfginfo.buildcfg, out cfg);

                _p(1, "<ItemDefinitionGroup " + if_config_and_platform() + ">", Helpers.esc(cfginfo.name));
				    clcompile(cfg);
				    resource_compile(cfg);
				    item_def_lib(cfg);
				    item_link(cfg);
				    event_hooks(cfg);
			    _p(1,"</ItemDefinitionGroup>");
            }
        }

	    public void write_file_type_block(List<string> files, string group_type)
        {
		    if (files.Count > 0)
            {
                _p(1,"<ItemGroup>");
			    foreach(string current_file in files)
                {
				    _p(2,"<%s Include=\"%s\" />", group_type, current_file);
			    }
			    _p(1,"</ItemGroup>");
		    }
        }

        public void write_file_compile_block(List<string> files, Project prj, List<Project.ConfigInfo> configs)
        {
		    if (files.Count > 0)
            {
		        Dictionary<string, string> config_mappings = new Dictionary<string,string>();
			    foreach (Project.ConfigInfo cfginfo in configs)
				{
                    Project.Config cfg;
                    prj.configs.TryGetValue(cfginfo.buildcfg, out cfg);

				    if (!String.IsNullOrEmpty(cfg.pchheader) && !String.IsNullOrEmpty(cfg.pchsource) && !cfg.NoPCH)
                    {
					    config_mappings[cfginfo.name] = Helpers.fix_slash(cfg.pchsource);
				    }
                }
			
			    _p(1,"<ItemGroup>");
			    foreach(string current_file in files)
                {
                    bool written = false;
                    foreach (Project.ConfigInfo cfginfo in configs)
                    {
					    if (config_mappings.ContainsKey(cfginfo.name) && current_file == config_mappings[cfginfo.name])
                        {
							// only one source file per pch
							config_mappings.Remove(cfginfo.name);

        				    _p(2,"<ClCompile Include=\"%s\" />", current_file);
							    _p(3,"<PrecompiledHeader " + if_config_and_platform() + ">Create</PrecompiledHeader>", Helpers.esc(cfginfo.name));
					        _p(2,"</ClCompile>");
                            written = true;
                            break;
                        }
				    }
                    if (!written)
				    {
    				    _p(2,"<ClCompile Include=\"%s\" />", current_file);
                    }
			    }
			    _p(1,"</ItemGroup>");
		    }
	    }

        public void vcxproj_files(Project prj)
        {
            Dictionary<string, List<string>> sorted = new Dictionary<string, List<string>>();
            sorted.Add("ClCompile", new List<string>());
            sorted.Add("ClInclude", new List<string>());
            sorted.Add("None", new List<string>());
            sorted.Add("ResourceCompile", new List<string>());
		
		    Project.Config cfg = premake.getconfig(prj);
		    Helpers.sort_input_files(cfg.files, sorted);

		    write_file_type_block(sorted["ClInclude"],"ClInclude");
		    write_file_compile_block(sorted["ClCompile"], prj, prj.vstudio_configs);
		    write_file_type_block(sorted["None"], "None");
		    write_file_type_block(sorted["ResourceCompile"], "ResourceCompile");
	    }

        public void write_filter_includes(Dictionary<string, List<string>> sorted_table)
        {
		    string[] directories = Helpers.table_of_file_filters(sorted_table);

		    // I am going to take a punt here that the ItemGroup is missing if no files!!!!
		    if (directories.Length > 0)
            {
			    _p(1,"<ItemGroup>");
			    foreach (string dir in directories)
                {
				    _p(2,"<Filter Include=\"%s\">", dir);
					    _p(3,"<UniqueIdentifier>{%s}</UniqueIdentifier>", Guid.NewGuid().ToString());
				    _p(2,"</Filter>");
			    }
			    _p(1,"</ItemGroup>");
		    }
	    }

        public void write_file_filter_block(List<string> files, string group_type)
        {
		    if (files.Count > 0)
            {
			    _p(1,"<ItemGroup>");
			    foreach(string current_file in files)
                {
				    string path_to_file = Helpers.file_path(current_file);
				    if (!String.IsNullOrEmpty(path_to_file))
                    {
                        _p(2,"<%s Include=\"%s\">", group_type, Helpers.fix_slash(current_file));
						    _p(3,"<Filter>%s</Filter>", path_to_file);
					    _p(2,"</%s>",group_type);
				    }
                    else
                    {
					    _p(2,"<%s Include=\"%s\" />", group_type, Helpers.fix_slash(current_file));
				    }
			    }
			    _p(1,"</ItemGroup>");
		    }
        }
        
	    public void vcxproj_filter_files(Project prj)
        {
            Dictionary<string, List<string>> sorted = new Dictionary<string, List<string>>();
            sorted.Add("ClCompile", new List<string>());
            sorted.Add("ClInclude", new List<string>());
            sorted.Add("None", new List<string>());
            sorted.Add("ResourceCompile", new List<string>());
		
		    Project.Config cfg = premake.getconfig(prj);
		    Helpers.sort_input_files(cfg.files, sorted);

		    _p(0, xml_version_and_encoding);
		    _p(0, "<Project " + tool_version_and_xmlns + ">");
			    write_filter_includes(sorted);
                write_file_filter_block(sorted["ClInclude"], "ClInclude");
			    write_file_filter_block(sorted["ClCompile"], "ClCompile");
			    write_file_filter_block(sorted["None"], "None");
			    write_file_filter_block(sorted["ResourceCompile"], "ResourceCompile");
		    _p(0, "</Project>");
        }

	    public void vcxproj(Project prj)
        {
		    _p(0, xml_version_and_encoding);
		    _p(0, "<Project DefaultTargets=\"Build\" " + tool_version_and_xmlns + ">");
			    config(prj);
			    globals(prj);
			
			    _p(1,"<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.Default.props\" />");
			
			    config_type_block(prj);
			
			    _p(1,"<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.props\" />");
			
			    // check what this section is doing
			    _p(1,"<ImportGroup Label=\"ExtensionSettings\">");
			    _p(1,"</ImportGroup>");
						
			    import_props(prj);
			
			    // what type of macros are these?
			    _p(1,"<PropertyGroup Label=\"UserMacros\" />");
			
			    intermediate_and_out_dirs(prj);
			
			    item_definitions(prj);
			
			    vcxproj_files(prj);

			    _p(1,"<Import Project=\"$(VCTargetsPath)\\Microsoft.Cpp.targets\" />");
			    _p(1,"<ImportGroup Label=\"ExtensionTargets\">");
			    _p(1,"</ImportGroup>");

		    _p(0,"</Project>");
        }
	

	    public void vcxproj_user(Project prj)
        {
		    _p(0, xml_version_and_encoding);
		    _p(0, "<Project " + tool_version_and_xmlns + ">");
		    _p(0, "</Project>");
        }
	
	    public void vcxproj_filters(Project prj)
        {
            vcxproj_filter_files(prj);
        }
    }
}
