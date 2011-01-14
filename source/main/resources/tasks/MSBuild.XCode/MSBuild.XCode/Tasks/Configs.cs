using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System;
using System.IO;
using System.Collections.Generic;
using System.Security.Cryptography;
using System.Linq;
using System.Text;
using System.Runtime;
using Ionic.Zip;
using Ionic.Zlib;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageConfigs : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string Platform { get; set; }
        [Required]
        public string ProjectGroup { get; set; }
        [Required]
        public string TemplateDir { get; set; }

        [Output]
        public string[] Configurations { get; set; }

        public override bool Execute()
        {
            bool success = false;
            Loggy.TaskLogger = Log;

            RootDir = RootDir.EndWith('\\');

            Environment.CurrentDirectory = RootDir;

            PackageInstance.TemplateDir = TemplateDir;
            PackageInstance.Initialize();

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            if (package.IsValid)
            {
                // Get all platforms and configs, e.g: DevDebug|Win32;DevRelease|Win32;DevFinal|Win32
                string[] configs = package.Pom.GetConfigsForPlatformsForGroup(Platform, ProjectGroup);
                Configurations = configs;
                success = true;
            }
            else
            {
                Loggy.Error(String.Format("Error: Loading package failed in Package::Configs"));
            }

            return success;
        }
    }
}
