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
        private static int mConsoleCursorLeft = -1;
        private static int mConsoleCursorTop = -1;
        private static int mConsoleCursorLastLineLeft = -1;

        public static int LastLineCursorLeft
        {
            get
            {
                return mConsoleCursorLastLineLeft;
            }
        }

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
        public static void RestoreConsoleCursor()
        {
            if (mConsoleCursorLeft != -1 && mConsoleCursorTop != -1)
                Console.SetCursorPosition(mConsoleCursorLeft, mConsoleCursorTop);
        }
        private static void SaveConsoleCursor()
        {
            mConsoleCursorLeft = Console.CursorLeft;
            mConsoleCursorTop = Console.CursorTop;
        }

         public static void Info(string line)
        {
            if (ToConsole)
            {
                PushConsoleColor(ConsoleColor.Green);
                RestoreConsoleCursor();

                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
                
                PopConsoleColor();
                SaveConsoleCursor();
            }
            else if (TaskLogger != null)
            {
                RestoreConsoleCursor();
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;
                TaskLogger.LogMessage(line);
                SaveConsoleCursor();
            }
        }

        public static void Warning(string line)
        {
            if (ToConsole)
            {
                PushConsoleColor(ConsoleColor.Yellow);
                RestoreConsoleCursor();

                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);

                PopConsoleColor();
                SaveConsoleCursor();
            }
            else if (TaskLogger != null)
            {
                RestoreConsoleCursor();
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;

                TaskLogger.LogWarning(line);
                SaveConsoleCursor();
            }
        }

        public static void Error(string line)
        {
            if (ToConsole)
            {
                PushConsoleColor(ConsoleColor.Red);
                RestoreConsoleCursor();

                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);

                PopConsoleColor();
                SaveConsoleCursor();
            }
            else if (TaskLogger != null)
            {
                RestoreConsoleCursor();
                for (int i = 0; i < Indent; ++i)
                    line = Indentor + line;

                TaskLogger.LogError(line);
                SaveConsoleCursor();
            }
        }
    }
}
