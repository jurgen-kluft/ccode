using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial.Extensions.CaseGuard
{
    /// <summary>
    /// This class contains logic for the caseguard Mercurial extension.
    /// </summary>
    public static class CaseGuardExtension
    {
        /// <summary>
        /// Gets whether the caseguard extension is installed and active.
        /// </summary>
        public static bool IsInstalled
        {
            get
            {
                return Client.Configuration.ValueExists("extensions", "caseguard");
            }
        }
    }
}