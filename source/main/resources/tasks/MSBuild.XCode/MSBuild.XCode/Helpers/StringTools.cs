using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;

namespace MSBuild.Cod.Helpers
{
	#region String Formatting Documentation
	/// <summary>
	/// The {0} in the string above is replaced with the value of nError, but what
	/// if you want to specify the number of digits to use? Or the base (hexadecimal etc)? 
	/// The framework supports all this, but where it seemd confusing is that it's not
	/// the String.Format function that does the string formatting, but rather the types
	/// themselves. Every object has a method called ToString that returns a string 
	/// representation of the object. The ToString method can accept a string parameter, 
	/// which tells the object how to format itself - in the String.Format call, the 
	/// formatting string is passed after the position, for example, "{0:##}"
	/// 
	/// The text inside the curly braces is {index[,alignment][:formatString]}. If 
	/// alignment is positive, the text is right-aligned in a field the given number 
	/// of spaces; if it's negative, it's left-aligned.

	/// Strings
	/// There really isn't any formatting within a string beyond it's alignment. Alignment works
	/// for any argument being printed in a String.Format call.
	/// 
	/// Sample 										Generates
	/// <code>
	/// 	String.Format(" ->{1,10}<-", "Hello");		-> Hello<-
	///		String.Format(" ->{1,-10}<-", "Hello"); 	->Hello <-
	/// </code>
	/// 
	/// Numbers
	/// 
	/// Basic number formatting specifiers:
	/// 
	/// Specifier 	Type 								Format 	Output (Passed Double 1.42) 	Output (Passed Int -12400)
	/// c 			Currency							{0:c} 	$1.42 							-$12,400
	/// d 			Decimal	(Whole number) 				{0:d} 	System.FormatException 			-12400
	/// e 			Scientific 							{0:e} 	1.420000e+000 					-1.240000e+004
	/// f 			Fixed point							{0:f} 	1.42 							-12400.00
	/// g 			General								{0:g} 	1.42 							-12400
	/// n 			Number with commas for thousands 	{0:n} 	1.42 							-12,400
	/// r 			Round trippable 					{0:r} 	1.42 							System.FormatException
	/// v 			Hexadecimal 						{0:x4} 	System.FormatException 			cf90
	/// 
	/// Custom number formatting:
	/// 
	/// Specifier 	Type 					Example 		Output (Passed Double 1500.42) 		Note
	/// 0 			Zero placeholder 		{0:00.0000} 	1500.4200 							Pads with zeroes.
	/// # 			Digit placeholder 		{0:(#).##} 		(1500).42 							
	/// . 			Decimal point 			{0:0.0} 		1500.4 	
	/// , 			Thousand separator 		{0:0,0} 		1,500 								Must be between two zeroes.
	/// ,. 			Number scaling 			{0:0,.} 		2 									Comma adjacent to Period scales by 1000.
	/// % 			Percent 				{0:0%} 			150042% 							Multiplies by 100, adds % sign.
	/// e 			Exponent placeholder 	{0:00e+0} 		15e+2 								Many exponent formats available.
	/// ; 			Group separator 	see below 		
	/// 
	/// The group separator is especially useful for formatting currency values which 
	/// require that negative values be enclosed in parentheses. This currency formatting 
	/// example at the bottom of this document makes it obvious:
	/// 
	/// Dates
	/// 
	/// Note that date formatting is especially dependant on the system's regional settings; the
	/// example strings here are from my local locale.
	/// 
	/// Specifier 	Type 								Example (Passed System.DateTime.Now)
	/// d 			Short date 							10/12/2002
	/// D 			Long date 							December 10, 2002
	/// t 			Short time 							10:11 PM
	/// T 			Long time 							10:11:29 PM
	/// f 			Full date & time 					December 10, 2002 10:11 PM
	/// F 			Full date & time (long) 			December 10, 2002 10:11:29 PM
	/// g 			Default date & time 				10/12/2002 10:11 PM
	/// G 			Default date & time (long) 			10/12/2002 10:11:29 PM
	/// M 			Month day pattern 					December 10
	/// r 			RFC1123 date string 				Tue, 10 Dec 2002 22:11:29 GMT
	/// s 			Sortable date string 				2002-12-10T22:11:29
	/// u 			Universal sortable, local time 		2002-12-10 22:13:50Z
	/// U 			Universal sortable, GMT 			December 11, 2002 3:13:50 AM
	/// Y 			Year month pattern 					December, 2002
	/// 
	/// The 'U' specifier seems broken; that string certainly isn't sortable.
	/// 
	/// Custom date formatting:
	/// Specifier 		Type 						Example 		Example Output
	/// dd 				Day 						{0:dd} 			10
	/// ddd 			Day name 					{0:ddd} 		Tue
	/// dddd 			Full day name 				{0:dddd} 		Tuesday
	/// f, ff, ... 	 	Second fractions 			{0:fff} 		932
	/// gg, ...			Era							{0:gg} 			A.D.
	/// hh 				2 digit hour 				{0:hh} 			10
	/// HH 				2 digit hour, 24hr format 	{0:HH} 			22
	/// mm 				Minute 00-59 				{0:mm} 			38
	/// MM 				Month 01-12 				{0:MM} 			12
	/// MMM 			Month abbreviation 			{0:MMM} 		Dec
	/// MMMM 			Full month name 			{0:MMMM} 		December
	/// ss 				Seconds 00-59 				{0:ss} 			46
	/// tt 				AM or PM 					{0:tt} 			PM
	/// yy 				Year, 2 digits 				{0:yy} 			02
	/// yyyy 			Year 						{0:yyyy} 		2002
	/// zz 				Timezone offset, 2 digits 	{0:zz} 			-05
	/// zzz 			Full timezone offset 		{0:zzz} 		-05:00
	/// : 				Separator 					{0:hh:mm:ss} 	10:43:20
	/// / 				Separator 					{0:dd/MM/yyyy} 	10/12/2002
	/// 
	/// Enumerations
	/// 
	/// Specifier 	Type
	/// g 			Default (Flag names if available, otherwise decimal)
	/// f 			Flags always
	/// d 			Integer always
	/// v 			Eight digit hex.
	/// 
	/// Some Useful Examples
	/// 
	/// String.Format("{0:$#,##0.00;($#,##0.00);Zero}", value);
	/// 
	///     This will output "$1,240.00" if passed 1243.50. It will output the same
	///		format but in parentheses if the number is negative, and will output the 
	///		string "Zero" if the number is zero.
	/// 
	/// String.Format("{0:(###) ###-####}", 8005551212);
	/// 
	///     This will output "(800) 555-1212".
	/// 
	/// </summary>
	#endregion

