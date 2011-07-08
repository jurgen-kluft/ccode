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
    /// This class encapsulates a &lt;logentry...&gt; node in the log output.
    /// </summary>
    [XmlType("logentry")]
    [EditorBrowsable(EditorBrowsableState.Never)]
    public class LogEntryNode
    {
        private readonly List<LogEntryParentNode> _Parents = new List<LogEntryParentNode>();
        private readonly List<LogEntryPathNode> _PathActions = new List<LogEntryPathNode>();
        private readonly List<LogEntryTagNode> _Tags = new List<LogEntryTagNode>();

        /// <summary>
        /// The local revision number of this log entry.
        /// </summary>
        [XmlAttribute("revision")]
        public int Revision
        {
            get;
            set;
        }

        /// <summary>
        /// The hash of this log entry.
        /// </summary>
        [XmlAttribute("node")]
        public string Hash
        {
            get;
            set;
        }

        /// <summary>
        /// The tags of this log entry.
        /// </summary>
        [XmlElement("tag")]
        public Collection<LogEntryTagNode> Tags
        {
            get
            {
                return new Collection<LogEntryTagNode>(_Tags);
            }
        }

        /// <summary>
        /// The commit message of this log entry.
        /// </summary>
        [XmlElement("msg")]
        public string CommitMessage
        {
            get;
            set;
        }

        /// <summary>
        /// The timestamp of this log entry.
        /// </summary>
        [XmlElement("date")]
        public DateTime Timestamp
        {
            get;
            set;
        }

        /// <summary>
        /// The author of this log entry.
        /// </summary>
        [XmlElement("author")]
        public LogEntryAuthorNode Author
        {
            get;
            set;
        }

        /// <summary>
        /// A parent of this log entry.
        /// </summary>
        [XmlElement("parent")]
        public Collection<LogEntryParentNode> Parents
        {
            get
            {
                return new Collection<LogEntryParentNode>(_Parents);
            }
        }

        /// <summary>
        /// The named branch this log entry is on.
        /// </summary>
        [XmlElement("branch")]
        public string Branch
        {
            get;
            set;
        }

        /// <summary>
        /// Individual path actions of this log entry.
        /// </summary>
        [XmlArray("paths")]
        public Collection<LogEntryPathNode> PathActions
        {
            get
            {
                return new Collection<LogEntryPathNode>(_PathActions);
            }
        }
    }
}