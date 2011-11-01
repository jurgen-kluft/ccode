using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Xml.Serialization;

namespace Mercurial.XmlSerializationTypes
{
    /// <summary>
    /// This class encapsulates a &lt;author...&gt; node in the log output.
    /// </summary>
    [XmlType("author")]
    [EditorBrowsable(EditorBrowsableState.Never)]
    public class LogEntryAuthorNode
    {
        /// <summary>
        /// The name of the author of the log entry.
        /// </summary>
        [XmlText]
        public string Name
        {
            get;
            set;
        }

        /// <summary>
        /// The email address of the author of the log entry.
        /// </summary>
        [XmlAttribute("email")]
        public string Email
        {
            get;
            set;
        }
    }
}