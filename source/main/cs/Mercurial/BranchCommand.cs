using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg branch" command (http://www.selenic.com/mercurial/hg.1.html#branch)
    /// </summary>
    public sealed class BranchCommand : CommandBase<BranchCommand>, IMercurialCommand<string>
    {
        private string _Name = String.Empty;
        private string _Result = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="BranchCommand"/> class.
        /// </summary>
        public BranchCommand()
            : base("branch")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets the new name to assign to the branch.
        /// If left empty, only return the current branch name.
        /// Default is empty.
        /// </summary>
        [NullableArgument]
        [DefaultValue("")]
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

        /// <summary>
        /// Gets or sets whether to force a new name for the branch, even if that
        /// shadows an existing branch elsewhere.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--force")]
        [DefaultValue(false)]
        public bool Force
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to clean the branch name of the current working
        /// folder, resetting it back to the branch name of the parent changeset.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--clean")]
        [DefaultValue(false)]
        public bool Clean
        {
            get;
            set;
        }

        #region IMercurialCommand<string> Members

        /// <summary>
        /// Gets the branch name that the 
        /// </summary>
        public string Result
        {
            get
            {
                return _Result;
            }
            private set
            {
                _Result = (value ?? string.Empty).Trim();
            }
        }

        #endregion

        /// <summary>
        /// Sets the <see cref="Name"/> property to the specified value and
        /// returns this <see cref="BranchCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Name"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="BranchCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public BranchCommand WithName(string value)
        {
            Name = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Force"/> property to the specified value and
        /// returns this <see cref="BranchCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Force"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="BranchCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public BranchCommand WithForce(bool value = true)
        {
            Force = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Clean"/> property to the specified value and
        /// returns this <see cref="BranchCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Clean"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="BranchCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public BranchCommand WithClean(bool value = true)
        {
            Clean = value;
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

            if (StringEx.IsNullOrWhiteSpace(standardOutput))
                Result = Name;
            else
                Result = standardOutput;
        }
    }
}