using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text.RegularExpressions;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg showconfig" command (http://www.selenic.com/mercurial/hg.1.html#showconfig)
    /// </summary>
    public sealed class ShowConfigCommand : CommandBase<ShowConfigCommand>, IMercurialCommand<IEnumerable<ConfigurationEntry>>
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ShowConfigCommand"/> class.
        /// </summary>
        public ShowConfigCommand()
            : base("showconfig")
        {
        }

        #region IMercurialCommand<IEnumerable<ConfigurationEntry>> Members

        /// <summary>
        /// The result from the command line execution, as an appropriately typed
        /// value.
        /// </summary>
        public IEnumerable<ConfigurationEntry> Result
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
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all.
        /// </remarks>
        protected override void ParseStandardOutputForResults(int exitCode, string standardOutput)
        {
            var re = new Regex(@"^(?<section>[^.]+)\.(?<name>[^=]+)=(?<value>.*)$", RegexOptions.None);

            var entries = new List<ConfigurationEntry>();
            using (var reader = new StringReader(standardOutput))
            {
                string line;

                while ((line = reader.ReadLine()) != null)
                {
                    Match ma = re.Match(line);
                    if (ma.Success)
                    {
                        entries.Add(new ConfigurationEntry(ma.Groups["section"].Value.Trim(), ma.Groups["name"].Value.Trim(),
                            ma.Groups["value"].Value.Trim()));
                    }
                }
            }
            Result = entries;
        }
    }
}