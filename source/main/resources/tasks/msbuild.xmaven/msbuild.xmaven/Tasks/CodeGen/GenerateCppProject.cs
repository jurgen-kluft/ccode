using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Reflection;
using System.Text;
using System.Text.RegularExpressions;
using msbuild.xmaven.helpers;

namespace msbuild.xmaven.tasks
{
    /// <summary>
    /// </summary>
    public class GenerateCppProject : Task
    {
        #region Private instance fields

        private string _templatePath;
        private string _outputPath;

        private ITaskItem[] _configs = new ITaskItem[0];
        private ITaskItem[] _platforms = new ITaskItem[0];
        private ITaskItem[] _configurations = new ITaskItem[0];

        private string _guid;
        private string _name;

        #endregion
        #region Public properties

        /// <summary>
        /// Path to the T4 template to be executed
        /// </summary>
        [Required()]
        public string TemplatePath
        {
            get { return this._templatePath; }
            set { this._templatePath = value; }
        }

        /// <summary>
        /// Path where the executed T4 output should be created
        /// </summary>
        [Required()]
        public string OutputPath
        {
            get { return this._outputPath; }
            set { this._outputPath = value; }
        }

        /// <summary>
        /// The configurations, Debug,Release,Profile,Final
        /// </summary>
        [Required()]
        public ITaskItem[] Configs
        {
            get { return this._configs; }
            set { this._configs = value; }
        }

        /// <summary>
        /// The platforms, Win32,NintendoWII,Nintendo3DS,...
        /// </summary>
        [Required()]
        public ITaskItem[] Platforms
        {
            get { return this._platforms; }
            set { this._platforms = value; }
        }

        /// <summary>
        /// The configuration blocks
        /// </summary>
        [Required()]
        public ITaskItem[] Configurations
        {
            get { return this._configurations; }
            set { this._configurations = value; }
        }

        /// <summary>
        /// The project name
        /// </summary>
        [Required()]
        public string ProjectName
        {
            get { return this._name; }
            set { this._name = value; }
        }

        /// <summary>
        /// The project GUID
        /// </summary>
        [Required()]
        public string ProjectGuid
        {
            get { return this._guid; }
            set { this._guid = value; }
        }

        #endregion
        #region Public instance methods

        /// <summary>
        /// Replaces property and item group tags in the template with the resolved values
        /// from MSBuild then executes the template using TextTransform.exe
        /// </summary>
        /// <returns>Whether the execution was successful</returns>
        public override bool Execute()
        {
            bool success = false;
#if DEBUG
            Log.LogMessage("Template path = " + TemplatePath);
            Log.LogMessage("Output path = " + OutputPath);
            Log.LogMessage("Project name = " + ProjectName);
            Log.LogMessage("Project guid = " + ProjectGuid);
#endif

            int i = 0;
            string[] configs = new string[_configs.Length];
#if DEBUG
            Log.LogMessage("Configs " + _configs.Length.ToString());
#endif
            foreach (ITaskItem item in _configs)
            {
                Log.LogMessage(item.ToString());
                configs[i++] = item.ToString();
            }

            i = 0;
            string[] platforms = new string[_platforms.Length];
#if DEBUG
            Log.LogMessage("Platforms " + _platforms.Length.ToString());
#endif
            foreach (ITaskItem item in _platforms)
            {
                Log.LogMessage(item.ToString());
                platforms[i++] = item.ToString();
            }

#if DEBUG
            Log.LogMessage("Configurations " + _configurations.Length.ToString());
#endif

            // Build
            CppProject project = new CppProject(ProjectName, ProjectGuid, configs, platforms);
            project.Preprocess(TemplatePath);

            foreach (ITaskItem item in _configurations)
            {
                string key;
                Dictionary<string, string> variables = project.ConstructVars(item, out key);
                project.AddVars(key, variables);

#if DEBUG
                Log.LogMessage("ID:" + key);
                foreach (KeyValuePair<string, string> kvp in variables)
                {
                    Log.LogMessage("Key:" + kvp.Key + " = " + kvp.Value);
                }
#endif
            }

            _outputPath = OutputPath;
            if (!_outputPath.EndsWith("\\"))
                _outputPath = _outputPath + "\\";
            _outputPath = _outputPath + ProjectName + ".vcxproj";
            project.Generate(_outputPath);

            return success;
        }


        #endregion
    }

    class CppProject
    {
        private string[] mConfigs;
        private string[] mPlatforms;

        private List<string> mProject;
        private List<string> mBlockNames;

        private Dictionary<string, string> mGlobalVariables;
        private Dictionary<string, Dictionary<string, string>> mVariables;
        private Dictionary<string, string[]> mBlocks;

        /// <summary>
        /// This will generate a C++ project (vcxproj) file
        /// </summary>
        /// <param name="name">The project name: xbase</param>
        /// <param name="guid">The project guid: {8da40dc6-5c53-4b55-8e5e-387ab7b87263}</param>
        /// <param name="type">The project type; StaticLibrary, Application, DynamicLibrary</param>
        /// <param name="configs">The project configurations: Debug,Release,Profile,Final</param>
        /// <param name="platforms">The project platforms: Win32, NintendoWII, SonyPS3</param>
        public CppProject(string name, string guid, string[] configs, string[] platforms)
        {
            mConfigs = configs;
            mPlatforms = platforms;

            mGlobalVariables = new Dictionary<string, string>();
            mGlobalVariables.Add("ProjectName", name);
            mGlobalVariables.Add("ProjectGuid", guid);
        }

