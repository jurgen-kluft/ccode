using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode.MsDev
{
    class CppSolution : ISolution
    {
        public enum EVersion
        {
            VS2010,
        }

        private string mRootDir = string.Empty;
        private List<FileSystemInfo> m_Projects;
        private EVersion mVersion = EVersion.VS2010;

        private Dictionary<string, HashSet<string>> m_Configs;
        private Dictionary<string, Guid> m_ProjectGuids;

        public CppSolution(EVersion version)
        {
            mVersion = version;
            m_Projects = new List<FileSystemInfo>();
            m_Configs = new Dictionary<string, HashSet<string>>();
        }

        private string ProjectTypeGuid()
        {
            string guid;
            guid = "8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942";
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
            foreach (FileSystemInfo project in m_Projects)
            {
                Guid projectGuid = GetProjectGuid(project);
                writer.Write(string.Format(@"Project(""{{{0}}}"") = ", ProjectTypeGuid()));
                writer.WriteLine(string.Format(@"""{0}"", ""{1}"", ""{{{2}}}""",
                    project.Name.Substring(0, project.Name.Length - project.Extension.Length),
                    GetRelativePath(mRootDir, project.FullName),
                    projectGuid.ToString().ToUpper()));

                // TODO: write dependencies
                /// ProjectSection(ProjectDependencies) = postProject
                /// 	{62068C48-9011-4E7F-A282-1D0F91023756} = {62068C48-9011-4E7F-A282-1D0F91023756}
                /// EndProjectSection
                string[] dependencyProjectFiles;
                if (mProjectDependencies.TryGetValue(project.Name, out dependencyProjectFiles))
                {
                    writer.WriteLine(string.Format("\tProjectSection(ProjectDependencies) = postProject"));
                    foreach (string projectFile in dependencyProjectFiles)
                    {
                        FileSystemInfo info = GetProjectByName(projectFile);
                        if (info != null)
                        {
                            Guid guid = GetProjectGuid(info);
                            writer.WriteLine("\t\t{{{0}}} = {{{0}}}", guid.ToString().ToUpper());
                        }
                    }
                    writer.WriteLine("\tEndProjectSection");
                }

                writer.WriteLine("EndProject");
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
            switch (_GlobalSection)
            {
                case EGlobalSection.Solution:
                    {
                        writer.WriteLine("\tGlobalSection(SolutionConfigurationPlatforms) = preSolution");
                        foreach (KeyValuePair<string, HashSet<string>> p in m_Configs)
                        {
                            string c = p.Key;
                            writer.WriteLine("\t\t" + c + " = " + c);
                        }
                        writer.WriteLine("\tEndGlobalSection");
                    } break;
                case EGlobalSection.Project:
                    {
                        writer.WriteLine("\tGlobalSection(ProjectConfigurationPlatforms) = postSolution");

                        foreach (FileSystemInfo project in m_Projects)
                        {
                            Guid projectGuid = GetProjectGuid(project);
                            foreach (KeyValuePair<string, HashSet<string>> p in m_Configs)
                            {
                                string c = p.Key;
                                writer.WriteLine("\t\t{" + projectGuid.ToString().ToUpper() + "}." + c + ".ActiveCfg = " + c);
                                if (p.Value.Contains(project.FullName))
                                    writer.WriteLine("\t\t{" + projectGuid.ToString().ToUpper() + "}." + c + ".Build.0 = " + c);
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

        private Dictionary<string, string[]> mProjectDependencies = new Dictionary<string, string[]>();
        public void AddDependencies(string projectFile, string[] dependencyProjectFiles)
        {
            if (mProjectDependencies.ContainsKey(projectFile))
                mProjectDependencies.Remove(projectFile);

            mProjectDependencies.Add(projectFile, dependencyProjectFiles);
        }

        public int Save(string _SolutionFile, List<string> _ProjectFiles)
        {
            mRootDir = Path.GetDirectoryName(_SolutionFile);
            if (!mRootDir.EndsWith("\\"))
                mRootDir = mRootDir + "\\";
            if (!Directory.Exists(mRootDir))
                return -1;

            m_ProjectGuids = new Dictionary<string, Guid>();
            foreach (string projectFilename in _ProjectFiles)
            {
                FileInfo fi = new FileInfo(mRootDir + projectFilename);
                if (fi.Exists)
                {
                    m_Projects.Add(fi);
                    m_ProjectGuids.Add(fi.FullName, Guid.NewGuid());
                }
            }

            // Analyze the configurations
            Dictionary<string, HashSet<string>> sln_configs = new Dictionary<string, HashSet<string>>();
            foreach (FileSystemInfo project in m_Projects)
            {
                Dictionary<string,bool> project_configs = GetProjectConfigs(project);
                foreach (KeyValuePair<string, bool> p in project_configs)
                {
                    HashSet<string> projects;
                    if (!sln_configs.TryGetValue(p.Key, out projects))
                    {
                        projects = new HashSet<string>();
                        sln_configs.Add(p.Key, projects);
                    }
                    projects.Add(project.FullName);
                }
            }
            foreach (KeyValuePair<string, HashSet<string>> p in sln_configs)
            {
                m_Configs.Add(p.Key, p.Value);
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

        private FileSystemInfo GetProjectByName(string projectFile)
        {
            foreach (FileSystemInfo info in m_Projects)
            {
                if (String.Compare(info.Name, projectFile, true) == 0)
                    return info;
            }
            return null;
        }

        private Guid GetProjectGuid(FileSystemInfo file)
        {
            if (mVersion == EVersion.VS2010)
            {
                Guid guid;
                if (m_ProjectGuids.TryGetValue(file.FullName, out guid))
                    return guid;
                return Guid.NewGuid();
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
            return Guid.Empty;
        }

        private Dictionary<string, bool> GetProjectConfigs(FileSystemInfo file)
        {
            Dictionary<string, bool> configs = new Dictionary<string, bool>();


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
            return configs;
        }

        private string GetRelativePath(string rootDirPath, string absoluteFilePath)
        {
            string[] firstPathParts = rootDirPath.Trim(Path.DirectorySeparatorChar).Split(Path.DirectorySeparatorChar);
            string[] secondPathParts = absoluteFilePath.Trim(Path.DirectorySeparatorChar).Split(Path.DirectorySeparatorChar);

            int sameCounter = 0;
            for (int i = 0; i < System.Math.Min(firstPathParts.Length, secondPathParts.Length); i++)
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
