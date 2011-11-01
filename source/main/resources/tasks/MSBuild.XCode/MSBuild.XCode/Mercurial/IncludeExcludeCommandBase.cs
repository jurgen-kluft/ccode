using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This is the base class for option classes for various commands for
    /// the Mercurial client that supports the "--include PATTERN" and
    /// "--exclude PATTERN" options.
    /// </summary>
    /// <typeparam name="T">
    /// The actual type descending from <see cref="IncludeExcludeCommandBase{T}"/>, used to generate type-correct
    /// methods in this base class.
    /// </typeparam>
    public abstract class IncludeExcludeCommandBase<T> : CommandBase<T>
        where T : IncludeExcludeCommandBase<T>
    {
        private readonly List<string> _ExcludePatterns = new List<string>();
        private readonly List<string> _IncludePatterns = new List<string>();

        /// <summary>
        /// Initializes a new instance of the <see cref="IncludeExcludeCommandBase{T}"/> class.
        /// </summary>
        /// <param name="command">The command.</param>
        protected IncludeExcludeCommandBase(string command)
            : base(command)
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets the collection of filename patterns to include for
        /// the command.
        /// </summary>
        [RepeatableArgument(Option = "-I")]
        public Collection<string> IncludePatterns
        {
            get
            {
                return new Collection<string>(_IncludePatterns);
            }
        }

        /// <summary>
        /// Gets the collection of filename patterns to exclude
        /// for the command.
        /// </summary>
        [RepeatableArgument(Option = "-X")]
        public Collection<string> ExcludePatterns
        {
            get
            {
                return new Collection<string>(_ExcludePatterns);
            }
        }

        /// <summary>
        /// Adds the value to the <see cref="IncludePatterns"/> collection property and
        /// returns this instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="IncludePatterns"/> collection property.
        /// </param>
        /// <returns>
        /// This instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public T WithIncludePattern(string value)
        {
            IncludePatterns.Add(value);
            return (T) this;
        }

        /// <summary>
        /// Adds the value to the <see cref="ExcludePatterns"/> collection property and
        /// returns this instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="ExcludePatterns"/> collection property.
        /// </param>
        /// <returns>
        /// This instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public T WithExcludePattern(string value)
        {
            ExcludePatterns.Add(value);
            return (T) this;
        }
    }
}