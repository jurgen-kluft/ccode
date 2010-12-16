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
        public string LocalRepoDir { get; set; }
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

            XGlobal.TemplateDir = TemplateDir;
            XGlobal.LocalRepoDir = LocalRepoDir;
            XGlobal.RemoteRepoDir = RemoteRepoDir;

            if (!Directory.Exists(RootDir))
            {
                string DstPath = RootDir + Name + "\\";
                Directory.CreateDirectory(DstPath);

                // pom.targets.template ==> pom.targets
                // pom.props.template ==> pom.props
                // pom.xml.template ==> pom.xml
                FileCopy(TemplateDir + "pom.targets.template", DstPath + "pom.targets");
                FileCopy(TemplateDir + "pom.props.template", DstPath + "pom.props");
                FileCopy(TemplateDir + "pom.xml.template", DstPath + "pom.xml");
            }
            else if (File.Exists(RootDir + "pom.xml"))
            {
                XGlobal.Initialize();

                XPom pom = new XPom();
                pom.Load(RootDir + "pom.xml");
                pom.PostLoad();

                // Check directory structure
                foreach (XAttribute xa in pom.DirectoryStructure)
                {
                    if (xa.Name == "Folder")
                    {
                        if (!Directory.Exists(RootDir + xa.Value))
                            Directory.CreateDirectory(RootDir + xa.Value);
                    }
                }

                string[] categories = pom.GetCategories();

                foreach (string category in categories)
                {
                    string[] platforms = pom.GetPlatformsForCategory(category);

                    Console.WriteLine(String.Format("Project, Category={0}", category));

                    foreach (string platform in platforms)
                    {
                        pom.BuildDependencies(platform, XGlobal.LocalRepo, XGlobal.RemoteRepo);
                        pom.PrintDependencies(platform);
                        pom.CheckoutDependencies(RootDir, platform, XGlobal.LocalRepo);
                    }

                    foreach (string platform in platforms)
                    {
                        string[] configs = pom.GetConfigsForPlatformsForCategory(category, platform);
                        foreach (string config in configs)
                            pom.CollectProjectInformation(category, platform, config);
                    }
                }

                // Generate the projects and solution
                pom.GenerateProjects(RootDir);
                pom.GenerateSolution(RootDir);
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
