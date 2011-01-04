using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial.Extensions.CaseGuard
{
    /// <summary>
    /// This class adds extension methods to the <see cref="AddCommand"/> class, for
    /// the CaseGuard extension.
    /// </summary>
    public static class CaseGuardAddCommandExtensions
    {
        /// <summary>
        /// Add files regardless of possible case-collision problems.
        /// </summary>
        public static AddCommand WithOverrideCaseCollision(this AddCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");
            if (!CaseGuardExtension.IsInstalled)
                throw new InvalidOperationException("The caseguard extension is not installed and active");

            command.AddArgument("--override");
            return command;
        }

        /// <summary>
        /// Do not check filenames for Windows incompatibilities.
        /// </summary>
        public static AddCommand WithoutWindowsFileNameChecks(this AddCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");
            if (!CaseGuardExtension.IsInstalled)
                throw new InvalidOperationException("The caseguard extension is not installed and active");

            command.AddArgument("--nowincheck");
            return command;
        }

        /// <summary>
        /// Completely skip checks related to case-collision problems.
        /// </summary>
        /// <param name="command"></param>
        /// <returns></returns>
        public static AddCommand WithoutCaseGuarding(this AddCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");
            if (!CaseGuardExtension.IsInstalled)
                throw new InvalidOperationException("The caseguard extension is not installed and active");

            command.AddArgument("--unguard");
            return command;
        }
    }
}