	/// <summary>
	/// UtilityClass with some additional functions for string operations
	/// </summary>
	public static class StringTools
	{
        private static Regex NullStringRegex = new Regex("\0");

        public static string MD5ToString(byte[] hash)
        {
            string str = string.Empty;
            for (int i = 0; i < 16; ++i)
            {
                byte b = hash[i];
                str += StringTools.NibbleToHex((byte)(b & 0xF));
                b = (byte)(b >> 4);
                str += StringTools.NibbleToHex((byte)(b & 0xF));
            }
            return str;
        }

        public static bool Contains(string text, string fragment)
        {
            return text.IndexOf(fragment) > -1;
        }

        public static bool EqualsIgnoreCase(string a, string b)
        {
            return CaseInsensitiveComparer.Default.Compare(a, b) == 0;
        }

        public static string JoinUnique(string delimiter, params string[][] fragmentArrays)
        {
            SortedList list = new SortedList();
            foreach (string[] fragmentArray in fragmentArrays)
            {
                foreach (string fragment in fragmentArray)
                {
                    if (!list.Contains(fragment))
                        list.Add(fragment, fragment);
                }
            }
            StringBuilder buffer = new StringBuilder();
            foreach (string value in list.Values)
            {
                if (buffer.Length > 0)
                {
                    buffer.Append(delimiter);
                }
                buffer.Append(value);
            }
            return buffer.ToString();
        }

        public static int GenerateHashCode(params string[] values)
        {
            int hashcode = 0;
            foreach (string value in values)
            {
                if (value != null)
                {
                    hashcode += value.GetHashCode();
                }
            }
            return hashcode;
        }

