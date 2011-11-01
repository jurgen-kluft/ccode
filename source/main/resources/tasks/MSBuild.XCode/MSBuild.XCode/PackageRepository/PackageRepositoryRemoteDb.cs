using System;
using System.Collections.Generic;
using MSBuild.XCode.Helpers;
using xpackage_repo;

namespace MSBuild.XCode
{
    public class RemoteFilename : IPackageFilename
    {
        public RemoteFilename()
            : this(string.Empty)
        {
        }

        public RemoteFilename(string storagekey)
        {
            Name = storagekey;
            Version = null;
            DateTime = DateTime.Now;
            Branch = string.Empty;
            Platform = string.Empty;
            Extension = string.Empty;
        }

        public string Name { get; set; }
        public ComparableVersion Version { get; set; }
        public DateTime DateTime { get; set; }
        public string Branch { get; set; }
        public string Platform { get; set; }
        public string Extension { get; set; }

        public string FilenameWithoutExtension  { get { return Name; } }
        public string Filename { get { return Name; } }

        public override string ToString()
        {
            return Filename;
        }
    }

    public class PackageRepositoryRemoteDb : IPackageRepository
    {
        private readonly string mDatabaseURL;
        private readonly string mStorageURL;
        private readonly PackageRepo mPackageRepo;

        public PackageRepositoryRemoteDb(string databaseURL, string storageURL)
        {
            mDatabaseURL = databaseURL;
            mStorageURL = storageURL;
            Location = ELocation.Remote;

            RepoURL = "Remote::MySQL+Storage";

            // Initialize PackageRepo (MySQL database & FileSystem storage)
            mPackageRepo = new PackageRepo();
            mPackageRepo.initialize(mDatabaseURL, mStorageURL);
        }

        public PackageRepositoryRemoteDb(string database_and_storage_URL)
        {
            Valid = false;

            // Format
            // db::  |fs::
            string[] db_fs = database_and_storage_URL.Split(new string[] { "|" }, StringSplitOptions.RemoveEmptyEntries);

            if (db_fs.Length == 2)
            {
                mDatabaseURL = db_fs[0];
                mStorageURL = db_fs[1];

                // Initialize PackageRepo (MySQL database & FileSystem storage)
                mPackageRepo = new PackageRepo();
                Valid = mPackageRepo.initialize(mDatabaseURL, mStorageURL);
            }
        }

        public bool Valid { get; set; }
        public string RepoURL { get; set; }
        public ELocation Location { get; set; }

        public bool Query(PackageState package)
        {
            // Query means that we have to supply the information from the database about
            // the last version.
            Int64 version;
            if (mPackageRepo.find(package.Name, package.Group, package.Language, package.Platform, package.Branch, out version))
            {
                Dictionary<string, object> vars;
                if (mPackageRepo.find(package.Name, package.Group, package.Language, package.Platform, package.Branch, version, out vars))
                {
                    object storageKey;
                    if (vars.TryGetValue("Location", out storageKey))
                    {
                        package.RemoteStorageKey = storageKey as string;

                        object dateTime;
                        if (vars.TryGetValue("DateTime", out dateTime))
                        {
                            package.SetVersion(Location, new ComparableVersion(version));
                        
                            PackageFilename pf = new PackageFilename();
                            pf.Name = package.Name;
                            pf.Platform = package.Platform;
                            pf.Branch = package.Branch;
                            pf.DateTime = (DateTime)dateTime;
                            pf.Version = new ComparableVersion(version);

                            package.SetFilename(Location, pf);
                            package.SetURL(Location, mStorageURL);
                            package.SetSignature(Location, (DateTime)dateTime);

                            // We can fill in the signature here since we have gotten the latest 
                            // version and this has a date-time stamp which acts as the signature.

                            return true;
                        }
                    }
                }
            }
            return false;
        }

        public bool Query(PackageState package, VersionRange versionRange)
        {
            Int64 version;
            if (mPackageRepo.find(package.Name, package.Group, package.Language, package.Platform, package.Branch, out version))
            {
                ComparableVersion cv = new ComparableVersion(version);
                if (versionRange.IsInRange(cv))
                {
                    Dictionary<string, object> vars;
                    if (mPackageRepo.find(package.Name, package.Group, package.Language, package.Platform, package.Branch, version, out vars))
                    {
                        object storageKey;
                        if (vars.TryGetValue("Location", out storageKey))
                        {
                            package.RemoteStorageKey = storageKey as string;

                            object dateTime;
                            if (vars.TryGetValue("DateTime", out dateTime))
                            {
                                package.SetVersion(Location, cv);

                                PackageFilename pf = new PackageFilename();
                                pf.Name = package.Name;
                                pf.Platform = package.Platform;
                                pf.Branch = package.Branch;
                                pf.DateTime = (DateTime)dateTime;
                                pf.Version = new ComparableVersion(version);

                                package.SetFilename(Location, pf);
                                package.SetURL(Location, mStorageURL);
                                package.SetSignature(Location, (DateTime)dateTime);

                                // We can fill in the signature here since we have gotten a specific
                                // version and this has a date-time stamp which acts as the signature.

                                return true;
                            }
                        }
                    }
                }
            }
            return false;
        }

        public bool Link(PackageState package, out string filename)
        {
            filename = string.Empty;
            return false;
        }

        public bool Download(PackageState package, string to_filename)
        {
            // Here we are asked to download a package to the local machine
            // The package.GetFilename(Location) will give you a storage key.
            // Here we know where and how to get it and we will copy it to
            // the to_filename.
            string key = package.RemoteStorageKey;
            return mPackageRepo.download(key, to_filename);
        }

        public bool Submit(PackageState package, IPackageRepository from)
        {
            // 'From' actually should always be 'cache' which is the local package repository
            // (not to be confused with LocalRepo).
            Int64 version = package.GetVersion(from.Location).ToInt();

            string fromRepoDirectFilename;
            if (from.Link(package, out fromRepoDirectFilename))
            {
                List<KeyValuePair<Package, Int64>> dependencies;
                if (PackageArchive.RetrieveDependencies(fromRepoDirectFilename, out dependencies))
                {
                    if (mPackageRepo.upLoad(package.Name, package.Group, package.Language, package.Platform, package.Branch, version, package.GetFilename(from.Location).DateTime.ToSql(), package.Changeset, dependencies, fromRepoDirectFilename))
                    {
                        package.RemoteVersion = new ComparableVersion(version);
                        return true;
                    }
                }
            }

            return false;
        }

    }
}
