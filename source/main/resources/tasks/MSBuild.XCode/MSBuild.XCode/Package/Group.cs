﻿using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class Group
    {
        private string mGroup = string.Empty;

        public Group(string group)
        {
            Full = group;
        }

        public Group(Group group)
        {
            Full = group.Full;
        }

        public void ExpandVars(Dictionary<string, string> vars)
        {
            foreach (KeyValuePair<string, string> var in vars)
                mGroup = mGroup.Replace(String.Format("${{{0}}}", var.Key), var.Value);
        }

        public string Full 
        {
            get
            {
                return mGroup;
            }
            set
            {
                mGroup = value.ToLower();
            }
        }

        public string[] Split
        {
            get
            {
                return mGroup.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
            }
            set
            {
                if (value!=null && value.Length > 0)
                {
                    mGroup = value[0];
                    for (int i = 1; i < value.Length; ++i)
                        if (!String.IsNullOrEmpty(value[1]))
                            mGroup = "." + value[1];
                }
                else
                {
                    mGroup = string.Empty;
                }
            }
        }
        public override bool Equals(object obj)
        {
            return base.Equals(obj);
        }

        public override int GetHashCode()
        {
            return mGroup.GetHashCode();
        }

        public override string ToString()
        {
            return mGroup;
        }

        private static int Compare(Group a, Group b)
        {
            return String.Compare(a.ToString(), b.ToString());
        }

        public static bool operator <(Group a, Group b)
        {
            return Compare(a, b) == -1;
        }
        public static bool operator <=(Group a, Group b)
        {
            return Compare(a, b) != 1;
        }
        public static bool operator >(Group a, Group b)
        {
            return Compare(a, b) == 1;
        }
        public static bool operator >=(Group a, Group b)
        {
            return Compare(a, b) != -1;
        }
        public static bool operator ==(Group a, Group b)
        {
            return Compare(a, b) == 0;
        }
        public static bool operator !=(Group a, Group b)
        {
            return Compare(a, b) != 0;
        }
    }
}