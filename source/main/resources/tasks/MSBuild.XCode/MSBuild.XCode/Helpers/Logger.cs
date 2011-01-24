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

        public static void Warning(string line)
        {
            if (TaskLogger != null && TaskLogger.TaskResources != null)
            {
                TaskLogger.LogWarning(line);
            }
            else if (ToConsole)
            {
                ConsoleColor oldColor = Console.ForegroundColor;
                Console.ForegroundColor = ConsoleColor.Yellow;
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
                Console.ForegroundColor = oldColor;
            }
        }

        public static void Error(string line)
        {
            if (TaskLogger != null && TaskLogger.TaskResources!=null)
            {
                TaskLogger.LogError(line);
            }
            else if (ToConsole)
            {
                ConsoleColor oldColor = Console.ForegroundColor;
                Console.ForegroundColor = ConsoleColor.Red;
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
                Console.ForegroundColor = oldColor;
            }
        }
    }
}
