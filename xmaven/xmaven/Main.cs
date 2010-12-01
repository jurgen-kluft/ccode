using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Windows.Forms;

namespace xmaven
{
    public partial class Main : Form
    {
        public Main()
        {
            InitializeComponent();
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

        private void btExit_Click(object sender, EventArgs e)
        {
            this.Close();
        }

        private void btVerify_Click(object sender, EventArgs e)
        {
            AddToLog("Verifying...\n", Color.Black);

            string local_repo_envvar = Environment.GetEnvironmentVariable("SCM_LOCAL_REPO_ROOT");
            string remote_repo_envvar = Environment.GetEnvironmentVariable("SCM_REMOTE_REPO_ROOT");

            if (String.IsNullOrEmpty(local_repo_envvar))
            {
                AddToLog("Environment variable \"SCM_LOCAL_REPO_ROOT\" doesn't exist on this machine!\n", Color.Red);
            }
            else
            {
                AddToLog("Environment variable \"SCM_LOCAL_REPO_ROOT\" on this machine is set to " + Environment.GetEnvironmentVariable("SCM_LOCAL_REPO_ROOT") + "!\n", Color.DarkGreen);
            }

            if (String.IsNullOrEmpty(remote_repo_envvar))
            {
                AddToLog("Environment variable \"SCM_REMOTE_REPO_ROOT\" doesn't exist on this machine!\n", Color.Red);
            }
            else
            {
                AddToLog("Environment variable \"SCM_REMOTE_REPO_ROOT\" on this machine is set to " + Environment.GetEnvironmentVariable("SCM_REMOTE_REPO_ROOT") + "!\n", Color.DarkGreen);
            }
            AddToLog("Done---\n", Color.Black);
        }

        private void btInstall_Click(object sender, EventArgs ea)
        {
            AddToLog("Installing...\n", Color.Black);

            try
            {
                AddToLog("Trying to set environment variable \"SCM_LOCAL_REPO_ROOT\" to " + this.tbLocalRepo.Text + "\n", Color.DarkGreen);
                Environment.SetEnvironmentVariable("SCM_LOCAL_REPO_ROOT", this.tbLocalRepo.Text, EnvironmentVariableTarget.User);
            }
            catch (Exception e)
            {
                AddToLog("Unable to set environment variable \"SCM_LOCAL_REPO_ROOT\" to " + this.tbLocalRepo.Text + " (reason: " + e.Message + ")\n", Color.DarkRed);
            }

            try
            {
                AddToLog("Trying to set environment variable \"SCM_REMOTE_REPO_ROOT\" to " + this.tbRemoteRepo.Text + "\n", Color.DarkGreen);
                Environment.SetEnvironmentVariable("SCM_REMOTE_REPO_ROOT", this.tbRemoteRepo.Text, EnvironmentVariableTarget.User);
            }
            catch (Exception e)
            {
                AddToLog("Unable to set environment variable \"SCM_REMOTE_REPO_ROOT\" to " + this.tbRemoteRepo.Text + " (reason: " + e.Message + ")\n", Color.DarkRed);
            }
            AddToLog("Done -----\n", Color.Black);


            // Copy 
        }
    }
}
