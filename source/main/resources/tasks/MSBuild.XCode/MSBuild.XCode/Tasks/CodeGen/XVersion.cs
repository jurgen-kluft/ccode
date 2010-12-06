using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    /// <!--
    /// Features:
    /// 
    ///     # Mixing of '-' (dash) and '.' (dot) separators
    ///     # Transition between characters and digits also constitutes a separator:
    ///          o 1.0alpha1 => [1, 0, alpha, 1]; This fixes '1.0alpha10 < 1.0alpha2'
    ///     # Unlimited number of version components
    ///     * Version components in the text can be digits or strings
    ///     * Strings are checked for well-known qualifiers and the qualifier ordering is used for version ordering
    ///           o well-known qualifiers (case insensitive)
    ///                 + snapshot (NOTE; snapshot needs discussion)
    ///                 + alpha or a
    ///                 + beta or b
    ///                 + milestone or m
    ///                 + rc or cr
    ///                 + (the empty string) or ga or final
    ///                 + sp
    ///     * Version components prefixed with '-' will result in a sub-list of version components.
    ///       A dash usually precedes a qualifier, and is always less important than something preceded with a dot.
    ///       We need to somehow record the separators themselves, which is done by sublists.
    ///       Parse examples:
    ///           o 1.0-alpha1 => [1, 0, ["alpha", 1]]
    ///           o 1.0-rc-2 => [1, 0, ["rc", [2]]]
    /// 
    /// Parsing versions
    /// 
    /// The version string is examined one character at a time.
    /// There's a buffer containing the current text - all characters are appended, except for '.' and '-'.
    /// Below, when it's stated 'append buffer to list', the buffer is first converted to an Integer item 
    /// if that's possible, otherwise left alone as a String. 
    /// It will only be appended if it's length is not 0.
    /// 
    ///     * If a '.' is encountered, the current buffer is appended to the current list, either as a IntegerItem (if it's a number) or a StringItem.
    ///     * If a '-' is encountered, do the same as when a '.' is encountered, then create a new sublist, append it to the current list and replace 
    ///       the current list with the new sub-list.
    ///     * If the last character was a digit:
    ///           o and the current one is too, append it to the buffer.
    ///           o otherwise append the current buffer to the list, reset the buffer with the current char as content
    ///     * If the last character was NOT a digit:
    ///           o if the last character was also NOT a digit, append it to the buffer
    ///           o if it is a digit, append buffer to list, set buffers content to the digit
    ///     * Finally, append the buffer to the list
    /// 
    /// Some examples:
    /// 
    ///     * 1.0 => [1, 0]
    ///     * 1.0.1 => [1, 0, 1]
    ///     * 1-SNAPSHOT => [1, ["SNAPSHOT"]]
    ///     * 1-alpha10-SNAPSHOT => [1, ["alpha", "10", ["SNAPSHOT"]]]
    /// 
    /// Ordering algorithm
    /// 
    /// Internally 3 version component types are used:
    /// 
    ///     * integer (IntegerItem)
    ///     * string (StringItem) (knows if it's a qualifier or not)
    ///     * sublist (ListItem)
    /// 
    /// Elements from both versions are compared one at a time; first the first element of both, then the second, etc.
    /// 
    /// (Note: 'item' and 'component' are used interchangeably)
    /// 
    /// Table: Ordering rules when comparing version components
    ///   	          | Integer 	              | String 	                      | List 	                                      | null
    /// --------------+---------------------------+-------------------------------+-----------------------------------------------+-----------------------------------------
    /// Integer 	    Highest is newer 	        Integer is newer 	            Integer is newer 	                            If integer==0 then equal,
    ///                                                                                                                             otherwise integer is newer
    ///
    /// String 	        Integer is newer 	        order by well-known             List is newer 	                                Compare with ""
    ///                                             qualifiers and lexically
    ///                                             (see below) 	
    ///
    /// List 	        Integer is newer 	        List is newer  	                Version itself is a list;  	                    Compare with empty list item (recursion)
    ///                                                                             compare item by item                           this will finally result in String==?null or
    ///                                                                                                                             Integer==?null
    ///
    /// null 	        If integer==0 then equal,   Compare with ""                 Compare with empty list item (recursion)
    ///                 otherwise integer is newer 	                                this will finally result in String==?null or  	
    ///                                                                             Integer==?null 	
    ///                                                                              
    /// 
    /// Special note on string comparing:
    /// A predefined list of well-known qualifiers is present. For comparison, the string is converted to another string, as follows:
    /// 
    ///     # First, the well-known qualifier list is consulted for presence of the string
    ///     # If the string is present, the index in the list is returned, as a string
    ///     # If the string is not present, then qualifiers.Count + "-" + string is returned.
    /// 
    /// Then the strings are lexically compared.
    /// 
    /// Examples:
    /// 
    ///     # "alpha" yields "1"
    ///     # "" yields "4"
    ///     # "abc" yields "7-abc"
    ///     # "xyz" yields "7-xyz"
    /// 
    /// String Compare examples:
    /// 
    ///     # 1.0 ==? 1.0-alpha: "" (or null) ==? "alpha" -> "4" ==? "1" -> 1.0 is newer
    ///     # 1 ==? 1.0: equal
    ///     # 1-beta ==? 1-xyz: "2" ==? "7-xyz" -> 1-xyz is newer
    /// 
    /// Some comparisons that yield different results from the current implementation:
    /// 
    ///     # 1-beta ==? 1-abc: "2" ==? "7-abc" -> 1-abc is newer
    ///     # 1.0 ==? 1.0-abc: "4" ==? "7-abc" -> 1.0-abc is newer
    ///     # 1.0-alpha-10 ==? 1.0-alpha-2: 10 > 2, so '1.0-alpha-10' is newer
    ///     # 1.0-alpha-1.0 ==? 1.0-alpha-1: equal
    ///     # 1.0-alpha-1.2 ==? 1.0-alpha-2: 1.0-alpha-2 is newer
    /// -->
    ///
    /// Generic implementation of version comparison.
    ///
    public class XVersion : IComparable
    {
        private String mValue;
        private String mCanonical;
        private ListItem mItems;

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

        /**
         * Represents a numeric item in the version item list.
         */
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

        /**
         * Represents a string in the version item list, usually a qualifier.
         */
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

            /**
             * A comparable for the empty-string qualifier. This one is used to determine if a given qualifier makes the
             * version older than one without a qualifier, or more recent.
             */
            private static string RELEASE_VERSION_INDEX = _QUALIFIERS.IndexOf("").ToString();

            private String mValue;

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

            /**
             * Returns a comparable for a qualifier.
             *
             * This method both takes into account the ordering of known qualifiers as well as lexical ordering for unknown
             * qualifiers.
             *
             * just returning an Integer with the index here is faster, but requires a lot of if/then/else to check for -1
             * or QUALIFIERS.size and then resort to lexical ordering. Most comparisons are decided by the first character,
             * so this is still fast. If more characters are needed then it requires a lexical sort anyway.
             *
             * @param qualifier
             * @return
             */
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

        /**
         * Represents a version list item. This class is used both for the global item list and for sub-lists (which start
         * with '-(number)' in the version specification).
         */
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
                            int j = 0;

                            ListItem list = item as ListItem;

                            while (i < Count && j < list.Count)
                            {
                                Item l = this[i];
                                Item r = list[j];

                                // if this is shorter, then invert the compare and mul with -1
                                int result = l == null ? -1 * r.CompareTo(l) : l.CompareTo(r);
                                if (result != 0)
                                {
                                    return result;
                                }
                                ++i;
                                ++j;
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

        public XVersion(string version)
        {
            ParseVersion(version);
        }

        public void ParseVersion(String version)
        {
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

        public override String ToString()
        {
            return mValue;
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
            return a.CompareTo(b) == -1;
        }
        public static bool operator <=(XVersion a, XVersion b)
        {
            return a.CompareTo(b) != 1;
        }
        public static bool operator >(XVersion a, XVersion b)
        {
            return a.CompareTo(b) == 1;
        }
        public static bool operator >=(XVersion a, XVersion b)
        {
            return a.CompareTo(b) != -1;
        }
        public static bool operator ==(XVersion a, XVersion b)
        {
            return a.CompareTo(b) == 0;
        }
        public static bool operator !=(XVersion a, XVersion b)
        {
            return a.CompareTo(b) != 0;
        }
    }

}