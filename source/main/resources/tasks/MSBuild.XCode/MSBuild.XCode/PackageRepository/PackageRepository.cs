using System;
using System.IO;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class PackageRepository
    {
        public enum ELayout
        {
            /// <summary>
            /// Default Layout:
            ///     Package VersionPath = RepositoryPath\GroupH\GroupM\GroupL\Package_Name\Version\Major\Minor\Fix
            ///     Package Filename    = Package_Name+Version+Branch+Platform.zip
            /// </summary>
            Default,
        }

        private IPackageRepository mRepository;

        public PackageRepository(string path, ELocation location)
        {
            mRepository = new PackageRepositoryFileSystem(path, location);
        }

        public string RepoPath { get; set; }

        public bool Update(PackageInstance package, VersionRange range)
        {
            return mRepository.Update(package, range);
        }
        public bool Update(PackageInstance package)
        {
            return mRepository.Update(package);
        }
        public bool Add(PackageInstance package, ELocation from)
        {
            return mRepository.Add(package, from);
        }

    }
}
