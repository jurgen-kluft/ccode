using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace ProjectFileGenerator
{
    public partial class ProjectFileTemplate
    {
        public ProjectFileTemplate(string[] platforms, string[] configs)
        {
            mPlatforms = platforms;
            mConfigs = configs;
        }

        public void Load(string filename)
        {
            InternalLoad(filename);
        }

        public List<string> GetGroupElementsFor(string platform, string config, string group)
        {
            return InternalGetGroupElementsFor(platform, config, group);
        }

    }
}