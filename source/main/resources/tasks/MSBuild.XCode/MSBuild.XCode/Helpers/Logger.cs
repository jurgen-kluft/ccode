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

        public static void Info(string line)
        {
            if (ToConsole)
            {
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
            }
            if (TaskLogger != null && TaskLogger.TaskResources != null)
            {
                TaskLogger.LogMessage(line);
            }
        }

        public static void Warning(string line)
        {
            if (ToConsole)
            {
                ConsoleColor oldColor = Console.ForegroundColor;
                Console.ForegroundColor = ConsoleColor.Yellow;
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
                Console.ForegroundColor = oldColor;
            }
            if (TaskLogger != null && TaskLogger.TaskResources != null)
            {
                TaskLogger.LogWarning(line);
            }
        }

        public static void Error(string line)
        {
            if (ToConsole)
            {
                ConsoleColor oldColor = Console.ForegroundColor;
                Console.ForegroundColor = ConsoleColor.Red;
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
                Console.ForegroundColor = oldColor;
            }
            if (TaskLogger != null && TaskLogger.TaskResources != null)
            {
                TaskLogger.LogError(line);
            }
        }
    }
}
