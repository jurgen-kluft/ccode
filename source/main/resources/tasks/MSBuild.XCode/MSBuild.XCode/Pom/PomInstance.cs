using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PomInstance
    {
        private PomResource mResource;

        public EType Type { get; set; }

        public string Name { get { return mResource.Name; } }
        public Group Group { get { return mResource.Group; } }

        public bool Main { get { return Type == EType.Main; } set { if (value) Type = EType.Main; else Type = EType.Dependency; } }

        public List<Attribute> DirectoryStructure { get { return mResource.DirectoryStructure; } }

        public Dictionary<string, List<KeyValuePair<string, string>>> Content { get { return mResource.Content; } }
        public List<DependencyResource> Dependencies { get { return mResource.Dependencies; } }
        public List<Project> Projects { get { return mResource.Projects; } }
        public List<string> Platforms { get { return mResource.Platforms; } }
        public Versions Versions { get { return mResource.Versions; } }

        public enum EType
        {
            Main,
            Dependency,
        }

        public PomInstance(PomResource resource, EType type)
        {
            mResource = resource;
            Type = type;
        }

        public bool Info()
        {

            return mResource.Info();
        }

        public Project GetProjectByGroup(string group)
        {
            foreach (Project p in Projects)
            {
                if (String.Compare(p.Group, group, true) == 0)
                    return p;
            }
            return null;
        }

        public Project GetProjectByName(string name)
        {
            foreach (Project p in Projects)
            {
                if (String.Compare(p.Name, name, true) == 0)
                    return p;
            }
            return null;
        }

        public string[] GetGroups()
        {
            List<string> categories = new List<string>();
            foreach (Project prj in Projects)
            {
                categories.Add(prj.Group);
            }
            return categories.ToArray();
        }

        public string[] GetPlatformsForGroup(string inGroup)
        {
            Project project = GetProjectByGroup(inGroup);
            if (project != null)
                return project.GetPlatforms();
            return new string[0];
        }

        public string[] GetConfigsForPlatformsForGroup(string Platform, string inGroup)
        {
            Project project = GetProjectByGroup(inGroup);
            if (project!=null)
                return project.GetConfigsForPlatform(Platform);
            return new string[0];
        }

        public void OnlyKeepPlatformSpecifics(string platform)
        {
            foreach (Project prj in Projects)
            {
                prj.OnlyKeepPlatformSpecifics(platform);
            }
        }


    }
}
