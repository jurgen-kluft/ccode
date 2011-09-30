using System;

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
