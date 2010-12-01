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
    /// Need:
    /// - Location of top package main and test project files
    /// - Guid of top package main and test project files
    /// - Project configurations of top package main and test project files
    /// - Dependencies of top package
    /// 
    /// Using the dependencies we need to collect
    /// 
    /// </summary>
    public class GenerateCppSolution : Task
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

    class CppSolution
    {
   }
}