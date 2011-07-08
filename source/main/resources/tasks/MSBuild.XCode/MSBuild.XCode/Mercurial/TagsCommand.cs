using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg tags" command (http://www.selenic.com/mercurial/hg.1.html#tags):
    /// retrieve the tags.
    /// </summary>
    public sealed class TagsCommand : CommandBase<TagsCommand>, IMercurialCommand<IEnumerable<Tag>>
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="TagsCommand"/> class.
        /// </summary>
        public TagsCommand()
            : base("tags")
        {
        }

        #region IMercurialCommand<string> Members

        /// <summary>
        /// The result from the command line execution, as an appropriately typed
        /// value.
        /// </summary>
        public IEnumerable<Tag> Result
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
            base.ParseStandardOutputForResults(exitCode, standardOutput);

            if (StringEx.IsNullOrWhiteSpace(standardOutput))
            {
                List<Tag> tags = new List<Tag>();
                Result = tags;
            }
            else
            {
                List<Tag> tags = new List<Tag>();
                string[] lines = standardOutput.Split(new char[] { '\n' }, StringSplitOptions.RemoveEmptyEntries);
                foreach (string line in lines)
                {
                    Tag tag = Tag.Parse(line);
                    tags.Add(tag);
                }
                Result = tags;
            }
        }
    }
}