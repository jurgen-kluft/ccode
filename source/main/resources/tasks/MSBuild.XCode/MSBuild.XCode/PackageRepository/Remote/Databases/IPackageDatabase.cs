using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace xpackage_repo
{
    public interface IPackageDatabase
    {
        bool submit(PackageVersion_pv package, List<KeyValuePair<string, int>> dependencies);

        bool findUniqueVersion(PackageVersion_pv package);
        bool findLatestVersion(PackageVersion_pv package, out int outVersion);
        bool findLatestVersion(PackageVersion_pv package, int start_version, bool include_start, int end_version, bool include_end, out int outVersion);

        bool retrieveVarsOf(PackageVersion_pv package, out Dictionary<string, object> vars);
    }
}
