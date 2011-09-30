using System;
using System.Collections.Generic;

namespace xpackage_repo
{
    public interface IPackageDatabase
    {
        bool submit(PackageVersion_pv package, List<PackageVersion_pv> dependencies);

        bool findUniqueVersion(PackageVersion_pv package);
        bool findLatestVersion(PackageVersion_pv package, out Int64 outVersion);
        bool findLatestVersion(PackageVersion_pv package, Int64 start_version, bool include_start, Int64 end_version, bool include_end, out Int64 outVersion);

        bool retrieveVarsOf(PackageVersion_pv package, out Dictionary<string, object> vars);
    }
}
