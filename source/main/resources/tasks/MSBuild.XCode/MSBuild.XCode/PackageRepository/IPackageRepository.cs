using System;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public interface IPackageRepository
    {
        bool Valid { get; }
        string RepoURL { get; }
        ELocation Location { get; }

        bool Query(Package package);
        bool Query(Package package, VersionRange versionRange);
        bool Link(Package package, out string filename);
        bool Download(Package package, string to_filename);

        bool Submit(Package package, IPackageRepository from);
    }

}
