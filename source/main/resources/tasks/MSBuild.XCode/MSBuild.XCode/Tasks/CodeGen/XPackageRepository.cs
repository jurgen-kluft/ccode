using System;
using System.IO;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XPackageRepository
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

        public XPackageRepository(string path)
        {
            mRepository = new XPackageRepositoryFileSystem(path);
        }

        public string RepoPath { get; set; }

        public bool Checkout(XPackage package, XVersionRange range)
        {
            return mRepository.Checkout(package, range);
        }
        public bool Checkout(XPackage package)
        {
            return mRepository.Checkout(package);
        }
        public bool Checkin(XPackage package)
        {
            return mRepository.Checkin(package);
        }

    }
}
