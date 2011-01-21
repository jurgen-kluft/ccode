using System;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class ProjectInstance
    {
        private ProjectResource mResource;
        private CppProject mMsDevProject;
        private bool mIsFinalProject;

        public Dictionary<string, StringItems> Configs { get { return mResource.Configs; } }

        public string Name { get { return mResource.Name; } }
        public bool Main { get; set; }
        public string Scope { get { return mResource.Scope; } }
        public string Group { get { return mResource.Group; } }
        public string Language { get { return mResource.Language; } }
        public string Location{ get { return mResource.Location; } }
        public string DependsOn { get { return mResource.DependsOn; } }

        public bool IsCpp { get { return mResource.IsCpp; } }
        public bool IsCs { get { return mResource.IsCs; } }

        public bool IsPrivate { get { return mResource.IsPrivate; } }
        public string Extension { get { if (IsCs) return ".csproj"; return ".vcxproj"; } }

        public ProjectInstance(bool main, ProjectResource resource)
        {
            mResource = resource;
            Main = main;

            mIsFinalProject = false;
            mMsDevProject = new CppProject();
            mMsDevProject.Copy(mResource.MsDevProject);
            if (Main)
                mMsDevProject.FilterItems(new string[] { "#" }, new string[] { "@" });
            else
                mMsDevProject.FilterItems(new string[] { "@" }, new string[] { "#" });
        }

        public void Info()
        {
            Loggy.Info(String.Format("Project                    : {0}", Name));
            Loggy.Info(String.Format("Category                   : {0}", Group));
            Loggy.Info(String.Format("Language                   : {0}", Language));
            Loggy.Info(String.Format("Location                   : {0}", Location));
        }

        private static bool ContainsPlatform(List<string> platforms, string platform)
        {
            foreach (string p in platforms)
                if (String.Compare(p, platform, true) == 0)
                    return true;
            return false;
        }

        public void ConstructFullMsDevProject(List<string> platforms)
        {
            if (mIsFinalProject)
                return;

            if (IsCpp)
            {
                CppProject finalProject = new CppProject();
                finalProject.Copy(PackageInstance.CppTemplateProject);
                finalProject.Merge(mMsDevProject, true, true, true);

                Dictionary<string, StringItems> platform_configs = new Dictionary<string, StringItems>();
                foreach (KeyValuePair<string, StringItems> pair in Configs)
                {
                    if (ContainsPlatform(platforms, pair.Key))
                        platform_configs.Add(pair.Key, pair.Value);
                }

                finalProject.RemoveAllBut(platform_configs);
                mMsDevProject = finalProject;
                mIsFinalProject = true;
            }
            else if (IsCs)
            {
                //CsProject finalProject = new CsProject();
                //finalProject.Copy(Global.CsTemplateProject);
                //finalProject.Merge(mMsDevProject, true, true, true);
                //finalProject.RemoveAllBut(Configs);
                //mMsDevProject = finalProject;
            }
        }

        public void MergeWithDependencyProject(ProjectInstance dependencyProject)
        {
            mMsDevProject.Merge(dependencyProject.mMsDevProject, false, false, false);
        }

        public void OnlyKeepPlatformSpecifics(string platform)
        {
            if (IsCpp)
            {
                mMsDevProject.RemoveAllPlatformsBut(platform);
            }
            else if (IsCs)
            {
                mMsDevProject.RemoveAllPlatformsBut(platform);
            }
        }

        public string[] GetPlatforms()
        {
            return mResource.GetPlatforms();
        }

        public string[] GetConfigsForPlatform(string platform)
        {
            return mResource.GetConfigsForPlatform(platform);
        }

        public bool HasPlatformWithConfig(string platform, string config)
        {
            return mResource.HasPlatformWithConfig(platform, config);
        }

        public void Save(string rootdir, string filename)
        {
            string reldir = rootdir + Path.GetDirectoryName(filename);
            reldir = reldir.EndWith('\\');

            mMsDevProject.ExpandGlobs(rootdir, reldir);
            mMsDevProject.Save(rootdir + filename);
        }
    }
}