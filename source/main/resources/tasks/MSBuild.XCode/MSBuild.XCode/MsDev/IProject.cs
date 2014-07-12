using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode.MsDev
{
	public enum EProjectVersion
	{
		VS2010,
		VS2012,
		VS2013
	}

	public static class ProjectUtils
	{
		public static EProjectVersion FromString(string IDE)
		{
			if (string.Compare(IDE, "VS2010", true) == 0)
				return EProjectVersion.VS2010;
			if (string.Compare(IDE, "VS2012", true) == 0)
				return EProjectVersion.VS2012;
			if (string.Compare(IDE, "VS2013", true) == 0)
				return EProjectVersion.VS2013;

			//default
			return EProjectVersion.VS2012;
		}
	}

    public interface IProject
    {
		EProjectVersion Version { get; set; }
        string Extension { get; }

        XmlDocument Xml { get; set; }

        bool Construct(EProjectVersion version, IProject template);
        void MergeDependencyProject(IProject project);

        string[] GetPlatforms();
        string[] GetPlatformConfigs(string platform);

        void RemoveAllBut(Dictionary<string, StringItems> platformConfigs);
        void RemoveAllPlatformsBut(string platformToKeep);

        bool FilterItems(string[] to_remove, string[] to_keep);
        bool ExpandVars(PackageVars vars);
        bool ExpandGlobs(string rootdir, string reldir);

        bool Save(string filename);
    }
}
