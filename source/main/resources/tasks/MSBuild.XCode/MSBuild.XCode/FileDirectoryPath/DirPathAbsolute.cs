using System.IO;
using System.Collections.Generic;
using System;

namespace FileDirectoryPath
{
    public sealed class DirPathAbsolute : DirPath
    {
        public DirPathAbsolute(string path)
            : base(path, true)
        {
        }


        //
        //  Absolute/Relative path conversion
        //
        public DirPathRelative GetPathRelativeFrom(DirPathAbsolute path)
        {
            if (path == null)
            {
                throw new ArgumentNullException();
            }
            if (PathHelper.IsEmpty(this) || PathHelper.IsEmpty(path))
            {
                throw new ArgumentException("Cannot compute a relative path from an empty path.");
            }
            return new DirPathRelative(BasePath.GetPathRelative(path, this));
        }


        public bool CanGetPathRelativeFrom(DirPathAbsolute path)
        {
            try
            {
                this.GetPathRelativeFrom(path);
                return true;
            }
            catch { }
            return false;
        }


        //
        //  Path Browsing facilities
        //
        public new DirPathAbsolute ParentDirectoryPath
        {
            get
            {
                string parentPath = InternalStringHelper.GetParentDirectory(this.Path);
                return new DirPathAbsolute(parentPath);
            }
        }

        public FilePathAbsolute GetBrotherFileWithName(string fileName)
        {
            if (fileName == null) { throw new ArgumentNullException("filename"); }
            if (fileName.Length == 0) { throw new ArgumentException("Can't get brother of an empty file", "filename"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get brother of an empty file"); }
            return this.ParentDirectoryPath.GetChildFileWithName(fileName);
        }

        public DirPathAbsolute GetBrotherDirectoryWithName(string fileName)
        {
            if (fileName == null) { throw new ArgumentNullException("filename"); }
            if (fileName.Length == 0) { throw new ArgumentException("Can't get brother of an empty file", "filename"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get brother of an empty file"); }
            return this.ParentDirectoryPath.GetChildDirectoryWithName(fileName);
        }

        public FilePathAbsolute GetChildFileWithName(string fileName)
        {
            if (fileName == null) { throw new ArgumentNullException("filename"); }
            if (fileName.Length == 0) { throw new ArgumentException("Empty filename not accepted", "filename"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get a child file name from an empty path"); }
            return new FilePathAbsolute(this.Path + System.IO.Path.DirectorySeparatorChar + fileName);
        }

        public DirPathAbsolute GetChildDirectoryWithName(string directoryName)
        {
            if (directoryName == null) { throw new ArgumentNullException("directoryName"); }
            if (directoryName.Length == 0) { throw new ArgumentException("Empty directoryName not accepted", "directoryName"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get a child directory name from an empty path"); }
            if (IsNetworkPath)
            {

            }
            return new DirPathAbsolute(this.Path + System.IO.Path.DirectorySeparatorChar + directoryName);
        }

        public bool IsChildDirectoryOf(DirPathAbsolute parentDir)
        {
            if (parentDir == null) { throw new ArgumentNullException("parentDir"); }
            if (parentDir.IsEmpty) { throw new ArgumentException("Empty parentDir not accepted", "parentDir"); }
            string parentPathUpperCase = parentDir.Path.ToUpper();
            string thisPathUpperCase = this.Path.ToUpper();
            return thisPathUpperCase.Contains(parentPathUpperCase);
        }



        //
        //  Operations that requires physical access
        //
        public string Drive { get { return this.DriveProtected; } }

        public bool Exists
        {
            get
            {
                return Directory.Exists(this.Path);
            }
        }

        public DirectoryInfo DirectoryInfo
        {
            get
            {
                if (!this.Exists)
                {
                    throw new FileNotFoundException(this.Path);
                }
                return new DirectoryInfo(this.Path);
            }
        }

        public List<FilePathAbsolute> ChildrenFilesPath
        {
            get
            {
                DirectoryInfo directoryInfo = this.DirectoryInfo;
                FileInfo[] filesInfos = directoryInfo.GetFiles();
                List<FilePathAbsolute> childrenFilesPath = new List<FilePathAbsolute>();
                foreach (FileInfo fileInfo in filesInfos)
                {
                    childrenFilesPath.Add(new FilePathAbsolute(fileInfo.FullName));
                }
                return childrenFilesPath;
            }
        }
        public List<DirPathAbsolute> ChildrenDirectoriesPath
        {
            get
            {
                DirectoryInfo directoryInfo = this.DirectoryInfo;
                DirectoryInfo[] directoriesInfos = directoryInfo.GetDirectories();
                List<DirPathAbsolute> childrenDirectoriesPath = new List<DirPathAbsolute>();
                foreach (DirectoryInfo childDirectoryInfo in directoriesInfos)
                {
                    childrenDirectoriesPath.Add(new DirPathAbsolute(childDirectoryInfo.FullName));
                }
                return childrenDirectoriesPath;
            }
        }


        //
        //  Empty DirectoryPathAbsolute
        //
        private DirPathAbsolute() : base() { }
        private static DirPathAbsolute s_Empty = new DirPathAbsolute();
        public static DirPathAbsolute Empty { get { return s_Empty; } }


        public override bool IsAbsolutePath { get { return true; } }
        public override bool IsRelativePath { get { return false; } }
        public override bool IsNetworkPath { get { return Path.Length > 2 && Path[0] == '\\' && Path[1] == '\\'; } }
    }
}
