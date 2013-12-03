using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg update" command (http://www.selenic.com/mercurial/hg.1.html#update)
    /// </summary>
    public sealed class UpdateCommand : CommandBase<UpdateCommand>
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="UpdateCommand"/> class.
        /// </summary>
        public UpdateCommand()
            : base("update")
        {
            // Do nothing here
        }

        /// <summary>
        /// Gets or sets the <see cref="RevSpec"/> to update the working copy to.
        /// </summary>
        [NullableArgument]
        [DefaultValue(null)]
        public RevSpec Revision
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether to do a clean update, discarding uncommitted changes in the process (no backup.)
        /// Default is <c>false</c>.
        /// </summary>
        [BooleanArgument(TrueOption = "--clean")]
        [DefaultValue(false)]
        public bool Clean
        {
            get;
            set;
        }

        /// <summary>
        /// Sets the <see cref="Revision"/> property to the specified value and
        /// returns this <see cref="UpdateCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Revision"/> property.
        /// </param>
        /// <returns>
        /// This <see cref="UpdateCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public UpdateCommand WithRevision(RevSpec value)
        {
            Revision = value;
            return this;
        }

        /// <summary>
        /// Sets the <see cref="Clean"/> property to the specified value and
        /// returns this <see cref="UpdateCommand"/> instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Clean"/> property,
        /// defaults to <c>true</c>.
        /// </param>
        /// <returns>
        /// This <see cref="UpdateCommand"/> instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public UpdateCommand WithClean(bool value = true)
        {
            Clean = value;
            return this;
        }
    }
}