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
        string RepoDir { get; set; }
        ELocation Location { get; set; }
        ILayout Layout { get; set; }

        bool Update(PackageInstance package, VersionRange versionRange);
        bool Update(PackageInstance package);
        bool Add(PackageInstance package, ELocation from);
    }

}
