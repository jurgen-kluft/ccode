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

        private XPackageRepositoryLayout mLayout;

        public XPackageRepository(string path)
        {
            mLayout = new XPackageRepositoryLayoutDefault(path);
        }

        public XPackageRepository(string path, ELayout layout)
        {
            if (layout == ELayout.Default)
                mLayout = new XPackageRepositoryLayoutDefault(path);
            else
                mLayout = new XPackageRepositoryLayoutDefault(path);
        }

        public string RepoPath { get; set; }

        public bool Checkout(string group, string package_path, string package_name, string branch, string platform, XVersionRange range)
        {
            return mLayout.Checkout(group, package_path, package_name, branch, platform, range);
        }
        public bool Commit(string group, string package_path, string package_name, string branch, string platform, XVersion version)
        {
            return mLayout.Commit(group, package_path, package_name, branch, platform, version);
        }
        public XVersion Sync(string group, string package_name, string branch, string platform, XVersionRange range, XPackageRepository to)
        {
            return mLayout.Sync(group, package_name, branch, platform, range, to.mLayout);
        }
        public XPackage Info(string group, string package_name, string branch, string platform, XVersion version)
        {
            return mLayout.Info(group, package_name, branch, platform, version);
        }
    }
}
