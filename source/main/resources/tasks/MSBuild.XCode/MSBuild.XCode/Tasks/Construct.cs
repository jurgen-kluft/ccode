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
        public string Platform { get; set; }
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
                Language = "C++";

            RootDir = RootDir.EndWith('\\');
            TemplateDir = TemplateDir.EndWith('\\');

            if (!Directory.Exists(TemplateDir))
            {
                Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct since template directory {1} doesn't exist", Action, TemplateDir));
                return false;
            }

            PackageInstance.TemplateDir = TemplateDir;
            PackageInstance.CacheRepoDir = CacheRepoDir;
            PackageInstance.RemoteRepoDir = RemoteRepoDir;

            if (Action.StartsWith("vs2010"))
            {
                if (!PackageInstance.Initialize())
                    return false;

                PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
                if (package.IsValid)
                {
                    PackageDependencies dependencies = new PackageDependencies(package);

                    List<string> platforms = new List<string>(package.Pom.Platforms);
                    if (!String.IsNullOrEmpty(Platform) && String.Compare(Platform, "all", true) != 0)
                    {
                        platforms.Clear();
                        platforms.Add(Platform);
                    }

                    if (dependencies.BuildForPlatforms(platforms))
                    {
                        dependencies.PrintForPlatforms(platforms);

                        // Generate the projects and solution
                        package.GenerateProjects(dependencies, platforms);
                        if (!package.GenerateSolution())
                        {
                            Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in saving solution (.sln)", Action));
                            return false;
                        }
                    }
                    else
                    {
                        Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in building dependencies", Action));
                        return false;
                    }
                }
                else
                {
                    Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in loading pom.xml", Action));
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
                    Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in loading pom.xml", Action));
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
                                if (String.Compare(Language, "C++", true) == 0 || String.Compare(Language, "CPP", true) == 0)
                                {
                                    if (FileCopy(TemplateDir + "pom.xml.template", DstPath + "pom.xml"))
                                    {
                                        Loggy.Info(String.Format("Info: Generated pom.targets, pom.props and pom.xml C++ files"));
                                        file_copy_result = true;
                                    }
                                }
                                else
                                {
                                    if (FileCopy(TemplateDir + "pom.xml.cs.template", DstPath + "pom.xml"))
                                    {
                                        Loggy.Info(String.Format("Info: Generated pom.targets, pom.props and pom.xml C++ files"));
                                        file_copy_result = true;
                                    }
                                }
                            }
                        }
                        if (!file_copy_result)
                        {
                            Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct to copy the template (pom.targets, pom.props and pom.xml) files", Action));
                        }

                        // Init the Mercurial repository, add the above files and commit
                        Mercurial.Repository hg_repo = new Mercurial.Repository(DstPath);
                        hg_repo.Init();
                        hg_repo.Add(".");
                        hg_repo.Commit("init (xcode)");
                    }
                    else
                    {
                        Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct since directory already exists", Action));
                        return false;
                    }
                }
                else
                {
                    Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct since 'Name' was not specified", Action));
                    return false;
                }
            }
            else
            {
                Loggy.Error(String.Format("Error: Action {0} is not recognized by Package::Construct (Available actions: Init, Dir, MsDev2010)", Action));
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
                        string l = line.Replace("${PackageName}", Name);
                        l = l.Replace("${PackageLanguage}", Language);
                        if (l.Contains("${PackageGUID}"))
                        {
                            string uuid = Guid.NewGuid().ToString();
                            l = l.Replace("${PackageGUID}", uuid);
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
