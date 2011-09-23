using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    ///
    /// Generic implementation of a comparable version (see Version.docx in docs\manuals)
    /// 
    public partial class ComparableVersion : IComparable<ComparableVersion>
    {
        private string mValue;

        public ComparableVersion(string version)
        {
            FromString(version);
        }

        public ComparableVersion(ComparableVersion version)
        {
            FromString(version.ToString());
        }

        public ComparableVersion(int v) /// MajorMinorBuild
        {
            FromString(String.Format("{0}.{1}.{2}", v / (1000 * 1000), ((v / 1000) % 1000), (v % 1000)));
        }

        public ComparableVersion(int major, int minor, int build)
        {
            FromString(String.Format("{0}.{1}.{2}", major, minor, build));
        }

        public bool IsNull
        {
            get
            {
                return String.IsNullOrEmpty(mValue);
            }
        }

        public void FromStrings(string[] versionItems)
        {
            string version = string.Empty;
            if (versionItems.Length > 0)
            {
                version = versionItems[0];
                for (int i = 1; i < versionItems.Length; ++i)
                    version += "." + versionItems[i];
            }
            Parse(version);
        }

        public void FromString(string version)
        {
            Parse(version);
        }
        
        public int CompareTo(Object o)
        {
            return mItems.CompareTo((o as ComparableVersion).mItems);
        }

        public bool LessThan(ComparableVersion v, bool include)
        {
            if (include) return this <= v;
            else return this < v;
        }

        public int GetMajor()
        {
            string[] items = mValue.Split('.');
            int major = 1;
            if (items.Length > 0)
                major = Int32.Parse(items[0]);
            return major;
        }

        public int GetMinor()
        {
            string[] items = mValue.Split('.');
            int minor = 0;
            if (items.Length > 1)
                minor = Int32.Parse(items[1]);
            return (minor);
        }

        public int GetBuild()
        {
            string[] items = mValue.Split('.');
            int build = 0;
            if (items.Length > 2)
                build = Int32.Parse(items[2]);
            return build;
        }

        public int ToInt()
        {
            string[] items = mValue.Split('.');
            int major = 1;
            if (items.Length > 0)
                major = Int32.Parse(items[0]);
            int minor = 0;
            if (items.Length > 1)
                minor = Int32.Parse(items[1]);
            int build = 0;
            if (items.Length > 2)
                build = Int32.Parse(items[2]);
            return (major * 1000000) + (minor * 1000) + build;
        }

        public int ToInt(int major, int minor, int build)
        {
            return (major * 1000000) + (minor * 1000) + build;
        }

        public override string ToString()
        {
            return mValue;
        }
        
        public string ToString(int n)
        {
            if (mItems.Count == 0)
                return string.Empty;
            List<string> strings = new List<string>();
            ToStringsRecursive(mItems, strings);
            if (strings.Count == 0)
                return string.Empty;
            string version = strings[0];
            for (int i = 1; i < strings.Count; ++i)
                version += "." + strings[i];
            return version;
        }

        public string[] ToStrings()
        {
            List<string> strings = new List<string>();
            ToStringsRecursive(mItems, strings);
            return strings.ToArray();
        }

        public string[] ToStrings(int n)
        {
            if (n == 0)
                return new string[0];

            List<string> strings = new List<string>();
            ToStringsRecursive(mItems, strings);

            int i = 0;
            string[] outStr = new string[n];
            foreach (string str in strings)
            {
                if (i == n) break;
                outStr[i++] = str;
            }
            while (i < n)
                outStr[i++] = "0";
            return outStr;
        }

        public override bool Equals(Object o)
        {
            return (o is ComparableVersion) && mCanonical.Equals((o as ComparableVersion).mCanonical);
        }

        public override int GetHashCode()
        {
            return mCanonical.GetHashCode();
        }

        public static bool operator <(ComparableVersion a, ComparableVersion b)
        {
            return Compare(a, b) == -1;
        }
        public static bool operator <=(ComparableVersion a, ComparableVersion b)
        {
            return Compare(a, b) != 1;
        }
        public static bool operator >(ComparableVersion a, ComparableVersion b)
        {
            return Compare(a, b) == 1;
        }
        public static bool operator >=(ComparableVersion a, ComparableVersion b)
        {
            return Compare(a, b) != -1;
        }
        public static bool operator ==(ComparableVersion a, ComparableVersion b)
        {
            return Compare(a, b) == 0;
        }
        public static bool operator !=(ComparableVersion a, ComparableVersion b)
        {
            return Compare(a, b) != 0;
        }

        public static int Compare(ComparableVersion a, ComparableVersion b)
        {
            bool a_null = (object)a == null;
            bool b_null = (object)b == null;
            if (a_null && b_null)
                return 0;
            if (a_null != b_null)
                return a_null ? -1 : 1;
            return a.mItems.CompareTo(b.mItems);
        }

        #region IComparable<XVersion> Members

        public int CompareTo(ComparableVersion b)
        {
            if ((object)b == null)
                return 1;
            return mItems.CompareTo(b.mItems);
        }

        #endregion
    }

}