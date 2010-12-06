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
        public string Path { get; set; }
        public string LocalRepoPath { get; set; }
        public string RemoteRepoPath { get; set; }

        public override bool Execute()
        {
            // Load prj.xml of main package
            // Add dependencies to the dependency list
            // 1
            // For every dependency copy it from remote to local repository if necessary
            // For every dependency load its prj.xml and add it to the dependency list
            // Go back to 1 until no new dependencies (watch out for cyclic dependencies)
            // Analyze the dependency list and resolve version conflicts
            // Install the dependency packages in the target folder
            // Verify the installed packages
            // Done

            if (!Path.EndsWith("\\"))
                Path = Path + "\\";

            if (!File.Exists(Path + "package.xml"))
                return false;

            XPackage package = new XPackage();
            package.Load(Path + "package.xml");

            return false;
        }
    }
}
