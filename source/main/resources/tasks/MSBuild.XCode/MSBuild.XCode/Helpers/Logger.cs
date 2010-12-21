using System;
using System.IO;
using System.Linq;
using System.Collections.Generic;

namespace MSBuild.XCode.Helpers
{
    public static class Logger
    {
        public static bool ToConsole { get; set; }
        public static int Indent { get; set; }
        public static string Indentor { get; set; }

        public static void Add(string line)
        {
            if (ToConsole)
            {
                for (int i = 0; i < Indent; ++i)
                    Console.Write(Indentor);
                Console.WriteLine(line);
            }
        }
    }
}
