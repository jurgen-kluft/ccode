using System;

namespace FileDirectoryPath
{
    public sealed class DirPathRelative : DirPath
    {
        public DirPathRelative(string path)
            : base(path, false)
        {
        }

        //
        //  Absolute/Relative path conversion
        //
        public DirPathAbsolute GetAbsolutePathFrom(DirPathAbsolute path)
        {
            if (path == null)
            {
                throw new ArgumentNullException();
            }
            if (PathHelper.IsEmpty(this) || PathHelper.IsEmpty(path))
            {
                throw new ArgumentException("Cannot compute an absolute path from an empty path.");
            }
            return new DirPathAbsolute(BasePath.GetAbsolutePathFrom(path, this));
        }

        public bool CanGetAbsolutePathFrom(DirPathAbsolute path)
        {
            try
            {
                this.GetAbsolutePathFrom(path);
                return true;
            }
            catch { }
            return false;
        }

        //
        //  Path Browsing facilities
        //
        public new DirPathRelative ParentDirectoryPath
        {
            get
            {
                string parentPath = InternalStringHelper.GetParentDirectory(this.Path);
                return new DirPathRelative(parentPath);
            }
        }

        public FilePathRelative GetBrotherFileWithName(string fileName)
        {
            if (fileName == null) { throw new ArgumentNullException("filename"); }
            if (fileName.Length == 0) { throw new ArgumentException("Can't get brother of an empty file", "filename"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get brother of an empty file"); }
            return this.ParentDirectoryPath.GetChildFileWithName(fileName);
        }

        public DirPathRelative GetBrotherDirectoryWithName(string fileName)
        {
            if (fileName == null) { throw new ArgumentNullException("filename"); }
            if (fileName.Length == 0) { throw new ArgumentException("Can't get brother of an empty file", "filename"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get brother of an empty file"); }
            return this.ParentDirectoryPath.GetChildDirectoryWithName(fileName);
        }

        public FilePathRelative GetChildFileWithName(string fileName)
        {
            if (fileName == null) { throw new ArgumentNullException("filename"); }
            if (fileName.Length == 0) { throw new ArgumentException("Empty filename not accepted", "filename"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get a child file name from an empty path"); }
            return new FilePathRelative(this.Path + System.IO.Path.DirectorySeparatorChar + fileName);
        }

        public DirPathRelative GetChildDirectoryWithName(string directoryName)
        {
            if (directoryName == null) { throw new ArgumentNullException("directoryName"); }
            if (directoryName.Length == 0) { throw new ArgumentException("Empty directoryName not accepted", "directoryName"); }
            if (this.IsEmpty) { throw new InvalidOperationException("Can't get a child directory name from an empty path"); }
            return new DirPathRelative(this.Path + System.IO.Path.DirectorySeparatorChar + directoryName);
        }


        //
        //  Empty DirectoryPathRelative
        //
        private DirPathRelative() : base() { }
        private static DirPathRelative s_Empty = new DirPathRelative();
        public static DirPathRelative Empty { get { return s_Empty; } }

        public override bool IsAbsolutePath { get { return false; } }
        public override bool IsRelativePath { get { return true; } }
        public override bool IsNetworkPath { get { return false; } }
    }
}
