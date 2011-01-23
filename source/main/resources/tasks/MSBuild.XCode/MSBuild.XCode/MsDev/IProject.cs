using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode.MsDev
{
    public interface IProject
    {
        string Extension { get; }

        XmlDocument Xml { get; set; }

        bool Construct(IProject template);
        void MergeDependencyProject(IProject project);

        string[] GetPlatforms();
        string[] GetPlatformConfigs(string platform);

        void RemoveAllBut(Dictionary<string, StringItems> platformConfigs);
        void RemoveAllPlatformsBut(string platformToKeep);

        bool FilterItems(string[] to_remove, string[] to_keep);
        bool ExpandVars(Dictionary<string, string> vars);
        bool ExpandGlobs(string rootdir, string reldir);

        bool Save(string filename);
    }
}
