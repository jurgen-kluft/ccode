using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    ///
    /// Generic implementation of version range (see Version.docx in docs\manuals)
    /// Examples:
    ///   [1.2.0,], is version 1.2.0 and higher
    public class VersionRange
    {
        private List<Range> mRanges;

        class Range
        {
            private static string[,] sDelimiters = new string[2, 2] { { "(", ")" }, { "[", "]" } };

            private int mIFrom;
            private ComparableVersion mFrom;
            private int mITo;
            private ComparableVersion mTo;

            public Range()
            {
                mIFrom = 1;
                mFrom = new ComparableVersion("1.0");
                mITo = 1;
                mTo = new ComparableVersion("");
            }

            public Range(string range)
            {
                FromString(range);
            }

            public Range(ComparableVersion unique)
            {
                mFrom = new ComparableVersion(unique);
                IncludeFrom = true;
                mTo = new ComparableVersion(string.Empty);
                IncludeTo = false;
                IsRange = false;
            }

            public Range(ComparableVersion from, ComparableVersion to, bool includeFrom, bool includeTo)
            {
                mFrom = from;
                mTo = to;
                IncludeFrom = includeFrom;
                IncludeTo = includeTo;

                if (mFrom == mTo)
                {
                    mTo = new ComparableVersion(string.Empty);
                    IsRange = false;
                }
                else
                {
                    IsRange = true;
                }
            }

            public bool IsRange { get; set; }

            public bool IncludeFrom { get { return mIFrom == 1; } set { mIFrom = value ? 1 : 0; } }
            public bool IncludeTo { get { return mITo == 1; } set { mITo = value ? 1 : 0; } }

            public ComparableVersion From { get { return mFrom; } set { mFrom = value; } }
            public ComparableVersion To { get { return mTo; } set { mTo = value; } }
            public bool IsFromNull { get { return mFrom.IsNull; } }
            public bool IsToNull { get { return mTo.IsNull; } }


            public void Split(ComparableVersion lowest, ComparableVersion highest, out VersionRange outRange, out ComparableVersion outVersion)
            {
                // Range cases (5):
                // 1) Unique version
                // 2) X >= Version                    ==> A bounded range can be created
                // 3) X <= Version                    ==> A bounded range can be created
                // 4) X >= VersionA AND X <= VersionB ==> A bounded range can be created
                outRange = null;
                outVersion = null;

                if (!IsRange)
                {
                    // #1
                    outVersion = mFrom;
                }
                else
                {
                    if (!mFrom.IsNull && mTo.IsNull)
                    {
                        // #2, From=Version ==> To=Unbounded
                        if (lowest < mFrom && highest > mFrom)  // Clamp lowest
                            outRange = new VersionRange(new ComparableVersion(mFrom), new ComparableVersion(highest), IncludeFrom, true);
                        else if (lowest > mFrom)
                            outRange = new VersionRange(new ComparableVersion(lowest), new ComparableVersion(highest), true, true);
                    }
                    else if (mFrom.IsNull && !mTo.IsNull)
                    {
                        // #3, From=Unbounded ==> To=Version
                        if (lowest < mTo && highest > mTo)  // Clamp highest
                            outRange = new VersionRange(lowest, mTo, true, IncludeTo);
                        else if (highest < mTo)
                            outRange = new VersionRange(lowest, highest, true, true);
                    }
                    else if (!mFrom.IsNull && !mTo.IsNull)
                    {
                        // #4, From=Version ==> To=Version
                        if (lowest > mTo || highest < mFrom)
                        {
                            // Out of range
                        }
                        else
                        {
                            ComparableVersion f, t;
                            bool fi = true, ti = true;
                            if (lowest < mFrom)
                            {
                                f = mFrom;
                                fi = IncludeFrom;
                            }
                            else
                            {
                                f = lowest;
                            }

                            if (highest > mTo)
                            {
                                t = mTo;
                                ti = IncludeTo;
                            }
                            else
                            {
                                t = highest;
                            }

                            outRange = new VersionRange(new ComparableVersion(f), new ComparableVersion(t), fi, ti);
                        }
                    }
                }
            }

            public bool IsInRange(ComparableVersion version)
            {
                bool from = false, to = false;
                if (IsRange)
                {
                    if (!mFrom.IsNull && !mTo.IsNull)
                    {
                        from = IncludeFrom ? (version >= mFrom) : (version > mFrom);
                        to = from ? (IncludeTo ? (version <= mTo) : (version < mTo)) : false;
                    }
                    else if (!mFrom.IsNull && mTo.IsNull)
                    {
                        from = IncludeFrom ? (version >= mFrom) : (version > mFrom);
                        to = true;
                    }
                    else if (mFrom.IsNull && !mTo.IsNull)
                    {
                        from = true;
                        to = IncludeTo ? (version <= mTo) : (version < mTo);
                    }
                    else
                    {
                        from = true;
                        to = true;
                    }
                }
                else
                {
                    from = (version == mFrom);
                    to = from;
                }
                return from && to;
            }

            public void FromString(string version)
            {
                string[] from_to = version.Substring(1, version.Length - 2).Split(new char[] { ',' }, StringSplitOptions.None);
                if (from_to.Length == 0)
                {
                    // Default to x >= 1.0
                    IncludeFrom = true;
                    IncludeTo = true;
                    IsRange = false;
                    mFrom = new ComparableVersion("1.0");
                    mTo = new ComparableVersion("");
                }
                else if (from_to.Length == 1)
                {
                    // Unique version
                    IncludeFrom = true;
                    IncludeTo = true;
                    IsRange = false;
                    mFrom = new ComparableVersion(from_to[0]);
                    mTo = new ComparableVersion("");
                }
                else if (from_to.Length == 2)
                {
                    // A real range
                    IncludeFrom = version.StartsWith("[");
                    IncludeTo = version.EndsWith("]");
                    mFrom = new ComparableVersion(from_to[0]);
                    mTo = new ComparableVersion(from_to[1]);

                    if (mFrom == mTo)
                    {
                        mTo = new ComparableVersion(string.Empty);
                        IsRange = false;
                    }
                    else
                    {
                        IsRange = true;
                    }
                }
            }

            public override String ToString()
            {
                if (mFrom != null && mTo != null)
                    return String.Format("{0}{1},{2}{3}", sDelimiters[mIFrom, 0], mFrom, mTo, sDelimiters[mITo, 1]);
                else
                    return String.Format("{0}{1}{2}", sDelimiters[mIFrom, 0], mFrom, sDelimiters[mITo, 1]);
            }
        }

        public VersionRange()
        {
            mRanges = new List<Range>();
        }

        public VersionRange(string range)
        {
            mRanges = new List<Range>();
            FromString(range);
        }

        public VersionRange(ComparableVersion unique)
        {
            mRanges = new List<Range>();
            mRanges.Add(new Range(unique));
        }

        public VersionRange(ComparableVersion from, ComparableVersion to, bool includeFrom, bool includeTo)
        {
            mRanges = new List<Range>();
            mRanges.Add(new Range(from, to, includeFrom, includeTo));
            SetKind();
        }

        public ComparableVersion From
        {
            get
            {
                switch (Kind)
                {
                    default:
                    case EKind.Invalid:
                        return null;
                    case EKind.UniqueVersion:
                    case EKind.VersionToUnbound:
                    case EKind.UnboundToVersion:
                    case EKind.VersionToVersion:
                        return mRanges[0].From;
                    case EKind.UnboundToVersionOrVersionToUnbound:
                        return mRanges[0].To;
                }
            }
            set
            {
                mRanges[0].From = value;
            }
        }

        public bool IncludeFrom
        {
            get
            {
                switch (Kind)
                {
                    default:
                    case EKind.Invalid:
                        return false;
                    case EKind.UniqueVersion:
                    case EKind.VersionToUnbound:
                    case EKind.UnboundToVersion:
                    case EKind.VersionToVersion:
                        return mRanges[0].IncludeFrom;
                    case EKind.UnboundToVersionOrVersionToUnbound:
                        return mRanges[0].IncludeTo;
                }
            }
        }

        public ComparableVersion To
        {
            get
            {
                switch (Kind)
                {
                    default:
                    case EKind.Invalid:
                        return null;
                    case EKind.UniqueVersion:
                        return mRanges[0].From;
                    case EKind.VersionToUnbound:
                    case EKind.UnboundToVersion:
                    case EKind.VersionToVersion:
                        return mRanges[0].To;
                    case EKind.UnboundToVersionOrVersionToUnbound:
                        return mRanges[1].From;
                }
            }
            set
            {
                mRanges[0].To = value;
            }
        }

        public bool IncludeTo
        {
            get
            {
                switch (Kind)
                {
                    default:
                    case EKind.Invalid:
                        return false;
                    case EKind.UniqueVersion:
                    case EKind.VersionToUnbound:
                    case EKind.UnboundToVersion:
                    case EKind.VersionToVersion:
                        return mRanges[0].IncludeTo;
                    case EKind.UnboundToVersionOrVersionToUnbound:
                        return mRanges[1].IncludeFrom;
                }
            }
        }

        public enum EKind
        {
            Invalid,
            UniqueVersion,                          // Case 1
            VersionToUnbound,                       // Case 2, X >= Version
            UnboundToVersion,                       // Case 3, X <= Version
            VersionToVersion,                       // Case 4, X >= VersionA AND X <= VersionB
            UnboundToVersionOrVersionToUnbound,     // Case 5, X <= VersionA OR X >= VersionB
        }

        public EKind Kind { get; set; }

        public void Add(ComparableVersion from, ComparableVersion to, bool includeFrom, bool includeTo)
        {
            mRanges = new List<Range>();
            mRanges.Add(new Range(from, to, includeFrom, includeTo));
            SetKind();
        }

        public bool IsInRange(ComparableVersion version)
        {
            bool in_range = false;

            // Multiple ranges are evaluated using the OR operator
            foreach (Range r in mRanges)
            {
                if (r.IsInRange(version))
                {
                    in_range = true;
                    break;
                }
            }
            return in_range;
        }

        public void Split(ComparableVersion lowest, ComparableVersion highest, out VersionRange[] outRanges, out ComparableVersion outVersion)
        {
            // XVersionRange cases (5):
            // 1) Unique version
            // 2) X >= Version                    ==> A bounded range can be created
            // 3) X <= Version                    ==> A bounded range can be created
            // 4) X >= VersionA AND X <= VersionB ==> A bounded range can be created
            // 5) X <= VersionA OR X >= VersionB  ==> One or two bounded ranges can be created

            outVersion = null;

            // Case #5 ?
            bool isCase5 = mRanges.Count == 2;

            List<VersionRange> ranges = new List<VersionRange>();
            foreach (Range r in mRanges)
            {
                ComparableVersion version;
                VersionRange range;
                r.Split(lowest, highest, out range, out version);
                if (version != null)
                {
                    // #1 Unique version
                    outVersion = version;
                }
                else if (range != null)
                {
                    ranges.Add(range);
                }
            }

            outRanges = ranges.ToArray();
        }

        private void SetKind()
        {
            // See if we have 2 ranges which are actually a case #4, if so combine them into one range
            if (mRanges.Count == 2)
            {
                if ((!mRanges[0].IsFromNull && mRanges[0].IsToNull && mRanges[1].IsFromNull && !mRanges[1].IsToNull))
                {
                    Range combined = new Range(mRanges[0].From, mRanges[1].To, mRanges[0].IncludeFrom, mRanges[1].IncludeTo);
                    mRanges.Clear();
                    mRanges.Add(combined);
                    Kind = EKind.VersionToVersion;
                }
                else if (mRanges[0].IsFromNull && !mRanges[0].IsToNull && !mRanges[1].IsFromNull && mRanges[1].IsToNull)
                {
                    Kind = EKind.UnboundToVersionOrVersionToUnbound;
                }
            }
            else if (mRanges.Count == 1)
            {
                if (!mRanges[0].IsRange)
                    Kind = EKind.UniqueVersion;
                else if (!mRanges[0].IsFromNull && mRanges[0].IsToNull)
                    Kind = EKind.VersionToUnbound;
                else if (mRanges[0].IsFromNull && !mRanges[0].IsToNull)
                    Kind = EKind.UnboundToVersion;
                else if (!mRanges[0].IsFromNull && !mRanges[0].IsToNull)
                    Kind = EKind.VersionToVersion;
            }
        }

        public bool Merge(VersionRange other)
        {
            // @TODO: finalize
            bool merged = false;
            if (Kind == EKind.VersionToUnbound)
            {
                if (other.Kind == EKind.UnboundToVersion)
                {
                    // this wins!
                }
                else if (other.Kind == EKind.VersionToUnbound)
                {
                    if (IncludeFrom == other.IncludeFrom)
                    {
                        if (From < other.From)
                        {
                            // this wins!
                        }
                        else
                        {
                            From = other.From;
                            merged = true;
                        }
                    }
                }
                else if (other.Kind == EKind.VersionToVersion)
                {
                    if (From < other.From)
                    {
                        // this wins!
                    }
                    else
                    {
                        From = other.From;
                        merged = true;
                    }
                }
                else if (other.Kind == EKind.UnboundToVersionOrVersionToUnbound)
                {
                    if (other.To.LessThan(From, false))
                    {
                        // this wins!
                    }
                    else
                    {
                        From = other.To;
                        merged = true;
                    }
                }
            }
            return merged;
        }

        public void FromString(string range)
        {
            int cursor = 0;
            int begin = 0;
            int end = 0;
            bool open = false;
            while (cursor < range.Length)
            {
                char c = range[cursor];
                if (!open)
                {
                    if (c == ',')
                    {
                        // Range delimiter
                    }
                    else if (c == '(' || c == '[')
                    {
                        begin = cursor;
                        open = true;
                    }
                }
                else
                {
                    if (c == ')' || c == ']')
                    {
                        end = cursor + 1;
                        mRanges.Add(new Range(range.Substring(begin, end - begin)));

                        open = false;
                        begin = -1;
                        end = -1;
                    }
                }
                ++cursor;
            }
            SetKind();
        }

        public override String ToString()
        {
            string range = string.Empty;
            foreach (Range r in mRanges)
            {
                if (range.Length == 0)
                {
                    range = r.ToString();
                }
                else
                {
                    range = range + "," + r.ToString();
                }
            }
            return range;
        }

        public override bool Equals(object obj)
        {
            return base.Equals(obj);
        }

        public override int GetHashCode()
        {
            return ToString().GetHashCode();   
        }

        public static int Compare(VersionRange a, VersionRange b)
        {
            bool a_null = (object)a == null;
            bool b_null = (object)b == null;
            if (a_null && b_null)
                return 0;
            if (a_null != b_null)
                return a_null ? -1 : 1;

            string a_str = a.ToString();
            string b_str = b.ToString();
            if (String.Compare(a_str, b_str)==0)
                return 0;

            return -1;
        }

        public static bool operator <(VersionRange a, VersionRange b)
        {
            return Compare(a, b) == -1;
        }
        public static bool operator <=(VersionRange a, VersionRange b)
        {
            return Compare(a, b) != 1;
        }
        public static bool operator >(VersionRange a, VersionRange b)
        {
            return Compare(a, b) == 1;
        }
        public static bool operator >=(VersionRange a, VersionRange b)
        {
            return Compare(a, b) != -1;
        }
        public static bool operator ==(VersionRange a, VersionRange b)
        {
            return Compare(a, b) == 0;
        }
        public static bool operator !=(VersionRange a, VersionRange b)
        {
            return Compare(a, b) != 0;
        }
    }
}