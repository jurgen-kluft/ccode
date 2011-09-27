using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading;
using Mercurial.Configuration;

namespace Mercurial
{
    /// <summary>
    /// This class encapsulates the Mercurial client application.
    /// </summary>
    public static class Client
    {
        private static readonly string _ClientPath;
        private static readonly ClientConfigurationCollection _ConfigurationCollection;

        static Client()
        {
            _ClientPath = LocateClient();
            if (!StringEx.IsNullOrWhiteSpace(_ClientPath))
                _ConfigurationCollection = new ClientConfigurationCollection();
        }

        /// <summary>
        /// Gets the path to the Mercurial client executable, on first read of this property
        /// the client will be located on the system.
        /// </summary>
        public static string ClientPath
        {
            get
            {
                return _ClientPath;
            }
        }

        /// <summary>
        /// Gets whether the Mercurial client executable could be located or not.
        /// </summary>
        public static bool CouldLocateClient
        {
            get
            {
                return ClientPath.Length > 0;
            }
        }

        /// <summary>
        /// Gets the current client configuration.
        /// </summary>
        public static ClientConfigurationCollection Configuration
        {
            get
            {
                if (!CouldLocateClient)
                    throw new InvalidOperationException(
                        "The Mercurial client configuration is not available because the client executable could not be located");
                return _ConfigurationCollection;
            }
        }

        /// <summary>
        /// Executes the given <see cref="IMercurialCommand"/> command without
        /// a repository.
        /// </summary>
        /// <param name="command">
        /// The <see cref="IMercurialCommand"/> command to execute.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public static void Execute(IMercurialCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Execute(Path.GetTempPath(), command);
        }

        /// <summary>
        /// Executes the given <see cref="IMercurialCommand"/> command against
        /// the Mercurial repository.
        /// </summary>
        /// <param name="repositoryPath">
        /// The root path of the repository to execute the command in.
        /// </param>
        /// <param name="command">
        /// The <see cref="IMercurialCommand"/> command to execute.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="repositoryPath"/> is <c>null</c>.</para>
        /// <para>- or -</para>
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public static void Execute(string repositoryPath, IMercurialCommand command)
        {
            if (StringEx.IsNullOrWhiteSpace(repositoryPath))
                throw new ArgumentNullException("repositoryPath");
            if (command == null)
                throw new ArgumentNullException("command");

            command.Validate();
            command.Before();

            string argumentsString = String.Join(" ",
                new[] { "--noninteractive", "--encoding", "UTF-8", }.Concat(command.Arguments.Where(a => !StringEx.IsNullOrWhiteSpace(a))).Concat(
                    command.AdditionalArguments.Where(a => !StringEx.IsNullOrWhiteSpace(a))).ToArray());

            var psi = new ProcessStartInfo
                {
                    FileName = ClientPath,
                    WorkingDirectory = repositoryPath,
                    RedirectStandardInput = true,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    CreateNoWindow = true,
                    WindowStyle = ProcessWindowStyle.Hidden,
                    UseShellExecute = false,
                    ErrorDialog = false,
                    Arguments = command.Command + " " + argumentsString,
                };
            psi.EnvironmentVariables["LANGUAGE"] = "EN";
            psi.EnvironmentVariables["HGENCODING"] = "UTF-8";
            psi.StandardErrorEncoding = Encoding.UTF8;
            psi.StandardOutputEncoding = Encoding.UTF8;

            if (command.Observer != null)
                command.Observer.Executing(command.Command, argumentsString);

            Process process = Process.Start(psi);
            try
            {
                Thread outputThread;
                Thread errorThread;
                string standardOutput = string.Empty;
                string errorOutput = string.Empty;

                if (command.Observer != null)
                {
                    outputThread = new Thread(() =>
                    {
                        var output = new StringBuilder();
                        string line;
                        while ((line = process.StandardOutput.ReadLine()) != null)
                        {
                            command.Observer.Output(line);
                            if (output.Length > 0)
                                output.Append(Environment.NewLine);
                            output.Append(line);
                        }
                        standardOutput = output.ToString();
                    });
                    errorThread = new Thread(() =>
                    {
                        var output = new StringBuilder();
                        string line;
                        while ((line = process.StandardError.ReadLine()) != null)
                        {
                            command.Observer.ErrorOutput(line);
                            if (output.Length > 0)
                                output.Append(Environment.NewLine);
                            output.Append(line);
                        }
                        errorOutput = output.ToString();
                    });
                }
                else
                {
                    outputThread = new Thread(() =>
                    {
                        standardOutput = process.StandardOutput.ReadToEnd();
                    });
                    errorThread = new Thread(() =>
                    {
                        errorOutput = process.StandardError.ReadToEnd();
                    });
                }

                outputThread.Name = "Mercurial.Net Standard Output Thread";
                errorThread.Name = "Mercurial.Net Standard Error Thread";

                outputThread.Start();
                errorThread.Start();

                if (!process.WaitForExit(1000*command.Timeout))
                {
                    if (command.Observer != null)
                        command.Observer.Executed(psi.FileName, psi.Arguments, 0, String.Empty, String.Empty);
                    throw new MercurialException("HG did not complete within the allotted time");
                }

                outputThread.Join();
                errorThread.Join();

                if (command.Observer != null)
                    command.Observer.Executed(command.Command, argumentsString, process.ExitCode, standardOutput, errorOutput);

                command.After(process.ExitCode, standardOutput, errorOutput);
            }
            finally
            {
                process.Dispose();
            }
        }

        private static string LocateClient()
        {
            string[] ppaths = Environment.GetEnvironmentVariable("PATH").Split(';');
            foreach (string path in ppaths)
            {
                try
                {
                    string hgpath = Path.Combine(path.Trim(), "hg.exe");
                    if (File.Exists(hgpath))
                    {
                        return hgpath;
                    }
                }
                catch (System.Exception e)
                {
                }
            }
            return string.Empty;
        }
    }
}