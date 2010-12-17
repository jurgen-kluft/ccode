using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;

namespace MSBuild.XCode
{
    public partial class MsDevProjectFileGenerator
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

        public MsDevProjectFileGenerator(string name, string guid, EVersion version, ELanguage language, Project project)
        {
            mProjectName = name;
            mProjectGuid = guid;
            mVersion = version;
            mLanguage = language;

            List<string> platforms = new List<string>();
            foreach (KeyValuePair<string, Platform> p in project.Platforms)
                platforms.Add(p.Key);
            mPlatforms = platforms.ToArray();
            List<string> configs = new List<string>();
            foreach (KeyValuePair<string, Config> c in project.configs)
                configs.Add(c.Key);
            mConfigs = configs.ToArray();

            mXProjectWriter = new XProjectWriter(project, mPlatforms, mConfigs);
        }

        public void Save(string filename)
        {
            _Save(filename);
        }

    }
}
