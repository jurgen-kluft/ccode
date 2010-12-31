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
        public string Category { get; set; }
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

            Global.TemplateDir = TemplateDir;
            Global.CacheRepoDir = string.Empty;
            Global.RemoteRepoDir = string.Empty;
            Global.Initialize();

            Package package = new Package();
            package.IsRoot = true;
            package.RootDir = RootDir;
            if (package.LoadPom())
            {
                // Get all platforms and configs, e.g: DevDebug|Win32;DevRelease|Win32;DevFinal|Win32
                string[] configs = package.Pom.GetConfigsForPlatformsForGroup(Platform, Category);
                Configurations = configs;
                success = true;
            }
            else
            {
                Loggy.Add(String.Format("Error: Loading 'pom.xml' failed in Package::Targets"));
            }

            return success;
        }
    }
}
