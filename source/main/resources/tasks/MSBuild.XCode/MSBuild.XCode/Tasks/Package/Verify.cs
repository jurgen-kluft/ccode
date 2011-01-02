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
        public string Name { get; set; }
        public string Platform { get; set; }

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;

            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            RootDir = RootDir.EndWith('\\');

            bool ok = false;
            PackageInstance package = PackageInstance.LoadFromTarget(RootDir + "target\\" + Name + "\\" + Platform + "\\");
            if (package.IsValid)
            {
                ok = package.Verify();
            }
            return ok;
        }
    }
}
