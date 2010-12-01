using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

namespace MsBuild.MsMaven.Tasks
{
    /// <summary>
    ///	Will copy a new package release to the local-package-repository. 
    ///	Also updates 'latest'.
    ///
    ///  Actions:
    ///	1) Determine destination device and path for version
    ///	2) Determine destination device and path for latest
    ///	3) Determine the file that is now latest in the repository
    ///	4) Copy the package to it's version location
    ///	5) Generate a file with the filename of the package + '.latest' to latest location and delete the previous latest
    ///	6) Done
    ///
    /// </summary>
    class Sync : Task
    {
        public string Cmd { get; set; }

        public override bool Execute()
        {

            return false;
        }
    }
}
