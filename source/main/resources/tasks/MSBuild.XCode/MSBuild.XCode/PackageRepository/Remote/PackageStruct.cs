using System;

namespace xpackage_repo
{
    public class Package_pk
    {
        private string m_PackageName;
        private string m_PackageGroup;
        private string m_PackageLanguage;

        public Package_pk()
        {
            m_PackageName = "unknown";
            m_PackageGroup = "com.virtuos.tnt";
            m_PackageLanguage = "C++";
        }

        public string Name
        {
            get { return m_PackageName; }
            set { m_PackageName = value; }
        }

        public string Group
        {
            get { return m_PackageGroup; }
            set { m_PackageGroup = value; }
        }

        public string Language
        {
            get { return m_PackageLanguage; }
            set { m_PackageLanguage = value; }
        }
    }

    public class PackageVersion_pv : Package_pk
    {
        private string m_PackagePlatform;
        private string m_PackageBranch;
        private Int64 m_PackageVersion;

        private string m_Datetime;
        private string m_Location;
        private string m_Changeset;

        public PackageVersion_pv()
        {
            m_PackagePlatform = "Win32";
            m_PackageBranch = "default";
            m_PackageVersion = 1000000;

            m_Datetime = DateTime.Now.ToString();
            m_Location = "";
            m_Changeset = "";
        }

        public static Int64 BuildVersion(int major, int minor, int build)
        {
            UInt64 version = 0;
            version = version | (((UInt64)major & 0x0007ffff) << 44);
            version = version | (((UInt64)minor & 0x000fffff) << 24);
            version = version | (((UInt64)build & 0x00ffffff));
            return (Int64)version;
        }

        public static Int64 LowestVersion
        {
            get
            {
                Int64 lowestVersion = PackageVersion_pv.BuildVersion(1, 0, 0);
                return lowestVersion;
            }
        }

        public static Int64 HighestVersion
        {
            get
            {
                Int64 highestVersion = PackageVersion_pv.BuildVersion((1 << 19) - 1, (1 << 20) - 1, (1 << 24) - 1);
                return highestVersion;
            }
        }

        public static void SplitVersion(Int64 version, out int major, out int minor, out int build)
        {
            major = (int)(((UInt64)version & 0x7ffff00000000000) >> 44);
            minor = (int)(((UInt64)version & 0x00000fffff000000) >> 24);
            build = (int)(((UInt64)version & 0x0000000000ffffff) >> 0);
        }

        public string Platform
        {
            get { return m_PackagePlatform; }
            set { m_PackagePlatform = value; }
        }

        public string Branch
        {
            get { return m_PackageBranch; }
            set { m_PackageBranch = value; }
        }
        
        public Int64 Version
        {
            get { return m_PackageVersion; }
            set { m_PackageVersion = value; }
        }

        public string Datetime
        {
            get { return m_Datetime; }
            set { m_Datetime = value; }
        }

        public string Location
        {
            get { return m_Location; }
            set { m_Location = value; }
        }

        public string Changeset
        {
            get { return m_Changeset; }
            set { m_Changeset = value; }
        }
    }
}
