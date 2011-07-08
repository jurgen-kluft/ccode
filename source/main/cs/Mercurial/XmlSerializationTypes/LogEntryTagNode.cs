using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Xml.Serialization;

namespace Mercurial.XmlSerializationTypes
{
    /// <summary>
    /// This class encapsulates a &lt;tag&gt;...&lt;/tag&gt; node in the log output.
    /// </summary>
    [XmlType("tag")]
    [EditorBrowsable(EditorBrowsableState.Never)]
    public class LogEntryTagNode
    {
        private string _Name = String.Empty;

        /// <summary>
        /// Gets or sets the name of the tag.
        /// </summary>
        [XmlText]
        public string Name
        {
            get
            {
                return _Name;
            }
            set
            {
                _Name = (value ?? String.Empty).Trim();
            }
        }
    }
}