using System;
using System.IO;
using System.Collections.Generic;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public partial class Package
    {
        public string SourcePath { get; set; }
        public string SourceFilename { get; set; }
        public string OldLatest { get; set; }
        public string RepoPath { get; set; }
        public string VersionPath { get; set; }
        public string LatestPath { get; set; }

        public Package()
        {
            Versions = new List<string>();
        }

        private string mVersion = string.Empty;
        public string Version { get { ResolveVersions(); return mVersion; } }
        private List<string> Versions { get; set; }
        private void ResolveVersions()
        {
        }

        private string mType = string.Empty;
        public string Type { get { ResolveTypes(); return mType; } }
        private List<string> Types { get; set; }
        private void ResolveTypes()
        {
        }

        private string mGroup = string.Empty;
        public string Group { get { return mGroup;} set { SetGroup(value); } }
        private void SetGroup(string group)
        {
        }

        public void Load(string filename)
        {
            _Load(filename);
        }

        public bool Install()
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
                foreach (string f in files)
                    File.Delete(f);
                File.Create(RepoPath + LatestPath + SourceFilename + ".latest");
            }
            catch (Exception)
            {
                ok = false;
            }

            return ok;
        }

        public bool Sync(string RemoteRepoPath)
        {
            // Synchronize Remote Repo with Local Repo

            return false;
        }
    }
}
