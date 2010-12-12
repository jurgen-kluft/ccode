using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Globalization;
using System.IO;
using System.Linq;
using System.Text.RegularExpressions;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg annotate" command (http://www.selenic.com/mercurial/hg.1.html#annotate)
    /// </summary>
    public sealed class AnnotateCommand : CommandBase<AnnotateCommand>, IMercurialCommand<IEnumerable<Annotation>>
    {
        private string _Path = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="AnnotateCommand"/> class.
        /// </summary>
        public AnnotateCommand()
            : base("annotate")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets the path to the item to annotate.
        /// </summary>
        [NullableArgument]
        [DefaultValue("")]
        public string Path
        {
            get
            {
                return _Path;
            }
            set
            {
                _Path = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets or sets whether to follow renames and copies when limiting the log.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(FalseOption = "--no-follow")]
        [DefaultValue(false)]
        public bool FollowRenamesAndMoves
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets which revision of the specified item to annotate.
        /// If <c>null</c>, annotate the revision in the working folder.
        /// Default is <c>null</c>.
        /// </summary>
        [NullableArgument(NonNullOption = "--rev")]
        [DefaultValue(null)]
        public RevSpec RevSpec
        {
            get;
            set;
        }

        #region IMercurialCommand<IEnumerable<Annotation>> Members

        /// <summary>
        /// Validates the command configuration. This method should throw the necessary
        /// exceptions to signal missing or incorrect configuration (like attempting to
        /// add files to the repository without specifying which files to add.)
        /// </summary>
        public override void Validate()
        {
            base.Validate();

            if (StringEx.IsNullOrWhiteSpace(Path))
                throw new InvalidOperationException("The 'annotate' command requires Path to be set");
        }

        /// <summary>
        /// The result from the command line execution, as an appropriately typed
        /// value.
        /// </summary>
        public IEnumerable<Annotation> Result
        {
            get;
            private set;
        }

        #endregion

        /// <summary>
        /// Sets the <see cref="Path"/> property to the specified value and
        /// returns this <see cref="AnnotateCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Path"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="AnnotateCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public AnnotateCommand WithPath(string value)
        {
            Path = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="FollowRenamesAndMoves"/> property to the specified value and
        /// returns this <see cref="AnnotateCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="FollowRenamesAndMoves"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="AnnotateCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public AnnotateCommand WithFollowRenamesAndMoves(bool value = true)
        {
            FollowRenamesAndMoves = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="RevSpec"/> property to the specified value and
        /// returns this <see cref="AnnotateCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="RevSpec"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="AnnotateCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public AnnotateCommand WithRevSpec(RevSpec value)
        {
            RevSpec = value;
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

            var result = new List<Annotation>();
            using (var reader = new StringReader(standardOutput))
            {
                var re = new Regex(@"^(?<rev>\d+): (?<line>.*)$", RegexOptions.None);

                string line;
                int lineNumber = 0;
                while ((line = reader.ReadLine()) != null)
                {
                    Match ma = re.Match(line);
                    if (ma.Success)
                        result.Add(new Annotation(lineNumber, Int32.Parse(ma.Groups["rev"].Value, CultureInfo.InvariantCulture),
                            ma.Groups["line"].Value));

                    lineNumber++;
                }
            }
            Result = result;
        }
    }
}