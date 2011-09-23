using System;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class DependencyInstance
    {
        private string mPlatform;
        private DependencyResource mResource;
        private VersionRange mVersionRange;

        public DependencyInstance(string platform, DependencyResource resource)
        {
            mPlatform = platform;            
            mResource = resource;
            mVersionRange = mResource.GetVersionRange(Platform);
        }

        public string Name { get { return mResource.Name; } }
        public string Platform { get { return mPlatform; } }
        public string Branch { get { return mResource.GetBranch(Platform, "default"); } }
        public Group Group { get { return mResource.Group; } }
        public string Type { get { return mResource.Type; } }
        public DependencyResource Resource { get { return mResource; } }

        public VersionRange VersionRange { get { return mVersionRange; } }

        public void Info()
        {
            Loggy.Info(String.Format("Dependency                 : {0}", Name));
            Loggy.Info(String.Format("Group                      : {0}", Group.ToString()));
            Loggy.Info(String.Format("Platform                   : {0}", Platform));
            Loggy.Info(String.Format("Branch                     : {0}", Branch));
            Loggy.Info(String.Format("Type                       : {0}", Type));
            Loggy.Info(String.Format("VersionRange               : {0}", VersionRange.ToString()));
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

            return true;
        }

   }
}