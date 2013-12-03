using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class LayoutTarget : ILayout
    {
        public string VersionToDir(ComparableVersion version)
        {
            return string.Empty;
        }

        public string VersionToFilename(string package_name, string branch, string platform, string toolset, ComparableVersion version)
        {
            return VersionToFilenameWithoutExtension(package_name, branch, platform, toolset, version) + ".zip";
        }

        public string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, string toolset, ComparableVersion version)
        {
            string versionStr = (version == null) ? "1.0.0" : version.ToString();
            return String.Format("{0}+{1}+{2}+{3}+{4}", package_name, versionStr, branch, platform, toolset);
        }

        public string FilenameToVersion(string filename)
        {
            string[] parts = filename.Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
            return parts[1];
        }

        public string PackageRootDir(string repoPath, string group, string package_name, string platform, string toolset)
        {
            // Path = package_name \ toolset \
            string fullPath = repoPath + package_name + "\\" + toolset + "\\";
            return fullPath;
        }

        public string PackageVersionDir(string repoPath, string group, string package_name, string platform, string toolset, string branch, ComparableVersion version)
        {
            // Path = package_name \  toolset \
            string fullPath = repoPath + package_name + "\\" + toolset + "\\";
            return fullPath;
        }
    }
}
