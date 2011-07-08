using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This interface must be implemented by an object that will act as an observer
    /// of events related to Mercurial command execution.
    /// </summary>
    public interface IMercurialCommandObserver
    {
        /// <summary>
        /// This method will be called once for each line of normal output from the command. Note that this method will be called on
        /// a different thread than the thread that called the <see cref="Client.Execute(string,IMercurialCommand)"/> method.
        /// </summary>
        void Output(string line);

        /// <summary>
        /// This method will be called once for each line of error output from the command. Note that this method will be called on
        /// a different thread than the thread that called the <see cref="Client.Execute(string,IMercurialCommand)"/> method.
        /// </summary>
        void ErrorOutput(string line);

        /// <summary>
        /// This method will be called before the command starts executing.
        /// </summary>
        void Executing(string command, string arguments);

        /// <summary>
        /// This method will be called after the command has terminated (either timed out or completed by itself.)
        /// </summary>
        void Executed(string command, string arguments, int exitCode, string output, string errorOutput);
    }
}