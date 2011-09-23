using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public static partial class MyExtensions
    {
        public static string ToSql(this DateTime dt)
        {
            string str = String.Format("{0:yyyy-M-d H:m:s}", dt);
            return str;
        }
    }
}
