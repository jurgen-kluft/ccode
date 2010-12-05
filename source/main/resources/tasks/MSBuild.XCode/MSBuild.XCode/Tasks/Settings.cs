using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

namespace MSBuild.XCode
{
    public class Settings : Task
    {
        public string PackagePath { get; set; }     ///< D:\Dev\xbase\package.xml

        [Output]
        public string Name { get; set; }
        [Output]
        public string Group { get; set; }
        [Output]
        public string GroupH { get; set; }
        [Output]
        public string GroupM { get; set; }
        [Output]
        public string GroupL { get; set; }
        [Output]
        public ITaskItem[] Dependencies { get; set; }
        [Output]
        public ITaskItem[] Configurations { get; set; }

        public override bool Execute()
        {
            bool result = false;
            return result;
        }
    }
}
