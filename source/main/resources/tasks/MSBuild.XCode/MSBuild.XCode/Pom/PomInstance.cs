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
        private List<ProjectInstance> mProjects;

        public string Name { get { return mResource.Name; } }
        public Group Group { get { return mResource.Group; } }

        public List<Attribute> DirectoryStructure { get { return mResource.DirectoryStructure; } }

        public Dictionary<string, List<KeyValuePair<string, string>>> Content { get { return mResource.Content; } }
        public List<DependencyResource> Dependencies { get { return mResource.Dependencies; } }
        public List<ProjectInstance> Projects { get { return mProjects; } }
        public List<string> Platforms { get { return mResource.Platforms; } }
        public Versions Versions { get { return mResource.Versions; } }

        public PomInstance(bool main, PomResource resource)
        {
            mResource = resource;
            mProjects = new List<ProjectInstance>();
            foreach (ProjectResource projectResource in resource.Projects)
            {
                ProjectInstance projectInstance = projectResource.CreateInstance(main);
                mProjects.Add(projectInstance);
            }
        }

        public bool IsCpp
        {
            get
            {
                bool isCpp = true;
                foreach (ProjectInstance prj in Projects)
                    isCpp = isCpp && prj.IsCpp;
                return isCpp;
            }
        }

        public bool IsCs
        {
            get
            {
                bool isCs = true;
                foreach (ProjectInstance prj in Projects)
                    isCs = isCs && prj.IsCs;
                return isCs;
            }
        }

        public bool Info()
        {

            return mResource.Info();
        }

        public ProjectInstance GetProjectByName(string name)
        {
            foreach (ProjectInstance p in Projects)
            {
                if (String.Compare(p.Name, name, true) == 0)
                    return p;
            }
            return null;
        }

        public bool IsDependencyForPlatform(string DependencyName, string Platform)
        {
            foreach (DependencyResource dependencyResource in Dependencies)
            {
                if (String.Compare(dependencyResource.Name, DependencyName, true) == 0)
                {
                    return (dependencyResource.IsForPlatform(Platform));
                }
            }
            return false;
        }

        public void OnlyKeepPlatformSpecifics(string platform)
        {
            foreach (ProjectInstance prj in Projects)
                prj.OnlyKeepPlatformSpecifics(platform);
        }
    }
}
