namespace MSBuild.XCode.Helpers
{
    partial class ProgressForm
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
            System.ComponentModel.ComponentResourceManager resources = new System.ComponentModel.ComponentResourceManager(typeof(ProgressForm));
            this.total_progress = new MSBuild.XCode.Helpers.AutoProgress();
            this.tbTotal = new System.Windows.Forms.TextBox();
            this.current_progress = new MSBuild.XCode.Helpers.AutoProgress();
            this.tbCurrent = new System.Windows.Forms.TextBox();
            this.SuspendLayout();
            // 
            // total_progress
            // 
            this.total_progress.Location = new System.Drawing.Point(88, 12);
            this.total_progress.Name = "total_progress";
            this.total_progress.Size = new System.Drawing.Size(645, 20);
            this.total_progress.TabIndex = 0;
            // 
            // tbTotal
            // 
            this.tbTotal.Location = new System.Drawing.Point(13, 12);
            this.tbTotal.Name = "tbTotal";
            this.tbTotal.ReadOnly = true;
            this.tbTotal.Size = new System.Drawing.Size(69, 20);
            this.tbTotal.TabIndex = 1;
            this.tbTotal.TabStop = false;
            this.tbTotal.Text = "Total";
            this.tbTotal.TextAlign = System.Windows.Forms.HorizontalAlignment.Center;
            // 
            // current_progress
            // 
            this.current_progress.Location = new System.Drawing.Point(88, 38);
            this.current_progress.Name = "current_progress";
            this.current_progress.Size = new System.Drawing.Size(645, 20);
            this.current_progress.TabIndex = 2;
            // 
            // tbCurrent
            // 
            this.tbCurrent.Location = new System.Drawing.Point(13, 38);
            this.tbCurrent.Name = "tbCurrent";
            this.tbCurrent.ReadOnly = true;
            this.tbCurrent.Size = new System.Drawing.Size(69, 20);
            this.tbCurrent.TabIndex = 3;
            this.tbCurrent.TabStop = false;
            this.tbCurrent.Text = "Current";
            this.tbCurrent.TextAlign = System.Windows.Forms.HorizontalAlignment.Center;
            // 
            // ProgressForm
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(745, 71);
            this.ControlBox = false;
            this.Controls.Add(this.tbCurrent);
            this.Controls.Add(this.current_progress);
            this.Controls.Add(this.tbTotal);
            this.Controls.Add(this.total_progress);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedToolWindow;
            this.Icon = ((System.Drawing.Icon)(resources.GetObject("$this.Icon")));
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "ProgressForm";
            this.SizeGripStyle = System.Windows.Forms.SizeGripStyle.Hide;
            this.Text = "XCode Progress";
            this.TopMost = true;
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private AutoProgress total_progress;
        private System.Windows.Forms.TextBox tbTotal;
        private AutoProgress current_progress;
        private System.Windows.Forms.TextBox tbCurrent;
    }
}