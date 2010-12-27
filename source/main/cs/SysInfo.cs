using System;
using System.Windows;
using System.Diagnostics;
using System.IO;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Windows.Forms;
using Microsoft.Win32;

namespace xcode
{
    public class SysInfo
    {
        public bool DotNet4IsInstalled { get; set; }
        public bool MsBuildInstalled { get; set; }
        public bool MercurialInstalled { get; set; }
        public bool RemoteRepoExists { get; set; }
        public bool CacheRepoExists { get; set; }

        public void Collect()
        {
            if (FrameworkVersionDetection.IsInstalled(FrameworkVersion.Fx40))
                DotNet4IsInstalled = true;
            else
                DotNet4IsInstalled = false;

            string windir = Environment.GetEnvironmentVariable("windir");
            string msbuildDir = windir + @"\Microsoft.NET\Framework\v4.0.30319";
            if (File.Exists(msbuildDir + "\\MSBuild.exe"))
            {
                MsBuildInstalled = true;

                string path = Environment.GetEnvironmentVariable("path");
                if (path.IndexOf(msbuildDir, StringComparison.CurrentCultureIgnoreCase) < 0)
                {
                    path = path + ";" + msbuildDir;
                    Environment.SetEnvironmentVariable("path", path);
                }
            }
            else
            {
                MsBuildInstalled = false;
            }


            Process p = new Process();
            
            p.StartInfo = new ProcessStartInfo("hg.exe", "--version");
            p.StartInfo.WindowStyle = ProcessWindowStyle.Hidden;
            p.StartInfo.CreateNoWindow = true;
            p.StartInfo.RedirectStandardOutput = true;
            p.StartInfo.UseShellExecute = false;
            try
            {
                if (p.Start())
                {
                    if (p.WaitForExit(1000))
                    {
                        string msg = p.StandardOutput.ReadToEnd();
                        if (msg.StartsWith("Mercurial Distributed SCM"))
                        {
                            string beginStr = "(version ";
                            int begin = msg.IndexOf(beginStr);
                            if (begin >= 0)
                            {
                                string endStr = ")";
                                int end = msg.IndexOf(endStr, begin + beginStr.Length);
                                string hg_version = msg.Substring(begin + beginStr.Length, end - (begin + beginStr.Length));
                                Version v = new Version(hg_version);
                                Console.WriteLine(v);
                                if (v < new Version(1, 7))
                                {
                                    MercurialInstalled = false;
                                }
                                else
                                {
                                    MercurialInstalled = true;
                                }
                            }
                        }
                    }
                }
            }
            catch (SystemException e)
            {
                MercurialInstalled = false;
            }
        }
    }
}
