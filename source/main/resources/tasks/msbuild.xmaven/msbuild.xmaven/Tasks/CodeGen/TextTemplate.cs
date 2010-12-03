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

namespace msbuild.xmaven
{
    /// <summary>
    /// </summary>
    public class TextTemplate : Task
    {
        #region Private instance fields

        private string _templatePath;
        private string _outputPath;

        private ITaskItem[] _vars;

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
        public ITaskItem[] Vars
        {
            get { return this._vars; }
            set { this._vars = value; }
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


            return success;
        }


        #endregion
    }

}