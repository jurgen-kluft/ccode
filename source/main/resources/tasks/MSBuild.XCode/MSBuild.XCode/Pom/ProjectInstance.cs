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
            
            mMsDevProject = new CppProject();
            mMsDevProject.Merge(mResource.MsDevProject, true, true);

            // Filter items, this should be moved to ProjectInstance!
            if (Main)
                mMsDevProject.FilterItems(new string[] { "#" }, new string[] { "@" });
            else
                mMsDevProject.FilterItems(new string[] { "@" }, new string[] { "#" });
        }

        public void Info()
        {
            Loggy.Add(String.Format("Project                    : {0}", Name));
            Loggy.Add(String.Format("Category                   : {0}", Group));
            Loggy.Add(String.Format("Language                   : {0}", Language));
            Loggy.Add(String.Format("Location                   : {0}", Location));
        }

        public void ConstructFullMsDevProject()
        {
            if (IsCpp)
            {
                CppProject finalProject = new CppProject();
                finalProject.Merge(Global.CppTemplateProject, true, false);
                finalProject.Merge(mMsDevProject, true, true);
                finalProject.RemoveAllBut(Configs);
                mMsDevProject = finalProject;
            }
            else if (IsCs)
            {
                //CsProject finalProject = new CsProject();
                //finalProject.Merge(Global.CsTemplateProject, true, false);
                //finalProject.Merge(mMsDevProject, true, true);
                //mMsDevProject = finalProject;
            }
        }

        public void MergeWithDependencyProject(ProjectInstance dependencyProject)
        {
            mMsDevProject.Merge(dependencyProject.mMsDevProject, false, false);
        }

        public void OnlyKeepPlatformSpecifics(string platform)
        {
            if (IsCpp)
            {
                CppProject finalProject = new CppProject();
                finalProject.Merge(mMsDevProject, true, true);
                finalProject.RemoveAllPlatformsBut(platform);
                mMsDevProject = finalProject;
            }
            else if (IsCs)
            {
                CppProject finalProject = new CppProject();
                finalProject.Merge(mMsDevProject, true, true);
                finalProject.RemoveAllPlatformsBut(platform);
                mMsDevProject = finalProject;
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