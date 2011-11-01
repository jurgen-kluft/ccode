using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode.Helpers
{
    public class TextFile
    {
        private FileStream mStream = null;
        private StreamWriter mWriter = null;

        public int Indent { get; set; }

        public bool Open(string _filename)
        {
            mStream = new FileStream(_filename, FileMode.Create, FileAccess.Write);
            mWriter = new StreamWriter(mStream);
            return true;
        }

        public void Close()
        {
            if (mWriter != null)
            {
                mWriter.Close();
                mWriter = null;
            }
            if (mStream != null)
            {
                mStream.Close();
                mStream = null;
            }
        }

        private static string IndentStr(int _indent)
        {
            string indent = string.Empty;
            for (int i = 0; i < _indent; i++)
                indent += "\t";
            return indent;
        }

        public void WriteLine(int indent, string line)
        {
            mWriter.WriteLine("{0}{1}", IndentStr(indent), line);
        }

        public void WriteLine(string line)
        {
            mWriter.WriteLine("{0}{1}", IndentStr(Indent), line);
        }

        public void WriteLine(int indent, string line, params object[] args)
        {
            mWriter.Write(IndentStr(indent));
            mWriter.WriteLine(line, args);
        }

        public void WriteLine(string line, params object[] args)
        {
            mWriter.Write(IndentStr(Indent));
            mWriter.WriteLine(line, args);
        }

    }
}
