using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public static partial class MyExtensions
    {
        public static string EndWith(this string str, char e)
        {
            if (!str.EndsWith("" + e))
                str = str + e;
            return str;
        }
    }
}
