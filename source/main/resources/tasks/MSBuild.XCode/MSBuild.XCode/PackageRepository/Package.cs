using System;

namespace MSBuild.XCode
{
    public enum ELocation
    {
        Remote,     ///< Remote package repository
        Cache,      ///< Cache package repository (on local machine)
        Local,      ///< Local package, a 'Created' package of the Root
        Target,     ///< Target package, an 'Extracted' package in the target folder of a root package
        Share,      ///< Share package, an 'Extracted' package in the shared package repo
        Root,       ///< Root package
        Build,      ///< Build package
    }

    public class Package
    {
        public string Name { get; set; }
        public string Group { get; set; }
        public string Branch { get; set; }
        public string Platform { get; set; }
        public string Language { get; set; }
        public string Changeset { get; set; }

        public Package()
        {
            Changeset = "?";
            Language = "C++";
        }

        public static Package From(string name, string group, string branch, string platform, string language, string changeset)
        {
            Package instance = new Package();
            instance.Name = name;
            instance.Group = group;
            instance.Branch = branch;
            instance.Platform = platform;
            instance.Language = language;
            instance.Changeset = changeset;
            return instance;
        }
    }

    public class PackageState : Package
    {
        public DateTime RemoteSignature { get; set; }
        public DateTime CacheSignature { get; set; }
        public DateTime ShareSignature { get; set; }
        public DateTime TargetSignature { get; set; }
        public DateTime LocalSignature { get; set; }

        public IPackageFilename RemoteFilename { get; set; }
        public string RemoteStorageKey { get; set; }
        public IPackageFilename CacheFilename { get; set; }
        public IPackageFilename ShareFilename { get; set; }
        public IPackageFilename TargetFilename { get; set; }
        public IPackageFilename LocalFilename { get; set; }

        public ComparableVersion RemoteVersion { get; set; }
        public ComparableVersion CacheVersion { get; set; }
        public ComparableVersion ShareVersion { get; set; }
        public ComparableVersion TargetVersion { get; set; }
        public ComparableVersion LocalVersion { get; set; }
        public ComparableVersion RootVersion { get; set; }
        public ComparableVersion CreateVersion { get; set; }

        public bool RemoteExists { get { return !String.IsNullOrEmpty(RemoteURL); } }
        public bool CacheExists { get { return !String.IsNullOrEmpty(CacheURL); } }
        public bool LocalExists { get { return !String.IsNullOrEmpty(LocalURL); } }
        public bool ShareExists { get { return !String.IsNullOrEmpty(ShareURL); } }
        public bool TargetExists { get { return !String.IsNullOrEmpty(TargetURL); } }
        public bool RootExists { get { return !String.IsNullOrEmpty(RootURL); } }

        public string RemoteURL { get; set; }
        public string CacheURL { get; set; }
        public string LocalURL { get; set; }
        public string ShareURL { get; set; }
        public string TargetURL { get; set; }
        public string RootURL { get; set; }

        public bool IncrementVersion { get; set; }

        public PackageState()
        {
            IncrementVersion = false;
        }

        public static PackageState From(string name, string group, string branch, string platform)
        {
            PackageState instance = new PackageState();
            instance.Name = name;
            instance.Group = group;
            instance.Branch = branch;
            instance.Platform = platform;
            return instance;
        }

        public void SetURL(ELocation location, string url)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteURL = url; break;
                case ELocation.Cache: CacheURL = url; break;
                case ELocation.Local: LocalURL = url; break;
                case ELocation.Share: ShareURL = url; break;
                case ELocation.Target: TargetURL = url; break;
                case ELocation.Root: RootURL = url; break;
            }
        }

        public void SetFilename(ELocation location, IPackageFilename filename)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteFilename = filename; break;
                case ELocation.Cache: CacheFilename = filename; break;
                case ELocation.Local: LocalFilename = filename; break;
                case ELocation.Share: ShareFilename = filename; break;
                case ELocation.Target: TargetFilename = filename; break;
                case ELocation.Root: break;
            }
        }

        public void SetSignature(ELocation location, DateTime signature)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteSignature = signature; break;
                case ELocation.Cache: CacheSignature = signature; break;
                case ELocation.Local: LocalSignature = signature; break;
                case ELocation.Share: ShareSignature = signature; break;
                case ELocation.Target: TargetSignature = signature; break;
                case ELocation.Root: break;
            }
        }


        public void SetVersion(ELocation location, ComparableVersion version)
        {
            switch (location)
            {
                case ELocation.Remote: RemoteVersion = version; break;
                case ELocation.Cache: CacheVersion = version; break;
                case ELocation.Local: LocalVersion = version; break;
                case ELocation.Share: ShareVersion = version; break;
                case ELocation.Target: TargetVersion = version; break;
                case ELocation.Root: RootVersion = version; break;
            }
        }

        public string GetLocalURL()
        {
            if (RootExists)
            {
                return RootURL;
            }
            else if (TargetExists)
            {
                if (ShareExists)
                {
                    return ShareURL;
                }
                return TargetURL;
            }
            else if (ShareExists)
            {
                return ShareURL;
            }
            return string.Empty;
        }

        public string GetURL(ELocation location)
        {
            string url = string.Empty;
            switch (location)
            {
                case ELocation.Remote: url = RemoteURL; break;
                case ELocation.Cache: url = CacheURL; break;
                case ELocation.Local: url = LocalURL; break;
                case ELocation.Share: url = ShareURL; break;
                case ELocation.Target: url = TargetURL; break;
                case ELocation.Root: url = RootURL; break;
            }
            return url;
        }

        public bool HasURL(ELocation location)
        {
            bool has = false;
            switch (location)
            {
                case ELocation.Remote: has = RemoteExists; break;
                case ELocation.Cache: has = CacheExists; break;
                case ELocation.Local: has = LocalExists; break;
                case ELocation.Share: has = ShareExists; break;
                case ELocation.Target: has = TargetExists; break;
                case ELocation.Root: has = RootExists; break;
            }
            return has;
        }

        public IPackageFilename GetFilename(ELocation location)
        {
            IPackageFilename filename = null;
            switch (location)
            {
                case ELocation.Remote: filename = RemoteFilename; break;
                case ELocation.Cache: filename = CacheFilename; break;
                case ELocation.Local: filename = LocalFilename; break;
                case ELocation.Share: filename = ShareFilename; break;
                case ELocation.Target: filename = TargetFilename; break;
                case ELocation.Root: break;
            }
            return filename;
        }

        public DateTime GetSignature(ELocation location)
        {
            DateTime signature = DateTime.MinValue;
            switch (location)
            {
                case ELocation.Remote: signature = RemoteSignature; break;
                case ELocation.Cache: signature = CacheSignature; break;
                case ELocation.Local: signature = LocalSignature; break;
                case ELocation.Share: signature = ShareSignature; break;
                case ELocation.Target: signature = TargetSignature; break;
                case ELocation.Root: break;
            }
            return signature;
        }

        public ComparableVersion GetVersion(ELocation location)
        {
            ComparableVersion version = null;
            switch (location)
            {
                case ELocation.Remote: version = RemoteVersion; break;
                case ELocation.Cache: version = CacheVersion; break;
                case ELocation.Local: version = LocalVersion; break;
                case ELocation.Share: version = ShareVersion; break;
                case ELocation.Target: version = TargetVersion; break;
                case ELocation.Root: version = RootVersion; break;
            }
            return version;
        }
    }
}