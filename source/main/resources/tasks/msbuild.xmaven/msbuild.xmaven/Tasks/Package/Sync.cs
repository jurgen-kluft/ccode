﻿using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using Microsoft.Build.Evaluation;

namespace msbuild.xmaven
{
    /// <summary>
    ///	Will sync the local-package-repository with the remote-package-repository. 
    ///	Will sync dependencies specified into the target folder
    /// </summary>
    public class PackageSync : Task
    {
        public string Name { get; set; }
        public string Path { get; set; }
        public string Dep { get; set; }
        public string LocalRepoPath { get; set; }
        public string RemoteRepoPath { get; set; }

        private class Dependency
        {
            public string Group { get; set; }
            public string Platforms { get; set; }
            public string Version { get; set; }
            public string Type { get; set; }
        }
        private Dictionary<string, List<Dependency>> mDependencies;

        private void AnalyzeAndGatherDependencies(Project project)
        {

        }

        private void ObtainDepFromPackage(string name, string version, string p)
        {

        }

        public override bool Execute()
        {
            // Load dep.props of main package
            // Add dependencies to the dependency list
            // 1
            // For every dependency copy it from remote to local repository if necessary
            // For every dependency load its dep.props and add it to the dependency list
            // Go back to 1 until no new dependencies (watch out for cyclic dependencies)
            // Analyze the dependency list and resolve versioning conflicts
            // Install the dependency packages in the target folder
            // Verify the installed packages
            // Done

            if (!File.Exists(Path + Dep))
                return false;

            Project deps = new Project(Path + Dep);
            AnalyzeAndGatherDependencies(deps);

            return false;
        }
    }
}
