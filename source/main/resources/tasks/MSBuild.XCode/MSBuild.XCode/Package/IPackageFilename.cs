using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace MSBuild.XCode
{
    public interface IPackageFilename
    {
        string Name { get; set; }
        ComparableVersion Version { get; set; }
        DateTime? DateTime { get; set; }
        string Branch { get; set; }
        string Platform { get; set; }
    }

    public class PackageFilename : IPackageFilename
    {
        public PackageFilename(string name, ComparableVersion version, string branch, string platform)
        {
            Name = name;
            Version = version;
            DateTime = null;
            Branch = branch;
            Platform = platform;
        }
        public PackageFilename(string name, ComparableVersion version, DateTime dateTime, string branch, string platform)
            : this(name, version, branch, platform)
        {
            DateTime = dateTime;
        }

        public string Name { get; set; }
        public ComparableVersion Version { get; set; }
        public DateTime? DateTime { get; set; }
        public string Branch { get; set; }
        public string Platform { get; set; }

        public override string ToString()
        {
            string datetime = string.Empty;
            if (DateTime.HasValue)
            {
                DateTime dt = DateTime.Value;
                datetime = String.Format(".{0}.{1}.{2}.{3}.{4}.{5}", dt.Year, dt.Month, dt.Day, dt.Hour, dt.Minute, dt.Second);
            }
            return String.Format("{0}+{1}{2}+{3}+{4}", Name, Version.ToString(), datetime, Branch, Platform);
        }
    }
}
