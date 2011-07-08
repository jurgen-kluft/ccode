using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial.Extensions.Queues
{
    /// <summary>
    /// This class contains logic for the Mercurial Queues extension.
    /// </summary>
    public static class QueueExtension
    {
        /// <summary>
        /// Gets whether the Mercurial Queues (mq) extension is installed and active.
        /// </summary>
        public static bool IsInstalled
        {
            get
            {
                return Client.Configuration.ValueExists("extensions", "mq");
            }
        }
    }
}