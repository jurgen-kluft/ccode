using System;
using System.Collections.Generic;
using System.Text;
using System.Diagnostics;
using Microsoft.Build.Utilities;

namespace msbuild.xmaven.helpers
{
    /// <summary>
    /// Helper class to run an executable
    /// </summary>
    public class ProcessHelper
    {
        /// <summary>
        /// Runs the given exe with the given arguments
        /// </summary>
        /// <param name="executingTask">MSBuild task executing the call</param>
        /// <param name="toolPath">Path to the exe to run</param>
        /// <param name="exeName">Name of the exe to run</param>
        /// <param name="arguments">Arguments to pass to the exe</param>
        public static void Run(Task executingTask, string toolPath, string exeName, params string[] arguments)
        {
            ProcessStartInfo startInfo = new ProcessStartInfo();
            startInfo.WorkingDirectory = toolPath;
            startInfo.FileName = exeName;
            startInfo.Arguments = string.Join(" ", arguments);
            executingTask.Log.LogMessage("Process path: {0} filename: {1}, arguments: {2}", startInfo.WorkingDirectory, startInfo.FileName, startInfo.Arguments);                       
            Process process = new Process();
            process.StartInfo = startInfo;
            process.Start();
            process.WaitForExit();
        }
    }
}
