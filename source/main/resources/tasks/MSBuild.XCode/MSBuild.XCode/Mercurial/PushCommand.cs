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
    /// This class implements the "hg push" command (http://www.selenic.com/mercurial/hg.1.html#push)
    /// </summary>
    public sealed class PushCommand : CommandBase<PushCommand>
    {
        private readonly List<RevSpec> _Revisions = new List<RevSpec>();
        private string _Destination = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="PushCommand"/> class.
        /// </summary>
        public PushCommand()
            : base("push")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets the destination to pull from. If <see cref="String.Empty"/>, push to the
        /// default destination. Default is <see cref="String.Empty"/>.
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
        /// Gets or sets whether to force push to the destination, even if
        /// the repositories are unrelated, or pushing would create new heads in the
        /// destination repository. Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--force")]
        [DefaultValue(false)]
        public bool Force
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to allow creating a new branch in the destination
        /// repository. Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--new-branch")]
        [DefaultValue(false)]
        public bool AllowCreatingNewBranch
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to push large repositories in chunks.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--chunked")]
        [DefaultValue(false)]
        public bool ChunkedTransfer
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the collection of revisions to include when pushing.
        /// If empty, push all changes. Default is empty.
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
        /// Sets the <see cref="Destination"/> property to the specified value and
        /// returns this <see cref="PushCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Destination"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="PushCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PushCommand WithDestination(string value)
        {
            Destination = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Force"/> property to the specified value and
        /// returns this <see cref="PushCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Force"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="PushCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PushCommand WithForce(bool value = true)
        {
            Force = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="AllowCreatingNewBranch"/> property to the specified value and
        /// returns this <see cref="PushCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="AllowCreatingNewBranch"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="PushCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PushCommand WithAllowCreatingNewBranch(bool value = true)
        {
            AllowCreatingNewBranch = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="ChunkedTransfer"/> property to the specified value and
        /// returns this <see cref="PushCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="ChunkedTransfer"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="PushCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PushCommand WithChunkedTransfer(bool value)
        {
            ChunkedTransfer = value;
            return this;
        }

        /// <summary>
        /// Adds the value to the <see cref="Revisions"/> collection property and
        /// returns this <see cref="PushCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Revisions"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="PushCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public PushCommand WithRevision(RevSpec value)
        {
            Revisions.Add(value);
            return this;
        }
    }
}