        private static List<string> LoadMain(string filename)
        {
            List<string> vcxProj = new List<string>();

            FileStream fs = new FileStream(filename, FileMode.Open);
            StreamReader reader = new StreamReader(fs);
            while (!reader.EndOfStream)
            {
                string line = reader.ReadLine();
                vcxProj.Add(line);
            }

            reader.Close();
            fs.Close();
            return vcxProj;
        }

        private static List<string> ExtractBlockNames(List<string> main)
        {
            List<string> blockNames = new List<string>();
            foreach (string l in main)
            {
                string line = l.Trim();
                if (line.StartsWith("@"))
                {
                    string[] filenames = StringTools.Between(line, "@(", ")@");
                    foreach (string f in filenames)
                        blockNames.Add(f.Trim());
                }
            }
            return blockNames;
        }

        private static string[] LoadBlock(string name, string filename)
        {
            List<string> lines = new List<string>();
            {
                FileStream fs = new FileStream(filename, FileMode.Open);
                StreamReader reader = new StreamReader(fs);

                while (!reader.EndOfStream)
                {
                    string line = reader.ReadLine().Trim();
                    lines.Add(line);
                }

                reader.Close();
                fs.Close();
            }
            return lines.ToArray();
        }

        public Dictionary<string, string> ConstructVars(ITaskItem item, out string key)
        {
            key = item.ToString();

            Dictionary<string, string> variables = new Dictionary<string, string>();

            string[] config_platform = key.Split(new char[] { '|' }, StringSplitOptions.RemoveEmptyEntries);
            string config = config_platform[0];
            string platform = config_platform[1];

            variables.Add("ProjectConfig", config);
            variables.Add("ProjectPlatform", platform);

            foreach (string name in item.MetadataNames)
            {
                string value = string.Empty;
                try
                {
                    value = item.GetMetadata(name);
                    variables.Add(name, value);
                }
                catch (Exception)
                {
                }
            }
            return variables;
        }

        public void AddVars(string key, Dictionary<string, string> variables)
        {
            if (!mVariables.ContainsKey(key))
                mVariables.Add(key, variables);
        }

        public void Preprocess(string path)
        {
            mProject = LoadMain(path + "\\vcxproj.main");
            mBlockNames = ExtractBlockNames(mProject);
            mVariables = new Dictionary<string, Dictionary<string, string>>();

            mBlocks = new Dictionary<string, string[]>();
            foreach (string f in mBlockNames)
            {
                if (File.Exists(path + "\\" + f + ".block"))
                {
                    string[] block = LoadBlock(f, path + "\\" + f + ".block");
                    foreach (string c in mConfigs)
                    {
                        foreach (string p in mPlatforms)
                        {
                            string key = f + "." + c + "|" + p;
                            // Register this block for all config|platform
                            if (mBlocks.ContainsKey(key))
                                mBlocks.Remove(key);
                            mBlocks.Add(key, block);
                        }
                    }
                }

                foreach (string c in mConfigs)
                {
                    if (File.Exists(path + "\\" + f + "." + c + ".block"))
                    {
                        string[] block = LoadBlock(f, path + "\\" + f + "." + c + ".block");
                        foreach (string p in mPlatforms)
                        {
                            string key = f + "." + c + "|" + p;
                            // Register this block for all platforms
                            if (mBlocks.ContainsKey(key))
                                mBlocks.Remove(key);
                            mBlocks.Add(key, block);
                        }
                    }

                    foreach (string p in mPlatforms)
                    {
                        if (File.Exists(path + "\\" + f + "." + c + "." + p + ".block"))
                        {
                            string[] block = LoadBlock(f, path + "\\" + f + "." + c + "." + p + ".block");

                            string key = f + "." + c + "|" + p;
                            if (mBlocks.ContainsKey(key))
                                mBlocks.Remove(key);
                            mBlocks.Add(key, block);
                        }
                    }
                }
            }

        }

        public string WriteBlockLine(string line, Dictionary<string, string> vars)
        {
            if (line.Contains("##"))
            {
                foreach (KeyValuePair<string, string> p in vars)
                {
                    line = line.Replace("##" + p.Key + "##", p.Value);
                }
            }
            return line;
        }

        public void WriteBlock(List<string> outLines, string[] block, Dictionary<string, string> vars)
        {
            foreach (string l in block)
            {
                string line = WriteBlockLine(l, vars);
                outLines.Add(line);
            }
        }

        public void Generate(string filename)
        {
            List<string> outLines = new List<string>();

            foreach (string l in mProject)
            {
                string line = l.Trim();
                if (line.StartsWith("@"))
                {
                    string[] blockName = StringTools.Between(line, "@(", ")@");
                    if (blockName.Length == 1)
                    {
                        string f = blockName[0];
                        foreach (string c in mConfigs)
                        {
                            foreach (string p in mPlatforms)
                            {
                                string[] block;
                                string key1 = f + "." + c + "|" + p;
                                mBlocks.TryGetValue(key1, out block);

                                Dictionary<string, string> vars;
                                string key2 = c + "|" + p;
                                mVariables.TryGetValue(key2, out vars);
                                WriteBlock(outLines, block, vars);
                            }
                        }
                    }
                }
                else
                {
                    if (line.Contains("##"))
                    {
                        foreach (KeyValuePair<string, string> p in mGlobalVariables)
                        {
                            line = line.Replace("##" + p.Key + "##", p.Value);
                        }
                    }

                    outLines.Add(line);
                }
            }

            FileStream fs = new FileStream(filename, FileMode.Create, FileAccess.Write);
            StreamWriter writer = new StreamWriter(fs);
            foreach (string l in outLines)
            {
                writer.WriteLine(l);
            }
            writer.Close();
            fs.Close();
        }
    }
}