using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Diagnostics;
using System.Globalization;
using System.Linq;
using Mercurial.Attributes;

namespace Mercurial
{
    /// <summary>
    /// This is the base class for option classes for various commands for
    /// the Mercurial client.
    /// </summary>
    /// <typeparam name="T">
    /// The actual type descending from <see cref="CommandBase{T}"/>, used to generate type-correct
    /// methods in this base class.
    /// </typeparam>
    public abstract class CommandBase<T> : IMercurialCommand
        where T : CommandBase<T>
    {
        private readonly List<string> _AdditionalArguments = new List<string>();
        private readonly string _Command;

        private int _RawExitCode;
        private string _RawStandardErrorOutput = String.Empty;
        private string _RawStandardOutput = String.Empty;
        private int _Timeout = 60;

        /// <summary>
        /// Initializes a new instance of the <see cref="CommandBase{T}"/> class.
        /// </summary>
        /// <param name="command">
        /// The name of the command that will be passed to the Mercurial command line client.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c> or empty.</para>
        /// </exception>
        protected CommandBase(string command)
        {
            if (StringEx.IsNullOrWhiteSpace(command))
                throw new ArgumentNullException("command");

            _Command = command;
        }

        /// <summary>
        /// Gets the raw standard output from executing the command line client.
        /// </summary>
        public string RawStandardOutput
        {
            get
            {
                return _RawStandardOutput;
            }
        }

        /// <summary>
        /// Gets the raw standard error output from executing the command line client.
        /// </summary>
        public string RawStandardErrorOutput
        {
            get
            {
                return _RawStandardErrorOutput;
            }
        }

        /// <summary>
        /// Gets the raw exit code from executing the command line client.
        /// </summary>
        public int RawExitCode
        {
            get
            {
                return _RawExitCode;
            }
        }

        #region IMercurialCommand Members

        /// <summary>
        /// Gets the collection which additional arguments can be added into. This collection
        /// is exposed for extensions, so that they have a place to add all their
        /// extra arguments to the Mercurial command line client.
        /// </summary>
        /// <remarks>
        /// Note that all of these arguments will be appended to the end of the command line,
        /// after all the normal arguments supported by the command classes.
        /// </remarks>
        public Collection<string> AdditionalArguments
        {
            get
            {
                return new Collection<string>(_AdditionalArguments);
            }
        }

        /// <summary>
        /// Gets or sets the object that will act as an observer of command execution.
        /// </summary>
        [DefaultValue(null)]
        public IMercurialCommandObserver Observer
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the timeout to use when executing Mercurial commands, in
        /// seconds. Default is 60.
        /// </summary>
        /// <exception cref="ArgumentOutOfRangeException">
        /// <para><see cref="Timeout"/> cannot be less than 1.</para>
        /// </exception>
        [DefaultValue(60)]
        public int Timeout
        {
            get
            {
                return _Timeout;
            }
            set
            {
                if (value < 1)
                    throw new ArgumentOutOfRangeException("value", value, "Timeout cannot be lower than 1");
                _Timeout = value;
            }
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
        public virtual void Validate()
        {
            // Do nothing by default
        }

        /// <summary>
        /// Gets the command to execute with the Mercurial command line client.
        /// </summary>
        /// <remarks>
        /// Note that this property is required to return a non-null non-empty (including whitespace) string,
        /// as it will be used to specify which command to execute for the Mercurial command line client. You're
        /// not required to call the base property though, as long as you descend from <see cref="CommandBase{T}"/>.
        /// </remarks>
        public virtual string Command
        {
            get
            {
                return _Command;
            }
        }

        /// <summary>
        /// Gets all the arguments to the <see cref="Command"/>, or an
        /// empty array if there are none.
        /// </summary>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to access
        /// the base property at all, but you are required to return a non-<c>null</c> array reference,
        /// even for an empty array.
        /// </remarks>
        public virtual IEnumerable<string> Arguments
        {
            get
            {
                return from prop in GetType().GetProperties()
                       let attributes = prop.GetCustomAttributes(typeof (ArgumentAttribute), true)
                       where attributes.Length > 0
                       let attr = attributes[0] as ArgumentAttribute
                       where attr != null
                       let options = attr.GetOptions(prop.GetValue(this, null))
                       where options != null
                       from option in options
                       select option;
            }
        }

        /// <summary>
        /// This method is called before the command is executed. You can use this to
        /// store temporary files (like a commit message or similar) that the
        /// <see cref="Arguments"/> refer to, before the command is executed.
        /// </summary>
        public void Before()
        {
            Prepare();
        }

        /// <summary>
        /// This method is called after the command has been executed. You can use this to
        /// clean up after the command execution (like removing temporary files), and to
        /// react to the exit code from the command line client. If the exit code is
        /// considered a failure, this method should throw the correct exception.
        /// </summary>
        /// <param name="exitCode">The exit code from the command line client. Typically 0 means success, but this
        /// can vary from command to command.</param>
        /// <param name="standardOutput">The standard output of the execution, or <see cref="string.Empty"/> if there
        /// was none.</param>
        /// <param name="standardErrorOutput">The standard error output of the execution, or <see cref="string.Empty"/> if
        /// there was none.</param>
        public void After(int exitCode, string standardOutput, string standardErrorOutput)
        {
            _RawStandardOutput = standardOutput;
            _RawStandardErrorOutput = standardErrorOutput;
            _RawExitCode = exitCode;

            ThrowOnUnsuccessfulExecution(exitCode, standardOutput, standardErrorOutput);
            ParseStandardOutputForResults(exitCode, standardOutput);
            Cleanup();
        }

        #endregion

        /// <summary>
        /// Adds the value to the <see cref="AdditionalArguments"/> collection property and
        /// returns this instance.
        /// </summary>
        /// <param name="value">
        /// The value to add to the <see cref="AdditionalArguments"/> collection property.
        /// </param>
        /// <returns>
        /// This instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="value"/> is <c>null</c> or empty.</para>
        /// </exception>
        public T WithAdditionalArgument(string value)
        {
            if (StringEx.IsNullOrWhiteSpace(value))
                throw new ArgumentNullException("value");
            AdditionalArguments.Add(value);
            return (T) this;
        }

        /// <summary>
        /// Sets the <see cref="Observer"/> property to the specified value and
        /// returns this instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Observer"/> property.
        /// </param>
        /// <returns>
        /// This instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public T WithObserver(IMercurialCommandObserver value)
        {
            Observer = value;
            return (T) this;
        }

        /// <summary>
        /// Sets the <see cref="Timeout"/> property to the specified value and
        /// returns this instance.
        /// </summary>
        /// <param name="value">
        /// The new value for the <see cref="Timeout"/> property.
        /// </param>
        /// <returns>
        /// This instance.
        /// </returns>
        /// <remarks>
        /// This method is part of the fluent interface.
        /// </remarks>
        public T WithTimeout(int value)
        {
            Timeout = value;
            return (T) this;
        }

        /// <summary>
        /// This method should throw the appropriate exception depending on the contents of
        /// the <paramref name="exitCode"/> and <paramref name="standardErrorOutput"/>
        /// parameters, or simply return if the execution is considered successful.
        /// </summary>
        /// <param name="exitCode">
        /// The exit code from executing the command line client.
        /// </param>
        /// <param name="standardOutput">
        /// The standard output from executing the command line client.
        /// </param>
        /// <param name="standardErrorOutput">
        /// The standard error output from executing the command client.
        /// </param>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all. The default behavior is to throw a <see cref="MercurialExecutionException"/>
        /// if <paramref name="exitCode"/> is not zero. If you require different behavior, don't call the base
        /// method.
        /// </remarks>
        /// <exception cref="MercurialExecutionException">
        /// <para><paramref name="exitCode"/> is not <c>0</c>.</para>
        /// </exception>
        protected virtual void ThrowOnUnsuccessfulExecution(int exitCode, string standardOutput, string standardErrorOutput)
        {
            if (exitCode == 0)
                return;

            throw new MercurialExecutionException(standardErrorOutput);
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
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all.
        /// </remarks>
        protected virtual void ParseStandardOutputForResults(int exitCode, string standardOutput)
        {
            // Do nothing by default
        }

        /// <summary>
        /// Override this method to implement code that will execute before command
        /// line execution.
        /// </summary>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all.
        /// </remarks>
        protected virtual void Prepare()
        {
            // Do nothing by default
        }

        /// <summary>
        /// Override this method to implement code that will execute after command
        /// line execution.
        /// </summary>
        /// <remarks>
        /// Note that as long as you descend from <see cref="CommandBase{T}"/> you're not required to call
        /// the base method at all.
        /// </remarks>
        protected virtual void Cleanup()
        {
            // Do nothing by default
        }

        /// <summary>
        /// Adds the specified argument to the <see cref="AdditionalArguments"/> collection,
        /// unless it is already present.
        /// </summary>
        /// <param name="argument">
        /// The argument to add to <see cref="AdditionalArguments"/>.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="argument"/> is <c>null</c> or empty.</para>
        /// </exception>
        [EditorBrowsable(EditorBrowsableState.Never)]
        public void AddArgument(string argument)
        {
            if (StringEx.IsNullOrWhiteSpace(argument))
                throw new ArgumentNullException("argument");

            if (!_AdditionalArguments.Contains(argument))
                _AdditionalArguments.Add(argument);
        }

        /// <summary>
        /// Adds a configuration override specification to the <see cref="AdditionalArguments"/>
        /// collection in the form of <c>section.name=value</c>.
        /// </summary>
        /// <param name="sectionName">
        /// The name of the section.
        /// </param>
        /// <param name="name">
        /// The name of the value.
        /// </param>
        /// <param name="value">
        /// The value.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="sectionName"/> is <c>null</c> or empty.</para>
        /// <para>- or -</para>
        /// <para><paramref name="name"/> is <c>null</c> or empty.</para>
        /// <para>- or -</para>
        /// <para><paramref name="value"/> is <c>null</c>.</para>
        /// </exception>
        public void WithConfigurationOverride(string sectionName, string name, string value)
        {
            if (StringEx.IsNullOrWhiteSpace(sectionName))
                throw new ArgumentNullException("sectionName");
            if (StringEx.IsNullOrWhiteSpace(name))
                throw new ArgumentNullException("name");
            if (value == null)
                throw new ArgumentNullException("value");

            AdditionalArguments.Add("--config");
            AdditionalArguments.Add(String.Format(CultureInfo.InvariantCulture, "{0}.{1}=\"{2}\"", sectionName, name, value));
        }
    }
}