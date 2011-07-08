namespace xcode
{
    partial class Main
    {
        /// <summary>
        /// Required designer variable.
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        /// Clean up any resources being used.
        /// </summary>
        /// <param name="disposing">true if managed resources should be disposed; otherwise, false.</param>
        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows Form Designer generated code

        /// <summary>
        /// Required method for Designer support - do not modify
        /// the contents of this method with the code editor.
        /// </summary>
        private void InitializeComponent()
        {
            this.btInstall = new System.Windows.Forms.Button();
            this.tbLocalRepo = new System.Windows.Forms.TextBox();
            this.btLocalRepo = new System.Windows.Forms.Button();
            this.btRemoteRepo = new System.Windows.Forms.Button();
            this.tbRemoteRepo = new System.Windows.Forms.TextBox();
            this.rtbLog = new System.Windows.Forms.RichTextBox();
            this.groupBox1 = new System.Windows.Forms.GroupBox();
            this.btVerify = new System.Windows.Forms.Button();
            this.groupBox2 = new System.Windows.Forms.GroupBox();
            this.pictureBox1 = new System.Windows.Forms.PictureBox();
            this.btExit = new System.Windows.Forms.Button();
            this.mFolderBrowser = new System.Windows.Forms.FolderBrowserDialog();
            this.btLocalWorkDir = new System.Windows.Forms.Button();
            this.tbLocalWorkDir = new System.Windows.Forms.TextBox();
            this.tbXCodeRepo = new System.Windows.Forms.TextBox();
            this.btXCodeRepoUrl = new System.Windows.Forms.Button();
            this.groupBox1.SuspendLayout();
            this.groupBox2.SuspendLayout();
            ((System.ComponentModel.ISupportInitialize)(this.pictureBox1)).BeginInit();
            this.SuspendLayout();
            // 
            // btInstall
            // 
            this.btInstall.Location = new System.Drawing.Point(6, 48);
            this.btInstall.Name = "btInstall";
            this.btInstall.Size = new System.Drawing.Size(159, 23);
            this.btInstall.TabIndex = 0;
            this.btInstall.Text = "Install";
            this.btInstall.UseVisualStyleBackColor = true;
            this.btInstall.Click += new System.EventHandler(this.btInstall_Click);
            // 
            // tbLocalRepo
            // 
            this.tbLocalRepo.Location = new System.Drawing.Point(189, 43);
            this.tbLocalRepo.Name = "tbLocalRepo";
            this.tbLocalRepo.Size = new System.Drawing.Size(484, 20);
            this.tbLocalRepo.TabIndex = 1;
            this.tbLocalRepo.Text = "D:\\Packages\\PACKAGE_REPO\\";
            // 
            // btLocalRepo
            // 
            this.btLocalRepo.Location = new System.Drawing.Point(12, 41);
            this.btLocalRepo.Name = "btLocalRepo";
            this.btLocalRepo.Size = new System.Drawing.Size(170, 23);
            this.btLocalRepo.TabIndex = 2;
            this.btLocalRepo.Text = "Local Package Repository...";
            this.btLocalRepo.UseVisualStyleBackColor = true;
            this.btLocalRepo.Click += new System.EventHandler(this.btLocalRepo_Click);
            // 
            // btRemoteRepo
            // 
            this.btRemoteRepo.Location = new System.Drawing.Point(13, 12);
            this.btRemoteRepo.Name = "btRemoteRepo";
            this.btRemoteRepo.Size = new System.Drawing.Size(170, 23);
            this.btRemoteRepo.TabIndex = 3;
            this.btRemoteRepo.Text = "Remote Package Repository...";
            this.btRemoteRepo.UseVisualStyleBackColor = true;
            this.btRemoteRepo.Click += new System.EventHandler(this.btRemoteRepo_Click);
            // 
            // tbRemoteRepo
            // 
            this.tbRemoteRepo.Location = new System.Drawing.Point(189, 14);
            this.tbRemoteRepo.Name = "tbRemoteRepo";
            this.tbRemoteRepo.Size = new System.Drawing.Size(484, 20);
            this.tbRemoteRepo.TabIndex = 4;
            this.tbRemoteRepo.Text = "\\\\cnshasap2\\Hg_Repo\\PACKAGE_REPO\\";
            // 
            // rtbLog
            // 
            this.rtbLog.BackColor = System.Drawing.SystemColors.ControlLightLight;
            this.rtbLog.ForeColor = System.Drawing.Color.Green;
            this.rtbLog.Location = new System.Drawing.Point(6, 19);
            this.rtbLog.Name = "rtbLog";
            this.rtbLog.ReadOnly = true;
            this.rtbLog.Size = new System.Drawing.Size(472, 215);
            this.rtbLog.TabIndex = 5;
            this.rtbLog.Text = "";
            this.rtbLog.WordWrap = false;
            // 
            // groupBox1
            // 
            this.groupBox1.Controls.Add(this.rtbLog);
            this.groupBox1.Location = new System.Drawing.Point(189, 139);
            this.groupBox1.Name = "groupBox1";
            this.groupBox1.Size = new System.Drawing.Size(484, 241);
            this.groupBox1.TabIndex = 6;
            this.groupBox1.TabStop = false;
            this.groupBox1.Text = "Log";
            // 
            // btVerify
            // 
            this.btVerify.Location = new System.Drawing.Point(6, 19);
            this.btVerify.Name = "btVerify";
            this.btVerify.Size = new System.Drawing.Size(159, 23);
            this.btVerify.TabIndex = 7;
            this.btVerify.Text = "Verify";
            this.btVerify.UseVisualStyleBackColor = true;
            this.btVerify.Click += new System.EventHandler(this.btVerify_Click);
            // 
            // groupBox2
            // 
            this.groupBox2.Controls.Add(this.pictureBox1);
            this.groupBox2.Controls.Add(this.btExit);
            this.groupBox2.Controls.Add(this.btInstall);
            this.groupBox2.Controls.Add(this.btVerify);
            this.groupBox2.Location = new System.Drawing.Point(12, 140);
            this.groupBox2.Name = "groupBox2";
            this.groupBox2.Size = new System.Drawing.Size(171, 240);
            this.groupBox2.TabIndex = 7;
            this.groupBox2.TabStop = false;
            this.groupBox2.Text = "Actions";
            // 
            // pictureBox1
            // 
            this.pictureBox1.Image = global::xcode.Properties.Resources.XCode_Large;
            this.pictureBox1.Location = new System.Drawing.Point(7, 78);
            this.pictureBox1.Name = "pictureBox1";
            this.pictureBox1.Size = new System.Drawing.Size(158, 126);
            this.pictureBox1.SizeMode = System.Windows.Forms.PictureBoxSizeMode.StretchImage;
            this.pictureBox1.TabIndex = 9;
            this.pictureBox1.TabStop = false;
            // 
            // btExit
            // 
            this.btExit.DialogResult = System.Windows.Forms.DialogResult.Cancel;
            this.btExit.Location = new System.Drawing.Point(6, 210);
            this.btExit.Name = "btExit";
            this.btExit.Size = new System.Drawing.Size(159, 23);
            this.btExit.TabIndex = 8;
            this.btExit.Text = "Exit";
            this.btExit.UseVisualStyleBackColor = true;
            this.btExit.Click += new System.EventHandler(this.btExit_Click);
            // 
            // btLocalWorkDir
            // 
            this.btLocalWorkDir.Location = new System.Drawing.Point(13, 70);
            this.btLocalWorkDir.Name = "btLocalWorkDir";
            this.btLocalWorkDir.Size = new System.Drawing.Size(170, 23);
            this.btLocalWorkDir.TabIndex = 8;
            this.btLocalWorkDir.Text = "Local Working Directory...";
            this.btLocalWorkDir.UseVisualStyleBackColor = true;
            this.btLocalWorkDir.Click += new System.EventHandler(this.btLocalWorkDir_Click);
            // 
            // tbLocalWorkDir
            // 
            this.tbLocalWorkDir.Location = new System.Drawing.Point(189, 72);
            this.tbLocalWorkDir.Name = "tbLocalWorkDir";
            this.tbLocalWorkDir.Size = new System.Drawing.Size(484, 20);
            this.tbLocalWorkDir.TabIndex = 9;
            this.tbLocalWorkDir.Text = "D:\\Packages";
            // 
            // tbXCodeRepo
            // 
            this.tbXCodeRepo.Location = new System.Drawing.Point(189, 101);
            this.tbXCodeRepo.Name = "tbXCodeRepo";
            this.tbXCodeRepo.Size = new System.Drawing.Size(484, 20);
            this.tbXCodeRepo.TabIndex = 11;
            this.tbXCodeRepo.Text = "\\\\cnshasap2\\Hg_Repo\\PACKAGE_REPO\\com\\virtuos\\xcode\\publish\\";
            // 
            // btXCodeRepoUrl
            // 
            this.btXCodeRepoUrl.Location = new System.Drawing.Point(13, 99);
            this.btXCodeRepoUrl.Name = "btXCodeRepoUrl";
            this.btXCodeRepoUrl.Size = new System.Drawing.Size(170, 23);
            this.btXCodeRepoUrl.TabIndex = 10;
            this.btXCodeRepoUrl.Text = "XCode Repository...";
            this.btXCodeRepoUrl.UseVisualStyleBackColor = true;
            // 
            // Main
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.CancelButton = this.btExit;
            this.ClientSize = new System.Drawing.Size(680, 393);
            this.Controls.Add(this.tbXCodeRepo);
            this.Controls.Add(this.btXCodeRepoUrl);
            this.Controls.Add(this.tbLocalWorkDir);
            this.Controls.Add(this.btLocalWorkDir);
            this.Controls.Add(this.groupBox2);
            this.Controls.Add(this.groupBox1);
            this.Controls.Add(this.tbRemoteRepo);
            this.Controls.Add(this.btRemoteRepo);
            this.Controls.Add(this.btLocalRepo);
            this.Controls.Add(this.tbLocalRepo);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedToolWindow;
            this.KeyPreview = true;
            this.Name = "Main";
            this.Text = "..::XCode::..";
            this.groupBox1.ResumeLayout(false);
            this.groupBox2.ResumeLayout(false);
            ((System.ComponentModel.ISupportInitialize)(this.pictureBox1)).EndInit();
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.Button btInstall;
        private System.Windows.Forms.TextBox tbLocalRepo;
        private System.Windows.Forms.Button btLocalRepo;
        private System.Windows.Forms.Button btRemoteRepo;
        private System.Windows.Forms.TextBox tbRemoteRepo;
        private System.Windows.Forms.RichTextBox rtbLog;
        private System.Windows.Forms.GroupBox groupBox1;
        private System.Windows.Forms.Button btVerify;
        private System.Windows.Forms.GroupBox groupBox2;
        private System.Windows.Forms.FolderBrowserDialog mFolderBrowser;
        private System.Windows.Forms.Button btExit;
        private System.Windows.Forms.Button btLocalWorkDir;
        private System.Windows.Forms.TextBox tbLocalWorkDir;
        private System.Windows.Forms.PictureBox pictureBox1;
        private System.Windows.Forms.TextBox tbXCodeRepo;
        private System.Windows.Forms.Button btXCodeRepoUrl;
    }
}

