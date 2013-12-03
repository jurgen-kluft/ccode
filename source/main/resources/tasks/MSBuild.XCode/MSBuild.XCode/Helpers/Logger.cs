using System;
using System.Collections.Generic;
using Microsoft.Build.Utilities;

namespace MSBuild.XCode.Helpers
{
    public static class Loggy
    {
        public static bool ToConsole { get; set; }
        public static int Indent { get; set; }
        public static string Indentor { get; set; }

        public static TaskLoggingHelper TaskLogger { get; set; }

        private static Stack<ConsoleColor> mConsoleColorStack;

        static Loggy()
        {
            ToConsole = false;
            Indent = 0;
            Indentor = "\t";
            mConsoleColorStack = new Stack<ConsoleColor>();
        }

        private static void PushConsoleColor(ConsoleColor c)
        {
            mConsoleColorStack.Push(Console.ForegroundColor);
            Console.ForegroundColor = c;
        }
        private static void PopConsoleColor()
        {
            ConsoleColor c = mConsoleColorStack.Peek();
            Console.ForegroundColor = c;
            mConsoleColorStack.Pop();
        }

        public static void Info(string line)
        {
            if (ToConsole)
            {
                PushConsoleColor(ConsoleColor.Green);

                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);

                PopConsoleColor();
            }
            else if (TaskLogger != null)
            {
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;
                TaskLogger.LogMessage(line);
            }
            Console.Out.Flush();
        }

        public static void Warning(string line)
        {
            if (ToConsole)
            {
                PushConsoleColor(ConsoleColor.Yellow);

                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);

                PopConsoleColor();
            }
            else if (TaskLogger != null)
            {
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;

                TaskLogger.LogWarning(line);
            }
            Console.Out.Flush();
        }

        public static void Error(string line)
        {
            if (ToConsole)
            {
                PushConsoleColor(ConsoleColor.Red);

                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);

                PopConsoleColor();
            }
            else if (TaskLogger != null)
            {
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;

                TaskLogger.LogError(line);
            }
            Console.Out.Flush();
        }
    }
}
