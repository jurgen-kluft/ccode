using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Globalization;
using System.Linq;
using System.Text.RegularExpressions;

namespace Mercurial
{
    /// <summary>
    /// Specifies a set of revisions, typically used to extract only
    /// a portion of the log, or specifying diff ranges
    /// </summary>
    public sealed class RevSpec : IEquatable<RevSpec>
    {
        private static readonly Regex _SafeBranchName = new Regex(@"^\s*[a-z][a-z0-9_]*\s*$", RegexOptions.IgnoreCase);
        private static readonly Regex _HashRegex = new Regex(@"^[a-f0-9]{1,40}$", RegexOptions.IgnoreCase);
        private static readonly RevSpec _Null = new RevSpec("null");
        private static readonly RevSpec _Closed = new RevSpec("closed()");
        private static readonly RevSpec _WorkingDirectoryParent = new RevSpec(".");
        private static readonly RevSpec _All = new RevSpec("all()");
        private readonly string _Value;

        private RevSpec(string value)
        {
            Debug.Assert(!StringEx.IsNullOrWhiteSpace(value), "value cannot be null or empty here");

            _Value = value.Trim();
        }

        /// <summary>
        /// Select the null revision, the revision that is
        /// the initial, empty repository, revision, the parent of
        /// revision 0.
        /// </summary>
        /// <returns>
        /// The revision specification for the empty repository revision.
        /// </returns>
        public static RevSpec Null
        {
            get
            {
                return _Null;
            }
        }

        /// <summary>
        /// Select the parent revision of the working directory. If an
        /// uncommitted merge is in progress, pick the first
        /// parent.
        /// </summary>
        /// <returns>
        /// The revision specification for the parent of the
        /// working directory.
        /// </returns>
        public static RevSpec WorkingDirectoryParent
        {
            get
            {
                return _WorkingDirectoryParent;
            }
        }

        /// <summary>
        /// All changesets in the repository.
        /// </summary>
        /// <value>
        /// The revision specification for all the changesets
        /// in the repository.
        /// </value>
        public static RevSpec All
        {
            get
            {
                return _All;
            }
        }

        /// <summary>
        /// Selects all changesets that belongs to branches found in this <see cref="RevSpec"/>.
        /// </summary>
        /// <value>
        ///     A revision specification that selects all changesets that belongs
        ///     to branches found in this &lt;see cref=&quot;RevSpec&quot;/&gt;.
        /// </value>
        public RevSpec Branches
        {
            get
            {
                return new RevSpec(string.Format(CultureInfo.InvariantCulture, "branch({0})", this));
            }
        }

        /// <summary>
        /// Selects all changesets that are child changesets of changesets in this <see cref="RevSpec"/>.
        /// </summary>
        /// <value>
        ///     A revision specification that selects all changesets that are child
        ///     changesets of changesets in this &lt;see cref=&quot;RevSpec&quot;/&gt;.
        /// </value>
        public RevSpec Children
        {
            get
            {
                return new RevSpec(string.Format(CultureInfo.InvariantCulture, "children({0})", this));
            }
        }

        /// <summary>
        /// Selects all changesets that close a branch.
        /// </summary>
        /// <value>
        /// The revision specification for all changesets that close a branch.
        /// </value>
        public static RevSpec Closed
        {
            get
            {
                return _Closed;
            }
        }

        #region IEquatable<RevSpec> Members

        /// <summary>
        /// Indicates whether the current object is equal to another object of the same type.
        /// </summary>
        /// <returns>
        /// true if the current object is equal to the <paramref name="other"/> parameter; otherwise, false.
        /// </returns>
        /// <param name="other">An object to compare with this object.
        ///                 </param>
        public bool Equals(RevSpec other)
        {
            if (ReferenceEquals(null, other)) return false;
            if (ReferenceEquals(this, other)) return true;
            return Equals(other._Value, _Value);
        }

        #endregion

