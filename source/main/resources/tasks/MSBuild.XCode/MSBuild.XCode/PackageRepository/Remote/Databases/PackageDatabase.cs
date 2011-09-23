using System;
using System.Collections;
using System.Collections.Generic;

namespace xpackage_repo
{
    public class PackageDatabase
    {
        private IPackageDatabase mDatabase;

        public PackageDatabase()
        {

        }

        public bool connect(string databaseURL)
        {
            string prefix = "db::";
            if (databaseURL.StartsWith(prefix))
            {
                databaseURL = databaseURL.Remove(0, prefix.Length);
                PackageDatabaseMySQL db = new PackageDatabaseMySQL();
                if (db.connect(databaseURL))
                {
                    mDatabase = db;
                    return true;
                }
            }

            mDatabase = null;
            return false;
        }

        public bool submit(PackageVersion_pv package, List<KeyValuePair<string, int>> dependencies)
        {
            return mDatabase.submit(package, dependencies);
        }

        public bool findUniqueVersion(PackageVersion_pv package)
        {
            return mDatabase.findUniqueVersion(package);
        }

        public bool findLatestVersion(PackageVersion_pv package, out int outVersion)
        {
            return mDatabase.findLatestVersion(package, out outVersion);
        }

        public bool findLatestVersion(PackageVersion_pv package, int start_version, bool include_start, int end_version, bool include_end, out int outVersion)
        {
            return mDatabase.findLatestVersion(package, start_version, include_start, end_version, include_end, out outVersion);
        }

        public bool retrieveVarsOf(PackageVersion_pv package, out Dictionary<string, object> vars)
        {
            return mDatabase.retrieveVarsOf(package, out vars);
        }
    }
}