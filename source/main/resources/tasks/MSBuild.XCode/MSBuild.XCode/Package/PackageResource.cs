using System;
using System.IO;
using System.Collections.Generic;

namespace MSBuild.XCode
{
    public class PackageResource
    {
        private PomResource mPom;

        public string Name { get { return mPom.Name; } }
        public Group Group { get { return mPom.Group; } }

        public bool IsValid { get { return mPom.IsValid; } }

        public PackageContent Content { get { return mPom.Content; } }
        public PackageVars Vars { get { return mPom.Vars; } }
        public List<DependencyResource> Dependencies { get { return mPom.Dependencies; } }
        public PackageStructure DirectoryStructure { get { return mPom.DirectoryStructure; } }
        public List<ProjectResource> Projects { get { return mPom.Projects; } }
        public List<string> Platforms { get { return mPom.Platforms; } }
        public Versions Versions { get { return mPom.Versions; } }

        private PackageResource()
        {

        }

        public bool Info()
        {
            return mPom.Info();
        }

        internal PackageInstance CreateInstance(bool main)
        {
            PackageInstance instance = new PackageInstance(main, this, new PomInstance(main, mPom));
            return instance;
        }

        internal PomInstance CreatePomInstance(bool main)
        {
            PomInstance pom = new PomInstance(main, mPom);
            return pom;
        }

        public static PackageResource From(string name, string group)
        {
            PackageResource resource = new PackageResource();
            resource.mPom = PomResource.From(name, group);
            return resource;
        }

        public static PackageResource From(string group, IPackageFilename filename)
        {
            PackageResource resource = new PackageResource();
            resource.mPom = PomResource.From(filename.Name, group);
            return resource;
        }

        private static PackageResource From(IPackageFilename filename)
        {
            return From(string.Empty, filename);
        }

        public static PackageResource LoadFromFile(string url)
        {
            PackageResource resource = new PackageResource();

            if (!String.IsNullOrEmpty(url) && File.Exists(url + "pom.xml"))
            {
                resource.mPom = new PomResource();
                resource.mPom.LoadFile(url + "pom.xml");
            }
            else
            {
                resource.mPom = new PomResource();
            }
            return resource;
        }

        public static PackageResource LoadFromPackage(string url, IPackageFilename filename)
        {
            PackageResource resource = PackageResource.From(filename);
            if (File.Exists(url + filename.ToString()))
            {
                PackageZipper zip = PackageZipper.Open(url + filename.ToString(), FileAccess.Read);
                if (zip.Files.Count > 0)
                {
                    PackageZipper.ZipFileEntry entry;
                    if (zip.FindEntryByFilename("pom.xml", out entry))
                    {
                        using (MemoryStream stream = new MemoryStream())
                        {
                            zip.ExtractFile(entry, stream);
                            stream.Position = 0;
                            using (StreamReader reader = new StreamReader(stream))
                            {
                                string xml = reader.ReadToEnd();
                                reader.Close();
                                stream.Close();
                                resource.mPom = new PomResource();
                                resource.mPom.LoadXml(xml);
                                return resource;
                            }
                        }
                    }
                }
            }
            resource.mPom = new PomResource();
            return resource;
        }


    }
}