        public static string LastWord(string input)
        {
            return LastWord(input, " .,;!?:");
        }

        public static string LastWord(string input, string separators)
        {
            if (input == null)
            {
                return null;
            }
            string[] tokens = input.Split(separators.ToCharArray());
            for (int i = tokens.Length - 1; i >= 0; i--)
            {
                if (IsWhitespace(tokens[i]) == false)
                {
                    return tokens[i].Trim();
                }
            }
            return String.Empty;
        }

        public static bool IsBlank(string input)
        {
            return (input == null || input.Length == 0);
        }

        public static bool IsWhitespace(string input)
        {
            return (input == null || input.Length == 0 || input.Trim().Length == 0);
        }

        public static string Strip(string input, params string[] removals)
        {
            string revised = input;
            foreach (string removal in removals)
            {
                int i = 0;
                while ((i = revised.IndexOf(removal)) > -1)
                {
                    revised = revised.Remove(i, removal.Length);
                }
            }
            return revised;
        }

        public static string[] Insert(string[] input, string insert, int index)
        {
            ArrayList list = new ArrayList(input);
            list.Insert(index, insert);
            return (string[])list.ToArray(typeof(string));
        }

        public static string Join(string separator, params string[] strings)
        {
            StringBuilder builder = new StringBuilder();
            foreach (string s in strings)
            {
                if (IsBlank(s)) 
                    continue;
                if (builder.Length > 0) 
                    builder.Append(separator);

                builder.Append(s.ToString());
            }
            return builder.ToString();
        }

        public static string RemoveNulls(string s)
        {
            return NullStringRegex.Replace(s, string.Empty).TrimStart();
        }

        public static string StripQuotes(string filename)
        {
            return filename == null ? null : filename.Trim('"');
        }

        public static string SurroundInQuotesIfContainsSpace(string value, string quote)
        {
            if (!StringTools.IsBlank(value) && value.IndexOf(' ') >= 0)
                return string.Format(@"{0}{1}{0}", quote, value);

            return value;
        }

        public static string SurroundInQuotesIfContainsSpace(string value)
        {
            return SurroundInQuotesIfContainsSpace(value, "\"");
        }

        public static bool IsNumber(char c)
        {
            return (c >= '0' && c <= '9');
        }

        public static bool IsHexNumber(char c)
        {
            return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f');
        }

        public static bool IsDecimalNumber(string str)
        {
            foreach (char c in str)
            {
                if (!IsNumber(c))
                    return false;
            }
            return true;
        }

        public static byte HexToNibble(char c)
        {
            byte b = 0;
            if (c >= '0' && c <= '9')
                b = (byte)(c - '0');
            else if (c >= 'a' && c <= 'f')
                b = (byte)(10 + (c - 'a'));
            else if (c >= 'A' && c <= 'F')
                b = (byte)(10 + (c - 'A'));
            return b;
        }

        public static Int32 HexToInt32(string str)
        {
            if (str.StartsWith("0x"))
                str = str.Substring(2);

            Int32 value = 0;
            foreach (char c in str)
            {
                if (value == 0 && c == '0')
                    continue;

                byte b = 0;
                if (c >= '0' && c <= '9')
                    b = (byte)(c - '0');
                else if (c >= 'a' && c <= 'f')
                    b = (byte)(10 + (c - 'a'));
                else if (c >= 'A' && c <= 'F')
                    b = (byte)(10 + (c - 'A'));
                value = value << 4;
                value = value | b;
            }
            return value;
        }

        public static Int64 HexToInt64(string str)
        {
            if (str.StartsWith("0x"))
                str = str.Substring(2);

            Int64 value = 0;
            foreach (char c in str)
            {
                if (value == 0 && c == '0')
                    continue;

                byte b = 0;
                if (c >= '0' && c <= '9')
                    b = (byte)(c - '0');
                else if (c >= 'a' && c <= 'f')
                    b = (byte)(10 + (c - 'a'));
                else if (c >= 'A' && c <= 'F')
                    b = (byte)(10 + (c - 'A'));
                value = value << 4;
                value = value | b;
            }
            return value;
        }

