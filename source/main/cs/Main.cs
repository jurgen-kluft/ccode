using System;
using System.IO;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Windows.Forms;

namespace xcode
{
    public partial class Main : Form
    {
        public Main()
        {
            Bitmap bitmap = xcode.Properties.Resources.Xcode_icon;
            IntPtr Hicon = bitmap.GetHicon();
            Icon = Icon.FromHandle(Hicon);

            InitializeComponent();
        }

        private void ClearLog()
        {
            rtbLog.Clear();
        }

        private void AddToLog(string text, Color color)
        {
            int start = rtbLog.Text.Length;
            rtbLog.AppendText(text);
            int end = rtbLog.Text.Length;
            rtbLog.Select(start, end - start);
            rtbLog.SelectionColor = color;
        }

        private void btLocalRepo_Click(object sender, EventArgs e)
        {
            if (mFolderBrowser.ShowDialog() == DialogResult.OK)
            {
                this.tbLocalRepo.Text = mFolderBrowser.SelectedPath;
            }
        }

        private void btRemoteRepo_Click(object sender, EventArgs e)
        {
            if (mFolderBrowser.ShowDialog() == DialogResult.OK)
            {
                this.tbRemoteRepo.Text = mFolderBrowser.SelectedPath;
            }
        }


        private void btLocalWorkDir_Click(object sender, EventArgs e)
        {
            if (mFolderBrowser.ShowDialog() == DialogResult.OK)
            {
                this.tbLocalWorkDir.Text = mFolderBrowser.SelectedPath;
            }
        }

        private void btExit_Click(object sender, EventArgs e)
        {
            this.Close();
        }

        private void btVerify_Click(object sender, EventArgs e)
        {
            ClearLog();

            bool remoteRepoExists = Directory.Exists(this.tbRemoteRepo.Text);
            bool cacheRepoExists = Directory.Exists(this.tbLocalRepo.Text);

            SysInfo sysInfo = new SysInfo();
            sysInfo.Collect();

            if (!sysInfo.DotNet4IsInstalled)
            {
                AddToLog(".NET framework version 4.0 is not installed\n", Color.Red);
            }
            else
            {
                AddToLog(".NET framework version 4.0 is installed\n", Color.Green);
            }

            if (!sysInfo.MercurialInstalled)
            {
                AddToLog("Mercurial not found or version to old\n", Color.Red);
            }
            else
            {
                AddToLog("Mercurial located and version validated\n", Color.Green);
            }

            if (!sysInfo.MsBuildInstalled)
            {
                AddToLog("Unable to locate MsBuild.exe version 4.0\n", Color.Red);
            }
            else
            {
                AddToLog("MsBuild.exe located succesfully\n", Color.Green);
            }

            if (!remoteRepoExists)
            {
                AddToLog("Remote package repository (cache) doesn't exist\n", Color.Red);
            }
            else
            {
                AddToLog("Remote package repository present\n", Color.Green);
            }

            if (!cacheRepoExists)
            {
                AddToLog("Local package repository (cache) doesn't exist\n", Color.Red);
            }
            else
            {
                DriveInfo drive = new DriveInfo(this.tbLocalRepo.Text.Substring(0, 1));
                AddToLog("Local package repository present\n", Color.Green);
                if (drive.AvailableFreeSpace < (100 * 1024 * 1024))
                {
                    AddToLog("Local package repository, not enough space (<100MB)\n", Color.Red);
                }
            }
        }

        private void btInstall_Click(object sender, EventArgs ea)
        {
            ClearLog();
            AddToLog("Installing...\n", Color.Black);

            try
            {
                AddToLog("Initializing local package repository (cache) at " + this.tbLocalRepo.Text + "\n", Color.DarkGreen);
                
            }
            catch (Exception e)
            {
                AddToLog("Unable to initialize local package repository (cache) at " + this.tbLocalRepo.Text + " (reason: " + e.Message + ")\n", Color.DarkRed);
            }

            try
            {
                AddToLog("Validating remote package repository at " + this.tbRemoteRepo.Text + "\n", Color.DarkGreen);
                
            }
            catch (Exception e)
            {
                AddToLog("Failed to validate remote package repository at " + this.tbRemoteRepo.Text + " (reason: " + e.Message + ")\n", Color.DarkRed);
            }

            // Version 2
            // Hg Clone
            if (!tbLocalWorkDir.Text.EndsWith("\\"))
                tbLocalWorkDir.Text += "\\";
            if (!tbLocalRepo.Text.EndsWith("\\"))
                tbLocalRepo.Text += "\\";
            if (!tbRemoteRepo.Text.EndsWith("\\"))
                tbRemoteRepo.Text += "\\";
            if (!tbXCodeRepo.Text.EndsWith("\\"))
                tbXCodeRepo.Text += "\\";

            string sub_path = @"com\virtuos\xcode\publish";
            string dst_path = tbLocalRepo.Text;

            if (Directory.Exists(dst_path + sub_path))
            {
                Directory.Delete(dst_path + sub_path);
            }
            Directory.CreateDirectory(dst_path + sub_path);

            Mercurial.Repository hg_repo = new Mercurial.Repository(tbXCodeRepo.Text);
            if (hg_repo.Exists)
            {
                Mercurial.CloneCommand clone_cmd = new Mercurial.CloneCommand();
                clone_cmd.CompressedTransfer = false;
                clone_cmd.Source = tbXCodeRepo.Text.Replace('\\', '/');

                Mercurial.Repository new_hg_repo = new Mercurial.Repository(dst_path + sub_path);
                new_hg_repo.Clone(clone_cmd);
            }
            else
            {
                // Error
            }

            string src_path = tbLocalRepo.Text + @"com\virtuos\xcode\publish\";
            dst_path = tbLocalWorkDir.Text;
            
            if (!Directory.Exists(dst_path))
                Directory.CreateDirectory(dst_path);

            {
                AddToLog("Copy file from file://" + src_path + "templates\\dev.targets.template" + " to file://" + dst_path + "dev.targets" + "\n", Color.Blue);
                File.Copy(src_path + "templates\\dev.targets.template", dst_path + "dev.targets", true);
                AddToLog("Copy file from file://" + src_path + "templates\\dev.props.template" + " to file://" + dst_path + "dev.props" + "\n", Color.Blue);
                FileCopy(src_path + "templates\\dev.props.template", dst_path + "dev.props", tbLocalRepo.Text, tbRemoteRepo.Text, tbXCodeRepo.Text);
            }

            AddToLog("Done -----\n", Color.Black);

            // Copy 
        }

        private bool FileCopy(string srcfile, string dstfile, string cacheRepoDir, string remoteRepoDir, string xcodeRepoDir)
        {
            string[] lines = File.ReadAllLines(srcfile);

            using (FileStream wfs = new FileStream(dstfile, FileMode.Create, FileAccess.Write))
            {
                using (StreamWriter writer = new StreamWriter(wfs))
                {
                    foreach (string line in lines)
                    {
                        string l = line.Replace("${CacheRepoRoot}", cacheRepoDir);
                        l = l.Replace("${RemoteRepoRoot}", remoteRepoDir);
                        l = l.Replace("${XCodeRepoRoot}", xcodeRepoDir);
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
