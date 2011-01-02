using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

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
    /// 2) CachePackageRepository:  Init, Add, Get
    /// 3) TargetPackageRepository: Init, Add, Get
    /// 
    /// Packages go from 3 to 2 to 1, or from 1 to 2 to 3, or just from 2 to 3 and 3 to 2
    /// 
    /// PackageRepositoryActor
    /// - Init     (Analyze the target folder)
    /// - Update   (Using the state of the target folder, check the cache package repository for a better version, and last checking the remote package repository)
    /// - Install  (From the target folder, add a new package to the cache package repository)
    /// - Deploy   (From the cache package repository add the new package to the remote package repository)
    /// 
    /// </summary>
    public class PackageRepositoryActor
    {
        public IPackageRepository RemoteRepository { get; set; }
        public IPackageRepository CacheRepository { get; set; }
        public IPackageRepository TargetRepository { get; set; }

        public void Init(string rootPackageName, List<DependencyInstance> dependencies)
        {

        }

        public void Update(string packageName, VersionRange versionRange)
        {
            // Use Target Repository information and check Cache Repository for better versions
            // Use Cache Repository information and check Remote Repository for better versions
        }

        public bool Install()
        {
            // From Target Repository to Cache Repository

            return false;
        }

        public bool Deploy()
        {
            // From Cache Repository to Remote Repository

            return false;
        }
    }
}
