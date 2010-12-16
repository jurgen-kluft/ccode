using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    ///
    /// Generic implementation of a comparable version (see Version.docx in docs\manuals)
    /// 
    public partial class XVersion : IComparable<XVersion>
    {
        private string mValue;

        public XVersion(string version)
        {
            FromString(version);
        }

        public XVersion(XVersion version)
        {
            FromString(version.ToString());
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
            return mItems.CompareTo((o as XVersion).mItems);
        }

        public bool LessThan(XVersion v, bool include)
        {
            if (include) return this <= v;
            else return this < v;
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
            List<string> strings = new List<string>();
            ToStringsRecursive(mItems, strings);
            while (strings.Count < n)
                strings.Add("0");
            return strings.ToArray();
        }

        public override bool Equals(Object o)
        {
            return (o is XVersion) && mCanonical.Equals((o as XVersion).mCanonical);
        }

        public override int GetHashCode()
        {
            return mCanonical.GetHashCode();
        }

        public static bool operator <(XVersion a, XVersion b)
        {
            return Compare(a, b) == -1;
        }
        public static bool operator <=(XVersion a, XVersion b)
        {
            return Compare(a, b) != 1;
        }
        public static bool operator >(XVersion a, XVersion b)
        {
            return Compare(a, b) == 1;
        }
        public static bool operator >=(XVersion a, XVersion b)
        {
            return Compare(a, b) != -1;
        }
        public static bool operator ==(XVersion a, XVersion b)
        {
            return Compare(a, b) == 0;
        }
        public static bool operator !=(XVersion a, XVersion b)
        {
            return Compare(a, b) != 0;
        }

        public static int Compare(XVersion a, XVersion b)
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

        public int CompareTo(XVersion b)
        {
            if ((object)b == null)
                return 1;
            return mItems.CompareTo(b.mItems);
        }

        #endregion
    }

}