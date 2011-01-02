using System;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyInstance
    {
        private string mPlatform;
        private string mType;
        private DependencyResource mResource;
        private VersionRange mVersionRange;
        private ComparableVersion mVersion;

        public DependencyInstance(string platform, DependencyResource resource)
        {
            mPlatform = platform;            
            mResource = resource;
            mVersionRange = mResource.GetVersionRange(Platform);
            mVersion = new ComparableVersion("1.0.0");
        }

        public string Name { get { return mResource.Name; } }
        public string Platform { get { return mPlatform; } }
        public string Branch { get { return mResource.GetBranch(Platform, "default"); } }
        public Group Group { get { return mResource.Group; } }
        public string Type { get { return mType; } }
        public DependencyResource Resource { get { return mResource; } }

        public VersionRange VersionRange { get { return mVersionRange; } }
        public ComparableVersion Version { get { return mVersion; } }

        public void ChangeResource(DependencyResource resource)
        {
            mResource = resource;
            mVersionRange = mResource.GetVersionRange(Platform);
        }

        public void Info()
        {
            Loggy.Add(String.Format("Dependency                 : {0}", Name));
            Loggy.Add(String.Format("Group                      : {0}", Group.ToString()));
            Loggy.Add(String.Format("Platform                   : {0}", Platform));
            Loggy.Add(String.Format("Branch                     : {0}", Branch));
            Loggy.Add(String.Format("Type                       : {0}", Type));
            Loggy.Add(String.Format("VersionRange               : {0}", VersionRange.ToString()));
            Loggy.Add(String.Format("Version                    : {0}", Version.ToString()));
        }

        public bool IsEqual(DependencyInstance dependency)
        {
            if (String.Compare(dependency.Name, Name, true) != 0)
                return false;
            if (String.Compare(dependency.Platform, Platform, true) != 0)
                return false;
            if (String.Compare(dependency.Branch, Branch, true) != 0)
                return false;
            if (Group != dependency.Group)
                return false;
            if (mVersionRange != dependency.mVersionRange)
                return false;
            if (mVersion != dependency.mVersion)
                return false;

            return true;
        }

        // Merge with same package dependency
        // Return True when merge resulted in an updated dependency (A change in VersionRange)
        public bool Merge(DependencyInstance dependency)
        {
            bool modified = false;
            if (String.Compare(Name, dependency.Name, true) != 0)
                return modified;

            // Merge the type
            if (String.Compare(Type, dependency.Type, true) != 0)
            {
                // Currently there are only 2 types, Package and Source
                if (String.Compare(Type, "Source", true) == 0)
                {
                    mType = "Package";
                    modified = true;
                }
            }

            // Merge the version range
            if (mVersionRange.Merge(dependency.mVersionRange))
                modified = true;

            return modified;
        }
   }
}