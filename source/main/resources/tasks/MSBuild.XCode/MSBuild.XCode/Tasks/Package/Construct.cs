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

                PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
                if (package.IsValid)
                {
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
                PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
                if (package.IsValid)
                {
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
                        bool file_copy_result = false;
                        if (FileCopy(TemplateDir + "pom.targets.template", DstPath + "pom.targets"))
                        {
                            if (FileCopy(TemplateDir + "pom.props.template", DstPath + "pom.props"))
                            {
                                if (FileCopy(TemplateDir + "pom.xml.template", DstPath + "pom.xml"))
                                {
                                    file_copy_result = true;
                                }
                            }
                        }
                        if (!file_copy_result)
                        {
                            Loggy.Add(String.Format("Error: Action {0} failed in Package::Construct to copy the template (pom.targets, pom.props and pom.xml) files", Action));
                        }

                        // Init the Mercurial repository, add the above files and commit
                        Mercurial.Repository hg_repo = new Mercurial.Repository(DstPath);
                        hg_repo.Init();
                        hg_repo.Add(".");
                        hg_repo.Commit("init (xcode)");
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
