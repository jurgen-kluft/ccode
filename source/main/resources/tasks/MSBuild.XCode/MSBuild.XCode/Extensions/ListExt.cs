using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public static partial class MyExtensions
    {
        public static int IndexOfUsingBinarySearch<T>(this List<T> sortedCollection, T value) where T : IComparable<T>
        {
            if (sortedCollection == null)
                return -1;

            int begin = 0;
            int end = sortedCollection.Count - 1;
            int index = 0;
            while (end >= begin)
            {
                index = (begin + end) / 2;
                T val = sortedCollection[index];
                int compare = val.CompareTo(value);
                if (compare == 0)
                    return index;
                if (compare > 0)
                    end = index - 1;
                else
                    begin = index + 1;
            }

            return ~index;  // Not found, return bitwise complement of the index.
        }

        public static int LowerBound<T>(this List<T> sortedCollection, T value, bool lessOrEqual) where T : IComparable<T>
        {
            int index = IndexOfUsingBinarySearch(sortedCollection, value);
            if (index < 0)
                index = ~index;

            if (lessOrEqual)
            {
                while (index > 0 && value.CompareTo(sortedCollection[index - 1]) == -1)
                    --index;
                while (index < sortedCollection.Count && value.CompareTo(sortedCollection[index]) != -1)
                    ++index;
            }
            else
            {
                while (index > 0 && value.CompareTo(sortedCollection[index - 1]) != 1)
                    --index;
                while (index < sortedCollection.Count && value.CompareTo(sortedCollection[index]) == 1)
                    ++index;
            }

            return index;
        }
    }
}
