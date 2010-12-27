using System;
using System.Xml;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will sync the local-package-repository with the remote-package-repository. 
    ///	Will sync dependencies specified into the target folder
    /// </summary>
    public class PackageSync : Task
    {
        public string RootDir { get; set; }
        public string Platform { get; set; }
        public string CacheRepoDir { get; set; }
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;
            RootDir = RootDir.EndWith('\\');

            Package package = new Package();
            package.IsRoot = true;
            package.RootDir = RootDir;
            if (package.LoadFinalPom())
            {
                package.Name = package.Pom.Name;
                package.Group = new Group(package.Pom.Group);
                package.Version = package.Pom.Versions.GetForPlatform(Platform);
                package.Platform = Platform;
                {
                    string package_filename;
                    package.Create(out package_filename);

                    if (package.BuildDependencies(Platform))
                    {
                        if (package.SyncDependencies(Platform))
                        {
                            return true;
                        }
                    }
                }
            }

            return false;
        }
    }
}
