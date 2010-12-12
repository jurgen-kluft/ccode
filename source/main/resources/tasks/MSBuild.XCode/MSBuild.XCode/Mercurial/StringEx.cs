using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial
{
    internal static class StringEx
    {
        public static bool IsNullOrWhiteSpace(string value)
        {
            return value == null || value.Trim().Length == 0;
        }
    }
}