using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Globalization;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This class encapsulates a tag.
    /// </summary>
    [DebuggerDisplay("{Name}={RevisionNumber}:{Hash}")]
    public sealed class Tag : IEquatable<Tag>
    {
        public static bool IsHexDigit(char c)
        {
            c = Char.ToLower(c);
            return (c >= 'a' && c <= 'f') || (c >= '0' && c <= '9');
        }

        public static Tag Parse(string line)
        {
            // Scan backwards
            int revision_number_start = -1;
            int revision_number_end = -1;
            int hash_start = -1;
            int changeset_start = -1;
            int semi = line.Length - 1;
            while (semi > 0)
            {
                if (IsHexDigit(line[semi]))
                {

                }
                else if (line[semi] == ':')
                {
                    hash_start = semi + 1;
                    revision_number_end = semi;
                }
                else
                {
                    changeset_start = semi + 1;
                    revision_number_start = changeset_start;
                    break;
                }
                --semi;
            }
            string tag_name = line.Substring(0, semi).TrimEnd(' ');
            string tag_revision = line.Substring(revision_number_start, revision_number_end - revision_number_start);
            string tag_hash = line.Substring(hash_start, line.Length - hash_start).TrimEnd(' ');

            Tag tag = new Tag();
            tag.Name = tag_name;
            tag.RevisionNumber = Int32.Parse(tag_revision);
            tag.Hash = tag_hash;
            return tag;
        }

        /// <summary>
        /// The name of this <see cref="Tag"/>.
        /// </summary>
        public string Name
        {
            get;
            internal set;
        }

        /// <summary>
        /// The locally unique revision number of this <see cref="Tag"/>.
        /// </summary>
        public int RevisionNumber
        {
            get;
            internal set;
        }

        /// <summary>
        /// The unique hash of this <see cref="Tag"/>.
        /// </summary>
        public string Hash
        {
            get;
            internal set;
        }

        #region IEquatable<Tag> Members

        /// <summary>
        /// Indicates whether the current object is equal to another object of the same type.
        /// </summary>
        /// <returns>
        /// true if the current object is equal to the <paramref name="other"/> parameter; otherwise, false.
        /// </returns>
        /// <param name="other">An object to compare with this object.
        ///                 </param>
        public bool Equals(Tag other)
        {
            if (ReferenceEquals(null, other)) return false;
            if (ReferenceEquals(this, other)) return true;
            return other.Name.Equals(Name) && Equals(other.RevisionNumber, RevisionNumber);
        }

        #endregion

        /// <summary>
        /// Determines whether the specified <see cref="T:System.Object"/> is equal to the current <see cref="T:System.Object"/>.
        /// </summary>
        /// <returns>
        /// true if the specified <see cref="T:System.Object"/> is equal to the current <see cref="T:System.Object"/>; otherwise, false.
        /// </returns>
        /// <param name="obj">The <see cref="T:System.Object"/> to compare with the current <see cref="T:System.Object"/>. 
        ///                 </param><exception cref="T:System.NullReferenceException">The <paramref name="obj"/> parameter is null.
        ///                 </exception><filterpriority>2</filterpriority>
        public override bool Equals(object obj)
        {
            if (ReferenceEquals(null, obj)) return false;
            if (ReferenceEquals(this, obj)) return true;
            if (obj.GetType() != typeof (Tag)) return false;
            return Equals((Tag) obj);
        }

        /// <summary>
        /// Serves as a hash function for a particular type. 
        /// </summary>
        /// <returns>
        /// A hash code for the current <see cref="T:System.Object"/>.
        /// </returns>
        /// <filterpriority>2</filterpriority>
        public override int GetHashCode()
        {
            unchecked
            {
                int result = Name.GetHashCode();
                result = (result*397) ^ (RevisionNumber.GetHashCode());
                return result;
            }
        }

        /// <summary>
        /// Returns a <see cref="T:System.String"/> that represents the current <see cref="T:System.Object"/>.
        /// </summary>
        /// <returns>
        /// A <see cref="T:System.String"/> that represents the current <see cref="T:System.Object"/>.
        /// </returns>
        /// <filterpriority>2</filterpriority>
        public override string ToString()
        {
            return String.Format(CultureInfo.InvariantCulture, "{0}={1}:{2}", Name, RevisionNumber, Hash);
        }
    }
}