        public static char NibbleToHex(byte b)
        {
            if (b >= 0 && b <= 9)
                return (char)('0' + b);
            else if (b >= 10 && b <= 15)
                return (char)('A' - 10 + b);
            else
                return '?';
        }

        /// <summary>
        /// Checks if the string is of the form header('0x') + hexadecimal number
        /// </summary>
        /// <param name="str"></param>
        /// <returns>True if string contains a hexadecimal number</returns>
        public static bool IsHexadecimalNumber(string str, bool header)
        {
            int i = 0;
            foreach (char c in str)
            {
                if (header)
                {
                    if (i == 0)
                    {
                        if (c != '0')
                            return false;
                        continue;
                    }
                    else if (i == 1)
                    {
                        if (c != 'x')
                            return false;
                        continue;
                    }
                }
                if (!IsHexNumber(c))
                {
                    return false;
                }

                ++i;
            }
            return true;
        }

        public static string MultiToSingle(string s, char charToReplace, char replacementChar)
        {
            StringBuilder sb = new StringBuilder(s.Length);
            bool skipped = false;
            foreach (char c in s)
            {
                if (c == charToReplace)
                {
                    skipped = true;
                }
                else
                {
                    if (skipped)
                    {
                        skipped = false;
                        sb.Append(replacementChar);
                    }

                    sb.Append(c);
                }
            }

            if (skipped)
                sb.Append(replacementChar);

            return sb.ToString();
        }

        public static string SingleToMulti(string s, char charToReplace, char replacementChar, int replacementCount)
        {
            StringBuilder sb = new StringBuilder(s.Length);
            foreach (char c in s)
            {
                if (c == charToReplace)
                {
                    sb.Append(replacementChar, replacementCount);
                }
                else
                {
                    sb.Append(c);
                }
            }
            return sb.ToString();
        }

		/// <summary>
		/// Concatenate all the strings in array <paramref name="strings"/> and seperate them with <paramref name="delimiter"/>
		/// </summary>
		/// <param name="strings">The strings to concatenate.</param>
		/// <param name="delimiter">The delimiter to seperate the concatenated strings.</param>
		/// <returns>The concatenation of <paramref name="strings"/> seperated with <paramref name="delimiter"/>.</returns>
		public static string ConcatWith(string[] strings, string delimiter)
		{
			if (strings.Length == 0)
				return string.Empty;

		    int totalLength = 0;
            foreach(string s in strings)
                totalLength += s.Length;

            StringBuilder sb = new StringBuilder(totalLength + (strings.Length * delimiter.Length));
			string str = strings[0];
		    sb.Append(str);
            for (int i = 1; i < strings.Length; ++i)
            {
                sb.Append(delimiter);
                sb.Append(strings[i]);
            }
			return sb.ToString();
		}

		/// <summary>
		/// Return a left part of <paramref name="inString"/> with length <paramref name="inLength"/>
		/// </summary>
		/// <param name="inString">The full string.</param>
		/// <param name="inLength">Length of the left part to return.</param>
		/// <returns>The left part with a certain length of a string</returns>
		public static string Left(string inString, int inLength)
		{
            if (inLength >= inString.Length)
                return inString;
            
            if (inLength == 0)
                return string.Empty;

            return inString.Substring(0, inLength);
		}

		/// <summary>
		/// Return a right part of <paramref name="inString"/> with length <paramref name="inLength"/>
		/// </summary>
		/// <param name="inString">The full string.</param>
		/// <param name="inLength">Length of the right part to return.</param>
		/// <returns>The right part with a certain length of a string</returns>
		public static string Right(string inString, int inLength)
		{
            if (inLength == 0)
                return string.Empty;

            if (inLength >= inString.Length)
                inLength = inString.Length;

			return inString.Substring(inString.Length - inLength, inLength);
		}

