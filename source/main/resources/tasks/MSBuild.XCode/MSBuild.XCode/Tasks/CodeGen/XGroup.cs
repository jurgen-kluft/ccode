using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class XGroup
    {
        private string mGroup = string.Empty;
        private string[] mCategory = new string[0];

        public XGroup(string group)
        {
            Group = group;
        }

        public string Group 
        {
            get
            {
                return mGroup;
            }
            set
            {
                Split(value);
                mGroup = value;
            }
        }

        private void Split(string value)
        {
            mCategory = value.Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
        }
    }
}