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
        public ProjectFileTemplate(string[] all_platforms, string[] all_configs, string[] project_platforms, string[] project_configs)
        {
            mPlatforms = all_platforms;
            mConfigs = all_configs;
            mProjectPlatforms = project_platforms;
            mProjectConfigs = project_configs;
        }

        public void Load(string template_filename, string project_filename)
        {
            InternalLoad(template_filename, project_filename);
        }

        public List<string> GetGroupElementsFor(string platform, string config, string group)
        {
            return InternalGetGroupElementsFor(platform, config, group);
        }

    }
}