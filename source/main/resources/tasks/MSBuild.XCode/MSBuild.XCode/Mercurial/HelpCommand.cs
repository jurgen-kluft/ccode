using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg help" command (http://www.selenic.com/mercurial/hg.1.html#help)
    /// </summary>
    public sealed class HelpCommand : CommandBase<HelpCommand>, IMercurialCommand<string>
    {
        private string _Topic = String.Empty;

        /// <summary>
        /// Initializes a new instance of <see cref="HelpCommand"/>.
        /// </summary>
        public HelpCommand()
            : base("help")
        {
        }

        /// <summary>
        /// Gets or sets which topic (command name, or other topics) to request help on.
        /// If left empty, will request the main help text.
        /// Default is <see cref="String.Empty"/>.
        /// </summary>
        [NullableArgument]
        [DefaultValue("")]
        public string Topic
        {
            get
            {
                return _Topic;
            }
            set
            {
                _Topic = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets or sets whether to include global help information.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "-v")]
        [DefaultValue(false)]
        public bool IncludeGlobalHelp
        {
            get;
            set;
        }

        #region IMercurialCommand<string> Members

        /// <summary>
        /// The result from the command line execution, as an appropriately typed
        /// value.
        /// </summary>
        public string Result
        {
            get;
            private set;
        }

        #endregion

        /// <summary>
        /// Sets the <see cref="Topic"/> property to the specified value and
        /// returns this <see cref="HelpCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Topic"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="HelpCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public HelpCommand WithTopic(string value)
        {
            Topic = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="IncludeGlobalHelp"/> property to the specified value and
        /// returns this <see cref="HelpCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="IncludeGlobalHelp"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="HelpCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public HelpCommand WithIncludeGlobalHelp(bool value = true)
        {
            IncludeGlobalHelp = value;
            return this;
        }

        /// <summary>
        /// This method should parse and store the appropriate execution result output
        /// according to the type of data the command line client would return for
        /// the command.
        /// </summary>
        /// <param name="exitCode">
        /// The exit code from executing the command line client.
        /// </param>
        /// <param name="standardOutput">
        /// The standard output from executing the command line client.
        /// </param>
        protected override void ParseStandardOutputForResults(int exitCode, string standardOutput)
        {
            base.ParseStandardOutputForResults(exitCode, standardOutput);

            Result = standardOutput.Trim();
        }
    }
}