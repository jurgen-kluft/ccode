using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Xml.Serialization;

namespace Mercurial.XmlSerializationTypes
{
    /// <summary>
    /// This class encapsulates a &lt;path...&gt; node in the log output.
    /// </summary>
    [XmlType("path")]
    [EditorBrowsable(EditorBrowsableState.Never)]
    public class LogEntryPathNode
    {
        /// <summary>
        /// The type of action performed on the path.
        /// </summary>
        [XmlAttribute("action")]
        public string Action
        {
            get;
            set;
        }

        /// <summary>
        /// The path that was involved.
        /// </summary>
        [XmlText]
        public string Path
        {
            get;
            set;
        }
    }
}