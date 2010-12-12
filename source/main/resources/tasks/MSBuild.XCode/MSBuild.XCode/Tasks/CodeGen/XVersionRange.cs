using System;
using System.Text;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    ///
    /// Generic implementation of version range (see Version.docx in docs\manuals)
    /// Examples:
    ///   [1.2.0,], is version 1.2.0 and higher
    public class XVersionRange
    {
        private List<Range> mRanges;

        class Range
        {
            private static string[,] sDelimiters = new string[2, 2] { { "(", ")" }, { "[", "]" } };

            private int mIFrom;
            private XVersion mFrom;
            private int mITo;
            private XVersion mTo;

            public Range()
            {
                mIFrom = 1;
                mFrom = new XVersion("1.0");
                mITo = 1;
                mTo = new XVersion("");
            }

            public Range(string range)
            {
                FromString(range);
            }

            public Range(XVersion from, XVersion to, bool includeFrom, bool includeTo)
            {
                mFrom = from;
                mTo = to;
                IncludeFrom = includeFrom;
                IncludeTo = includeTo;

                if (mFrom == mTo)
                {
                    mTo = new XVersion(string.Empty);
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

            public XVersion From { get { return mFrom; } }
            public XVersion To { get { return mTo; } }
            public bool IsFromNull { get { return mFrom.IsNull; } }
            public bool IsToNull { get { return mTo.IsNull; } }


            public void Split(XVersion lowest, XVersion highest, out XVersionRange outRange, out XVersion outVersion)
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
                            outRange = new XVersionRange(new XVersion(mFrom), new XVersion(highest), IncludeFrom, true);
                        else if (lowest > mFrom)
                            outRange = new XVersionRange(new XVersion(lowest), new XVersion(highest), true, true);
                    }
                    else if (mFrom.IsNull && !mTo.IsNull)
                    {
                        // #3, From=Unbounded ==> To=Version
                        if (lowest < mTo && highest > mTo)  // Clamp highest
                            outRange = new XVersionRange(lowest, mTo, true, IncludeTo);
                        else if (highest < mTo)
                            outRange = new XVersionRange(lowest, highest, true, true);
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
                            XVersion f, t;
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

                            outRange = new XVersionRange(new XVersion(f), new XVersion(t), fi, ti);
                        }
                    }
                }
            }

            public bool IsInRange(XVersion version)
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
                    mFrom = new XVersion("1.0");
                    mTo = new XVersion("");
                }
                else if (from_to.Length == 1)
                {
                    // Unique version
                    IncludeFrom = true;
                    IncludeTo = true;
                    IsRange = false;
                    mFrom = new XVersion(from_to[0]);
                    mTo = new XVersion("");
                }
                else if (from_to.Length == 2)
                {
                    // A real range
                    IncludeFrom = version.StartsWith("[");
                    IncludeTo = version.EndsWith("]");
                    mFrom = new XVersion(from_to[0]);
                    mTo = new XVersion(from_to[1]);

                    if (mFrom == mTo)
                    {
                        mTo = new XVersion(string.Empty);
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

        public XVersionRange()
        {
            mRanges = new List<Range>();
        }

        public XVersionRange(string range)
        {
            mRanges = new List<Range>();
            FromString(range);
        }

        public XVersionRange(XVersion from, XVersion to, bool includeFrom, bool includeTo)
        {
            mRanges = new List<Range>();
            mRanges.Add(new Range(from, to, includeFrom, includeTo));
            SetKind();
        }

        public XVersion From
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

        public XVersion To
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

        public void Add(XVersion from, XVersion to, bool includeFrom, bool includeTo)
        {
            mRanges = new List<Range>();
            mRanges.Add(new Range(from, to, includeFrom, includeTo));
            SetKind();
        }

        public bool IsInRange(XVersion version)
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

        public void Split(XVersion lowest, XVersion highest, out XVersionRange[] outRanges, out XVersion outVersion)
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

            List<XVersionRange> ranges = new List<XVersionRange>();
            foreach (Range r in mRanges)
            {
                XVersion version;
                XVersionRange range;
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


        public static int Join(XVersionRange a, XVersionRange b)
        {
            bool a_null = (object)a == null;
            bool b_null = (object)b == null;
            if (a_null && b_null)
                return 0;
            if (a_null != b_null)
                return a_null ? -1 : 1;

            bool same_kind = (a.Kind == b.Kind);

            int cfrom = a.From.CompareTo(b.From);
            int cto = a.To.CompareTo(b.To);

        }

        public static bool operator <(XVersionRange a, XVersionRange b)
        {
            return Compare(a, b) == -1;
        }
        public static bool operator <=(XVersionRange a, XVersionRange b)
        {
            return Compare(a, b) != 1;
        }
        public static bool operator >(XVersionRange a, XVersionRange b)
        {
            return Compare(a, b) == 1;
        }
        public static bool operator >=(XVersionRange a, XVersionRange b)
        {
            return Compare(a, b) != -1;
        }
        public static bool operator ==(XVersionRange a, XVersionRange b)
        {
            return Compare(a, b) == 0;
        }
        public static bool operator !=(XVersionRange a, XVersionRange b)
        {
            return Compare(a, b) != 0;
        }
    }
}