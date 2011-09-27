using System;
using Microsoft.Build.Utilities;

namespace MSBuild.XCode.Helpers
{
    public static class Loggy
    {
        public static bool ToConsole { get; set; }
        public static int Indent { get; set; }
        public static string Indentor { get; set; }

        public static TaskLoggingHelper TaskLogger { get; set; }

        static Loggy()
        {
            ToConsole = false;
            Indent = 0;
            Indentor = "\t";
        }

        public static void Info(string line)
        {
            if (ToConsole)
            {
                ConsoleColor oldColor = Console.ForegroundColor;
                Console.ForegroundColor = ConsoleColor.Green;
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
                Console.ForegroundColor = oldColor;
            }
            else if (TaskLogger != null)
            {
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;
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
            else if (TaskLogger != null)
            {
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;

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
            else if (TaskLogger != null)
            {
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;

                TaskLogger.LogError(line);
            }
        }
    }
}
