using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using System;
using System.IO;
using System.Collections.Generic;
using System.Security.Cryptography;
using System.Linq;
using System.Text;
using System.Runtime;
using Ionic.Zip;
using Ionic.Zlib;
using MSBuild.XCode.Helpers;

namespace MSBuild.XCode
{
    public class PackageCreate : Task
    {
        [Required]
        public string RootDir { get; set; }
        [Required]
        public string Platform { get; set; }

        [Output]
        public string Filename { get; set; }

        public override bool Execute()
        {
            bool success = false;
            Loggy.TaskLogger = Log;

            if (String.IsNullOrEmpty(Platform))
                Platform = "Win32";

            RootDir = RootDir.EndWith('\\');

            Environment.CurrentDirectory = RootDir;

            // - Verify that there are no local changes 
            // - Verify that there are no outgoing changes
            Mercurial.Repository hg_repo = new Mercurial.Repository(RootDir);
            if (!hg_repo.Exists)
            {
                Loggy.Error(String.Format("Error: Package::Create failed since there is no Hg (Mercurial) repository!"));
                return false;
            }
            if (hg_repo.HasOutstandingChanges)
            {
                Loggy.Error(String.Format("Error: Package::Create failed since there are still outstanding (non commited) changes!"));
                return false;
            }

            Mercurial.Changeset hg_changeset = hg_repo.GetChangeSet();

            // Write a vcs.info file containing VCS information, this will be included in the package
            dynamic x = new MSBuild.XCode.Helpers.Xml();
            x.Vcs(MSBuild.XCode.Helpers.Xml.Fragment(u => 
            { 
                u.Type("Hg"); 
                u.Branch(hg_changeset.Branch);
                u.Revision(hg_changeset.Hash);
                u.AuthorName(hg_changeset.AuthorName);
                u.AuthorEmail(hg_changeset.AuthorEmailAddress);
            }));

            PackageInstance package = PackageInstance.LoadFromRoot(RootDir);
            // Write a dependency.info file containing dependency package info, this will be included in the package
            if (package.IsValid)
            {
                package.Branch = hg_changeset.Branch;
                package.Platform = Platform;

                string buildURL = PackageInstance.RootDir + "target\\" + package.Name + "\\build\\" + Platform + "\\";
                if (!Directory.Exists(buildURL))
                    Directory.CreateDirectory(buildURL);

                using (FileStream fs = new FileStream(buildURL + "vcs.info", FileMode.Create))
                {
                    using (StreamWriter sw = new StreamWriter(fs))
                    {
                        sw.Write(x.ToString(true));
                        sw.Close();
                        fs.Close();
                    }
                }

                PackageRepositoryLocal localRepo = new PackageRepositoryLocal(RootDir);
                localRepo.Update(package);

                if (localRepo.Add(package, ELocation.Root))
                {
                    IPackageFilename filename = package.LocalFilename;
                    Filename = filename.ToString();
                    success = true;
                }
                else
                {
                    Filename = string.Empty;
                    success = false;
                }
            }

            return success;
        }
    }
}
