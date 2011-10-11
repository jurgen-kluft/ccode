using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Windows.Forms;

namespace MSBuild.XCode.Helpers
{
    public partial class ProgressForm : Form
    {
        public ProgressForm()
        {
            InitializeComponent();
        }

        public void UpdateProgress(ProgressEventArgs total, ProgressEventArgs current)
        {
            this.total_progress.UpdateProgress(total);
            this.current_progress.UpdateProgress(current);
        }
    }
}
