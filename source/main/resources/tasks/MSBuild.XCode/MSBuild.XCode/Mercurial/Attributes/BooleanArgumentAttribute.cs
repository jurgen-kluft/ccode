using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial.Attributes
{
    /// <summary>
    /// This attribute can be applied to bool-properties in option classes,
    /// to specify the option to pass to the executable in case of
    /// a <c>false</c> or <c>true</c> value. By default, no options
    /// will be sent for either value.
    /// </summary>
    [AttributeUsage(AttributeTargets.Property, AllowMultiple = false, Inherited = true)]
    public sealed class BooleanArgumentAttribute : ArgumentAttribute
    {
        private string _FalseOption = String.Empty;
        private string _TrueOption = String.Empty;

        /// <summary>
        /// Gets or sets the option to pass to the Mercurial executable if
        /// the tagged bool property has a <c>false</c> value,
        /// or <see cref="String.Empty"/> if no option should be passed in
        /// that case.
        /// </summary>
        public string FalseOption
        {
            get
            {
                return _FalseOption;
            }
            set
            {
                _FalseOption = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets or sets the option to pass to the Mercurial executable if
        /// the tagged bool property has a <c>true</c> value,
        /// or <see cref="String.Empty"/> if no option should be passed in
        /// that case.
        /// </summary>
        public string TrueOption
        {
            get
            {
                return _TrueOption;
            }
            set
            {
                _TrueOption = (value ?? String.Empty).Trim();
            }
        }

        /// <summary>
        /// Gets a collection of options or arguments to pass to the Mercurial
        /// executable, based on the property value from the options class.
        /// </summary>
        /// <param name="propertyValue">
        /// The property value from the tagged property of the options class.
        /// </param>
        /// <returns>
        /// A collection of options or arguments, or an empty array or <c>null</c>
        /// for no options for the specified property value.
        /// </returns>
        public override string[] GetOptions(object propertyValue)
        {
            string result;
            if (propertyValue == null)
                result = FalseOption;
            else if (propertyValue is bool)
            {
                if ((bool) propertyValue)
                    result = TrueOption;
                else
                    result = FalseOption;
            }
            else
                throw new InvalidOperationException("BooleanArgumentAttribute applied to non-bool property");

            if (String.IsNullOrEmpty(result))
                return null;

            return new[] { result };
        }
    }
}