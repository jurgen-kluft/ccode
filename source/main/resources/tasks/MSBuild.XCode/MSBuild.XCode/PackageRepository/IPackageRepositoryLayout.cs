
namespace MSBuild.XCode
{
    public interface ILayout
    {
        string VersionToDir(ComparableVersion version);
        string VersionToFilename(string package_name, string branch, string platform, string toolset, ComparableVersion version);
        string VersionToFilenameWithoutExtension(string package_name, string branch, string platform, string toolset, ComparableVersion version);
        string FilenameToVersion(string filename);
        string PackageRootDir(string repoPath, string group, string package_name, string platform, string toolset);
        string PackageVersionDir(string repoPath, string group, string package_name, string platform, string toolset, string branch, ComparableVersion version);
    }

}
