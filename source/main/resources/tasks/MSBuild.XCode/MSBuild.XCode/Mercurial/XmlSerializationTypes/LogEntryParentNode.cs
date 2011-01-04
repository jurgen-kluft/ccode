using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Xml.Serialization;

namespace Mercurial.XmlSerializationTypes
{
    /// <summary>
    /// This class encapsulates a &lt;parent...&gt; node in the log output.
    /// </summary>
    [XmlType("parent")]
    [EditorBrowsable(EditorBrowsableState.Never)]
    public class LogEntryParentNode
    {
        /// <summary>
        /// The revision of the parent log entry.
        /// </summary>
        [XmlAttribute("revision")]
        public int Revision
        {
            get;
            set;
        }

        /// <summary>
        /// The hash of the parent log entry.
        /// </summary>
        [XmlAttribute("node")]
        public string Hash
        {
            get;
            set;
        }
    }
}