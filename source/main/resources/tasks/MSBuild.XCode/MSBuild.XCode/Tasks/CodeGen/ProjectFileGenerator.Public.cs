using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;

namespace MSBuild.Cod
{
    public partial class ProjectFileGenerator
    {
        public enum EVersion
        {
            VS2010,
        }
        public enum ELanguage
        {
            CS,
            CPP,
        }

        public ProjectFileGenerator(string name, string guid, EVersion version, ELanguage language, string[] platforms, string[] configs, XProject project)
        {
            mProjectName = name;
            mProjectGuid = guid;
            mVersion = version;
            mLanguage = language;
            mPlatforms = platforms;
            mConfigs = configs;
            mXProjectWriter = new XProjectWriter(project, platforms, configs);
        }

        public void Save(string filename)
        {
            _Save(filename);
        }

    }
}
