using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg forget" command (http://www.selenic.com/mercurial/hg.1.html#forget)
    /// </summary>
    public sealed class ForgetCommand : IncludeExcludeCommandBase<ForgetCommand>
    {
        private readonly List<string> _Paths = new List<string>();

        /// <summary>
        /// Initializes a new instance of <see cref="ForgetCommand"/>.
        /// </summary>
        public ForgetCommand()
            : base("forget")
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
        /// Adds the value to the <see cref="Paths"/> collection property and
        /// returns this <see cref="ForgetCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Paths"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="ForgetCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public ForgetCommand WithPath(string value)
        {
            Paths.Add(value);
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