        /// <summary>
        /// Create a <see cref="RevSpec"/> that includes a range
        /// of revisions, by simply selecting all changesets that has
        /// a revision number in the specified range.
        /// </summary>
        /// <param name="from">
        /// The first <see cref="RevSpec"/> to include.
        /// </param>
        /// <param name="to">
        /// The last <see cref="RevSpec"/> to include.
        /// </param>
        /// <returns>
        /// A <see cref="RevSpec"/> with the specified range.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="from"/> is <c>null</c>.</para>
        /// <para>- or -</para>
        /// <para><paramref name="to"/> is <c>null</c>.</para>
        /// </exception>
        public static RevSpec Range(RevSpec from, RevSpec to)
        {
            if (from == null)
                throw new ArgumentNullException("from");
            if (to == null)
                throw new ArgumentNullException("to");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "{0}:{1}", from, to));
        }

        /// <summary>
        /// Implements the operator ! by creating a <see cref="RevSpec"/>
        /// that includes all revisions except the ones in the specified set.
        /// </summary>
        /// <param name="set">The set.</param>
        /// <returns>The result of the operator.</returns>
        /// <exception cref="ArgumentNullException"><paramref name="set" /> is <c>null</c>.</exception>
        public static RevSpec operator !(RevSpec set)
        {
            if (set == null)
                throw new ArgumentNullException("set");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "not ({0})", set));
        }

        /// <summary>
        /// Create a <see cref="RevSpec"/> that includes a range
        /// of revisions, starting with the specified
        /// <see cref="RevSpec"/> and runs all the way up to and including
        /// the tip.
        /// </summary>
        /// <param name="revSpec">
        /// The first <see cref="RevSpec"/> to include.
        /// </param>
        /// <returns>
        /// A <see cref="RevSpec"/> with the specified range.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public static RevSpec From(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "{0}:", revSpec));
        }

        /// <summary>
        /// Create a <see cref="RevSpec"/> that includes a range
        /// of revisions, starting with the first revision in the
        /// repository, and ending with the specified
        /// <see cref="RevSpec"/>.
        /// the tip.
        /// </summary>
        /// <param name="revSpec">
        /// The last <see cref="RevSpec"/> to include.
        /// </param>
        /// <returns>
        /// A <see cref="RevSpec"/> with the specified range.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public static RevSpec To(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, ":{0}", revSpec));
        }

        /// <summary>
        /// Create a <see cref="RevSpec"/> that includes the revision
        /// specified and all descendant revisions.
        /// </summary>
        /// <param name="revSpec">
        /// The <see cref="RevSpec"/> to start from.
        /// </param>
        /// <returns>
        /// A <see cref="RevSpec"/> with the specified range.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public static RevSpec DescendantsOf(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "{0}::", revSpec));
        }

        /// <summary>
        /// Create a <see cref="RevSpec"/> that includes the revision
        /// specified and all ancestor revisions.
        /// </summary>
        /// <param name="revSpec">
        /// The <see cref="RevSpec"/> to end with.
        /// </param>
        /// <returns>
        /// A <see cref="RevSpec"/> with the specified range.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public static RevSpec AncestorsOf(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "::{0}", revSpec));
        }

        /// <summary>
        /// Create a <see cref="RevSpec"/> that includes the revisions
        /// specified, and all revisions that are both descendants of
        /// <paramref name="from"/> and ancestors of <paramref name="to"/>
        /// 
        /// </summary>
        /// <param name="from">
        /// The <see cref="RevSpec"/> to start from.
        /// </param>
        /// <param name="to">
        /// The <see cref="RevSpec"/> to end with.
        /// </param>
        /// <returns>
        /// A <see cref="RevSpec"/> with the specified range.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="from"/> is <c>null</c>.</para>
        /// <para>- or -</para>
        /// <para><paramref name="to"/> is <c>null</c>.</para>
        /// </exception>
        public static RevSpec Bracketed(RevSpec from, RevSpec to)
        {
            if (from == null)
                throw new ArgumentNullException("from");
            if (to == null)
                throw new ArgumentNullException("to");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "{0}::{1}", from, to));
        }

        /// <summary>
        /// Returns a <see cref="System.String"/> that represents this instance.
        /// </summary>
        /// <returns>
        /// A <see cref="System.String"/> that represents this instance.
        /// </returns>
        public override string ToString()
        {
            return _Value;
        }

        /// <summary>
        /// Select a revision based on its locally unique revision
        /// number.
        /// </summary>
        /// <param name="revision">
        /// The locally unique revision number.
        /// </param>
        /// <returns>
        /// The revision specification for a revision selected by
        /// its locally unique revision number.
        /// </returns>
        /// <exception cref="ArgumentOutOfRangeException">
        /// <para><paramref name="revision"/> is less than 0.</para>
        /// </exception>
        public static RevSpec Single(int revision)
        {
            if (revision < 0)
                throw new ArgumentOutOfRangeException("revision", revision, "revision cannot be negative");

            return new RevSpec(revision.ToString(CultureInfo.InvariantCulture));
        }

        /// <summary>
        /// Select a revision based on its globally unique hash.
        /// </summary>
        /// <param name="hash">
        /// The globally unique hash, a 1-40 digit hexadecimal number.
        /// </param>
        /// <returns>
        /// The revision specification for a revision selected by
        /// its globally unique hash.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="hash"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><paramref name="hash"/> does not contain a valid hexadecimal number of maximum 40 digits.</para>
        /// </exception>
        public static RevSpec Single(string hash)
        {
            if (StringEx.IsNullOrWhiteSpace(hash))
                throw new ArgumentNullException("hash");
            if (!_HashRegex.Match(hash.Trim()).Success)
                throw new ArgumentException("The hash parameter does not contain a valid hexadecimal hash");

            return new RevSpec(hash.Trim().ToLowerInvariant());
        }

        /// <summary>
        /// Select a revision based on its branch name, will select the
        /// tipmost revision that belongs to the named branch.
        /// </summary>
        /// <param name="name">
        /// The name of the branch to select the tipmost revision of.
        /// </param>
        /// <returns>
        /// The revision specification for the tipmost revision that
        /// belongs to the named branch.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="name"/> is <c>null</c> or empty.</para>
        /// </exception>
        public static RevSpec ByBranch(string name)
        {
            if (StringEx.IsNullOrWhiteSpace(name))
                throw new ArgumentNullException("name");

            if (_SafeBranchName.Match(name).Success)
                return new RevSpec(name.Trim());
            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "'{0}'", name.Trim()));
        }

        /// <summary>
        /// Select a revision based on tag.
        /// </summary>
        /// <param name="name">
        /// The name of the tag to select.
        /// </param>
        /// <returns>
        /// The revision specification for the revision that has
        /// the specified tag.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="name"/> is <c>null</c> or empty.</para>
        /// </exception>
        public static RevSpec ByTag(string name)
        {
            if (StringEx.IsNullOrWhiteSpace(name))
                throw new ArgumentNullException("name");

            return new RevSpec(name.Trim());
        }

        /// <summary>
        /// Performs an implicit conversion from <see cref="System.Int32"/> to <see cref="RevSpec"/>
        /// by using the number as the revision number.
        /// </summary>
        /// <param name="revisionNumber">The revision number.</param>
        /// <returns>The result of the conversion.</returns>
        public static implicit operator RevSpec(int revisionNumber)
        {
            return Single(revisionNumber);
        }

        /// <summary>
        /// Performs an implicit conversion from <see cref="System.String"/> to <see cref="RevSpec"/>
        /// by using the string as the revision hash.
        /// </summary>
        /// <param name="hash">The hash.</param>
        /// <returns>The result of the conversion.</returns>
        public static implicit operator RevSpec(string hash)
        {
            return Single(hash);
        }

        /// <summary>
        /// Performs an implicit conversion from <see cref="RevSpec"/> to <see cref="System.String"/>
        /// by calling the <see cref="ToString"/> method.
        /// </summary>
        /// <param name="revSpec">The revisions.</param>
        /// <returns>The result of the conversion.</returns>
        public static implicit operator string(RevSpec revSpec)
        {
            if (revSpec == null)
                return WorkingDirectoryParent;

            return revSpec.ToString();
        }

        /// <summary>
        /// Select all changesets in this <see cref="RevSpec"/> specification, but
        /// not in <paramref name="revSpec"/>.
        /// </summary>
        /// <param name="revSpec">
        /// The revisions of the changesets to exclude.
        /// </param>
        /// <returns>
        /// A revision specification that selects all the changesets in this
        /// <see cref="RevSpec"/>, but not in <paramref name="revSpec"/>.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public RevSpec Except(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "({0}) - ({1})", this, revSpec));
        }

        /// <summary>
        /// Selects all changesets that are both in this <see cref="RevSpec"/>
        /// and also in <paramref name="revSpec"/>.
        /// </summary>
        /// <param name="revSpec">
        /// The 2nd operand to the <c>and</c> operator.
        /// </param>
        /// <returns>
        /// A revision specification that selects all changesets that are both in this <see cref="RevSpec"/>
        /// and also in <paramref name="revSpec"/>.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public RevSpec And(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "({0}) and ({1})", this, revSpec));
        }

        /// <summary>
        /// Selects all changesets that either in this <see cref="RevSpec"/>
        /// or in <paramref name="revSpec"/>.
        /// </summary>
        /// <param name="revSpec">
        /// The 2nd operand to the <c>or</c> operator.
        /// </param>
        /// <returns>
        /// A revision specification that selects all changesets that are either in this <see cref="RevSpec"/>
        /// or in <paramref name="revSpec"/>.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revSpec"/> is <c>null</c>.</para>
        /// </exception>
        public RevSpec Or(RevSpec revSpec)
        {
            if (revSpec == null)
                throw new ArgumentNullException("revSpec");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "({0}) or ({1})", this, revSpec));
        }

        /// <summary>
        /// Selects all changesets committed by the specified <paramref name="name"/>.
        /// </summary>
        /// <param name="name">
        /// The name of the author to select changesets for.
        /// </param>
        /// <returns>
        /// A revision specification that selects all changesets committed
        /// by the specified <paramref name="name"/>.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="name"/> is <c>null</c> or empty.</para>
        /// </exception>
        public static RevSpec Author(string name)
        {
            if (StringEx.IsNullOrWhiteSpace(name))
                throw new ArgumentNullException("name");

            return new RevSpec(string.Format(CultureInfo.InvariantCulture, "author('{0}')", name));
        }

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
            if (obj.GetType() != typeof (RevSpec)) return false;
            return Equals((RevSpec) obj);
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
            return (_Value != null ? _Value.GetHashCode() : 0);
        }
    }
}