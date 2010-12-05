﻿using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using msbuild.xmaven.Helpers;

namespace msbuild.xmaven
{
    /// <summary>
    ///	Will copy a new package release to the local-package-repository. 
    ///	Also updates 'latest'.
    /// </summary>
    public class PackageInstall : Task
    {
        public string SourcePath { get; set; }
        public string SourceFilename { get; set; }
        public string OldLatest { get; set; }
        public string RepoPath { get; set; }
        public string VersionPath { get; set; }
        public string LatestPath { get; set; }

        public override bool Execute()
        {
            Package p = new Package();
            p.SourcePath = SourcePath;
            p.SourceFilename = SourceFilename;
            p.OldLatest = OldLatest;
            p.RepoPath = RepoPath;
            p.VersionPath = VersionPath;
            p.LatestPath = LatestPath;
            return p.Install();
        }
    }
}
