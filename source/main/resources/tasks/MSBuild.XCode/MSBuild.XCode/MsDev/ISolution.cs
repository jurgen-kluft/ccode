using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.MsDev
{
    public interface ISolution
    {
        string Extension { get; }

        void AddDependencies(string projectFile, string[] dependencyProjectFiles);

        int Save(string _SolutionFile, List<string> _ProjectFiles);
    }
}
