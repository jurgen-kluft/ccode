using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This class implements the "hg init" command (http://www.selenic.com/mercurial/hg.1.html#init)
    /// </summary>
    public sealed class InitCommand : CommandBase<InitCommand>
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="InitCommand"/> class.
        /// </summary>
        public InitCommand()
            : base("init")
        {
            // Do nothing here
        }
    }
}