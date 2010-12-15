using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XGroup
    {
        private string mGroup = string.Empty;

        public XGroup(string group)
        {
            Full = group;
        }

        public string Full 
        {
            get
            {
                return mGroup;
            }
            set
            {
                mGroup = value;
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
    }
}