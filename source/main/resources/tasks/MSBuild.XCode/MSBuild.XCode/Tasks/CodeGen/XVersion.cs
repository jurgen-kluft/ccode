using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    ///
    /// Generic implementation of a comparable version (see Version.docx in docs\manuals)
    /// 
    public class XVersion : IComparable<XVersion>
    {
        private string mValue;
        private string mCanonical;
        private ListItem mItems;

        public XVersion(string version)
        {
            ParseVersion(version);
        }

        public XVersion(XVersion version)
        {
            ParseVersion(version.ToString());
        }

        public bool IsNull
        {
            get
            {
                return String.IsNullOrEmpty(mValue);
            }
        }

        public enum EItem : int
        {
            Integer = 0,
            String = 1,
            List = 2,
        }

        private interface Item
        {
            int CompareTo(Item item);
            EItem GetItemType();
            bool IsNull();
        }

        /// <summary>
        ///  Represents a numeric item in the version item list.
        /// </summary>
        private class IntegerItem : Item
        {
            private int mValue;

            public IntegerItem(int i)
            {
                mValue = i;
            }

            public int Value
            {
                get
                {
                    return mValue;
                }
            }

            public EItem GetItemType()
            {
                return EItem.Integer;
            }

            public bool IsNull()
            {
                return (mValue == 0);
            }

            public int CompareTo(Item item)
            {
                if (item == null)
                {
                    return mValue == 0 ? 0 : 1; // 1.0 == 1, 1.1 > 1
                }

                switch (item.GetItemType())
                {
                    case EItem.Integer:
                        return mValue.CompareTo((item as IntegerItem).Value);

                    case EItem.String:
                        return 1; // 1.1 > 1-sp

                    case EItem.List:
                        return 1; // 1.1 > 1-1

                    default:
                        throw new SystemException("Invalid item: " + item.GetType());
                }
            }

            public override String ToString()
            {
                return mValue.ToString();
            }
        }

        /// <summary>
        /// Represents a string in the version item list, usually a qualifier.
        /// </summary>
        private class StringItem : Item
        {
            private readonly static String[] QUALIFIERS = { "snapshot", "alpha", "beta", "milestone", "rc", "", "sp" };
            private readonly static List<String> _QUALIFIERS = new List<String>() { "snapshot", "alpha", "beta", "milestone", "rc", "", "sp" };

            private readonly static Dictionary<string, string> ALIASES = new Dictionary<string, string>()
            {
                { "ga", "" },
                { "const", "" },
                { "cr", "rc" }
            };

            /// A comparable for the empty-string qualifier. This one is used to determine if a given qualifier makes the
            /// version older than one without a qualifier, or more recent.
            private static string RELEASE_VERSION_INDEX = _QUALIFIERS.IndexOf("").ToString();

            private string mValue;

            public StringItem(String value, bool followedByDigit)
            {
                if (followedByDigit && value.Length == 1)
                {
                    // a1 = alpha-1, b1 = beta-1, m1 = milestone-1
                    switch (value[0])
                    {
                        case 'a':
                            value = "alpha";
                            break;
                        case 'b':
                            value = "beta";
                            break;
                        case 'm':
                            value = "milestone";
                            break;
                    }
                }
                string v;
                if (ALIASES.TryGetValue(value, out v))
                    mValue = v;
                else
                    mValue = value;
            }

            public String Value
            {
                get
                {
                    return mValue;
                }
            }

            public EItem GetItemType()
            {
                return EItem.String;
            }

            public bool IsNull()
            {
                return (ComparableQualifier(mValue).CompareTo(RELEASE_VERSION_INDEX) == 0);
            }

            /// 
            /// Returns a comparable for a qualifier.
            /// 
            /// This method both takes into account the ordering of known qualifiers as well as lexical ordering for unknown
            /// qualifiers.
            /// 
            /// just returning an Integer with the index here is faster, but requires a lot of if/then/else to check for -1
            /// or QUALIFIERS.size and then resort to lexical ordering. Most comparisons are decided by the first character,
            /// so this is still fast. If more characters are needed then it requires a lexical sort anyway.
            /// 
            /// @param qualifier
            /// @return
            /// 
            public static string ComparableQualifier(String qualifier)
            {
                int i = _QUALIFIERS.IndexOf(qualifier);
                return i == -1 ? _QUALIFIERS.Count + "-" + qualifier : i.ToString();
            }

            public int CompareTo(Item item)
            {
                if (item == null)
                {
                    // 1-rc < 1, 1-ga > 1
                    return ComparableQualifier(mValue).CompareTo(RELEASE_VERSION_INDEX);
                }
                switch (item.GetItemType())
                {
                    case EItem.Integer:
                        return -1; // 1.any < 1.1 ?

                    case EItem.String:
                        return ComparableQualifier(mValue).CompareTo(ComparableQualifier(((StringItem)item).Value));

                    case EItem.List:
                        return -1; // 1.any < 1-1

                    default:
                        throw new SystemException("invalid item: " + item.GetType());
                }
            }

            public override String ToString()
            {
                return mValue;
            }
        }

        /// 
        /// Represents a version list item. This class is used both for the global item list and for sub-lists 
        /// (which start with '-(number)' in the version specification).
        /// 
        private class ListItem : Item
        {
            private List<Item> mItems = new List<Item>();

            public EItem GetItemType()
            {
                return EItem.List;
            }

            public int Count
            {
                get
                {
                    return mItems.Count;
                }
            }

            public void Add(Item item)
            {
                mItems.Add(item);
            }

            public Item this[int index]
            {
                get
                {
                    return mItems[index];
                }
            }

            public bool IsNull()
            {
                return (mItems.Count == 0);
            }

            internal void Normalize()
            {
                for (int i = mItems.Count - 1; i >= 0; )
                {
                    Item item = mItems[i];
                    if (item.IsNull())
                    {
                        mItems.RemoveAt(i); // remove null trailing items: 0, "", empty list
                        --i;
                    }
                    else
                    {
                        break;
                    }
                }
            }

            public int CompareTo(Item item)
            {
                if (item == null)
                {
                    if (IsNull())
                    {
                        return 0; // 1-0 = 1- (normalize) = 1
                    }
                    Item first = mItems[0];
                    return first.CompareTo(null);
                }
                switch (item.GetItemType())
                {
                    case EItem.Integer:
                        return -1; // 1-1 < 1.0.x

                    case EItem.String:
                        return 1; // 1-1 > 1-sp

                    case EItem.List:
                        {
                            int i = 0;

                            ListItem list = item as ListItem;
                            while (i < Count || i < list.Count)
                            {
                                Item l = i < Count ? this[i] : null;
                                Item r = i < list.Count ? list[i] : null;

                                // if this is shorter, then invert the compare and mul with -1
                                int result = (l == null) ? -1 * r.CompareTo(l) : l.CompareTo(r);
                                if (result != 0)
                                    return result;

                                ++i;
                            }

                            return 0;
                        }
                    default:
                        throw new SystemException("invalid item: " + item.GetType());
                }
            }

            public override String ToString()
            {
                int i = 0;
                StringBuilder buffer = new StringBuilder("(");
                foreach (Item item in mItems)
                {
                    if (i > 0)
                        buffer.Append(',');
                    buffer.Append(item.ToString());
                }
                buffer.Append(')');
                return buffer.ToString();
            }
        }

        public void ParseVersion(string version)
        {
            if (String.IsNullOrEmpty(version))
                mValue = string.Empty;
            else
                mValue = version;

            mItems = new ListItem();

            version = version.ToLower();

            ListItem list = mItems;

            Stack<Item> stack = new Stack<Item>();
            stack.Push(list);

            bool isDigit = false;
            int startIndex = 0;
            for (int i = 0; i < version.Length; i++)
            {
                char c = version[i];

                if (c == '.')
                {
                    if (i == startIndex)
                    {
                        list.Add(new IntegerItem(0));
                    }
                    else
                    {
                        list.Add(ParseItem(isDigit, version.Substring(startIndex, i - startIndex)));
                    }
                    startIndex = i + 1;
                }
                else if (c == '-')
                {
                    if (i == startIndex)
                    {
                        list.Add(new IntegerItem(0));
                    }
                    else
                    {
                        list.Add(ParseItem(isDigit, version.Substring(startIndex, i - startIndex)));
                    }
                    startIndex = i + 1;

                    if (isDigit)
                    {
                        list.Normalize(); // 1.0-* = 1-*

                        if ((i + 1 < version.Length) && Char.IsDigit(version[i + 1]))
                        {
                            // new ListItem only if previous were digits and new char is a digit,
                            // ie need to differentiate only 1.1 from 1-1
                            list.Add(list = new ListItem());

                            stack.Push(list);
                        }
                    }
                }
                else if (Char.IsDigit(c))
                {
                    if (!isDigit && i > startIndex)
                    {
                        list.Add(new StringItem(version.Substring(startIndex, i - startIndex), true));
                        startIndex = i;
                    }

                    isDigit = true;
                }
                else
                {
                    if (isDigit && i > startIndex)
                    {
                        list.Add(ParseItem(true, version.Substring(startIndex, i - startIndex)));
                        startIndex = i;
                    }

                    isDigit = false;
                }
            }

            if (version.Length > startIndex)
            {
                list.Add(ParseItem(isDigit, version.Substring(startIndex)));
            }

            while (stack.Count > 0)
            {
                list = stack.Pop() as ListItem;
                list.Normalize();
            }

            mCanonical = mItems.ToString();
        }

        private static Item ParseItem(bool isDigit, String buf)
        {
            if (isDigit)
                return new IntegerItem(Int32.Parse(buf));
            else
                return new StringItem(buf, false);
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

        public override String ToString()
        {
            return mValue;
        }

        private void ToStringsRecursive(ListItem list, List<string> strings)
        {
            for (int i = 0; i < list.Count; ++i)
            {
                Item item = list[i];
                ListItem listItem = item as ListItem;
                if (listItem != null)
                {
                    ToStringsRecursive(listItem, strings);
                }
                else
                {
                    strings.Add(item.ToString());
                }
            }
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