		/// <summary>
		/// Returns a sub part of a string
		/// </summary>
		/// <param name="inString">The full string.</param>
		/// <param name="inLeftPos">The left pos of the subpart.</param>
		/// <param name="inRightPos">The right pos of the subpart.</param>
		/// <returns>The subpart of the string</returns>
		public static string Mid(string inString, int inLeftPos, int inRightPos)
		{
            inLeftPos = Math.Clamp(inLeftPos, 0, inString.Length);
            inRightPos = Math.Clamp(inRightPos, 0, inString.Length);
            if (inLeftPos >= inRightPos)
                return string.Empty;
            
			return inString.Substring(inLeftPos, inRightPos - inLeftPos);
		}

		/// <summary>
		/// Removes the specified characters from a string and returns a string with those characters removed.
		/// </summary>
		/// <param name="inString">The string to remove the characters from.</param>
		/// <param name="inRemove">The characters to filter.</param>
		/// <returns>A copy of a string without the specified characters</returns>
		public static string Remove(string inString, char[] inRemove)
		{
			string s = inString;
			foreach (char c in inRemove)
			{
				int i = s.IndexOf(c);
				if (i >= 0)
					s = s.Remove(i, 1);
			}
			return s;
		}

		/// <summary>
		/// Replaces the specified place in <paramref name="inString"/> with <paramref name="inReplacementString"/>.
		/// </summary>
		/// <param name="inString">The original string.</param>
		/// <param name="inStartIndex">Start index of the subpart.</param>
		/// <param name="inLength">Length of the subpart.</param>
		/// <param name="inReplacementString">The replacement string.</param>
		/// <returns>A copy of inString where the subpart <paramref name="inStartIndex"/> to <paramref name="inLength"/>
		/// is replaced with <paramref name="inReplacementString"/></returns>
		public static string Replace(string inString, int inStartIndex, int inLength, string inReplacementString)
		{
			int endIndex = inStartIndex + inLength;
			string str = inString.Substring(0, inStartIndex) + inReplacementString + inString.Substring(endIndex, inString.Length - endIndex);
			return str;
		}

		/// <summary>
		/// Left of the first index where <paramref name="c"/> is found
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">Return everything to the left of this character.</param>
		/// <returns>String to the left of c, or the entire string.</returns>
		public static string LeftOf(string src, char c)
		{
			string ret = src;
			int idx = src.IndexOf(c);
			if (idx != -1)
			{
				ret = src.Substring(0, idx);
			}
			return ret;
		}

		/// <summary>
		/// Left of the first index where <paramref name="c"/> is found
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">Return everything to the left of this character.</param>
		/// <returns>String to the left of c, or the entire string.</returns>
		public static string LeftOf(string src, string c)
		{
			string ret = src;
			int idx = src.IndexOf(c);
			if (idx != -1)
			{
				ret = src.Substring(0, idx);
			}
			return ret;
		}

		/// <summary>
		/// Left of the nth occurrence where <paramref name="c"/> is found
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">Return everything to the left n'th occurrence of this character.</param>
		/// <param name="n">The nth occurrence of <paramref name="c"/>.</param>
		/// <returns>String to the left of c, or the entire string if not found or n is 0.</returns>
		public static string LeftOf(string src, char c, int n)
		{
			string ret = src;
			int idx = -1;
			while (n > 0)
			{
				idx = src.IndexOf(c, idx + 1);
				if (idx == -1)
				{
					break;
				}
				--n;
			}
			if (idx != -1)
			{
				ret = src.Substring(0, idx);
			}
			return ret;
		}

		/// <summary>
		/// Left of the nth occurrence where <paramref name="c"/> is found
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">Return everything to the left n'th occurrence of this string.</param>
		/// <param name="n">The nth occurrence of <paramref name="c"/>.</param>
		/// <returns>String to the left of <paramref name="c"/>, or the entire string if not found or <paramref name="n"/> is 0.</returns>
		public static string LeftOf(string src, string c, int n)
		{
			string ret = src;
			int idx = -1;
			while (n > 0)
			{
				idx = src.IndexOf(c, idx + 1);
				if (idx == -1)
				{
					break;
				}
				--n;
			}
			if (idx != -1)
			{
				ret = src.Substring(0, idx);
			}
			return ret;
		}

