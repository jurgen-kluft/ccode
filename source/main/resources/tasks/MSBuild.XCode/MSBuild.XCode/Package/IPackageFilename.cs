using System;

namespace MSBuild.XCode
{
    public interface IPackageFilename
    {
        string Name { get; set; }
        ComparableVersion Version { get; set; }
        DateTime DateTime { get; set; }
        string Branch { get; set; }
        string Platform { get; set; }
        string Extension { get; set; }

        string Filename { get; }
        string FilenameWithoutExtension { get; }
    }

    public class PackageFilename : IPackageFilename
    {
        private string mName;
        private string mBranch;
        private string mPlatform;
        private string mExtension;

        public PackageFilename()
        {
            Name = string.Empty;
            Version = new ComparableVersion("1.0.0");
            DateTime = System.DateTime.Now;
            Branch = "default";
            Platform = "Win32";
            Extension = ".zip";
        }

        public PackageFilename(IPackageFilename filename)
        {
            Name = filename.Name;
            Version = new ComparableVersion(filename.Version);
            DateTime = new DateTime(filename.DateTime.Ticks);
            Branch = filename.Branch;
            Platform = filename.Platform;
            Extension = filename.Extension;
        }
        public PackageFilename(string filename)
        {
            if (filename.EndsWith(".zip"))
                filename = System.IO.Path.GetFileNameWithoutExtension(filename);

            string[] parts = filename.Split(new char[] { '+' }, StringSplitOptions.RemoveEmptyEntries);
            Name = parts[0];

            // Here split the version and date time
            // Find the 
            string[] dparts = parts[1].Split(new char[] { '.' }, StringSplitOptions.RemoveEmptyEntries);
            Version = dparts.Length>2 ? new ComparableVersion(String.Format("{0}.{1}.{2}", dparts[0], dparts[1], dparts[2])) : new ComparableVersion("1.0.0");
            DateTime = dparts.Length>8 ? System.DateTime.Parse(String.Format("{0}-{1}-{2} {3}:{4}:{5}", dparts[3], dparts[4], dparts[5], dparts[6], dparts[7], dparts[8])) : System.DateTime.Now;
            Branch = parts.Length>2 ? parts[2] : "default";
            Platform = parts.Length>3 ? parts[3] : "Win32";
            Extension = ".zip";
        }
        public PackageFilename(string name, ComparableVersion version, string branch, string platform)
        {
            Name = name;
            Version = version;
            DateTime = System.DateTime.Now;
            Branch = branch;
            Platform = platform;
            Extension = ".zip";
        }
        public PackageFilename(string name, ComparableVersion version, DateTime dateTime, string branch, string platform)
            : this(name, version, branch, platform)
        {
            DateTime = dateTime;
        }

        public string Name { get { return mName; } set { mName = value; } }
        public ComparableVersion Version { get; set; }
        public DateTime DateTime { get; set; }
        public string Branch { get { return mBranch; } set { mBranch = value; } }
        public string Platform { get { return mPlatform; } set { mPlatform = value; } }
        public string Extension { get { return mExtension; } set { mExtension = value; } }

        public string VersionAndDateTime
        {
            get
            {
                DateTime dt = DateTime;
                string v = String.Format("{0}.{1:yyyy.M.d.H.m.s}", Version.ToString(), dt);
                return v;
            }
        }

        public string VersionAndDateTimeComparable
        {
            get
            {
                int major = Version.GetMajor();
                int minor = Version.GetMinor();
                int build = Version.GetBuild();

                DateTime dt = DateTime;
                string v = String.Format("{0:D7}.{1:D7}.{2:D8}.{3:yyyy.MM.dd.HH.mm.ss}", major, minor, build, dt);
                return v;
            }
        }

        public string FilenameWithoutExtension
        {
            get
            {
                return String.Format("{0}+{1}+{2}+{3}", Name, VersionAndDateTime, Branch, Platform);
            }
        }

        public string Filename
        {
            get
            {
                return FilenameWithoutExtension + Extension;
            }
        }

        public override string ToString()
        {
            return Filename;
        }
    }
}
