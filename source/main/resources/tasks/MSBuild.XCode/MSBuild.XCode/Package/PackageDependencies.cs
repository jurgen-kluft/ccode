using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    /// <summary>
    /// This object will handle the dependencies of the root package.
    /// 
    /// Compile
    ///  Compile will build the dependency tree incrementally, it will return 0 when done.
    ///  You need to call Compile() incrementally while it is returning 1, 0 means it is
    ///  complete and -1 means a fatal error.
    ///  
    /// Update
    ///  Update will walk the dependency tree and update all packages by negotiating with
    ///  the Cache and Remote Package Repositories.
    ///  Whenever Update() is called and returning 1, you need to call Compile() again since
    ///  a new package might have modified its dependencies.
    /// 
    /// Note:
    ///  Maybe we should pass an object down the dependency tree which acts as an actor and
    ///  container for the different repositories. Currently they are accessed through Global.
    /// 
    /// </summary>
    public class PackageDependencies
    {
        private PackageInstance mRootPackage;
        private Dictionary<string, DependencyTree> mDependencyTree;

        public PackageDependencies(PackageInstance rootPackage)
        {
            mRootPackage = rootPackage;
            mDependencyTree = new Dictionary<string, DependencyTree>();
        }

        public PackageInstance Package { get { return mRootPackage; } }

        public List<PackageInstance> GetAllDependencyPackages(string platform)
        {
            DependencyTree tree = GetDependencyTree(platform);
            List<PackageInstance> allDependencyPackages = tree.GetAllDependencyPackages();
            return allDependencyPackages;
        }

        public bool IsDependencyForPlatform(string DependencyName, string platform)
        {
            // It could be asking for ourselves, so check if this dependency name is the root package
            if (String.Compare(Package.Name, DependencyName, true) == 0)
                return true;

            // It was not the root package so it might be a dependency package, check the dependency tree
            DependencyTree tree;
            if (mDependencyTree.TryGetValue(platform.ToLower(), out tree))
            {
                return tree.ContainsDependencyForPlatform(DependencyName, platform);
            }
            return false;
        }

        private DependencyTree GetDependencyTree(string platform)
        {
            DependencyTree tree;
            if (!mDependencyTree.TryGetValue(platform.ToLower(), out tree))
            {
                List<DependencyInstance> dependencies = new List<DependencyInstance>();
                foreach (DependencyResource resource in Package.Dependencies)
                {
                    if (resource.IsForPlatform(platform))
                        dependencies.Add(new DependencyInstance(platform, resource));
                }
                tree = new DependencyTree(Package, platform, dependencies);
                mDependencyTree.Add(platform.ToLower(), tree);
            }
            return tree;
        }

        // 0=Done, -1=Error
        private int Compile(string platform)
        {
            // Will return True if the DependencyTree is completely build and all dependency packages are available in the Target Repository.
            // This doesn't mean that all dependencies are up-to-date, there can be newer (better) versions in the Cache/Remote Repository.

            DependencyTree tree = GetDependencyTree(platform);
            int result = tree.Compile();
            return result;
        }

        public bool BuildForPlatform(string platform)
        {
            int result = Compile(platform);
            return result == 0;
        }

        public bool BuildForPlatforms(List<string> platforms)
        {
            foreach (string platform in platforms)
            {
                if (!BuildForPlatform(platform))
                    return false;
            }
            return true;
        }

        public bool BuildForAllPlatforms()
        {
            foreach (string platform in Package.Pom.Platforms)
            {
                if (!BuildForPlatform(platform))
                    return false;
            }
            return true;
        }

        public void PrintForPlatform(string platform)
        {
            Loggy.Info(String.Format("Dependencies for platform : {0}", platform));
            DependencyTree tree = GetDependencyTree(platform);
            tree.Print();
        }

        public void PrintForPlatforms(List<string> platforms)
        {
            Loggy.Indent += 1;
            foreach (string platform in platforms)
                PrintForPlatform(platform);
            Loggy.Indent -= 1;
        }

        public void PrintForAllPlatforms()
        {
            Loggy.Indent += 1;
            foreach (string platform in Package.Pom.Platforms)
                PrintForPlatform(platform);
            Loggy.Indent -= 1;
        }

        public void SaveInfo(string platform, FileDirectoryPath.FilePathAbsolute filepath)
        {
            DependencyTree tree = GetDependencyTree(platform);
            tree.SaveInfo(filepath);
        }
    }
}
