using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will copy a new package release to the local-package-repository. 
    /// </summary>
    public class PackageInstall : Task
    {
        public string Path { get; set; }
        public string Filename { get; set; }
        public string RepoPath { get; set; }

        public override bool Execute()
        {
            if (!Path.EndsWith("\\"))
                Path = Path + "\\";

            // if (!File.Exists(Path + "package.xml"))
            //    return false;

            // - Commit version to local package repository

            return false;
        }
    }
}
