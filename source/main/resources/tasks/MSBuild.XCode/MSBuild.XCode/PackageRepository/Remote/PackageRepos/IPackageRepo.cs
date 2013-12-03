using System;
using System.Collections.Generic;
using MSBuild.XCode;

namespace xpackage_repo
{
    /// <summary>
    /// A PackageServer consists of a IPackageRepo and a StorageSystem(IStorage) instance.
    /// 
    /// We upload the package to the IPackageRepo which will validate if there is an entry
    /// in the database for that package, if there is no entry then it will create one.
    /// A new version is added for the package and the information is committed to the
    /// database upon which we submit the binary to the StorageSystem and the received
    /// key will be updated in the database.
    /// 
    /// </summary>
    public interface IPackageRepo
    {
        /// <summary>
        /// Initialize the package repository giving it a database URL and
        /// a storage URL.
        /// 
        /// The database URL can be:
        /// - fs::\\cnshasap2\Hg_Repo\PACKAGE_REPO
        /// - fs::D:\Packages\PACKAGE_REPO
        /// - db::cnshasap2:3307|username:password,developer:Fastercode189|schema,xcode_cpp
        /// 
        /// The storage URL can be:
        /// - fs::\\cnshasap2\Hg_Repo\PACKAGE_REPO
        /// - fs::D:\Packages\PACKAGE_REPO
        /// - p4::10.0.0.105:1666|username:password,xcode:112233
        /// 
        /// </summary>
        bool initialize(string databaseURL, string storageURL);

        /// <summary>
        /// This method is used for uploading package 
        /// </summary>
        bool upLoad(string Name, string Group, string Language, string Platform, string ToolSet, string Branch, Int64 Version, string Datetime, string Changeset, List<KeyValuePair<Package, Int64>> dependencies, string localFilename);

        /// <summary>
        /// This method is used to retrieve a package from the repository
        /// </summary>
        /// <param name="storageKey">The storage key of the package</param>
        /// <param name="destinationPath">Where to copy the package file</param>
        /// <returns>True if package has been copied at the destination path successfully</returns>
        bool download(string storageKey, string destinationPath);

        /// <summary>
        /// These methods are used for downloading a version or best version of the package
        /// </summary>
        /// <param name="destLocalPath"> The location you want to download in your local place</param>
        /// <returns>if the package is not exist , it will return false</returns>
        bool find(string package_name, string package_group, string package_language, string platform, string toolset, string branch, out Int64 version);
        bool find(string package_name, string package_group, string package_language, string platform, string toolset, string branch, Int64 version, out Dictionary<string, object> vars);
        bool find(string package_name, string package_group, string package_language, string platform, string toolset, string branch, Int64 start_version, bool include_start, Int64 end_version, bool include_end, out Dictionary<string, object> vars);
    }
}
