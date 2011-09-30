namespace MSBuild.XCode
{
    public interface IPackageRepository
    {
        bool Valid { get; }
        string RepoURL { get; }
        ELocation Location { get; }

        bool Query(PackageState package);
        bool Query(PackageState package, VersionRange versionRange);
        bool Link(PackageState package, out string filename);
        bool Download(PackageState package, string to_filename);

        bool Submit(PackageState package, IPackageRepository from);
    }

}
