using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System.Windows.Forms;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageConstruct : Task
    {
        public string Name { get; set; }
        [Required]
        public string Action { get; set; }      ///< init, dir, vs2010/vs2012/vs2013
        public string Platform { get; set; }
        public string IDE { get; set; }         ///< Eclipse, Visual Studio, XCODE
        public string ToolSet { get; set; }     ///< GCC, Visual Studio (v90, v100, v110, v120)
        public string Language { get; set; }
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string TemplateDir { get; set; }
        [Required]
        public string CacheRepoDir { get; set; }
        [Required]
        public string RemoteRepoDir { get; set; }
        
        private bool False()
        {
            return false;
        }
        private bool True()
        {
            return true;
        }

        public override bool Execute()
        {
            try
            {
                Loggy.TaskLogger = Log;

                if (String.IsNullOrEmpty(Action))
                    Action = "dir";

                Action = Action.ToLower();
                if (String.IsNullOrEmpty(Language))
                    Language = "C++";

                if (!String.IsNullOrEmpty(Platform))
                {
                    if (Platform.ToLower() == "all")
                        Platform = "*";
                }
                else
                {
                    Platform = "*";
                }

                IDE = !String.IsNullOrEmpty(IDE) ? IDE.ToLower() : "vs2012";
                ToolSet = !String.IsNullOrEmpty(ToolSet) ? ToolSet.ToLower() : "v110";

                RootDir = RootDir.EndWith('\\');
                TemplateDir = TemplateDir.EndWith('\\');

                if (!Directory.Exists(TemplateDir))
                {
                    Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct since template directory {1} doesn't exist", Action, TemplateDir));
                    return False();
                }

                PackageInstance.TemplateDir = TemplateDir;

                if (Action.StartsWith("genprj"))
                {
                    if (!PackageInstance.Initialize(RemoteRepoDir, CacheRepoDir, RootDir))
                        return False();

                    PackageVars vars = new PackageVars();
                    vars.Add("Name", Name);
                    vars.Add("Platform", Platform);
                    vars.Add("IDE", IDE);
                    vars.SetToolSet(Platform, ToolSet, true);

                    PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);

                    List<string> platforms = new List<string>(package.Pom.Platforms);
                    if (Platform != "*")
                    {
                        platforms.Clear();
                        string[] platforms_array = Platform.Split(new char[] { ',' }, StringSplitOptions.RemoveEmptyEntries);
                        foreach (string p in platforms_array)
                            platforms.Add(p);
                    }
                    else
                    {
                        // Retrieve the list from the POM which is the global list of platforms that the POM is using
                    }

                    if (platforms.Count == 0)
                    {
                        Loggy.Error(String.Format("Error: No platforms, are you sure you typed the platforms correctly?"));
                        return False();
                    }

                    package.SetPlatform(platforms[0]);
                    foreach (string p in platforms)
                    {
                        if (!package.HasPlatform(p))
                        {
                            Loggy.Error(String.Format("Error: Platform(s) '{0}' is not supported for this package, are you sure you typed it correctly?", p));
                            return False();
                        }
                    }

                    if (package.IsValid)
                    {
                        PackageDependencies dependencies = new PackageDependencies(package);

                        if (dependencies.BuildForPlatforms(platforms))
                        {
                            if (dependencies.SaveInfoForPlatforms(platforms, package.Vars))
                            {
                                dependencies.PrintForPlatforms(platforms);

                                // Generate the projects and solution
                                package.GenerateProjects(dependencies, platforms);
                                if (!package.GenerateSolution(platforms, IDE))
                                {
                                    Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in saving solution (.sln)", Action));
                                    return False();
                                }
                            }
                            else
                            {
                                Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in writing dependency information (dependencies.info)", Action));
                                return False();
                            }
                        }
                        else
                        {
                            Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in building dependencies", Action));
                            return False();
                        }
                    }
                    else
                    {
                        Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in loading pom.xml", Action));
                        return False();
                    }
                }
                else if (Action.StartsWith("dir"))
                {
                    PackageVars vars = new PackageVars();
                    PackageInstance package = PackageInstance.LoadFromRoot(RootDir, vars);
                    if (package.IsValid)
                    {
                        package.Pom.DirectoryStructure.Create(RootDir);
                    }
                    else
                    {
                        Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct due to failure in loading pom.xml", Action));
                        return False();
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
                                            Loggy.Info(String.Format("Generated pom.targets, pom.props and pom.xml files"));
                                            file_copy_result = true;
                                        }
                                    }
                                    else
                                    {
                                        if (FileCopy(TemplateDir + "pom.xml.cs.template", DstPath + "pom.xml"))
                                        {
                                            Loggy.Info(String.Format("Generated pom.targets, pom.props and pom.xml files"));
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
                            return False();
                        }
                    }
                    else
                    {
                        Loggy.Error(String.Format("Error: Action {0} failed in Package::Construct since 'Name' was not specified", Action));
                        return False();
                    }
                }
                else
                {
                    Loggy.Error(String.Format("Error: Action {0} is not recognized by Package::Construct (Available actions: Init, Dir, MsDev2010)", Action));
                    return False();
                }

                return True();
            }
            catch (Exception e)
            {
                Loggy.Error("Exception: " + e.Message);
                return False();
            }
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
