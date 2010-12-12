using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Text.RegularExpressions;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg status" command (http://www.selenic.com/mercurial/hg.1.html#status)
    /// </summary>
    public sealed class StatusCommand : CommandBase<StatusCommand>, IMercurialCommand<IEnumerable<FileStatus>>
    {
        private static readonly Dictionary<char, FileState> _FileStateCodes = new Dictionary<char, FileState>
            {
                { 'M', FileState.Modified },
                { 'A', FileState.Added },
                { 'R', FileState.Removed },
                { 'C', FileState.Clean },
                { '!', FileState.Missing },
                { '?', FileState.Unknown },
                { 'I', FileState.Ignored },
            };

        /// <summary>
        /// Initializes a new instance of the <see cref="StatusCommand"/> class.
        /// </summary>
        public StatusCommand()
            : base("status")
        {
            Include = FileStatusIncludes.Default;
        }

        /// <summary>
        /// Specify which kind of status codes to include. Default is
        /// <see cref="FileStatusIncludes.Default"/>.
        /// </summary>
        [DefaultValue(FileStatusIncludes.Default)]
        public FileStatusIncludes Include
        {
            get;
            set;
        }

        #region IMercurialCommand<IEnumerable<FileStatus>> Members

        /// <summary>
        /// Gets all the arguments to the <see cref="CommandBase{T}.Command"/>, or an
        /// empty array if there are none.
        /// </summary>
        /// <value></value>
        public override IEnumerable<string> Arguments
        {
            get
            {
                var result = new List<string>(base.Arguments);
                if ((Include & FileStatusIncludes.Added) != 0)
                    result.Add("--added");
                if ((Include & FileStatusIncludes.Clean) != 0)
                    result.Add("--clean");
                if ((Include & FileStatusIncludes.Ignored) != 0)
                    result.Add("--ignored");
                if ((Include & FileStatusIncludes.Missing) != 0)
                    result.Add("--deleted");
                if ((Include & FileStatusIncludes.Modified) != 0)
                    result.Add("--modified");
                if ((Include & FileStatusIncludes.Removed) != 0)
                    result.Add("--removed");
                if ((Include & FileStatusIncludes.Unknown) != 0)
                    result.Add("--unknown");
                return result.ToArray();
            }
        }

        /// <summary>
        /// The result from the command line execution, as an appropriately typed
        /// value.
        /// </summary>
        public IEnumerable<FileStatus> Result
        {
            get;
            private set;
        }

        #endregion

        /// <summary>
        /// Sets the <see cref="Include"/> property to the specified value and
        /// returns this <see cref="StatusCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Include"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="StatusCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public StatusCommand WithInclude(FileStatusIncludes value)
        {
            Include = value;
            return this;
        }

        /// <summary>
        /// Parses the standard output for results.
        /// </summary>
        /// <param name="exitCode">The exit code.</param>
        /// <param name="standardOutput">The standard output.</param>
        protected override void ParseStandardOutputForResults(int exitCode, string standardOutput)
        {
            base.ParseStandardOutputForResults(exitCode, standardOutput);

            var result = new List<FileStatus>();

            var re = new Regex(@"^(?<status>[MARC!?I ])\s+(?<path>.*)$");
            var statusEntries = from line in standardOutput.Split('\n', '\r')
                                where !StringEx.IsNullOrWhiteSpace(line)
                                let ma = re.Match(line)
                                where ma.Success
                                select new { status = ma.Groups["status"].Value[0], path = ma.Groups["path"].Value };
            foreach (var entry in statusEntries)
            {
                FileState state;
                if (_FileStateCodes.TryGetValue(entry.status, out state))
                    result.Add(new FileStatus(state, entry.path));
                else
                {
                    if (entry.status == ' ')
                    {
                        throw new InvalidOperationException("Status does not yet support the Added sub-state to show where the file was added from");
                    }
                    else
                        throw new InvalidOperationException("Unknown status code reported by Mercurial: '" + entry.status +
                                                            "', I do not know how to handle that");
                }
            }

            Result = result;
        }
    }
}