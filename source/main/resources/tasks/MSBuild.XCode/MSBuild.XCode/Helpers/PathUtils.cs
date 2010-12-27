using System;
using System.IO;
using System.Text.RegularExpressions;
using System.Collections;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public static class PathUtil
    {
        // getFiles, understand globbing (e.g: C:\data\*\*.dat
        public static List<string> getFiles(string path)
        {
            Match m = Regex.Match(path, @"(.*)\\(.*)");
            string FileMask = m.Groups[2].Value;

            List<string> ret = new List<string>();
            Stack<string> dirs = new Stack<string>();
            dirs.Push(m.Groups[1].ToString());
            while (dirs.Count > 0)
            {
                string dir = dirs.Pop();
                if (dir.IndexOf('*') < 0)
                {
                    ret.AddRange(Directory.GetFiles(dir, FileMask));
                }
                else
                {
                    m = Regex.Match(dir, @"([^*]*)\\([^\\]*\*[^\\]*)\\?(.*)");
                    string[] ds = Directory.GetDirectories(m.Groups[1].Value, m.Groups[2].Value, m.Groups[2].Value == "**" ? SearchOption.AllDirectories : SearchOption.TopDirectoryOnly);
                    foreach (string d in ds)
                        dirs.Push(d + '\\' + m.Groups[3].Value);
                }
            }

            return ret;
        }

        public static string EnsureDir(string dir)
        {
            if (!dir.EndsWith("\\"))
                return dir + "\\";
            return dir;
        }


        /// <summary>
        /// Creates a relative path from one file
        /// or folder to another.
        /// </summary>
        /// <param name="fromDirectory">
        /// Contains the directory that defines the
        /// start of the relative path.
        /// </param>
        /// <param name="toPath">
        /// Contains the path that defines the
        /// endpoint of the relative path.
        /// </param>
        /// <returns>
        /// The relative path from the start
        /// directory to the end path.
        /// </returns>
        /// <exception cref="ArgumentNullException"></exception>
        public static string RelativePathTo(string fromDirectory, string toPath)
        {
            if (fromDirectory == null)
                throw new ArgumentNullException("fromDirectory");

            if (toPath == null)
                throw new ArgumentNullException("toPath");

            bool isRooted = Path.IsPathRooted(fromDirectory)
                && Path.IsPathRooted(toPath);

            if (isRooted)
            {
                bool isDifferentRoot = string.Compare(Path.GetPathRoot(fromDirectory), Path.GetPathRoot(toPath), true) != 0;
                if (isDifferentRoot)
                    return toPath;
            }

            StringCollection relativePath = new StringCollection();
            string[] fromDirectories = fromDirectory.Split(Path.DirectorySeparatorChar);

            string[] toDirectories = toPath.Split(Path.DirectorySeparatorChar);

            int length = Math.Min(fromDirectories.Length,toDirectories.Length);

            int lastCommonRoot = -1;

            // find common root
            for (int x = 0; x < length; x++)
            {
                if (string.Compare(fromDirectories[x],toDirectories[x], true) != 0)
                    break;
                lastCommonRoot = x;
            }
            if (lastCommonRoot == -1)
                return toPath;

            // add relative folders in from path
            for (int x = lastCommonRoot + 1; x < fromDirectories.Length; x++)
                if (fromDirectories[x].Length > 0)
                    relativePath.Add("..");

            // add to folders to path
            for (int x = lastCommonRoot + 1; x < toDirectories.Length; x++)
                relativePath.Add(toDirectories[x]);

            // create relative path
            string[] relativeParts = new string[relativePath.Count];
            relativePath.CopyTo(relativeParts, 0);

            string newPath = string.Join(Path.DirectorySeparatorChar.ToString(),relativeParts);
            return newPath;
        }
    }

}
