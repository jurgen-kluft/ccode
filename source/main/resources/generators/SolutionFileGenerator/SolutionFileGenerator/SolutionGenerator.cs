using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;

namespace SolutionFileGenerator
{
    class SolutionGenerator
    {
        public enum EVersion
        {
            VS2010,
        }
        public enum ELanguage
        {
            CS,
            CPP,
        }

        private string mRootDir = string.Empty;
        private List<FileSystemInfo> m_Projects;
        private EVersion mVersion = EVersion.VS2010;
        private ELanguage mLanguage = ELanguage.CS;

        private List<string> m_Configs;

        public SolutionGenerator(EVersion version, ELanguage language)
        {
            mVersion = version;
            mLanguage = language;
            m_Projects = new List<FileSystemInfo>();

            m_Configs = new List<string>();
        }

        private string ProjectTypeGuid()
        {
            string guid;
            switch (mLanguage)
            {
                case ELanguage.CS:
                    {
                        guid = "FAE04EC0-301F-11D3-BF4B-00C04F79EFBC";
                    }
                    break;
                default:
                case ELanguage.CPP:
                    {
                        guid = "8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942";
                    }
                    break;
            }
            return guid;
        }

        private void WriteHeader(StreamWriter writer)
        {
            switch (mVersion)
            {
                default:
                case EVersion.VS2010:
                    {
                        writer.WriteLine("Microsoft Visual Studio Solution File, Format Version 11.00");
                        writer.WriteLine("# Visual Studio 2010");
                    }
                    break;
            }
        }

        private void WriteGlobalHeader(StreamWriter writer)
        {
            writer.WriteLine("Global");
        }
        private void WriteGlobalFooter(StreamWriter writer)
        {
            writer.WriteLine("EndGlobal");
        }

        private void WriteProjects(StreamWriter writer)
        {
            switch (mLanguage)
            {
                case ELanguage.CS:
                    {
                        foreach (FileSystemInfo project in m_Projects)
                        {
                            Guid projectGuid = GetProjectGuid(project);
                            writer.Write(string.Format(@"Project(""{{{0}}}"") = ", ProjectTypeGuid()));
                            writer.WriteLine(string.Format(@"""{0}"", ""{1}"", ""{{{2}}}""",
                                project.Name.Substring(0, project.Name.Length - project.Extension.Length),
                                GetRelativePath(mRootDir, project.FullName),
                                projectGuid.ToString().ToUpper()));
                            writer.WriteLine("EndProject");
                        }
                    } break;
                case ELanguage.CPP:
                    {
                        foreach (FileSystemInfo project in m_Projects)
                        {
                            Guid projectGuid = GetProjectGuid(project);
                            writer.Write(string.Format(@"Project(""{{{0}}}"") = ", ProjectTypeGuid()));
                            writer.WriteLine(string.Format(@"""{0}"", ""{1}"", ""{{{2}}}""",
                                project.Name.Substring(0, project.Name.Length - project.Extension.Length),
                                GetRelativePath(mRootDir, project.FullName),
                                projectGuid.ToString().ToUpper()));

                            // TODO: write dependencies

                            writer.WriteLine("EndProject");
                        }
                    } break;
            }
        }

        enum EGlobalSection
        {
            Solution,
            Project,
            Properties,
        }

