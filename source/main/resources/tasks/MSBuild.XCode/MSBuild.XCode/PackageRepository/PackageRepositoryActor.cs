using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    /// Update (PackageName, VersionRange/Version)
    ///   This will update Remote->Cache->Target.
    ///   This will fail when one or more dependency packages do not exist.
    /// 
    /// Load (TargetFolder)
    ///   This will load the pom.xml from Target for every dependency.
    ///   It will fail when a dependency package is not found.
    /// 
    /// 
    /// 1) RemotePackageRepository: Init, Add, Get
    /// 2) CachePackageRepository : Init, Add, Get
    /// 3) TargetPackageRepository: Init, Add, Get
    /// 4) LocalPackageRepository : Init, Add, Get
    /// 
    /// Package flow:
    /// - 4 to 2 (push)
    /// - 2 to 1 (push)
    /// 
    /// - 1 to 2 (pull)
    /// - 2 to 3 (pull)
    /// 
    /// PackageRepositoryActor
    /// - Init     (Analyze the target folder)
    /// - Update   (Using the state of the target folder, check the cache package repository for a better version, and last checking the remote package repository)
    /// - Install  (From the local folder, add a new package to the cache package repository)
    /// - Deploy   (From the cache package repository add the new package to the remote package repository)
    /// 
    /// </summary>
    public class PackageRepositoryActor
    {
        private Mercurial.Repository mHgRepo = null;

        private IPackageRepository RemoteRepo{ get; set; }
        private IPackageRepository CacheRepo{ get; set; }
        private IPackageRepository ShareRepo { get; set; }
        private IPackageRepository TargetRepo{ get; set; }
        private IPackageRepository LocalRepo{ get; set; }

        public static string RemoteRepoURL { get; set; }
        public static string CacheRepoURL { get; set; }
        
        public static string RootURL { get; set; }

        public bool Initialize(string _RemoteRepoURL, string _CacheRepoURL, string _RootURL)
        {
            RemoteRepoURL = _RemoteRepoURL;
            CacheRepoURL = _CacheRepoURL;
            RootURL = _RootURL;

            if (!String.IsNullOrEmpty(CacheRepoURL))
            {
                if (!Directory.Exists(CacheRepoURL))
                {
                    Loggy.Error(String.Format("Error: Cache repo {0} doesn't exist", CacheRepoURL));
                    return false;
                }
            }

            RemoteRepo = new PackageRepositoryRemoteDb(RemoteRepoURL);
            CacheRepo = new PackageRepositoryCache(CacheRepoURL, ELocation.Cache);
            ShareRepo = new PackageRepositoryShare(CacheRepoURL + ".share\\");
            LocalRepo = new PackageRepositoryLocal(RootURL);
            TargetRepo = new PackageRepositoryTarget(RootURL + "target\\");
            return RemoteRepo.Valid;
        }

        private bool CheckForUncommittedModifications()
        {
            if (mHgRepo == null)
                mHgRepo = new Mercurial.Repository(RootURL);
        
            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            if (!mHgRepo.Exists)
            {
                Loggy.Error(String.Format("Error: there is no Hg (Mercurial) repository!"));
                return true;
            }
            else if (mHgRepo.HasOutstandingChanges)
            {
                Loggy.Error(String.Format("Error: there are still outstanding (non committed) changes!"));
                return true;
            }
            return false;
        }

        private bool CheckForOutstandingModifications()
        {
            if (mHgRepo == null)
                mHgRepo = new Mercurial.Repository(RootURL);

            Mercurial.Changeset hg_changeset = mHgRepo.GetChangeSet();

            // - If there are outgoing or incoming change sets, then do not deploy
            bool any_incoming_changesets = !mHgRepo.Incoming().IsEmpty();
            bool any_outgoing_changesets = !mHgRepo.Outgoing().IsEmpty();

            if (any_outgoing_changesets && any_incoming_changesets)
            {
                Loggy.Error(String.Format("Error: Package::Deploy failed since there are incoming and outgoing changesets, pull, merge, build, test, commit and push before deploying!"));
                return true;
            }
            else if (any_incoming_changesets)
            {
                Loggy.Error(String.Format("Error: Package::Deploy failed since there are incoming changesets, pull, merge, build, test and commit before deploying!"));
                return true;
            }
            else if (any_outgoing_changesets)
            {
                Loggy.Error(String.Format("Error: Package::Deploy failed since there are outgoing changesets, push before deploying!"));
                return true;
            }
            return false;
        }

        public bool QueryBranch(Package package)
        {
            if (mHgRepo == null)
                mHgRepo = new Mercurial.Repository(RootURL);

            Mercurial.Changeset hg_changeset = mHgRepo.GetChangeSet();
            if (hg_changeset != null)
            {
                package.Branch = hg_changeset.Branch;
                return true;
            }
            return false;
        }

        public bool QueryBranchAndChangeset(Package package)
        {
            if (mHgRepo == null)
                mHgRepo = new Mercurial.Repository(RootURL);

            Mercurial.Changeset hg_changeset = mHgRepo.GetChangeSet();
            if (hg_changeset != null)
            {
                package.Branch = hg_changeset.Branch;
                package.Changeset = hg_changeset.Hash;
                return true;
            }
            return false;
        }

        public bool WriteVcsInformation(string filename)
        {
            if (mHgRepo == null)
                mHgRepo = new Mercurial.Repository(RootURL);

            Mercurial.Changeset hg_changeset = mHgRepo.GetChangeSet();

            // Write a vcs.info file containing VCS information, this will be included in the package
            dynamic x = new MSBuild.XCode.Helpers.Xml();
            x.Vcs(MSBuild.XCode.Helpers.Xml.Fragment(u =>
            {
                u.Type("Hg");
                u.Branch(hg_changeset.Branch);
                u.Revision(hg_changeset.Hash);
                u.AuthorName(hg_changeset.AuthorName);
                u.AuthorEmail(hg_changeset.AuthorEmailAddress);
            }));

            using (FileStream fs = new FileStream(filename, FileMode.Create))
            {
                using (StreamWriter sw = new StreamWriter(fs))
                {
                    sw.Write(x.ToString(true));
                    sw.Close();
                    fs.Close();

                    return true;
                }
            }

            return false;
        }


        public ComparableVersion MakeVersion(Package package, ComparableVersion root)
        {
            ComparableVersion remote;
            if (package.RemoteExists)
                remote = package.GetVersion(RemoteRepo.Location);
            else
                remote = new ComparableVersion("1.0.0");

            ComparableVersion cache;
            if (package.CacheExists)
                cache = package.GetVersion(CacheRepo.Location);
            else
                cache = new ComparableVersion("1.0.0");

            ComparableVersion local;
            if (package.LocalExists)
                local = package.GetVersion(LocalRepo.Location);
            else
                local = new ComparableVersion("1.0.0");

            //
            // Remote
            //

            int major = remote.GetMajor();
            int minor = remote.GetMinor();
            int build = remote.GetBuild();

            //
            // Cache
            //

            if (cache.GetMajor() > major)
            {
                major = cache.GetMajor();
                minor = cache.GetMinor();
                build = cache.GetBuild();
            }
            else if (cache.GetMajor() == major && cache.GetMinor() > minor)
            {
                minor = cache.GetMinor();
                build = cache.GetBuild();
            }
            else if (cache.GetMajor() == major && cache.GetMinor() == minor && cache.GetBuild() > build)
            {
                build = cache.GetBuild();
            }

            //
            // Local
            //

            if (local.GetMajor() > major)
            {
                major = local.GetMajor();
                minor = local.GetMinor();
                build = local.GetBuild();
            }
            else if (local.GetMajor() == major && local.GetMinor() > minor)
            {
                minor = local.GetMinor();
                build = local.GetBuild();
            }
            else if (local.GetMajor() == major && local.GetMinor() == minor && local.GetBuild() > build)
            {
                build = local.GetBuild();
            }

            //
            // Package (Root)
            //

            if (root.GetMajor() > major)
            {
                major = root.GetMajor();
                minor = root.GetMinor();
                build = root.GetBuild();
            }
            else if (root.GetMajor() == major && root.GetMinor() > minor)
            {
                minor = root.GetMinor();
                build = root.GetBuild();
            }
            else if (root.GetMajor() == major && root.GetMinor() == minor && root.GetBuild() > build)
            {
                build = root.GetBuild();
            }

            if (package.IncrementVersion)
                ++build;

            return new ComparableVersion(String.Format("{0}.{1}.{2}", major, minor, build));
        }

        public bool Create(Package package, PackageContent content, ComparableVersion rootVersion)
        {
            Environment.CurrentDirectory = RootURL;

            // First query the branch that we are working on
            if (!QueryBranch(package))
                return false;

            // Query the package at the Remote which will
            // give us the latest version
            if (package.IncrementVersion)
                RemoteRepo.Query(package);

            // Query the cache and local to get the latest versions
            // if there are any.
            CacheRepo.Query(package);
            LocalRepo.Query(package);
            {
                // Make version here?
                package.CreateVersion = MakeVersion(package, rootVersion);

                // Write a vcs.info file containing dependency package info, this will be included in the package
                {
                    string buildURL = String.Format("{0}target\\{1}\\build\\{2}\\", RootURL, package.Name, package.Platform);
                    if (!Directory.Exists(buildURL))
                        Directory.CreateDirectory(buildURL);

                    WriteVcsInformation(buildURL + "vcs.info");

                    if (PackageArchive.Create(package, content, package.RootURL))
                    {
                        if (LocalRepo.Submit(package, null))
                        {
                            return true;
                        }
                    }
                }
            }

            return false;
        }

        public int Update(Package package, VersionRange versionRange)
        {
            // Find best version on remote
            // Find best version in the cache
            // Determine the best version, remote or cache
            // If remote has the best version
            //    Download from remote to cache
            // End
            // Update cache to share
            // Update share to target

            TargetRepo.Query(package);

            // Try to get the package from the Cache to Target
            if (!package.RemoteExists)
                RemoteRepo.Query(package, versionRange);
            if (!package.CacheExists)
                CacheRepo.Query(package, versionRange);

            if (package.CacheExists)
            {
                ShareRepo.Query(package);
            }

            // Do a signature verification
            if (package.TargetExists && package.ShareExists && package.CacheExists)
            {
                if (package.RemoteExists)
                {
                    if (package.RemoteSignature == package.CacheSignature && package.CacheSignature == package.ShareSignature)
                    {
                        return 0;
                    }
                }
                else
                {
                    if (package.CacheSignature == package.TargetSignature)
                        return 0;
                }
            }

            int result = 0;

            // Check if the Remote has a better version, Remote may not exist
            if (package.RemoteExists)
            {
                if (!package.CacheExists || (package.RemoteVersion > package.CacheVersion))
                {
                    // Update from remote to cache
                    if (!CacheRepo.Submit(package, RemoteRepo))
                    {
                        // Failed to get package from Remote to Cache
                        result = -1;
                    }
                }
            }

            if (package.CacheExists)
            {
                if (!package.ShareExists || (package.CacheVersion > package.ShareVersion))
                {
                    if (ShareRepo.Submit(package, CacheRepo))
                    {
                        result = 1;
                    }
                    else
                    {
                        // Failed to get package from Cache to Share
                        result = -1;
                    }
                }
            }
            else
            {
                // Failed to get package from Remote to Cache
                result = -1;
            }

            if (package.ShareExists)
            {
                if (!package.TargetExists || (package.ShareVersion > package.TargetVersion))
                {
                    if (TargetRepo.Submit(package, ShareRepo))
                    {
                        result = 1;
                    }
                    else
                    {
                        // Failed to get package from Cache to Target
                        result = -1;
                    }
                }
            }
            else
            {
                // Failed to get package from Remote to Cache
                result = -1;
            }

            return result;
        }

        public bool Install(Package package)
        {
            if (CheckForUncommittedModifications())
                return false;

            // First query the branch that we are working on
            if (!QueryBranch(package))
                return false;

            // Query Local (and there should be a package)
            if (LocalRepo.Query(package))
            {
                // Submit the created package from Local to Cache
                if (CacheRepo.Submit(package, LocalRepo))
                    return true;
            }
            return false;
        }

        public bool Deploy(Package package)
        {
            if (CheckForOutstandingModifications())
                return false;

            // First query the branch and change set
            if (!QueryBranchAndChangeset(package))
                return false;
           
            // Query the package at the Remote which will
            // give us the latest version
            RemoteRepo.Query(package);

            // Query Cache (and there should be a package)
            if (CacheRepo.Query(package))
            {
                // Submit the created package from Cache to Remote
                return (RemoteRepo.Submit(package, CacheRepo));
            }

            return false;
        }
    }
}
