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
    /// This class implements the "hg pull" command (http://www.selenic.com/mercurial/hg.1.html#pull)
    /// </summary>
    public sealed class PullCommand : CommandBase<PullCommand>
    {
        private readonly List<string> _Branches = new List<string>();
        private readonly List<RevSpec> _Revisions = new List<RevSpec>();
        private string _Source = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="PullCommand"/> class.
        /// </summary>
        public PullCommand()
            : base("pull")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets whether to update to new branch head if changes were pulled.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--update")]
        [DefaultValue(false)]
        public bool Update
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to force pulling from the source, even if the
        /// source repository is unrelated.
        /// </summary>
        [BooleanArgument(TrueOption = "--force")]
        [DefaultValue(false)]
        public bool Force
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the source to pull from. If <see cref="String.Empty"/>, pull from the
        /// default source. Default is <see cref="String.Empty"/>.
        /// </summary>
        [NullableArgument]
        [DefaultValue("")]
        public string Source
        {
            get
            {
                return _Source;
            }
            set
            {
                _Source = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets the collection of branches to pull. If empty, pull all branches.
        /// Default is empty.
        /// </summary>
        [RepeatableArgument(Option = "--branch")]
        public Collection<string> Branches
        {
            get
            {
                return new Collection<string>(_Branches);
            }
        }

        /// <summary>
        /// Gets the collection of revisions to include from the <see cref="Source"/>.
        /// If empty, pull all changes. Default is empty.
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
        /// Sets the <see cref="Update"/> property to the specified value and
        /// returns this <see cref="PullCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Update"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="PullCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PullCommand WithUpdate(bool value = true)
        {
            Update = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Force"/> property to the specified value and
        /// returns this <see cref="PullCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Force"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="PullCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PullCommand WithForce(bool value = true)
        {
            Force = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Source"/> property to the specified value and
        /// returns this <see cref="PullCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Source"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="PullCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PullCommand WithSource(string value)
        {
            Source = value;
            return this;
        }

        /// <summary>
        /// Adds the value to the <see cref="Branches"/> collection property and
        /// returns this <see cref="PullCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Branches"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="PullCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PullCommand WithBranch(string value)
        {
            Branches.Add(value);
            return this;
        }

        /// <summary>
        /// Adds the value to the <see cref="Revisions"/> collection property and
        /// returns this <see cref="PullCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Revisions"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="PullCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PullCommand WithRevision(RevSpec value)
        {
            Revisions.Add(value);
            return this;
        }
    }
}