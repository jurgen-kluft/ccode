using System;
using System.IO;
using System.Linq;
using System.Collections.Generic;

using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

namespace MSBuild.XCode.Helpers
{
    public static class Loggy
    {
        public static bool ToConsole { get; set; }
        public static int Indent { get; set; }
        public static string Indentor { get; set; }

        public static TaskLoggingHelper TaskLogger { get; set; }

        public static void Add(string line)
        {
            if (TaskLogger != null && TaskLogger.TaskResources!=null)
            {
                TaskLogger.LogMessage(line);
            }
            else if (ToConsole)
            {
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
            }
        }
    }
}