		/// <summary>
		/// Right of the first occurrence of c
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">The search char.</param>
		/// <returns>Returns everything to the right of c, or an empty string if c is not found.</returns>
		public static string RightOf(string src, char c)
		{
			int idx = src.IndexOf(c);
			if (idx != -1)
				return src.Substring(idx + 1);
			return string.Empty;
		}

		/// <summary>
		/// Right of the first occurrence of c
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">The search char.</param>
		/// <returns>Returns everything to the right of c, or an empty string if c is not found.</returns>
		public static string RightOf(string src, string c)
		{
			int idx = src.IndexOf(c);
			if (idx != -1)
				return src.Substring(idx + 1);
			return string.Empty;
		}

		/// <summary>
		/// Right of the n'th occurrence of c
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">The search char.</param>
		/// <param name="n">The occurrence.</param>
		/// <returns>Returns everything to the right of c, or an empty string if c is not found.</returns>
		public static string RightOf(string src, char c, int n)
		{
			string ret = string.Empty;
			int idx = -1;
			while (n > 0)
			{
				idx = src.IndexOf(c, idx + 1);
				if (idx == -1)
				{
					break;
				}
				--n;
			}

			if (idx != -1)
			{
				ret = src.Substring(idx + 1);
			}

			return ret;
		}

		/// <summary>
		/// Returns everything to the left of the rightmost char c.
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">The search char.</param>
		/// <returns>Everything to the left of the rightmost char c, or the entire string.</returns>
		public static string LeftOfRightmostOf(string src, char c)
		{
			string ret = src;
			int idx = src.LastIndexOf(c);
			if (idx != -1)
			{
				ret = src.Substring(0, idx);
			}
			return ret;
		}

		/// <summary>
		/// Returns everything to the right of the rightmost char c.
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="c">The search char.</param>
		/// <returns>Returns everything to the right of the rightmost search char, or an empty string.</returns>
		public static string RightOfRightmostOf(string src, char c)
		{
			string ret = string.Empty;
			int idx = src.LastIndexOf(c);
			if (idx != -1)
			{
				ret = src.Substring(idx + 1);
			}
			return ret;
		}

		/// <summary>
		/// Returns everything between the start and end chars, exclusive.
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="start">The first char to find.</param>
		/// <param name="end">The end char to find.</param>
		/// <returns>The string between the start and stop chars, or an empty string if not found.</returns>
		public static string[] Between(string src, char start, char end)
		{
            int numStartEndOccurences = 0;
            int startEndOpen = 0;
            foreach (char c in src)
            {
                if (c == start)
                {
                    ++startEndOpen;
                }

                if (c == end)
                {
                    --startEndOpen;
                    if (startEndOpen == 0)
                    {
                        numStartEndOccurences++;
                    }
                }
            }
            string[] ret = new string[numStartEndOccurences];

            numStartEndOccurences = 0;

            int index = 0;
            int startIndex = 0;
            foreach (char c in src)
            {
                if (c == start)
                {
                    if (startEndOpen==0)
                        startIndex = index + 1;

                    ++startEndOpen;
                }
                if (c == end)
                {
                    --startEndOpen;
                    if (startEndOpen == 0)
                    {
                        ret[numStartEndOccurences] = src.Substring(startIndex, index - startIndex);
                        numStartEndOccurences++;
                    }
                }
                ++index;
            }

			return ret;
		}

