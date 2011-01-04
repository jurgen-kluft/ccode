using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public static partial class MyExtensions
    {
        public static bool IsEmpty<T>(this IEnumerable<T> enumerable)
        {
            return !enumerable.GetEnumerator().MoveNext();
        }
    }
}
