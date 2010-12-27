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
    /// This class implements the "hg add" command (http://www.selenic.com/mercurial/hg.1.html#add)
    /// </summary>
    public sealed class AddCommand : IncludeExcludeCommandBase<AddCommand>
    {
        private readonly List<string> _Paths = new List<string>();

        /// <summary>
        /// Initializes a new instance of <see cref="AddCommand"/>.
        /// </summary>
        public AddCommand()
            : base("add")
        {
        }

        /// <summary>
        /// Gets the collection of path patterns to add to the repository.
        /// </summary>
        [RepeatableArgument]
        public Collection<string> Paths
        {
            get
            {
                return new Collection<string>(_Paths);
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

        /// <summary>
        /// Sets the <see cref="RecurseSubRepositories"/> property to the specified value and
        /// returns this <see cref="AddCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="RecurseSubRepositories"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="AddCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public AddCommand WithRecurseSubRepositories(bool value)
        {
            RecurseSubRepositories = value;
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

            if (Paths.Count == 0)
                throw new InvalidOperationException("The 'add' command requires at least one path specified");
        }
    }
}