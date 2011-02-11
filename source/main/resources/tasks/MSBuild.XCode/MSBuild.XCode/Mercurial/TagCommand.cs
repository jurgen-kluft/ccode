using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Globalization;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg tag" command (http://www.selenic.com/mercurial/hg.1.html#tag):
    /// add one or more tags for the current or given revision.
    /// </summary>
    public sealed class TagCommand : CommandBase<TagCommand>
    {
        private TagAction _Action = TagAction.Add;
        private string _Message = String.Empty;
        private string _Name = String.Empty;
        private string _OverrideAuthor = String.Empty;

        /// <summary>
        /// Initializes a new instance of the <see cref="TagCommand"/> class.
        /// </summary>
        public TagCommand()
            : base("tag")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets the name of the tag to add or remove.
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
                _Name = (value ?? String.Empty);
            }
        }

        /// <summary>
        /// Gets or sets whether to add or remove a local tag. If <c>false</c>, a changeset
        /// will be committed that edits the .hgtags file accordingly.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--local")]
        [DefaultValue(false)]
        public bool IsLocal
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to add or remove the tag.
        /// Default is <see cref="TagAction.Add"/>.
        /// </summary>
        [DefaultValue(TagAction.Add)]
        public TagAction Action
        {
            get
            {
                return _Action;
            }
            set
            {
                _Action = value;
            }
        }

        /// <summary>
        /// Gets or sets whether to replace an existing tag, in effect moving the tag to
        /// a different changeset. Without this flag, adding a tag that already exists
        /// will result in a <see cref="MercurialExecutionException"/> being thrown.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--force")]
        [DefaultValue(true)]
        public bool Force
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to replace an existing tag, in effect moving the tag to
        /// a different changeset. Without this flag, adding a tag that already exists
        /// will result in a <see cref="MercurialExecutionException"/> being thrown.
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--remove")]
        [DefaultValue(false)]
        public bool Remove
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the commit message to use when committing the tag.
        /// </summary>
        [DefaultValue("")]
        [NullableArgument(NonNullOption = "--message")]
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
        /// Gets or sets which revision to tag, or <c>null</c> for the parent of the
        /// working folder.
        /// Default is <c>null</c>.
        /// </summary>
        [DefaultValue(null)]
        [NullableArgument(NonNullOption = "--rev")]
        public RevSpec Revision
        {
            get;
            set;
        }

        /// <summary>
        /// Gets all the arguments to the <see cref="CommandBase{T}.Command"/>, or an
        /// empty array if there are none.
        /// </summary>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to access
        /// the base property at all, but you are required to return a non-<c>null</c> array reference,
        /// even for an empty array.
        /// </remarks>
        public override IEnumerable<string> Arguments
        {
            get
            {
                List<string> result = base.Arguments.ToList();

                switch (Action)
                {
                    case TagAction.Add:
                        break;

                    case TagAction.Remove:
                        result.Add("--remove");
                        break;

                    default:
                        throw new InvalidOperationException(string.Format(CultureInfo.InvariantCulture,
                            "Unknown TagAction specified for TagCommand ({0})", Action));
                }

                return result;
            }
        }

        /// <summary>
        /// Sets the <see cref="IsLocal"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="IsLocal"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithIsLocal(bool value = true)
        {
            IsLocal = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Action"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Action"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithAction(TagAction value)
        {
            Action = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="ReplaceExisting"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="ReplaceExisting"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithForce(bool value = true)
        {
            Force = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Message"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Message"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithMessage(string value)
        {
            Message = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="OverrideAuthor"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="OverrideAuthor"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithOverrideAuthor(string value)
        {
            OverrideAuthor = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="OverrideTimestamp"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="OverrideTimestamp"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithOverrideTimestamp(DateTime value)
        {
            OverrideTimestamp = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Revision"/> property to the specified value and
        /// returns this <see cref="TagCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Revision"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="TagCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public TagCommand WithRevision(RevSpec value)
        {
            Revision = value;
            return this;
        }

        /// <summary>
        /// Validates the command configuration. This method should throw the necessary
        /// exceptions to signal missing or incorrect configuration (like attempting to
        /// add files to the repository without specifying which files to add.)
        /// </summary>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all.
        /// </remarks>
        public override void Validate()
        {
            base.Validate();

            if (StringEx.IsNullOrWhiteSpace(Name))
                throw new InvalidOperationException("The name of the tag to add or remove must be set for TagCommand");
        }
    }
}