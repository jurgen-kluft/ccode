using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Globalization;
using System.IO;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg commit" command (http://www.selenic.com/mercurial/hg.1.html#commit)
    /// </summary>
    public sealed class CommitCommand : IncludeExcludeCommandBase<CommitCommand>
    {
        private string _Message = String.Empty;
        private string _MessageFilePath;
        private string _OverrideAuthor = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="CommitCommand"/> class.
        /// </summary>
        public CommitCommand()
            : base("commit")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets the commit message to use when committing.
        /// </summary>
        [DefaultValue("")]
        public string Message
        {
            get
            {
                return _Message;
            }
            set
            {
                _Message = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets or sets whether to automatically add new files and remove missing files before committing.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--addremove")]
        [DefaultValue(false)]
        public bool AddRemove
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to mark a branch as closed, hiding it from the branch list.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--close-branch")]
        [DefaultValue(false)]
        public bool CloseBranch
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the username to use when committing;
        /// or <see cref="String.Empty"/> to use the username configured in the repository or by
        /// the current user. Default is <see cref="String.Empty"/>.
        /// </summary>
        [NullableArgument(NonNullOption = "--user")]
        [DefaultValue("")]
        public string OverrideAuthor
        {
            get
            {
                return _OverrideAuthor;
            }
            set
            {
                _OverrideAuthor = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets or sets the timestamp <see cref="DateTime"/> to use when committing;
        /// or <c>null</c> which means use the current date and time. Default is <c>null</c>.
        /// </summary>
        [DateTimeArgument(NonNullOption = "--date")]
        [DefaultValue(null)]
        public DateTime? OverrideTimestamp
        {
            get;
            set;
        }

        /// <summary>
        /// Gets all the arguments to the <see cref="CommandBase{T}.Command"/>, or an
        /// empty array if there are none.
        /// </summary>
        /// <value></value>
        public override IEnumerable<string> Arguments
        {
            get
            {
                return base.Arguments.Concat(new[] { "--logfile", string.Format(CultureInfo.InvariantCulture, "\"{0}\"", _MessageFilePath), });
            }
        }

        /// <summary>
        /// Sets the <see cref="Message"/> property to the specified value and
        /// returns this <see cref="CommitCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Message"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CommitCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CommitCommand WithMessage(string value)
        {
            Message = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="AddRemove"/> property to the specified value and
        /// returns this <see cref="CommitCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="AddRemove"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="CommitCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CommitCommand WithAddRemove(bool value = true)
        {
            AddRemove = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="CloseBranch"/> property to the specified value and
        /// returns this <see cref="CommitCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="CloseBranch"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="CommitCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CommitCommand WithCloseBranch(bool value = true)
        {
            CloseBranch = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="OverrideAuthor"/> property to the specified value and
        /// returns this <see cref="CommitCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="OverrideAuthor"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CommitCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CommitCommand WithOverrideAuthor(string value)
        {
            OverrideAuthor = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="OverrideTimestamp"/> property to the specified value and
        /// returns this <see cref="CommitCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="OverrideTimestamp"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CommitCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CommitCommand WithOverrideTimestamp(DateTime value)
        {
            OverrideTimestamp = value;
            return this;
        }

        /// <summary>
        /// Validates the command configuration. This method should throw the necessary
        /// exceptions to signal missing or incorrect configuration (like attempting to
        /// add files to the repository without specifying which files to add.)
        /// </summary>
        public override void Validate()
        {
            base.Validate();

            if (StringEx.IsNullOrWhiteSpace(Message))
                throw new InvalidOperationException("The 'commit' command requires Message to be specified");
        }

        /// <summary>
        /// Override this method to implement code that will execute before command
        /// line execution.
        /// </summary>
        protected override void Prepare()
        {
            _MessageFilePath = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString().Replace("-", "").ToLowerInvariant() + ".txt");
            File.WriteAllText(_MessageFilePath, Message);
        }

        /// <summary>
        /// Override this method to implement code that will execute after command
        /// line execution.
        /// </summary>
        protected override void Cleanup()
        {
            if (_MessageFilePath != null && File.Exists(_MessageFilePath))
                File.Delete(_MessageFilePath);
        }
    }
}