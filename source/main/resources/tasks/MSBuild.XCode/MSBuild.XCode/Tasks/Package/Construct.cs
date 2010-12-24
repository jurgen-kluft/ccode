using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageConstruct : Task
    {

        public string Name { get; set; }
        [Required]
        public string Action { get; set; }      ///< init, dir, vs2010

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
            Loggy.TaskLogger = Log;

            if (String.IsNullOrEmpty(Action))
                Action = "dir";
            Action = Action.ToLower();
            if (String.IsNullOrEmpty(Language))
                Language = "cpp";

            RootDir = RootDir.EndWith('\\');
            TemplateDir = TemplateDir.EndWith('\\');

            if (!Directory.Exists(TemplateDir))
            {
                Loggy.Add(String.Format("Error: Action {0} failed in Package::Construct since template directory {1} doesn't exist", Action, TemplateDir));
                return false;
            }

            Global.TemplateDir = TemplateDir;
            Global.CacheRepoDir = CacheRepoDir;
            Global.RemoteRepoDir = RemoteRepoDir;

            if (Action.StartsWith("vs2010"))
            {
                if (!Global.Initialize())
                    return false;

                Package package = new Package();
                package.IsRoot = true;
                package.RootDir = RootDir;
                if (package.LoadFinalPom())
                {
                    package.Name = package.Pom.Name;
                    package.Group = package.Pom.Group;
                    package.Version = null;
                    package.Branch = string.Empty;
                    package.Platform = string.Empty;

                    package.BuildAllDependencies();
                    package.SyncAllDependencies();
                    package.PrintAllDependencies();

                    // Generate the projects and solution
                    package.GenerateProjects();
                    package.GenerateSolution();
                }
                else
                {
                    Loggy.Add(String.Format("Error: Action {0} failed in Package::Construct due to failure in loading pom.xml", Action));
                    return false;
                }
            }
            else if (Action.StartsWith("dir"))
            {
                Package package = new Package();
                package.IsRoot = true;
                package.RootDir = RootDir;
                if (package.LoadPom())
                {
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
                }
                else
                {
                    Loggy.Add(String.Format("Error: Action {0} failed in Package::Construct due to failure in loading pom.xml", Action));
                    return false;
                }
            }
            else if (Action.StartsWith("init"))
            {
                if (!String.IsNullOrEmpty(Name))
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
                    else
                    {
                        Loggy.Add(String.Format("Error: Action {0} failed in Package::Construct since directory already exists", Action));
                        return false;
                    }
                }
                else
                {
                    Loggy.Add(String.Format("Error: Action {0} failed in Package::Construct since 'Name' was not specified", Action));
                    return false;
                }
            }
            else
            {
                Loggy.Add(String.Format("Error: Action {0} is not recognized by Package::Construct (Available actions: Dir, MsDev2010)", Action));
                return false;
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
