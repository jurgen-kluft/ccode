using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

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
            bool ok = true;

            if (!SourcePath.EndsWith("\\"))
                SourcePath = SourcePath + "\\";
            if (!RepoPath.EndsWith("\\"))
                RepoPath = RepoPath + "\\";
            if (!VersionPath.EndsWith("\\"))
                VersionPath = VersionPath + "\\";
            if (!LatestPath.EndsWith("\\"))
                LatestPath = LatestPath + "\\";

            if (File.Exists(RepoPath + VersionPath + SourceFilename) && File.Exists(RepoPath + LatestPath + SourceFilename + ".latest"))
            {
                // Do we need to do a binary compare to be sure ?
                return ok;
            }

            try
            {
                File.Copy(SourcePath + SourceFilename, RepoPath + VersionPath + SourceFilename, true);
                string[] files = Directory.GetFiles(RepoPath + LatestPath, OldLatest, SearchOption.TopDirectoryOnly);
                foreach(string f in files)
                    File.Delete(f);
                File.Create(RepoPath + LatestPath + SourceFilename + ".latest");
            }
            catch (Exception)
            {
                Log.LogMessage("PackageInstall FAILED...");
                ok = false;
            }

            return ok;
        }
    }
}
