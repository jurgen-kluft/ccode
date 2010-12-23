using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public class StringItems
    {
        private char mSeperator = ';';
        private HashSet<string> mItems;

        public StringItems()
        {
            mItems = new HashSet<string>();
        }
        public StringItems(string[] items)
        {
            mItems = new HashSet<string>();
            foreach(string item in items)
                AddOne(item, mItems);
        }

        public void Add(string value, bool concat)
        {
            Add(value, concat, mSeperator, mItems);
        }

        public void Add(StringItems other)
        {
            foreach (string v in other.mItems)
            {
                if (!mItems.Contains(v))
                    mItems.Add(v);
            }           
        }

        private string Filter(string str, string[] prefixes, bool remove)
        {
            foreach (string prefix in prefixes)
            {
                if (str.StartsWith(prefix))
                {
                    if (remove)
                        return null;

                    string f = str.Substring(prefix.Length, str.Length - prefix.Length);
                    return f;
                }
            }
            return str;
        }

        public void Filter(string[] remove, string[] keep)
        {
            HashSet<string> filteredItems = new HashSet<string>();
            foreach (string s in mItems)
            {
                string f = Filter(s, remove, true);
                if (!String.IsNullOrEmpty(f))
                {
                    f = Filter(f, keep, false);
                    if (!String.IsNullOrEmpty(f))
                    {
                        if (!filteredItems.Contains(f))
                            filteredItems.Add(f);
                    }
                }
            }
            mItems = filteredItems;
        }

        public string Get()
        {
            return Get(mItems, mSeperator);
        }
        
        public bool Contains(string item)
        {
            return mItems.Contains(item);
        }
        
        public string[] ToArray()
        {
            return mItems.ToArray();
        }

        private static void AddOne(string value, HashSet<string> content)
        {
            if (!content.Contains(value))
                content.Add(value);
        }

        private static void Add(string value, bool concat, char seperator, HashSet<string> content)
        {
            if (!String.IsNullOrEmpty(value))
            {
                string[] values = value.Split(new char[] { seperator }, StringSplitOptions.RemoveEmptyEntries);
                {
                    foreach (string v in values)
                        AddOne(v, content);
                }
            }
        }
        private static string Get(HashSet<string> content, char seperator)
        {
            string str = string.Empty;
            {
                foreach (string s in content)
                    str = str + seperator + s;
                str = str.TrimStart(seperator);
                str = str.TrimEnd(seperator);
            }
            return str;
        }

        public override string ToString()
        {
            return Get();
        }
    }
}
