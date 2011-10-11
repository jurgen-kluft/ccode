using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Text;
using System.Windows.Forms;

namespace MSBuild.XCode.Helpers
{
    /// <summary>
    /// Provides information about the progress of a save, read, or extract operation. 
    /// This is a base class; you will probably use one of the classes derived from this one.
    /// </summary>
    public class ProgressEventArgs : EventArgs
    {
        private bool _cancel;
        private Int32 _step;
        private Int32 _step_max;

        internal ProgressEventArgs()
        {
        }

        public ProgressEventArgs(int step, int step_max)
        {
            _step = step;
            _step_max = step_max;
        }

        /// <summary>
        /// In an event handler, set this to cancel the save or extract 
        /// operation that is in progress.
        /// </summary>
        public bool Cancel
        {
            get { return _cancel; }
            set { _cancel = _cancel || value; }
        }

        /// <summary>
        /// The total step
        /// </summary>
        public Int32 Step
        {
            get { return _step; }
            set { _step = value; }
        }

        /// <summary>
        /// Total step maximum
        /// </summary>
        public Int32 Total
        {
            get { return _step_max; }
            set { _step_max = value; }
        }
    }


    public class AutoProgress : System.Windows.Forms.UserControl
    {
        public delegate void UpdateProgressDelegate(ProgressEventArgs e);
        public UpdateProgressDelegate dUpdateProgress;
        internal System.Windows.Forms.ProgressBar myProgressBar;

        public AutoProgress()
        {
            this.myProgressBar = new System.Windows.Forms.ProgressBar();
            this.dUpdateProgress = new UpdateProgressDelegate(UpdateProgress);

            this.SuspendLayout();

            this.myProgressBar.Dock = System.Windows.Forms.DockStyle.Fill;
            this.myProgressBar.Location = new System.Drawing.Point(0, 0);
            this.myProgressBar.Name = "myProgressBar";
            this.myProgressBar.Size = this.Size;
            this.myProgressBar.TabIndex = 2;
            this.Controls.Add(this.myProgressBar);

            this.ResumeLayout(false);

        }

        public void UpdateProgress(ProgressEventArgs e)
        {
            if (this.myProgressBar.InvokeRequired)
            {
                this.myProgressBar.Invoke(new UpdateProgressDelegate(this.UpdateProgress), new object[] { e });
            }
            else if (e.Total>0)
            {
                this.myProgressBar.Maximum = e.Total;
                this.myProgressBar.Value = Convert.ToInt32(100 * e.Step / e.Total);
            }
        }

        public void Start()
        {
            myProgressBar.Maximum = 100;
            myProgressBar.Value = 0;
        }

        public void Stop()
        {
            myProgressBar.Value = 0;
        }

        public void Finish()
        {
            myProgressBar.Value = myProgressBar.Maximum;
        }


    }

}

