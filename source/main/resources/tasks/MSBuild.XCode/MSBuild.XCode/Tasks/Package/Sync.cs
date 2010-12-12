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
        public string Platform { get; set; }
        public string Branch { get; set; }
        public string LocalRepoPath { get; set; }
        public string RemoteRepoPath { get; set; }

        public override bool Execute()
        {
            // Load package.xml of main package
            // Add dependencies to the dependency list
            // 1
            // For every dependency copy it from remote to local repository if necessary
            // For every dependency load its package.xml and add it to the dependency list
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

            Dictionary<string, XVersionRange> dependencyVersionRangeMap = new Dictionary<string, XVersionRange>();
            Dictionary<string, XDependency> dependencyMap = new Dictionary<string, XDependency>();
            foreach (XDependency d in package.Dependencies)
            {
                if (dependencyMap.ContainsKey(d.Name))
                {
                    // Version conflict?
                    XVersionRange current;
                    dependencyVersionRangeMap.TryGetValue(d.Name, out current);
                    XVersionRange union = d.GetVersionRange(Platform, Branch).Union(current);
                    dependencyVersionRangeMap.Remove(d.Name);
                    dependencyVersionRangeMap.Add(d.Name, union);
                }
                else
                {
                    dependencyVersionRangeMap.Add(d.Name, d.GetVersionRange(Platform, Branch));
                    dependencyMap.Add(d.Name, d);
                }
            }

            Queue<string> dependencyQueue = new Queue<string>();
            foreach (XDependency d in dependencyMap.Values)
                dependencyQueue.Enqueue(d.Name);

            while (dependencyQueue.Count > 0)
            {

            }


            return false;
        }
    }
}
