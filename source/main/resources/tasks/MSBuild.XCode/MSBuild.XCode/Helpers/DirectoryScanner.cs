using System;
using System.Globalization;
using System.Collections.Generic;
using System.IO;

namespace MSBuild.Cod.Helpers
{
    public class DirectoryScanner
    {
        #region Fields

        private bool mScanSubDirs;

        private xDirname mBasePath;

        private Queue<string> mTDirectories;
        private List<xFilename> mFilenames;

        #endregion
        #region FilterEvent

        public class FilterEvent
        {
            #region Fields

            public enum EType
            {
                NONE,
                FILE,
                FOLDER
            }

            private EType mType = EType.FILE;
            private string mName = string.Empty;

            #endregion
            #region Constructors

            public FilterEvent()
            {
                mType = EType.NONE;
                mName = string.Empty;
            }

            public FilterEvent(xFilename f)
            {
                mType = EType.FILE;
                mName = f;
            }

            public FilterEvent(xDirname d)
            {
                mType = EType.FOLDER;
                mName = d;
            }

            #endregion
            #region Properties

            public bool isFile
            {
                get
                {
                    return mType == EType.FILE;
                }
            }

            public bool isFolder
            {
                get
                {
                    return mType == EType.FOLDER;
                }
            }

            public string name
            {
                get
                {
                    return mName;
                }
            }

            public xFilename filename
            {
                set
                {
                    mType = EType.FILE;
                    mName = value;
                }
            }

            public xDirname dir
            {
                set
                {
                    mType = EType.FOLDER;
                    mName = value;
                }
            }

            #endregion
        }

        #endregion
        #region FilterDelegate

        public delegate bool FilterDelegate(FilterEvent e);
        public static bool EmptyFilterDelegate(FilterEvent e) { return false; }
        public static bool NoSvnFilterDelegate(FilterEvent e) 
        {
            if (e.isFolder)
            {
                if (e.name.EndsWith(".svn", true, CultureInfo.InvariantCulture))
                    return true;
            }
            return false; 
        }

        #endregion
        #region Constructor

        public DirectoryScanner(xDirname basePath)
        {
            mBasePath = basePath;

            mTDirectories = new Queue<string>();
            mFilenames = new List<xFilename>();
        }

        #endregion
        #region Properties

        public xDirname basepath
        {
            get
            {
                return mBasePath;
            }
            set
            {
                mBasePath = value;
            }
        }

        public bool scanSubDirs
        {
            get
            {
                return mScanSubDirs;
            }
            set
            {
                mScanSubDirs = value;
            }
        }

        public List<xFilename> filenames
        {
            get 
            { 
                return mFilenames; 
            }
        }

        #endregion
        #region Private Methods

        private bool scan(string dirName, string extension, FilterDelegate filter)
        {
            try
            {
                FilterEvent fe = new FilterEvent();

                string[] files = Directory.GetFiles(dirName, extension);
                foreach (string f in files)
                {
                    xFilename filename = new xFilename(f);
                    filename = filename.MakeRelative(mBasePath);
                    fe.filename = filename;

                    if (!filter(fe))
                        mFilenames.Add(filename);
                }

                string[] dirs = Directory.GetDirectories(dirName);
                foreach (string d in dirs)
                {
                    xDirname dir = new xDirname(d);
                    dir = dir.MakeRelative(mBasePath);
                    fe.dir = dir;

                    if (!filter(fe))
                        mTDirectories.Enqueue(d);
                }
            }
            catch (Exception e)
            {
                Console.WriteLine(e.Message);
                return false;
            }

            return true;
        }

        #endregion
        #region Public Methods

        public void clear()
        {
            mFilenames.Clear();
        }

        public bool collect(xDirname subDirName, string extension, FilterDelegate filter)
        {
            try
            {
                DirectoryInfo dirInfo = new DirectoryInfo(mBasePath + subDirName);
                if (!dirInfo.Exists)
                    return true;

                mTDirectories.Clear();

                // Push main directory on first worker thread
                mTDirectories.Enqueue(dirInfo.FullName);
                while (mTDirectories.Count > 0)
                {
                    // Any thread free ?
                    // Any directories to do?
                    string dirName = string.Empty;
                    if (mTDirectories.Count > 0)
                    {
                        dirName = mTDirectories.Dequeue();

                        {
                            string directory = dirName;
                            scan(directory, extension, filter);
                        }
                    }
                }
            }
            catch (Exception e)
            {
                Console.WriteLine(e.Message);
                return false;
            }

            return true;
        }

        #endregion
    }
}
