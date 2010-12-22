using System;
using System.Xml;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Utilities;

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
            // Load pom.xml of main package
            // Add dependencies to the dependency list
            // 1
            // For every dependency copy it from remote to local repository if necessary
            // For every dependency load its pom.xml and add it to the dependency list
            // Go back to 1 until no new dependencies (watch out for cyclic dependencies)
            // Analyze the dependency list and resolve version conflicts
            // Install the dependency packages in the target folder
            // Verify the installed packages
            // Done

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

                    if (package.BuildDependencies(Platform, Global.CacheRepo, Global.RemoteRepo))
                    {
                        if (package.SyncDependencies(Platform, Global.CacheRepo))
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
