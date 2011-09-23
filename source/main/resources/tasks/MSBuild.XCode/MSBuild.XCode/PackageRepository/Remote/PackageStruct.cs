using System;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Xml;
using System.IO;

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
        private int m_PackageVersion;   // Major Minor Build - 999 999 999

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
        
        public int Version
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
