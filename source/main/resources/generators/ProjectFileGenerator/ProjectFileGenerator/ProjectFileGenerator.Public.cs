using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;

namespace ProjectFileGenerator
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

        public ProjectFileGenerator(string name, string guid, EVersion version, ELanguage language, string[] platforms, string[] configs, ProjectFileTemplate template)
        {
            mProjectName = name;
            mProjectGuid = guid;
            mVersion = version;
            mLanguage = language;
            mPlatforms = platforms;
            mConfigs = configs;
            mTemplate = template;
        }

        public void Save(string filename)
        {
            _Save(filename);
        }

    }
}