        private void WriteGlobalSection(EGlobalSection _GlobalSection, StreamWriter writer)
        {
            if (mLanguage == ELanguage.CS)
            {
                switch (_GlobalSection)
                {
                    case EGlobalSection.Solution:
                        {
                            writer.WriteLine("\tGlobalSection(SolutionConfigurationPlatforms) = preSolution");
                            foreach (string c in m_Configs)
                            {
                                writer.WriteLine("\t\t" + c + " = " + c);
                            }
                            writer.WriteLine("\tEndGlobalSection");
                        } break;
                    case EGlobalSection.Project:
                        {
                            writer.WriteLine("\tGlobalSection(ProjectConfigurationPlatforms) = postSolution");

                            // TODO Write configurations
                            foreach (FileSystemInfo project in m_Projects)
                            {
                                Guid projectGuid = GetProjectGuid(project);
                                foreach (string c in m_Configs)
                                {
                                    writer.WriteLine("\t\t{" + projectGuid + "}." + c + ".ActiveCfg = " + c);
                                    writer.WriteLine("\t\t{" + projectGuid + "}." + c + ".Build.0 = " + c);
                                }
                            }

                            writer.WriteLine("\tEndGlobalSection");
                        } break;
                    case EGlobalSection.Properties:
                        {
                            writer.WriteLine("\tGlobalSection(SolutionProperties) = preSolution");
                            writer.WriteLine("\t\tHideSolutionNode = FALSE");
                            writer.WriteLine("\tEndGlobalSection");
                        } break;
                }
            }
            else if (mLanguage == ELanguage.CPP)
            {
                switch (_GlobalSection)
                {
                    case EGlobalSection.Solution:
                        {
                            writer.WriteLine("\tGlobalSection(SolutionConfigurationPlatforms) = preSolution");
                            foreach (string c in m_Configs)
                            {
                                writer.WriteLine("\t\t" + c + " = " + c);
                            }
                            writer.WriteLine("\tEndGlobalSection");
                        } break;
                    case EGlobalSection.Project:
                        {
                            writer.WriteLine("\tGlobalSection(ProjectConfigurationPlatforms) = postSolution");

                            // TODO Write configurations
                            foreach (FileSystemInfo project in m_Projects)
                            {
                                Guid projectGuid = GetProjectGuid(project);
                                foreach (string c in m_Configs)
                                {
                                    writer.WriteLine("\t\t{" + projectGuid + "}." + c + ".ActiveCfg = " + c);
                                    writer.WriteLine("\t\t{" + projectGuid + "}." + c + ".Build.0 = " + c);
                                }
                            }

                            writer.WriteLine("\tEndGlobalSection");
                        } break;
                    case EGlobalSection.Properties:
                        {
                            writer.WriteLine("\tGlobalSection(SolutionProperties) = preSolution");
                            writer.WriteLine("\t\tHideSolutionNode = FALSE");
                            writer.WriteLine("\tEndGlobalSection");
                        } break;
                }
            }
        }

        public int Execute(string _SolutionFile, List<string> _ProjectFiles)
        {
            mRootDir = Path.GetDirectoryName(_SolutionFile);

            foreach (string projectFilename in _ProjectFiles)
            {
                FileInfo fi = new FileInfo(projectFilename);
                if (fi.Exists)
                    m_Projects.Add(fi);
            }

            // Analyze the configurations
            Dictionary<string, bool> sln_configs = new Dictionary<string, bool>();
            foreach (FileSystemInfo project in m_Projects)
            {
                Dictionary<string,bool> project_configs = GetProjectConfigs(project);
                foreach (KeyValuePair<string, bool> p in project_configs)
                {
                    if (!sln_configs.ContainsKey(p.Key))
                    {
                        sln_configs.Add(p.Key, true);
                    }
                }
            }
            foreach (KeyValuePair<string, bool> p in sln_configs)
            {
                m_Configs.Add(p.Key);
            }

            using (StreamWriter writer = File.CreateText(_SolutionFile))
            {
                WriteHeader(writer);
                WriteProjects(writer);
                WriteGlobalHeader(writer);
                
                // These 2 sections are generated by visual studio, however we need them to have msbuild be able to build.
                WriteGlobalSection(EGlobalSection.Solution, writer);
                WriteGlobalSection(EGlobalSection.Project, writer);

                WriteGlobalSection(EGlobalSection.Properties, writer);
                WriteGlobalFooter(writer);
            }

            return m_Projects.Count;
        }

