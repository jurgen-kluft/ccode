using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg version" command (http://www.selenic.com/mercurial/hg.1.html#version)
    /// </summary>
    public sealed class VersionCommand : CommandBase<VersionCommand>, IMercurialCommand<string>
    {
        /// <summary>
        /// Initializes a new instance of <see cref="VersionCommand"/>.
        /// </summary>
        public VersionCommand()
            : base("version")
        {
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