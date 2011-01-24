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
    /// 2) CachePackageRepository : Init, Add, Get
    /// 3) TargetPackageRepository: Init, Add, Get
    /// 4) LocalPackageRepository : Init, Add, Get
    /// 
    /// Package flow:
    /// - 4 to 2 (push)
    /// - 2 to 1 (push)
    /// - 4 to 1 (push)
    /// - 1 to 2 (pull)
    /// - 2 to 3 (pull)
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
        public IPackageRepository LocalRepository { get; set; }

        public void Init(PackageInstance package, List<DependencyInstance> dependencies)
        {

        }

        public void Update(PackageInstance package, VersionRange versionRange)
        {
            // Use Target Repository information and check Cache Repository for better versions
            // Use Cache Repository information and check Remote Repository for better versions
        }

        public bool Install(PackageInstance package)
        {
            // From Root to Local Repository (Create Package)
            // From Local Repository to Cache Repository

            return false;
        }

        public bool Deploy(PackageInstance package)
        {
            // From Cache Repository to Remote Repository

            return false;
        }
    }
}
