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

        private static void Add(string value, bool concat, char seperator, HashSet<string> content)
        {
            if (!String.IsNullOrEmpty(value))
            {
                string[] values = value.Split(new char[] { seperator }, StringSplitOptions.RemoveEmptyEntries);
                {
                    foreach (string v in values)
                    {
                        if (!content.Contains(v))
                            content.Add(v);
                    }
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
    }
}
