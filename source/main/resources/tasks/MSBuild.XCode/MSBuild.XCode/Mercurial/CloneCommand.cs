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
    /// This class implements the "hg clone" command (http://www.selenic.com/mercurial/hg.1.html#clone)
    /// </summary>
    public sealed class CloneCommand : CommandBase<CloneCommand>
    {
        private readonly List<string> _Branches = new List<string>();
        private readonly List<RevSpec> _Revisions = new List<RevSpec>();
        private string _Source = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="CloneCommand"/> class.
        /// </summary>
        public CloneCommand()
            : base("clone")
        {
            Update = true;
            CompressedTransfer = true;
        }

        /// <summary>
        /// Gets or sets the source path or Uri to clone from.
        /// </summary>
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
        /// Gets or sets whether to update the clone with a working folder.
        /// Default is <c>true</c>.
        /// </summary>
        [BooleanArgument(FalseOption = "--noupdate")]
        [DefaultValue(true)]
        public bool Update
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the <see cref="Revisions"/> to update the working
        /// folder to, or <c>null</c> to update to the tip. Default is <c>null</c>.
        /// </summary>
        [NullableArgument(NonNullOption = "--updaterev")]
        [DefaultValue(null)]
        public RevSpec UpdateToRevision
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to use compressed transfer or not. Over LAN, uncompressed is faster, otherwise
        /// compressed is most likely faster. Default is <c>true</c>.
        /// </summary>
        [BooleanArgument(FalseOption = "--uncompressed")]
        [DefaultValue(true)]
        public bool CompressedTransfer
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the collection of revisions to include in the clone. If empty, include every changeset
        /// from the source repository. Default is empty.
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
        /// Gets the collection of branches to include in the clone. If empty, include every branch
        /// from the source repository. Default is empty.
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
        /// Gets all the arguments to the <see cref="CommandBase{T}.Command"/>, or an
        /// empty array if there are none.
        /// </summary>
        /// <value></value>
        public override IEnumerable<string> Arguments
        {
            get
            {
                return base.Arguments.Concat(new[] { "\"" + Source + "\"", ".", });
            }
        }

        /// <summary>
        /// Sets the <see cref="Source"/> property to the specified value and
        /// returns this <see cref="CloneCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Source"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CloneCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CloneCommand WithSource(string value)
        {
            Source = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Update"/> property to the specified value and
        /// returns this <see cref="CloneCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Update"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CloneCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CloneCommand WithUpdate(bool value)
        {
            Update = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="UpdateToRevision"/> property to the specified value and
        /// returns this <see cref="CloneCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="UpdateToRevision"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CloneCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CloneCommand WithUpdateToRevision(RevSpec value)
        {
            UpdateToRevision = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="CompressedTransfer"/> property to the specified value and
        /// returns this <see cref="CloneCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="CompressedTransfer"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="CloneCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CloneCommand WithCompressedTransfer(bool value)
        {
            CompressedTransfer = value;
            return this;
        }

        /// <summary>
        /// Adds the value to the <see cref="Revisions"/> collection property and
        /// returns this <see cref="CloneCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Revisions"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="CloneCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CloneCommand WithRevision(RevSpec value)
        {
            Revisions.Add(value);
            return this;
        }

        /// <summary>
        /// Adds the value to the <see cref="Branches"/> collection property and
        /// returns this <see cref="CloneCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="Branches"/> collection property.
        /// </param>
        /// <returns>
        /// This <see cref="CloneCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public CloneCommand WithBranch(string value)
        {
            Branches.Add(value);
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

            if (StringEx.IsNullOrWhiteSpace(Source))
                throw new InvalidOperationException("The 'clone' command requires Source to be specified");
        }
    }
}