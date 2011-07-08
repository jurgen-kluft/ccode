using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Xml.Serialization;

namespace Mercurial.XmlSerializationTypes
{
    /// <summary>
    /// This class encapsulates a &lt;log...&gt; node in the log output.
    /// </summary>
    [XmlType("log")]
    [EditorBrowsable(EditorBrowsableState.Never)]
    public class LogNode
    {
        private readonly List<LogEntryNode> _LogEntries = new List<LogEntryNode>();

        /// <summary>
        /// The collection of log entries of the log.
        /// </summary>
        [XmlElement("logentry")]
        public Collection<LogEntryNode> LogEntries
        {
            get
            {
                return new Collection<LogEntryNode>(_LogEntries);
            }
        }
    }
}