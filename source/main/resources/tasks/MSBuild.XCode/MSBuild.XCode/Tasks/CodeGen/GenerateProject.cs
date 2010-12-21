using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Reflection;
using System.Text;
using System.Text.RegularExpressions;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    /// </summary>
    public class GenerateProject : Task
    {
        #region Public properties

        /// <summary>
        /// Path to the package
        /// </summary>
        [Required()]
        public string RootDir { get; set; }

        /// <summary>
        /// Path to the templates
        /// </summary>
        [Required()]
        public string TemplateDir { get; set; }
       
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
            Log.LogMessage("Template directory = " + TemplateDir);
#endif
            // Load package
            // Load full pom
            // Save project(s)

            return success;
        }


        #endregion
    }
}