using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

namespace MSBuild.XCode
{
    public class PackageConstruct : Task
    {
        [Required]
        public string Name { get; set; }
        [Required]
        public string Language { get; set; }
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string TemplateDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }

        public override bool Execute()
        {
            // Check for the availability of pom.xml, pom.props and pom.targets
            // If they do not exist then
            //   Initialize the package dir using a name
            //   Copy files from the template directory
            //   Replace patterns (${Name}, ${GUID}, ${Language})
            // Else
            //   Load pom.xml
            //   Verify that the directory structure exists
            //   If not then create them
            //   Sync all dependend packages
            //   Create Developer Studio project and solution files
            // Endif
            if (!RootDir.EndsWith("\\"))
                RootDir = RootDir + "\\";

            if (!TemplateDir.EndsWith("\\"))
                TemplateDir = TemplateDir + "\\";

            if (!Directory.Exists(TemplateDir))
                return false;

            Global.TemplateDir = TemplateDir;
            Global.CacheRepoDir = CacheRepoDir;
            Global.RemoteRepoDir = RemoteRepoDir;

            if (File.Exists(RootDir + "pom.xml"))
            {
                Global.Initialize();

                Package package = new Package();
                package.IsRoot = true;
                package.RootDir = RootDir;
                package.LoadFinalPom();

                package.Name = package.Pom.Name;
                package.Group = package.Pom.Group;
                package.Version = null;
                package.Branch = string.Empty;
                package.Platform = string.Empty;

                // Check directory structure
                foreach (Attribute xa in package.Pom.DirectoryStructure)
                {
                    if (xa.Name == "Folder")
                    {
                        if (!Directory.Exists(RootDir + xa.Value))
                            Directory.CreateDirectory(RootDir + xa.Value);
                    }
                }

                if (false)
                {
                    string[] categories = package.Pom.GetCategories();

                    foreach (string category in categories)
                    {
                        string[] platforms = package.Pom.GetPlatformsForCategory(category);

                        Console.WriteLine(String.Format("Project, Category={0}", category));

                        foreach (string platform in platforms)
                        {
                            package.Pom.BuildDependencies(platform, Global.CacheRepo, Global.RemoteRepo);
                            package.Pom.PrintDependencies(platform);
                            package.Pom.SyncDependencies(platform, Global.CacheRepo);
                        }

                        foreach (string platform in platforms)
                        {
                            string[] configs = package.Pom.GetConfigsForPlatformsForCategory(platform, category);
                            //foreach (string config in configs)
                            //  package.Pom.CollectProjectInformation(category, platform, config);
                        }
                    }

                    // Generate the projects and solution
                    package.GenerateProjects();
                    package.GenerateSolution();
                }
            }
            else
            {
                string DstPath = RootDir + Name + "\\";
                if (!Directory.Exists(DstPath))
                {
                    Directory.CreateDirectory(DstPath);

                    // pom.targets.template ==> pom.targets
                    // pom.props.template ==> pom.props
                    // pom.xml.template ==> pom.xml
                    FileCopy(TemplateDir + "pom.targets.template", DstPath + "pom.targets");
                    FileCopy(TemplateDir + "pom.props.template", DstPath + "pom.props");
                    FileCopy(TemplateDir + "pom.xml.template", DstPath + "pom.xml");
                }
            }

            return true;
        }

        private bool FileCopy(string srcfile, string dstfile)
        {
            string[] lines = File.ReadAllLines(srcfile);

            using (FileStream wfs = new FileStream(dstfile, FileMode.Create, FileAccess.Write))
            {
                using (StreamWriter writer = new StreamWriter(wfs))
                {
                    foreach (string line in lines)
                    {
                        string l = line.Replace("${Name}", Name);
                        l = l.Replace("${Language}", Language);
                        if (l.Contains("${GUID}"))
                        {
                            string uuid = Guid.NewGuid().ToString();
                            l = l.Replace("${GUID}", uuid);
                        }
                        writer.WriteLine(l);
                    }
                    writer.Close();
                    wfs.Close();
                    return true;
                }
            }
        }
    }
}
