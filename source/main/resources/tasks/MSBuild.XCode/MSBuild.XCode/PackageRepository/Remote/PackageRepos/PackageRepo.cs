using System;
using System.Collections.Generic;
using xstorage_system;
using MSBuild.XCode.Helpers;

namespace xpackage_repo
{
    public class PackageRepo : IPackageRepo
    {
        private StorageSystem mStorageSystem;
        private PackageDatabase mDatabaseSystem;

        public void initialize(string databaseURL, string storageURL)
        {
            mStorageSystem = new StorageSystem();
            mStorageSystem.connect(storageURL);

            mDatabaseSystem = new PackageDatabase();
            mDatabaseSystem.connect(databaseURL);
        }

        public bool upLoad(string Name, string Group, string Language, string Platform, string Branch, int Version, string Datetime, string Changeset, List<KeyValuePair<string, int>> dependencies, string localFilename)
        {
            string storage_key;
            if (mStorageSystem.submit(localFilename, out storage_key))
            {
                // First check/add the following data in the database (and get their id):
                //    - group
                //    - platform
                //    - language
                //    - branch
                // Then check/add for a package entry in the package table (and get the id):
                //    - name
                //    - group
                //    - language
                // Then add an entry in the database for this version:
                //    - package id
                //    - platform id
                //    - branch id
                //    - datetime
                //    - version
                //    - location
                //    - changeset
                PackageVersion_pv pv = new PackageVersion_pv();
                pv.Name = Name;
                pv.Group = Group;
                pv.Language = Language;
                pv.Platform = Platform;
                pv.Branch = Branch;
                pv.Version = Version;
                pv.Datetime = Datetime;
                pv.Location = storage_key;
                pv.Changeset = Changeset;
                return mDatabaseSystem.submit(pv, dependencies);
            }
            else
            {
                Loggy.Error(String.Format("Error: PackageRepo::upload, failed to store package"));
            }
            return false;
        }

        public bool download(string storageKey, string destinationPath)
        {
            return mStorageSystem.retrieve(storageKey, destinationPath);
        }

        public bool find(string package_name, string package_group, string package_language, string platform, string branch, out int version)
        {
            PackageVersion_pv pv = new PackageVersion_pv();
            pv.Name = package_name;
            pv.Group = package_group;
            pv.Language = package_language;
            pv.Branch = branch;
            pv.Platform = platform;

            int bestVersion;
            if (mDatabaseSystem.findLatestVersion(pv, 1000000, true, 999999999, true, out bestVersion))
            {
                version = bestVersion;
                return true;
            }
            else
            {
                version = 1000000;
                return false;
            }
        }

        public bool find(string package_name, string package_group, string package_language, string platform, string branch, int version, out Dictionary<string, object> vars)
        {
            PackageVersion_pv pv = new PackageVersion_pv();
            pv.Name = package_name;
            pv.Group = package_group;
            pv.Language = package_language;
            pv.Branch = branch;
            pv.Platform = platform;
            pv.Version = version;

            if (mDatabaseSystem.retrieveVarsOf(pv, out vars))
            {
                return true;
            }
            else
            {
                vars = new Dictionary<string, object>();
                return false;
            }
        }

        public bool find(string package_name, string package_group, string package_language, string platform, string branch, int start_version, bool include_start, int end_version, bool include_end, out Dictionary<string, object> vars)
        {
            PackageVersion_pv pv = new PackageVersion_pv();
            pv.Name = package_name;
            pv.Group = package_group;
            pv.Language = package_language;
            pv.Branch = branch;
            pv.Platform = platform;

            int bestVersion;
            if (mDatabaseSystem.findLatestVersion(pv, start_version, include_start, end_version, include_end, out bestVersion))
            {
                pv.Version = bestVersion;
                if (mDatabaseSystem.retrieveVarsOf(pv, out vars))
                {
                    return true;
                }
            }
            vars = new Dictionary<string, object>();
            return false;
        }
    }
}