        private Guid GetProjectGuid(FileSystemInfo file)
        {
            switch (mLanguage)
            {
                case ELanguage.CS:
                    {
                        using (StreamReader reader = File.OpenText(file.FullName))
                        {
                            string text = reader.ReadToEnd();
                            string pattern = "<ProjectGuid>";
                            int start = text.IndexOf(pattern);
                            if (start > 0)
                            {
                                start += pattern.Length;
                                pattern = "</ProjectGuid>";
                                int end = text.IndexOf(pattern);
                                if (end > 0)
                                {
                                    return new Guid(text.Substring(start + 1, end - start - 2));
                                }
                            }
                        }
                    } break;
                case ELanguage.CPP:
                    {
                        if (mVersion == EVersion.VS2010)
                        {
                            using (StreamReader reader = File.OpenText(file.FullName))
                            {
                                string text = reader.ReadToEnd();
                                string pattern = "<ProjectGuid>";
                                int start = text.IndexOf(pattern);
                                if (start > 0)
                                {
                                    start += pattern.Length;
                                    pattern = "</ProjectGuid>";
                                    int end = text.IndexOf(pattern);
                                    if (end > 0)
                                    {
                                        return new Guid(text.Substring(start + 1, end - start - 2));
                                    }
                                }
                            }
                        }
                        else
                        {
                            using (StreamReader reader = File.OpenText(file.FullName))
                            {
                                string text = reader.ReadToEnd();
                                string pattern = "ProjectGUID=\"";
                                int start = text.IndexOf(pattern);
                                if (start > 0)
                                {
                                    start += pattern.Length;
                                    pattern = "\" ";
                                    int end = text.IndexOf(pattern, start);
                                    if (end > 0)
                                    {
                                        return new Guid(text.Substring(start + 1, end - start - 2));
                                    }
                                }
                            }
                        }
                    } break;
            }

            return Guid.Empty;
        }

        private Dictionary<string, bool> GetProjectConfigs(FileSystemInfo file)
        {
            Dictionary<string, bool> configs = new Dictionary<string, bool>();

            switch (mLanguage)
            {
                case ELanguage.CS:
                    {
                        using (StreamReader reader = File.OpenText(file.FullName))
                        {
                            string text = reader.ReadToEnd();
                            int cursor = 0;
                            while (true)
                            {
                                string pattern = "$(Configuration)|$(Platform)";
                                cursor = text.IndexOf(pattern, cursor);
                                if (cursor > 0)
                                {
                                    cursor += pattern.Length;
                                    pattern = "\">";
                                    int end = text.IndexOf(pattern, cursor);
                                    if (end > 0)
                                    {
                                        string config = text.Substring(cursor + 1, end - cursor - 2).Trim();
                                        config = config.Replace("==", "");
                                        config = config.Replace("'", "");
                                        config = config.Trim();
                                        if (!configs.ContainsKey(config))
                                            configs.Add(config, true);
                                    }
                                    cursor = end + pattern.Length;
                                }
                                else
                                {
                                    break;
                                }
                            }
                        }
                    } break;
                case ELanguage.CPP:
                    {
                        using (StreamReader reader = File.OpenText(file.FullName))
                        {
                            string text = reader.ReadToEnd();
                            int cursor = 0;
                            while (true)
                            {
                                string pattern = "ProjectConfiguration Include=\"";
                                cursor = text.IndexOf(pattern, cursor);
                                if (cursor > 0)
                                {
                                    cursor += pattern.Length;
                                    pattern = "\">";
                                    int end = text.IndexOf(pattern, cursor);
                                    if (end > 0)
                                    {
                                        string config = text.Substring(cursor, end - cursor).Trim();
                                        if (!configs.ContainsKey(config))
                                            configs.Add(config, true);
                                    }
                                    cursor = end + pattern.Length;
                                }
                                else
                                {
                                    break;
                                }
                            }
                        }
                    } break;
            }

            return configs;
        }

        private string GetRelativePath(string rootDirPath, string absoluteFilePath)
        {
            string[] firstPathParts = rootDirPath.Trim(Path.DirectorySeparatorChar).Split(Path.DirectorySeparatorChar);
            string[] secondPathParts = absoluteFilePath.Trim(Path.DirectorySeparatorChar).Split(Path.DirectorySeparatorChar);

            int sameCounter = 0;
            for (int i = 0; i < Math.Min(firstPathParts.Length,
            secondPathParts.Length); i++)
            {
                if (!firstPathParts[i].ToLower().Equals(secondPathParts[i].ToLower()))
                {
                    break;
                }
                sameCounter++;
            }

            if (sameCounter == 0)
            {
                return absoluteFilePath;
            }

            string newPath = String.Empty;
            for (int i = sameCounter; i < firstPathParts.Length; i++)
            {
                if (i > sameCounter)
                {
                    newPath += Path.DirectorySeparatorChar;
                }
                newPath += "..";
            }
            if (newPath.Length == 0)
            {
                newPath = ".";
            }
            for (int i = sameCounter; i < secondPathParts.Length; i++)
            {
                newPath += Path.DirectorySeparatorChar;
                newPath += secondPathParts[i];
            }
            return newPath;
        }
    }
}
