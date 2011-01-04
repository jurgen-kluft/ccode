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

        bool Update(Package package, VersionRange versionRange);
        bool Update(Package package);
        bool Add(Package package, ELocation from);
    }

}
