using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This class implements <see cref="IMercurialCommandObserver"/>
    /// by simply writing everything to debug output through
    /// <see cref="Debug.WriteLine(string)"/>.
    /// </summary>
    public class DebugObserver : IMercurialCommandObserver
    {
        #region IMercurialCommandObserver Members

        /// <summary>
        /// This method will be called once for each line of normal output from the command. Note that this method will be called on
        /// a different thread than the thread that called the <see cref="Client.Execute(string,IMercurialCommand)"/> method.
        /// </summary>
        public void Output(string line)
        {
            Debug.WriteLine(line);
        }

        /// <summary>
        /// This method will be called once for each line of error output from the command. Note that this method will be called on
        /// a different thread than the thread that called the <see cref="Client.Execute(string,IMercurialCommand)"/> method.
        /// </summary>
        public void ErrorOutput(string line)
        {
            Debug.WriteLine("! " + line);
        }

        /// <summary>
        /// This method will be called before the command starts executing.
        /// </summary>
        public void Executing(string command, string arguments)
        {
            Debug.WriteLine("executing: " + command + " " + arguments);
        }

        /// <summary>
        /// This method will be called after the command has terminated (either timed out or completed by itself.)
        /// </summary>
        public void Executed(string command, string arguments, int exitCode, string output, string errorOutput)
        {
            Debug.WriteLine("executed: " + command + " " + arguments + " --> " + exitCode);
        }

        #endregion
    }
}