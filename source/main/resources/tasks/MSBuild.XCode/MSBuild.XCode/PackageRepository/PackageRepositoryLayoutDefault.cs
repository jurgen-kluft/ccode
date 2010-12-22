using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class LayoutDefault : ILayout
    {
        public string VersionToDir(ComparableVersion version)
        {
            string path = string.Empty;
            string[] components = version.ToStrings(3);
            // Keep it to X.Y.Z
            for (int i = 0; i < components.Length && i < 3; ++i)
                path = path + components[i] + "\\";
            return path;
        }

        public string VersionToFilename(string package_name, string branch, string platform, ComparableVersion version)
        {
            return VersionToFilenameWithoutExtension(package_name, branch, platform, version) + ".zip";
        }
        
        public string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, ComparableVersion version)
        {
            return String.Format("{0}+{1}+{2}+{3}", package_name, version.ToString(), branch, platform);
        }

        public string FilenameToVersion(string filename)
        {
            string[] parts = filename.Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
            return parts[1];
        }

        public string PackageRootDir(string repoPath, string group, string package_name)
        {
            // Path = group[] \ group[] ... \ package_name \ version.cache
            string[] splitted_group = group.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
            string groupPath = string.Empty;
            foreach (string g in splitted_group)
            {
                if (String.IsNullOrEmpty(groupPath))
                    groupPath = g + "\\";
                else
                    groupPath = groupPath + g + "\\";
            }
            string fullPath = repoPath + groupPath + package_name + "\\";
            return fullPath;
        }

        public string PackageVersionDir(string repoPath, string group, string package_name, ComparableVersion version)
        {
            // Path = group[] \ group[] ... \ package_name \ version.cache
            string[] splitted_group = group.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
            string groupPath = string.Empty;
            foreach (string g in splitted_group)
            {
                if (String.IsNullOrEmpty(groupPath))
                    groupPath = g + "\\";
                else
                    groupPath = groupPath + g + "\\";
            }
            string fullPath = repoPath + groupPath + package_name + "\\version\\" + VersionToDir(version);
            return fullPath;
        }
    }
}
