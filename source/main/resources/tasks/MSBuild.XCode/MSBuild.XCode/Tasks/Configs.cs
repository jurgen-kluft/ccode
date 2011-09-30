using System;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
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
            PackageInstance.Initialize(string.Empty, string.Empty, RootDir);

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            if (package.IsValid)
            {
                // Get all platforms and configs, e.g: DevDebug|Win32;DevRelease|Win32;DevFinal|Win32
                ProjectInstance project = package.Pom.GetProjectByName(package.Name);
                if (project != null)
                {
                    string[] configs = project.GetConfigsForPlatform(Platform);
                    Configurations = configs;
                    success = true;
                }
            }
            else
            {
                Loggy.Error(String.Format("Error: Loading package failed in Package::Configs"));
            }

            return success;
        }
    }
}