        public static string[] Between(string src, string start, string end)
        {
            int numStartEndOccurences = 0;
            int startCharIndex = 0;
            int endCharIndex = 0;
            int startEndOpen = 0;
            foreach (char c in src)
            {
                if (c == start[startCharIndex])
                {
                    startCharIndex++;
                    if (startCharIndex == start.Length)
                    {
                        startCharIndex = 0;
                        ++startEndOpen;
                    }
                }
                else
                {
                    startCharIndex = 0;
                }

                if (c == end[endCharIndex])
                {
                    ++endCharIndex;
                    if (endCharIndex == end.Length)
                    {
                        endCharIndex = 0;
                        --startEndOpen;
                        if (startEndOpen == 0)
                        {
                            numStartEndOccurences++;
                        }
                    }
                }
                else
                {
                    endCharIndex = 0;
                }
            }
            string[] ret = new string[numStartEndOccurences];

            numStartEndOccurences = 0;

            int index = 0;
            int startIndex = 0;

            startCharIndex = 0;
            endCharIndex = 0;
            startEndOpen = 0;
            foreach (char c in src)
            {
                if (c == start[startCharIndex])
                {
                    startCharIndex++;
                    if (startCharIndex == start.Length)
                    {
                        startCharIndex = 0;
                        if (startEndOpen == 0)
                            startIndex = index + 1;
                        ++startEndOpen;
                    }
                }
                else
                {
                    startCharIndex = 0;
                }

                if (c == end[endCharIndex])
                {
                    ++endCharIndex;
                    if (endCharIndex == end.Length)
                    {
                        endCharIndex = 0;
                        --startEndOpen;
                        if (startEndOpen == 0)
                        {
                            ret[numStartEndOccurences] = src.Substring(startIndex, index - startIndex - (end.Length - 1));
                            numStartEndOccurences++;
                        }
                    }
                }
                else
                {
                    endCharIndex = 0;
                }
                ++index;
            }

            return ret;
        }

		/// <summary>
		/// Returns the number of occurrences of "find".
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <param name="find">The search char.</param>
		/// <returns>The # of times the char occurs in the search string.</returns>
		public static int Count(string src, char find)
		{
			int ret = 0;
			foreach (char s in src)
			{
				if (s == find)
				{
					++ret;
				}
			}
			return ret;
		}

		/// <summary>
		/// Returns the rightmost char in src.
		/// </summary>
		/// <param name="src">The source string.</param>
		/// <returns>The rightmost char, or '\0' if the source has zero length.</returns>
		public static char Rightmost(string src)
		{
			char c = '\0';
			if (src.Length > 0)
			{
				c = src[src.Length - 1];
			}
			return c;
		}

        /// <summary>
        /// Remove characters from 'src' starting at 'index' until the 'stop' character is encountered
        /// </summary>
        /// <param name="src">The source string.</param>
        /// <returns>The rightmost char, or '\0' if the source has zero length.</returns>
        public static string RemoveUntil(string src, int index, char stop)
        {
            int i = index;
            while (src[i] != stop) { ++i; }
            src = src.Remove(index, i);
            return src;
        }

        /// <summary>
        /// Filter the lines with StartsWith(key) and hold them in outLines
        /// </summary>
        /// <param name="text">The source text.</param>
        /// <param name="format">The format to use for scanning/parsing the source text.</param>
        /// <returns>The values parsed from the text.</returns>
        public static bool FilterTextAsMultiLineUsingStartsWith(string text, string key, string newLine, bool stripKey, out List<string> filteredLines)
        {
            if (String.IsNullOrEmpty(text))
            {
                filteredLines = new List<string>();
                return false;
            }

            string[] l = text.Split(new string[] { newLine }, StringSplitOptions.RemoveEmptyEntries);
            if (l.Length == 0)
            {
                filteredLines = new List<string>();
                return false;
            }

            filteredLines = new List<string>();
            foreach (string s in l)
            {
                if (s.StartsWith(key))
                {
                    string f = stripKey ? s.Substring(key.Length) : s;
                    if (!String.IsNullOrEmpty(f))
                        filteredLines.Add(f);
                }
            }
            return true;
        }


        /// <summary>
        /// Scans the string 'text' by using fieldSpecification as the format description
        /// </summary>
        /// <param name="text">The source text.</param>
        /// <param name="format">The format to use for scanning/parsing the source text.</param>
        /// <returns>The values parsed from the text.</returns>
        public static bool Scan(string text, string format, out object[] targets)
        {
            return StringScan.Scan(text, format, out targets);
        }

        /// <summary>
        /// Scans the string 'text' by using fieldSpecification as the format description
        /// </summary>
        /// <param name="text">The source text.</param>
        /// <param name="format">The format to use for scanning/parsing the source text.</param>
        /// <param name="targets">The values parsed from the text, they should match the number of fields specified in 'format'.</param>
        public static bool Scan(string text, string format, params object[] targets)
        {
            return StringScan.Scan(text, format, targets);
        }
	}
}

