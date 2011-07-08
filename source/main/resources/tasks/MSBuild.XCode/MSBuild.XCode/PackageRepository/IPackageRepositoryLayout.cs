using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public interface ILayout
    {
        string VersionToDir(ComparableVersion version);
        string VersionToFilename(string package_name, string branch, string platform, ComparableVersion version);
        string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, ComparableVersion version);
        string FilenameToVersion(string filename);
        string PackageRootDir(string repoPath, string group, string package_name, string platform);
        string PackageVersionDir(string repoPath, string group, string package_name, string platform, string branch, ComparableVersion version);
    }

}
