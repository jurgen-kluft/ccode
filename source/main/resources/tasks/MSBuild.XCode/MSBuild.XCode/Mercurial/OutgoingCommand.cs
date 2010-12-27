using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg outgoing" command (http://www.selenic.com/mercurial/hg.1.html#outgoing)
    /// </summary>
    public sealed class OutgoingCommand : CommandBase<OutgoingCommand>, IMercurialCommand<IEnumerable<Changeset>>
    {
        private readonly List<RevSpec> _Revisions = new List<RevSpec>();
        private string _Destination = string.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="OutgoingCommand"/> class.
        /// </summary>
        public OutgoingCommand()
            : base("outgoing")
        {
        }

        /// <summary>
        /// Gets or sets the destination to pull from. If <see cref="string.Empty"/>, push to the
        /// default destination. Default is <see cref="string.Empty"/>.
        /// </summary>
        [NullableArgument]
        [DefaultValue("")]
        public string Destination
        {
            get
            {
                return _Destination;
            }
            set
            {
                _Destination = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets or sets whether to run even if the repositories are unrelated. Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--force")]
        [DefaultValue(false)]
        public bool Force
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the collection of revisions intended to be included in the destination.
        /// If empty, include all changes. Default is empty.
        /// </summary>
        [RepeatableArgument(Option = "--rev")]
        public Collection<RevSpec> Revisions
        {
            get
            {
                return new Collection<RevSpec>(_Revisions);
            }
        }

        /// <summary>
        /// Gets or sets whether to recurse into subrepositories.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--subrepos")]
        [DefaultValue(false)]
        public bool RecurseSubRepositories
        {
            get;
            set;
        }

        #region IMercurialCommand<IEnumerable<Changeset>> Members

        /// <summary>
        /// Gets all the arguments to the <see cref="CommandBase{T}.Command"/>, or an
        /// empty array if there are none.
        /// </summary>
        /// <value></value>
        public override IEnumerable<string> Arguments
        {
            get
            {
                return new[] { "--style=XML", "--quiet" }.Concat(base.Arguments).ToArray();
            }
        }

        /// <summary>
        /// The result from the command line execution, as an appropriately typed
        /// value.
        /// </summary>
        public IEnumerable<Changeset> Result
        {
            get;
            private set;
        }

        #endregion

        /// <summary>
        /// Sets the <see cref="Destination"/> property to the specified value and
        /// returns this <see cref="OutgoingCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Destination"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="OutgoingCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public OutgoingCommand WithDestination(string value)
        {
            Destination = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Force"/> property to the specified value and
        /// returns this <see cref="OutgoingCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Force"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="OutgoingCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public OutgoingCommand WithForce(bool value = true)
        {
            Force = value;
            return this;
        }

        /// <summary>
        /// Adds the value to the <see cref="Revisions"/> collection property and
        /// returns this <see cref="OutgoingCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Revisions"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="OutgoingCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public OutgoingCommand WithRevision(RevSpec value)
        {
            Revisions.Add(value);
            return this;
        }

        /// <summary>
        /// Sets the <see cref="RecurseSubRepositories"/> property to the specified value and
        /// returns this <see cref="OutgoingCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="RecurseSubRepositories"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="OutgoingCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public OutgoingCommand WithRecurseSubRepositories(bool value)
        {
            RecurseSubRepositories = value;
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
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all.
        /// </remarks>
        protected override void ParseStandardOutputForResults(int exitCode, string standardOutput)
        {
            Result = ChangesetXmlParser.Parse(standardOutput);
        }

        /// <summary>
        /// This method should throw the appropriate exception depending on the contents of
        /// the <paramref name="exitCode"/> and <paramref name="standardErrorOutput"/>
        /// parameters, or simply return if the execution is considered successful.
        /// </summary>
        /// <param name="exitCode">
        /// The exit code from executing the command line client.
        /// </param>
        /// <param name="standardOutput">
        /// The standard output from executing the command line client.
        /// </param>
        /// <param name="standardErrorOutput">
        /// The standard error output from executing the command client.
        /// </param>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all. The default behavior is to throw a <see cref="MercurialExecutionException"/>
        /// if <paramref name="exitCode"/> is not zero. If you require different behavior, don't call the base
        /// method.
        /// </remarks>
        protected override void ThrowOnUnsuccessfulExecution(int exitCode, string standardOutput, string standardErrorOutput)
        {
            switch (exitCode)
            {
                case 0:
                case 1:
                    break;

                default:
                    base.ThrowOnUnsuccessfulExecution(exitCode, standardOutput, standardErrorOutput);
                    break;
            }
        }
    }
}