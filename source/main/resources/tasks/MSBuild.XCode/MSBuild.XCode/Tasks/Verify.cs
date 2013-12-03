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
        public string IDE { get; set; }         ///< Eclipse, Visual Studio, XCODE
        public string ToolSet { get; set; }     ///< GCC, Visual Studio (v90, v100, v110, v120)

        public override bool Execute()
        {
            Loggy.TaskLogger = Log;

            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            IDE = !String.IsNullOrEmpty(IDE) ? IDE.ToLower() : "vs2012";
            ToolSet = !String.IsNullOrEmpty(IDE) ? ToolSet.ToLower() : "v110";

            RootDir = RootDir.EndWith('\\');

            bool ok = false;

            PackageVars vars = new PackageVars();
            vars.Add("IDE", IDE);
            vars.Add(Platform + "ToolSet", ToolSet);
            vars.SetToolSet(Platform, ToolSet, true);

            PackageInstance package = PackageInstance.LoadFromTarget(RootDir + "target\\" + Name + "\\" + Platform + "\\" + ToolSet + "\\", vars);
            if (package.IsValid)
            {
                ok = true;
            }
            return ok;
        }
    }
}
