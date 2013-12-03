using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using System.Xml;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageStructure
    {
        private Dictionary<string, string> mFolders;
        private Dictionary<string, string> mFiles;

        public PackageStructure()
        {
            mFolders = new Dictionary<string, string>();
            mFiles = new Dictionary<string, string>();
        }

        public void AddFolder(string value)
        {
            if (String.IsNullOrEmpty(value))
                return;

            if (!mFolders.ContainsKey(value.ToLower()))
                mFolders.Add(value.ToLower(), value);
        }

        public void AddFile(string value)
        {
            if (String.IsNullOrEmpty(value))
                return;

            if (!mFiles.ContainsKey(value.ToLower()))
                mFiles.Add(value.ToLower(), value);
        }

        public void Create(string rootDir)
        {
            foreach (KeyValuePair<string, string> folder in mFolders)
            {
                if (!Directory.Exists(rootDir + folder.Value))
                    Directory.CreateDirectory(rootDir + folder.Value);
            }

            foreach (KeyValuePair<string, string> file in mFiles)
            {
                if (!File.Exists(rootDir + file.Value))
                    File.Create(rootDir + file.Value).Close();
            }
        }

        public bool Read(XmlNode node, PackageVars vars)
        {
            if (node.Name == "DirectoryStructure")
            {
                if (node.HasChildNodes)
                {
                    foreach (XmlNode child in node.ChildNodes)
                    {
                        string folder = Attribute.Get("Folder", child, string.Empty);
                        if (!String.IsNullOrEmpty(folder))
                        {
                            folder = vars.ReplaceVars(folder);
                            AddFolder(folder);
                        }
                        string file = Attribute.Get("File", child, string.Empty);
                        if (!String.IsNullOrEmpty(file))
                        {
                            file = vars.ReplaceVars(file);
                            AddFile(file);
                        }
                    }
                }
                return true;
            }
            return false;
        }
    }
}