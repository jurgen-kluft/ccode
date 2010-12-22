using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using System.Security.Cryptography;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    ///	Will verify an 'extracted' package
    /// </summary>
    public class PackageVerify : Task
    {
        public string RootDir { get; set; }
        public string Platform { get; set; }
        public string Branch { get; set; }

        public override bool Execute()
        {
            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";
            if (String.IsNullOrEmpty(Branch))
                Branch = "default";

            RootDir = RootDir.EndWith('\\');

            Package package = new Package();
            package.RootDir = RootDir;
            package.LoadPom();
            package.Name = package.Pom.Name;
            package.Group = package.Pom.Group;
            package.Version = package.Pom.Versions.GetForPlatformWithBranch(Platform, Branch);
            package.Branch = Branch;
            package.Platform = Platform;

            bool ok = package.Verify();
            return ok;
        }
    }
}
