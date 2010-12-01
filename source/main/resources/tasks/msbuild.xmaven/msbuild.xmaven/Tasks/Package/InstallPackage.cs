using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

namespace MsBuild.MsMaven.Tasks
{
    /// <summary>
    ///	Will copy a new package release to the local-package-repository. 
    ///	Also updates 'latest'.
    ///
    ///  Actions:
    ///	1) Determine destination device and path for version
    ///	2) Determine destination device and path for latest
    ///	3) Determine the file that is now latest in the repository
    ///	4) Copy the package to it's version location
    ///	5) Generate a file with the filename of the package + '.latest' to latest location and delete the previous latest
    ///	6) Done
    ///
    /// </summary>
    class MvPackageInstallTask : Task
    {
        public string Cmd { get; set; }

        private bool ExtractVars(string str, out List<string> vars, out string scan)
        {
            vars = Var.Extract(str, out scan);
            return true;
        }


        private bool Decode(string _cmd, Var _vars)
        {
            ///< $(GroupH).$(GroupM).$(GroupL) ==> com.virtuos.tnt
            ///< $(Name)_$(Major).$(Minor).$(Year).$(Month).$(Branch).$(Revision)_$(Platform) ==> xbase_1.0.2010.11.default.0_Win32

            // Split at '==>', now we have a Left-Hand-Side and Right-Hand-Side
            // If RHS has vars then return false
            // If LHS has no string characters and only vars, var_scan = false
            // Else var_scan = true

            string[] lhs_rhs = _cmd.Split(new string[] { "<==" }, StringSplitOptions.RemoveEmptyEntries);
            if (lhs_rhs.Length == 2)
            {
                lhs_rhs[0] = lhs_rhs[0].Trim();
                lhs_rhs[1] = lhs_rhs[1].Trim();

                lhs_rhs[0] = lhs_rhs[0].TrimStart('$').TrimStart('(').TrimEnd(')');
                lhs_rhs[1] = _vars.Expand(lhs_rhs[1]);

                _vars.Add(lhs_rhs[0], lhs_rhs[1]);
                return true;
            }

            return false;
        }

        public void Test()
        {
            Var _vars = new Var();

            _vars.Add("mvpackage_group", "com.virtuos.tnt");
            _vars.Add("mvpackage_name", "xbase");
            _vars.Add("mvpackage_version_major", "1");
            _vars.Add("mvpackage_version_minor", "0");
            _vars.Add("mvpackage_version_year", "2010");
            _vars.Add("mvpackage_version_month", "11");
            _vars.Add("mvpackage_vcs_branch", "default");
            _vars.Add("mvpackage_vcs_version", "23453248");
            _vars.Add("mvpackage_platform", "NintendoWII");
            _vars.Add("mvpackage_dir", @"file://D:\Dev\xbase\");
            _vars.Add("mvpackage_local_url", @"file://D:\MV_PACKAGE_REPO\");

            string[] text = new string[] 
            {
                @"[split] $(mvpackage_group) ==> . ==> GroupH, GroupM, GroupL",
                @"$!(SrcFilename) <== $(mvpackage_name)_$(mvpackage_version_major).$(mvpackage_version_minor).$(mvpackage_version_year).$(mvpackage_version_month).$(mvpackage_vcs_branch).$(mvpackage_vcs_version)_$(mvpackage_platform)",
                @"$!(SrcFilenameExt) <== .zip",
                @"$!(SrcFolder) <== $(mvpackage_dir)target\",
                @"$!(DstFolder) <== $(mvpackage_local_url)$!(GroupH)\$!(GroupM)\$!(GroupL)\$(mvpackage_name)\",
                @"[copy] $!(SrcFolder)$!(SrcFilename)$!(SrcFilenameExt) ==> $!(DstFolder)$!(mvpackage_version_year)\$!(mvpackage_version_month)\",
                @"[delete] $!(DstFolder)latest\$!(mvpackage_name)_*$!(mvpackage_vcs_branch)*_$!(mvpackage_platform).latest",
                @"[create] $!(DstFolder)latest\$!(SrcFilename)$!(SrcFilenameExt).latest"
            };

            foreach (string line in text)
            {
                string cmd = line.Replace("$!(", "$(");

                if (cmd.StartsWith("[split]"))
                {
                    cmd = cmd.Replace("[split]", "").Trim();
                    cmd = _vars.Expand(cmd);

                    string[] text_seperator_groups = cmd.Split(new string[] { "==>" }, StringSplitOptions.RemoveEmptyEntries);

                    string seperator = text_seperator_groups[1].Trim();
                    string[] parts = text_seperator_groups[0].Split(new string[] { seperator }, StringSplitOptions.RemoveEmptyEntries);
                    string[] groups = text_seperator_groups[2].Split(new char[] { ',' }, StringSplitOptions.RemoveEmptyEntries);

                    if (parts.Length == groups.Length)
                    {
                        int i = 0;
                        foreach (string group in groups)
                        {
                            _vars.Add(group.Trim(), parts[i++].Trim());
                        }
                    }
                }
                else if (cmd.StartsWith("[copy]"))
                {
                    Console.WriteLine(_vars.Expand(cmd));
                }
                else if (cmd.StartsWith("[delete]"))
                {
                    Console.WriteLine(_vars.Expand(cmd));
                }
                else if (cmd.StartsWith("[create]"))
                {
                    Console.WriteLine(_vars.Expand(cmd));
                }
                else
                {
                    Decode(cmd, _vars);
                }
            }

            _vars.DumpToConsole();
        }

        public override bool Execute()
        {
            Var _vars = new Var();
            string[] lines = Cmd.Split(new string[] { Environment.NewLine }, StringSplitOptions.RemoveEmptyEntries);
            foreach (string line in lines)
            {
                string cmd = line.Replace("$!(", "$(");

                // For now we only support filesystem
                cmd = cmd.Replace("file://", "");

                if (cmd.StartsWith("[copy]"))
                {
                    cmd = _vars.Expand(cmd);
                    cmd = cmd.Replace("[copy]", "").Trim();

                    string[] src_dst = cmd.Split(new string[] { "==>" }, StringSplitOptions.RemoveEmptyEntries);

                    // Copy a file
                    try
                    {
                        File.Copy(src_dst[0], src_dst[1], true);
                    }
                    catch (Exception)
                    {

                    }
                }
                else if (cmd.StartsWith("[delete]"))
                {
                    cmd = _vars.Expand(cmd);
                    cmd = cmd.Replace("[delete]", "").Trim();

                    // Delete a file
                    try
                    {
                        File.Delete(cmd);
                    }
                    catch (Exception)
                    {
                    }
                }
                else if (cmd.StartsWith("[create]"))
                {
                    cmd = _vars.Expand(cmd);
                    cmd = cmd.Replace("[create]", "").Trim();

                    // Create a file
                    try
                    {
                        FileStream stream = File.Create(cmd);
                        stream.Close();
                    }
                    catch (Exception)
                    {

                    }
                }
                else
                {
                    Decode(cmd, _vars);
                }
            }
            return false;
        }
    }
}
