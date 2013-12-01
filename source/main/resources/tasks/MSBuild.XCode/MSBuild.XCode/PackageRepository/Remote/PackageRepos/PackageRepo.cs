using System;
using System.Collections.Generic;
using xstorage_system;
using MSBuild.XCode;
using MSBuild.XCode.Helpers;

namespace xpackage_repo
{
    public class PackageRepo : IPackageRepo
    {
        private StorageSystem mStorageSystem;
        private PackageDatabase mDatabaseSystem;

        public bool initialize(string databaseURL, string storageURL)
        {
            mStorageSystem = new StorageSystem();
            if (mStorageSystem.connect(storageURL))
            {
                mDatabaseSystem = new PackageDatabase();
                return mDatabaseSystem.connect(databaseURL);
            }
            else
            {
                return false;
            }
        }

        public bool upLoad(string Name, string Group, string Language, string Platform, string ToolSet, string Branch, Int64 Version, string Datetime, string Changeset, List<KeyValuePair<Package, Int64>> dependencies, string localFilename)
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
                pv.ToolSet = ToolSet;
                pv.Branch = Branch;
                pv.Version = Version;
                pv.Datetime = Datetime;
                pv.Location = storage_key;
                pv.Changeset = Changeset;

                List<PackageVersion_pv> deps = new List<PackageVersion_pv>();
                foreach (KeyValuePair<Package,Int64> d in dependencies)
                {
                    PackageVersion_pv p = new PackageVersion_pv();
                    p.Name = d.Key.Name;
                    p.Group = d.Key.Group;
                    p.Language = d.Key.Language;
                    p.Version = d.Value;
                    p.Platform = Platform;
                    p.ToolSet = ToolSet;
                    p.Branch = Branch;
                    p.Location = string.Empty;
                    p.Changeset = d.Key.Changeset;
                    deps.Add(p);
                }

                return mDatabaseSystem.submit(pv, deps);
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

        public bool find(string package_name, string package_group, string package_language, string platform, string toolset, string branch, out Int64 version)
        {
            PackageVersion_pv pv = new PackageVersion_pv();
            pv.Name = package_name;
            pv.Group = package_group;
            pv.Language = package_language;
            pv.Branch = branch;
            pv.Platform = platform;
            pv.ToolSet = toolset;

            Int64 lowestVersion = PackageVersion_pv.LowestVersion;
            Int64 highestVersion = PackageVersion_pv.HighestVersion;
            Int64 bestVersion;
            if (mDatabaseSystem.findLatestVersion(pv, lowestVersion, true, highestVersion, true, out bestVersion))
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

        public bool find(string package_name, string package_group, string package_language, string platform, string toolset, string branch, Int64 version, out Dictionary<string, object> vars)
        {
            PackageVersion_pv pv = new PackageVersion_pv();
            pv.Name = package_name;
            pv.Group = package_group;
            pv.Language = package_language;
            pv.Branch = branch;
            pv.Platform = platform;
            pv.ToolSet = toolset;
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

        public bool find(string package_name, string package_group, string package_language, string platform, string toolset, string branch, Int64 start_version, bool include_start, Int64 end_version, bool include_end, out Dictionary<string, object> vars)
        {
            PackageVersion_pv pv = new PackageVersion_pv();
            pv.Name = package_name;
            pv.Group = package_group;
            pv.Language = package_language;
            pv.Branch = branch;
            pv.Platform = platform;
            pv.ToolSet = toolset;

            Int64 bestVersion;
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
