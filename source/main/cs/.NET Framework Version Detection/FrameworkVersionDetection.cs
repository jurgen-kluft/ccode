using System;
using System.Diagnostics;
using System.Globalization;
using System.IO;
using Microsoft.Win32;
using System.Collections;
using System.Collections.Generic;
using System.Diagnostics.CodeAnalysis;

namespace xcode
{
    public static class FrameworkVersionDetection
    {
        #region events

        #endregion

        #region class-wide fields

        // Constants representing registry key names and value names
        private const string Netfx10RegKeyName = "Software\\Microsoft\\.NETFramework\\Policy\\v1.0";
        private const string Netfx10RegKeyValue = "3705";
        private const string Netfx10SPxMSIRegKeyName = "Software\\Microsoft\\Active Setup\\Installed Components\\{78705f0d-e8db-4b2d-8193-982bdda15ecd}";
        private const string Netfx10SPxOCMRegKeyName = "Software\\Microsoft\\Active Setup\\Installed Components\\{FDC11A6F-17D1-48f9-9EA3-9051954BAA24}";
        private const string Netfx11RegKeyName = "Software\\Microsoft\\NET Framework Setup\\NDP\\v1.1.4322";
        private const string Netfx20RegKeyName = "Software\\Microsoft\\NET Framework Setup\\NDP\\v2.0.50727";
        private const string Netfx30RegKeyName = "Software\\Microsoft\\NET Framework Setup\\NDP\\v3.0\\Setup";
        private const string Netfx30SpRegKeyName = "Software\\Microsoft\\NET Framework Setup\\NDP\\v3.0";
        private const string Netfx30RegValueName = "InstallSuccess";
        private const string Netfx35RegKeyName = "Software\\Microsoft\\NET Framework Setup\\NDP\\v3.5";
        private const string Netfx40RegKeyNameClient = "Software\\Microsoft\\NET Framework Setup\\NDP\\v4\\Client";
        private const string Netfx40RegKeyNameFull = "Software\\Microsoft\\NET Framework Setup\\NDP\\v4\\Full"; 
        private const string NetfxStandardRegValueName = "Install";
        private const string NetfxStandrdSpxRegValueName = "SP";
        private const string NetfxStandardVersionRegValueName = "Version";

        private const string Netfx20PlusBuildRegValueName = "Increment";
        private const string Netfx35PlusBuildRegValueName = "Build";

        private const string Netfx30PlusWCFRegKeyName = Netfx30RegKeyName + "\\Windows Communication Foundation";
        private const string Netfx30PlusWPFRegKeyName = Netfx30RegKeyName + "\\Windows Presentation Foundation";
        private const string Netfx30PlusWFRegKeyName = Netfx30RegKeyName + "\\Windows Workflow Foundation";
        private const string Netfx30PlusWFPlusVersionRegValueName = "FileVersion";
        private const string CardSpaceServicesRegKeyName = "System\\CurrentControlSet\\Services\\idsvc";
        private const string CardSpaceServicesPlusImagePathRegName = "ImagePath";

        private const string NetfxInstallRootRegKeyName = "Software\\Microsoft\\.NETFramework";
        private const string NetFxInstallRootRegValueName = "InstallRoot";

        private static readonly Version Netfx10Version = new Version(1, 0, 3705, 0);
        private static readonly Version Netfx11Version = new Version(1, 1, 4322, 573);
        private static readonly Version Netfx20Version = new Version(2, 0, 50727, 42);
        private static readonly Version Netfx30Version = new Version(3, 0, 4506, 26);
        private static readonly Version Netfx35Version = new Version(3, 5, 21022, 8);
        private static readonly Version Netfx40Version = new Version(4, 0, 30319);

        private const string Netfx10VersionString = "v1.0.3705";
        private const string Netfx11VersionString = "v1.1.4322";
        private const string Netfx20VersionString = "v2.0.50727";
        private const string NetfxMscorwks = "mscorwks.dll";
        #endregion

        #region private and internal properties and methods

        #region properties

        #endregion

        #region methods

        #region CheckFxVersion
        /// <summary>
        /// Retrieves the .NET Framework version number from the registry
        /// and validates that it is not a pre-release version number.
        /// </summary>
        /// <param name="frameworkVersion"></param>
        /// <returns><see langword="true"/> if the build number is greater than the 
        /// requested version; otherwise <see langword="false"/>.
        /// </returns>
        /// <remarks>If mscorwks.dll can be found the version number of the DLL (looking
        /// at the ProductVersion field) is also used.
        /// </remarks>
        private static bool CheckFxVersion(FrameworkVersion frameworkVersion)
        {
            bool valid = false;
            string installPath = GetMscorwksPath(frameworkVersion);
            FileVersionInfo fvi = null;
            Version fxVersion;

            if (!String.IsNullOrEmpty(installPath))
            {
                fvi = FileVersionInfo.GetVersionInfo(installPath);
            }

            switch (frameworkVersion)
            {
                case FrameworkVersion.Fx10:
                    fxVersion = GetNetfx10ExactVersion();
                    valid = (fvi != null) ? ((fxVersion >= Netfx10Version) && (new Version(fvi.ProductVersion) >= Netfx10Version)) : (fxVersion >= Netfx10Version);
                    break;
                case FrameworkVersion.Fx11:
                    fxVersion = GetNetfx11ExactVersion();
                    valid = (fvi != null) ? ((fxVersion >= Netfx11Version) && (new Version(fvi.ProductVersion) >= Netfx11Version)) : (fxVersion >= Netfx11Version);
                    break;
                case FrameworkVersion.Fx20:
                    fxVersion = GetNetfx20ExactVersion();
                    valid = (fvi != null) ? ((fxVersion >= Netfx20Version) && (new Version(fvi.ProductVersion) >= Netfx20Version)) : (fxVersion >= Netfx20Version);
                    break;
                case FrameworkVersion.Fx30:
                    fxVersion = GetNetfxExactVersion(Netfx30RegKeyName, NetfxStandardVersionRegValueName);
                    valid = (fvi != null) ? ((fxVersion >= Netfx30Version) && (new Version(fvi.ProductVersion) >= Netfx20Version)) : (fxVersion >= Netfx30Version);
                    break;
                case FrameworkVersion.Fx35:
                    fxVersion = GetNetfxExactVersion(Netfx35RegKeyName, NetfxStandardVersionRegValueName);
                    valid = (fvi != null) ? ((fxVersion >= Netfx35Version) && (new Version(fvi.ProductVersion) >= Netfx20Version)) : (fxVersion >= Netfx35Version);
                    break;
                case FrameworkVersion.Fx40:
                    fxVersion = GetNetfxExactVersion(Netfx40RegKeyNameFull, NetfxStandardVersionRegValueName);
                    valid = (fvi != null) ? ((fxVersion >= Netfx40Version) && (new Version(fvi.ProductVersion) >= Netfx35Version)) : (fxVersion >= Netfx40Version);
                    break;
                default:
                    valid = false;
                    break;
            }

            return valid;
        }
        #endregion

        #region GetInstallRoot
        /// <summary>
        /// Gets the installation root path for the .NET Framework.
        /// </summary>
        /// <returns>A <see cref="String"/> representing the installation root 
        /// path for the .NET Framework.</returns>
        private static string GetInstallRoot()
        {
            string installRoot = String.Empty;
            if (!GetRegistryValue(RegistryHive.LocalMachine, NetfxInstallRootRegKeyName, NetFxInstallRootRegValueName, RegistryValueKind.String, out installRoot))
            {
                throw new DirectoryNotFoundException("Installation root path of .NET framework cannot be found!");
            }
            return installRoot;
        }
        #endregion

        #region GetMscorwksPath
        /// <summary>
        /// Gets the path to the Mscorwks.DLL file.
        /// </summary>
        /// <param name="frameworkVersion"></param>
        /// <returns>The fully qualified path to the Mscorwks.DLL for the specified .NET
        /// Framework.
        /// </returns>
        private static string GetMscorwksPath(FrameworkVersion frameworkVersion)
        {
            string ret = String.Empty;

            switch (frameworkVersion)
            {
                case FrameworkVersion.Fx10:
                    ret = Path.Combine(Path.Combine(GetInstallRoot(), Netfx10VersionString), NetfxMscorwks);
                    break;

                case FrameworkVersion.Fx11:
                    ret = Path.Combine(Path.Combine(GetInstallRoot(), Netfx11VersionString), NetfxMscorwks);
                    break;

                case FrameworkVersion.Fx20:
                case FrameworkVersion.Fx30:
                case FrameworkVersion.Fx35:
                    ret = Path.Combine(Path.Combine(GetInstallRoot(), Netfx20VersionString), NetfxMscorwks);
                    break;

                default:
                    break;
            }

            return ret;
        }
        #endregion

        #region GetNetfxSPLevel functions

        #region GetNetfx10SPLevel
        /// <summary>
        /// Detects the service pack level for the .NET Framework 1.0.
        /// </summary>
        /// <returns>An <see cref="Int32"/> representing the service pack 
        /// level for the .NET Framework.</returns>
        /// <remarks>Uses the detection method recommended at
        /// http://blogs.msdn.com/astebner/archive/2004/09/14/229802.aspx 
        /// to determine what service pack for the .NET Framework 1.0 is 
        /// installed on the machine.
        /// </remarks>
        [System.Diagnostics.CodeAnalysis.SuppressMessage("Microsoft.Usage", "CA1806:DoNotIgnoreMethodResults", MessageId = "System.Int32.TryParse(System.String,System.Int32@)")]
        private static int GetNetfx10SPLevel()
        {
            bool foundKey = false;
            int servicePackLevel = -1;
            string regValue;

            if (IsTabletOrMediaCenter())
            {
                foundKey = GetRegistryValue(RegistryHive.LocalMachine, Netfx10SPxOCMRegKeyName, NetfxStandardVersionRegValueName, RegistryValueKind.String, out regValue);
            }
            else
            {
                foundKey = GetRegistryValue(RegistryHive.LocalMachine, Netfx10SPxMSIRegKeyName, NetfxStandardVersionRegValueName, RegistryValueKind.String, out regValue);
            }

            if (foundKey)
            {
                // This registry value should be of the format
                // #,#,#####,# where the last # is the SP level
                // Try to parse off the last # here
                int index = regValue.LastIndexOf(',');
                if (index > 0)
                {
                    Int32.TryParse(regValue.Substring(index + 1), out servicePackLevel);
                }
            }

            return servicePackLevel;
        }
        #endregion

        #region GetNetfxSPLevel
        /// <summary>
        /// Detects the service pack level for a version of .NET Framework.
        /// </summary>
        /// <param name="key">The registry key name.</param>
        /// <param name="value">The registry value name.</param>
        /// <returns>An <see cref="Int32"/> representing the service pack 
        /// level for the .NET Framework.</returns>
        private static int GetNetfxSPLevel(string key, string value)
        {
            int regValue = 0;

            // We can only get -1 if the .NET Framework is not
            // installed or there was some kind of error retrieving
            // the data from the registry
            int servicePackLevel = -1;

            if (GetRegistryValue(RegistryHive.LocalMachine, key, value, RegistryValueKind.DWord, out regValue))
            {
                servicePackLevel = regValue;
            }

            return servicePackLevel;
        }
        #endregion

        #endregion

        #region GetNetfxExactVersion functions

        #region GetNetfx10ExactVersion
        private static Version GetNetfx10ExactVersion()
        {
            bool foundKey = false;
            Version fxVersion = new Version();
            string regValue;

            if (IsTabletOrMediaCenter())
            {
                foundKey = GetRegistryValue(RegistryHive.LocalMachine, Netfx10SPxOCMRegKeyName, NetfxStandardVersionRegValueName, RegistryValueKind.String, out regValue);
            }
            else
            {
                foundKey = GetRegistryValue(RegistryHive.LocalMachine, Netfx10SPxMSIRegKeyName, NetfxStandardVersionRegValueName, RegistryValueKind.String, out regValue);
            }

            if (foundKey)
            {
                // This registry value should be of the format
                // #,#,#####,# where the last # is the SP level
                // Try to parse off the last # here
                int index = regValue.LastIndexOf(',');
                if (index > 0)
                {
                    string[] tokens = regValue.Substring(0, index).Split(',');
                    if (tokens.Length == 3)
                    {
                        fxVersion = new Version(Convert.ToInt32(tokens[0], NumberFormatInfo.InvariantInfo), Convert.ToInt32(tokens[1], NumberFormatInfo.InvariantInfo), Convert.ToInt32(tokens[2], NumberFormatInfo.InvariantInfo));
                    }
                }
            }

            return fxVersion;
        }
        #endregion

        #region GetNetfx11ExactVersion
        private static Version GetNetfx11ExactVersion()
        {
            int regValue = 0;

            // We can only get -1 if the .NET Framework is not
            // installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx11RegKeyName, NetfxStandardRegValueName, RegistryValueKind.DWord, out regValue))
            {
                if (regValue == 1)
                {
                    // In the strict sense, we are cheating here, but the registry key name itself
                    // contains the version number.
                    string[] tokens = Netfx11RegKeyName.Split(new string[] { "NDP\\v" }, StringSplitOptions.None);
                    if (tokens.Length == 2)
                    {
                        fxVersion = new Version(tokens[1]);
                    }
                }
            }

            return fxVersion;
        }
        #endregion

        #region GetNetfx20ExactVersion
        private static Version GetNetfx20ExactVersion()
        {
            string regValue = String.Empty;

            // We can only get -1 if the .NET Framework is not
            // installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            // If we have a Version registry value, use that.
            try
            {
                fxVersion = GetNetfxExactVersion(Netfx20RegKeyName, NetfxStandardVersionRegValueName);
            }
            catch (IOException)
            {
                // If we hit an exception here, the Version registry key probably doesn't exist so try
                // to get the version based on the registry key name itself.
                if (GetRegistryValue(RegistryHive.LocalMachine, Netfx20RegKeyName, Netfx20PlusBuildRegValueName, RegistryValueKind.String, out regValue))
                {
                    if (!String.IsNullOrEmpty(regValue))
                    {
                        string[] versionTokens = Netfx20RegKeyName.Split(new string[] { "NDP\\v" }, StringSplitOptions.None);
                        if (versionTokens.Length == 2)
                        {
                            string[] tokens = versionTokens[1].Split('.');
                            if (tokens.Length == 3)
                            {
                                fxVersion = new Version(Convert.ToInt32(tokens[0], NumberFormatInfo.InvariantInfo), Convert.ToInt32(tokens[1], NumberFormatInfo.InvariantInfo), Convert.ToInt32(tokens[2], NumberFormatInfo.InvariantInfo), Convert.ToInt32(regValue, NumberFormatInfo.InvariantInfo));
                            }
                        }
                    }
                }
            }

            return fxVersion;
        }
        #endregion

        #region GetNetfxExactVersion
        /// <summary>
        /// Retrieves the .NET Framework version number from the registry.
        /// </summary>
        /// <param name="key">The registry key name.</param>
        /// <param name="value">The registry value name.</param>
        /// <returns>A <see cref="Version"/> that represents the .NET 
        /// Framework version.</returns>
        private static Version GetNetfxExactVersion(string key, string value)
        {
            string regValue = String.Empty;

            // We can only get the default version if the .NET Framework
            // is not installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            if (GetRegistryValue(RegistryHive.LocalMachine, key, value, RegistryValueKind.String, out regValue))
            {
                if (!String.IsNullOrEmpty(regValue))
                {
                    fxVersion = new Version(regValue);
                }
            }

            return fxVersion;
        }
        #endregion

        #endregion

        #region GetRegistryValue
        private static bool GetRegistryValue<T>(RegistryHive hive, string key, string value, RegistryValueKind kind, out T data)
        {
            bool success = false;
            data = default(T);

            using (RegistryKey baseKey = RegistryKey.OpenRemoteBaseKey(hive, String.Empty))
            {
                if (baseKey != null)
                {
                    using (RegistryKey registryKey = baseKey.OpenSubKey(key, RegistryKeyPermissionCheck.ReadSubTree))
                    {
                        if (registryKey != null)
                        {
                            // If the key was opened, try to retrieve the value.
                            RegistryValueKind kindFound = registryKey.GetValueKind(value);
                            if (kindFound == kind)
                            {
                                object regValue = registryKey.GetValue(value, null);
                                if (regValue != null)
                                {
                                    data = (T)Convert.ChangeType(regValue, typeof(T), CultureInfo.InvariantCulture);
                                    success = true;
                                }
                            }
                        }
                    }
                }
            }
            return success;
        }
        #endregion

        #region IsNetfxInstalled functions

        #region IsNetfx10Installed
        /// <summary>
        /// Detects if the .NET 1.0 Framework is installed.
        /// </summary>
        /// <returns><see langword="true"/> if the .NET Framework 1.0 is 
        /// installed; otherwise <see langword="false"/></returns>
        /// <remarks>Uses the detection method recommended at
        /// http://msdn.microsoft.com/library/ms994349.aspx to determine
        /// whether the .NET Framework 1.0 is installed on the machine.
        /// </remarks>
        private static bool IsNetfx10Installed()
        {
            bool found = false;
            string regValue = string.Empty;
            found = GetRegistryValue(RegistryHive.LocalMachine, Netfx10RegKeyName, Netfx10RegKeyValue, RegistryValueKind.String, out regValue);

            return (found && CheckFxVersion(FrameworkVersion.Fx10));
        }
        #endregion

        #region IsNetfx11Installed
        /// <summary>
        /// Detects if the .NET 1.1 Framework is installed.
        /// </summary>
        /// <returns><see langword="true"/> if the .NET Framework 1.1 is 
        /// installed; otherwise <see langword="false"/></returns>
        /// <remarks>Uses the detection method recommended at
        /// http://msdn.microsoft.com/library/ms994339.aspx to determine
        /// whether the .NET Framework 1.1 is installed on the machine.
        /// </remarks>
        private static bool IsNetfx11Installed()
        {
            bool found = false;
            int regValue = 0;

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx11RegKeyName, NetfxStandardRegValueName, RegistryValueKind.DWord, out regValue))
            {
                if (regValue == 1)
                {
                    found = true;
                }
            }

            return (found && CheckFxVersion(FrameworkVersion.Fx11));
        }
        #endregion

        #region IsNetfx20Installed
        /// <summary>
        /// Detects if the .NET 2.0 Framework is installed.
        /// </summary>
        /// <returns><see langword="true"/> if the .NET Framework 2.0 is 
        /// installed; otherwise <see langword="false"/></returns>
        /// <remarks>Uses the detection method recommended at
        /// http://msdn.microsoft.com/library/aa480243.aspx to determine
        /// whether the .NET Framework 2.0 is installed on the machine.
        /// </remarks>
        private static bool IsNetfx20Installed()
        {
            bool found = false;
            int regValue = 0;

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx20RegKeyName, NetfxStandardRegValueName, RegistryValueKind.DWord, out regValue))
            {
                if (regValue == 1)
                {
                    found = true;
                }
            }

            return (found && CheckFxVersion(FrameworkVersion.Fx20));
        }
        #endregion

        #region IsNetfx30Installed
        /// <summary>
        /// Detects if the .NET 3.0 Framework is installed.
        /// </summary>
        /// <returns><see langword="true"/> if the .NET Framework 3.0 is 
        /// installed; otherwise <see langword="false"/></returns>
        /// <remarks>Uses the detection method recommended at
        /// http://msdn.microsoft.com/library/aa964979.aspx to determine
        /// whether the .NET Framework 3.0 is installed on the machine.
        /// </remarks>
        private static bool IsNetfx30Installed()
        {
            bool found = false;
            int regValue = 0;

            // The .NET Framework 3.0 is an add-in that installs on top of
            // the .NET Framework 2.0, so validate that both 2.0 and 3.0
            // are installed.
            if (IsNetfx20Installed())
            {
                // Check that the InstallSuccess registry value exists and equals 1.
                if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30RegKeyName, Netfx30RegValueName, RegistryValueKind.DWord, out regValue))
                {
                    if (regValue == 1)
                    {
                        found = true;
                    }
                }
            }

            // A system with a pre-release version of the .NET Fx 3.0 can have
            // the InstallSuccess value. As an added verification, check the
            // version number listed in the registry.
            return (found && CheckFxVersion(FrameworkVersion.Fx30));
        }
        #endregion

        #region IsNetfx35Installed
        /// <summary>
        /// Detects if the .NET 3.5 Framework is installed.
        /// </summary>
        /// <returns><see langword="true"/> if the .NET Framework 3.5 is 
        /// installed; otherwise <see langword="false"/></returns>
        /// <remarks>Uses the detection method recommended at
        /// http://msdn.microsoft.com/library/cc160716.aspx to determine
        /// whether the .NET Framework 3.5 is installed on the machine.
        /// Also uses the method described at 
        /// http://blogs.msdn.com/astebner/archive/2008/07/13/8729636.aspx.
        /// </remarks>
        private static bool IsNetfx35Installed()
        {
            bool found = false;
            int regValue = 0;

            // The .NET Framework 3.0 is an add-in that installs on top of
            // the .NET Framework 2.0 and 3.0, so validate that 2.0, 3.0,
            // and 3.5 are installed.
            if (IsNetfx20Installed() && IsNetfx30Installed())
            {
                // Check that the Install registry value exists and equals 1.
                if (GetRegistryValue(RegistryHive.LocalMachine, Netfx35RegKeyName, NetfxStandardRegValueName, RegistryValueKind.DWord, out regValue))
                {
                    if (regValue == 1)
                    {
                        found = true;
                    }
                }
            }

            // A system with a pre-release version of the .NET Fx 3.5 can have
            // the InstallSuccess value. As an added verification, check the
            // version number listed in the registry.
            return (found && CheckFxVersion(FrameworkVersion.Fx35));
        }
        #endregion

        #region IsNetfx40Installed
        /// <summary>
        /// Detects if the .NET 4.0 Framework is installed.
        /// </summary>
        /// <returns><see langword="true"/> if the .NET Framework 4.0 is 
        /// installed; otherwise <see langword="false"/></returns>
        /// <remarks>Uses the detection method recommended at
        /// http://msdn.microsoft.com/library/cc160716.aspx to determine
        /// whether the .NET Framework 4.0 is installed on the machine.
        /// Also uses the method described at 
        /// http://blogs.msdn.com/astebner/archive/2008/07/13/8729636.aspx.
        /// </remarks>
        private static bool IsNetfx40Installed()
        {
            bool found = false;
            int regValue = 0;

            // The .NET Framework 4.0 is an add-in that installs on top of
            // the .NET Framework 2.0, 3.0 and 3.5, so validate that 2.0, 3.0, 3.5
            // and 4.0 are installed.
            if (IsNetfx20Installed() && IsNetfx30Installed() && IsNetfx35Installed())
            {
                // Check that the Install registry value exists and equals 1.
                if (GetRegistryValue(RegistryHive.LocalMachine, Netfx40RegKeyNameClient, NetfxStandardRegValueName, RegistryValueKind.DWord, out regValue))
                {
                    if (regValue == 1)
                    {
                        // Check that the Install registry value exists and equals 1.
                        if (GetRegistryValue(RegistryHive.LocalMachine, Netfx40RegKeyNameFull, NetfxStandardRegValueName, RegistryValueKind.DWord, out regValue))
                        {
                            if (regValue == 1)
                            {
                                found = true;
                            }
                        }
                    }
                }
            }

            // A system with a pre-release version of the .NET Fx 4.0 can have
            // the InstallSuccess value. As an added verification, check the
            // version number listed in the registry.
            return (found && CheckFxVersion(FrameworkVersion.Fx40));
        }
        #endregion

        #endregion

        #region IsTabletOrMediaCenter
        private static bool IsTabletOrMediaCenter()
        {
            return ((SafeNativeMethods.GetSystemMetrics(SystemMetric.SM_TABLETPC) != 0) || (SafeNativeMethods.GetSystemMetrics(SystemMetric.SM_MEDIACENTER) != 0));
        }
        #endregion

        #region WindowsFounationLibrary functions

        #region CardSpace

        #region IsNetfx30CardSpaceInstalled
        private static bool IsNetfx30CardSpaceInstalled()
        {
            bool found = false;
            string regValue = String.Empty;

            if (GetRegistryValue(RegistryHive.LocalMachine, CardSpaceServicesRegKeyName, CardSpaceServicesPlusImagePathRegName, RegistryValueKind.ExpandString, out regValue))
            {
                if (!String.IsNullOrEmpty(regValue))
                {
                    found = true;
                }
            }

            return found;
        }
        #endregion

        #region GetNetfx30CardSpaceSPLevel
        // Currently, there are no service packs available for version 3.0 of 
        // the framework, so we always return -1. When a service pack does
        // become available, this method will need to be revised to correctly
        // determine the service pack level. Based on the current method for
        // determining if CardSpace is installed, it may not be possible to
        // correctly determine the Service Pack level for CardSpace.
        private static int GetNetfx30CardSpaceSPLevel()
        {
            int servicePackLevel = -1;
            return servicePackLevel;
        }
        #endregion

        #region GetNetfx30CardSpaceExactVersion
        private static Version GetNetfx30CardSpaceExactVersion()
        {
            string regValue = String.Empty;

            // We can only get the default version if the .NET Framework
            // is not installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            if (GetRegistryValue(RegistryHive.LocalMachine, CardSpaceServicesRegKeyName, CardSpaceServicesPlusImagePathRegName, RegistryValueKind.ExpandString, out regValue))
            {
                if (!String.IsNullOrEmpty(regValue))
                {
                    FileVersionInfo fileVersionInfo = FileVersionInfo.GetVersionInfo(regValue.Trim('"'));
                    int index = fileVersionInfo.FileVersion.IndexOf(' ');
                    fxVersion = new Version(fileVersionInfo.FileVersion.Substring(0, index));
                }
            }

            return fxVersion;
        }
        #endregion

        #endregion

        #region Windows Communication Foundation

        #region IsNetfx30WCFInstalled
        private static bool IsNetfx30WCFInstalled()
        {
            bool found = false;
            int regValue = 0;

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWCFRegKeyName, Netfx30RegValueName, RegistryValueKind.DWord, out regValue))
            {
                if (regValue == 1)
                {
                    found = true;
                }
            }

            return found;
        }
        #endregion

        #region GetNetfx30WCFSPLevel
        // This code is MOST LIKELY correct but will need to be verified.
        //
        // Currently, there are no service packs available for version 3.0 of 
        // the framework, so we always return -1. When a service pack does
        // become available, this method will need to be revised to correctly
        // determine the service pack level.
        private static int GetNetfx30WCFSPLevel()
        {
            //int regValue = 0;

            // We can only get -1 if the .NET Framework is not
            // installed or there was some kind of error retrieving
            // the data from the registry
            int servicePackLevel = -1;

            //if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWCFRegKeyName, Netfx11PlusSPxRegValueName, RegistryValueKind.DWord, out regValue))
            //{
            //    servicePackLevel = regValue;
            //}

            return servicePackLevel;
        }
        #endregion

        #region GetNetfx30WCFExactVersion
        private static Version GetNetfx30WCFExactVersion()
        {
            string regValue = String.Empty;

            // We can only get the default version if the .NET Framework
            // is not installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWCFRegKeyName, NetfxStandardVersionRegValueName, RegistryValueKind.String, out regValue))
            {
                if (!String.IsNullOrEmpty(regValue))
                {
                    fxVersion = new Version(regValue);
                }
            }

            return fxVersion;
        }
        #endregion

        #endregion

        #region Windows Presentation Foundation

        #region IsNetfx30WPFInstalled
        private static bool IsNetfx30WPFInstalled()
        {
            bool found = false;
            int regValue = 0;

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWPFRegKeyName, Netfx30RegValueName, RegistryValueKind.DWord, out regValue))
            {
                if (regValue == 1)
                {
                    found = true;
                }
            }

            return found;
        }
        #endregion

        #region GetNetfx30WPFSPLevel
        // This code is MOST LIKELY correct but will need to be verified.
        //
        // Currently, there are no service packs available for version 3.0 of 
        // the framework, so we always return -1. When a service pack does
        // become available, this method will need to be revised to correctly
        // determine the service pack level.
        private static int GetNetfx30WPFSPLevel()
        {
            //int regValue = 0;

            // We can only get -1 if the .NET Framework is not
            // installed or there was some kind of error retrieving
            // the data from the registry
            int servicePackLevel = -1;

            //if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWPFRegKeyName, Netfx11PlusSPxRegValueName, RegistryValueKind.DWord, out regValue))
            //{
            //    servicePackLevel = regValue;
            //}

            return servicePackLevel;
        }
        #endregion

        #region GetNetfx30WPFExactVersion
        private static Version GetNetfx30WPFExactVersion()
        {
            string regValue = String.Empty;

            // We can only get the default version if the .NET Framework
            // is not installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWPFRegKeyName, NetfxStandardVersionRegValueName, RegistryValueKind.String, out regValue))
            {
                if (!String.IsNullOrEmpty(regValue))
                {
                    fxVersion = new Version(regValue);
                }
            }

            return fxVersion;
        }
        #endregion

        #endregion

        #region Windows Workflow Foundation

        #region IsNetfx30WFInstalled
        private static bool IsNetfx30WFInstalled()
        {
            bool found = false;
            int regValue = 0;

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWFRegKeyName, Netfx30RegValueName, RegistryValueKind.DWord, out regValue))
            {
                if (regValue == 1)
                {
                    found = true;
                }
            }

            return found;
        }
        #endregion

        #region GetNetfx30WFSPLevel
        // This code is MOST LIKELY correct but will need to be verified.
        //
        // Currently, there are no service packs available for version 3.0 of 
        // the framework, so we always return -1. When a service pack does
        // become available, this method will need to be revised to correctly
        // determine the service pack level.
        private static int GetNetfx30WFSPLevel()
        {
            //int regValue = 0;

            // We can only get -1 if the .NET Framework is not
            // installed or there was some kind of error retrieving
            // the data from the registry
            int servicePackLevel = -1;

            //if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWFRegKeyName, Netfx11PlusSPxRegValueName, RegistryValueKind.DWord, out regValue))
            //{
            //    servicePackLevel = regValue;
            //}

            return servicePackLevel;
        }
        #endregion

        #region GetNetfx30WFExactVersion
        private static Version GetNetfx30WFExactVersion()
        {
            string regValue = String.Empty;

            // We can only get the default version if the .NET Framework
            // is not installed or there was some kind of error retrieving
            // the data from the registry
            Version fxVersion = new Version();

            if (GetRegistryValue(RegistryHive.LocalMachine, Netfx30PlusWFRegKeyName, Netfx30PlusWFPlusVersionRegValueName, RegistryValueKind.String, out regValue))
            {
                if (!String.IsNullOrEmpty(regValue))
                {
                    fxVersion = new Version(regValue);
                }
            }

            return fxVersion;
        }
        #endregion

        #endregion

        #endregion

        #endregion

        #endregion

        #region public properties and methods

        #region properties

        #region InstalledFrameworkVersions
        /// <summary>
        /// Gets an <see cref="IEnumerable"/> list of the installed .NET Framework 
        /// versions.
        /// </summary>
        public static IEnumerable InstalledFrameworkVersions
        {
            get
            {
                if (IsInstalled(FrameworkVersion.Fx10))
                {
                    yield return GetExactVersion(FrameworkVersion.Fx10);
                }
                if (IsInstalled(FrameworkVersion.Fx11))
                {
                    yield return GetExactVersion(FrameworkVersion.Fx11);
                }
                if (IsInstalled(FrameworkVersion.Fx20))
                {
                    yield return GetExactVersion(FrameworkVersion.Fx20);
                }
                if (IsInstalled(FrameworkVersion.Fx30))
                {
                    yield return GetExactVersion(FrameworkVersion.Fx30);
                }
                if (IsInstalled(FrameworkVersion.Fx35))
                {
                    yield return GetExactVersion(FrameworkVersion.Fx35);
                }
                if (IsInstalled(FrameworkVersion.Fx40))
                {
                    yield return GetExactVersion(FrameworkVersion.Fx40);
                }
            }
        }
        #endregion

        #endregion

        #region methods

        #region IsInstalled

        #region IsInstalled(FrameworkVersion frameworkVersion)
        ///<overloads>
        /// Determines if the specified .NET Framework or Foundation Library is 
        /// installed on the local computer.
        ///</overloads>
        /// <summary>
        /// Determines if the specified .NET Framework version is installed
        /// on the local computer.
        /// </summary>
        /// <param name="frameworkVersion">The version of the .NET Framework to test.
        /// </param>
        /// <returns><see langword="true"/> if the specified .NET Framework
        /// version is installed; otherwise <see langword="false"/>.</returns>
        public static bool IsInstalled(FrameworkVersion frameworkVersion)
        {
            bool ret = false;

            switch (frameworkVersion)
            {
                case FrameworkVersion.Fx10:
                    ret = IsNetfx10Installed();
                    break;

                case FrameworkVersion.Fx11:
                    ret = IsNetfx11Installed();
                    break;

                case FrameworkVersion.Fx20:
                    ret = IsNetfx20Installed();
                    break;

                case FrameworkVersion.Fx30:
                    ret = IsNetfx30Installed();
                    break;

                case FrameworkVersion.Fx35:
                    ret = IsNetfx35Installed();
                    break;

                case FrameworkVersion.Fx40:
                    ret = IsNetfx40Installed();
                    break;

                default:
                    break;
            }

            return ret;
        }
        #endregion

        #region IsInstalled(WindowsFoundationLibrary foundationLibrary)
        /// <summary>
        /// Determines if the specified .NET Framework Foundation Library is
        /// installed on the local computer.
        /// </summary>
        /// <param name="foundationLibrary">The Foundation Library to test.
        /// </param>
        /// <returns><see langword="true"/> if the specified .NET Framework
        /// Foundation Library is installed; otherwise <see langword="false"/>.</returns>
        public static bool IsInstalled(WindowsFoundationLibrary foundationLibrary)
        {
            bool ret = false;

            switch (foundationLibrary)
            {
                case WindowsFoundationLibrary.CardSpace:
                    ret = IsNetfx30CardSpaceInstalled();
                    break;

                case WindowsFoundationLibrary.WCF:
                    ret = IsNetfx30WCFInstalled();
                    break;

                case WindowsFoundationLibrary.WF:
                    ret = IsNetfx30WFInstalled();
                    break;

                case WindowsFoundationLibrary.WPF:
                    ret = IsNetfx30WPFInstalled();
                    break;

                default:
                    break;
            }

            return ret;
        }
        #endregion

        #endregion

        #region GetServicePackLevel

        #region GetServicePackLevel(FrameworkVersion frameworkVersion)
        ///<overloads>
        /// Retrieves the service pack level for the specified .NET Framework or  
        /// Foundation Library.
        ///</overloads>
        /// <summary>
        /// Retrieves the service pack level for the specified .NET Framework
        /// version.
        /// </summary>
        /// <param name="frameworkVersion">The .NET Framework whose service pack 
        /// level should be retrieved.</param>
        /// <returns>An <see cref="Int32">integer</see> value representing
        /// the service pack level for the specified .NET Framework version. If
        /// the specified .NET Frameowrk version is not found, -1 is returned.
        /// </returns>
        public static int GetServicePackLevel(FrameworkVersion frameworkVersion)
        {
            int servicePackLevel = -1;

            switch (frameworkVersion)
            {
                case FrameworkVersion.Fx10:
                    servicePackLevel = GetNetfx10SPLevel();
                    break;

                case FrameworkVersion.Fx11:
                    servicePackLevel = GetNetfxSPLevel(Netfx11RegKeyName, NetfxStandrdSpxRegValueName);
                    break;

                case FrameworkVersion.Fx20:
                    servicePackLevel = GetNetfxSPLevel(Netfx20RegKeyName, NetfxStandrdSpxRegValueName);
                    break;

                case FrameworkVersion.Fx30:
                    servicePackLevel = GetNetfxSPLevel(Netfx30SpRegKeyName, NetfxStandrdSpxRegValueName);
                    break;

                case FrameworkVersion.Fx35:
                    servicePackLevel = GetNetfxSPLevel(Netfx35RegKeyName, NetfxStandrdSpxRegValueName);
                    break;

                case FrameworkVersion.Fx40:
                    servicePackLevel = GetNetfxSPLevel(Netfx40RegKeyNameFull, NetfxStandrdSpxRegValueName);
                    break;

                default:
                    break;
            }

            return servicePackLevel;
        }
        #endregion

        #region GetServicePackLevel(WindowsFoundationLibrary foundationLibrary)
        /// <summary>
        /// Retrieves the service pack level for the specified .NET Framework
        /// Foundation Library.
        /// </summary>
        /// <param name="foundationLibrary">The Foundation Library whose service pack 
        /// level should be retrieved.</param>
        /// <returns>An <see cref="Int32">integer</see> value representing
        /// the service pack level for the specified .NET Framework Foundation
        /// Library. If the specified .NET Frameowrk Foundation Library is not
        /// found, -1 is returned.
        /// </returns>
        public static int GetServicePackLevel(WindowsFoundationLibrary foundationLibrary)
        {
            int servicePackLevel = -1;

            switch (foundationLibrary)
            {
                case WindowsFoundationLibrary.CardSpace:
                    servicePackLevel = GetNetfx30CardSpaceSPLevel();
                    break;

                case WindowsFoundationLibrary.WCF:
                    servicePackLevel = GetNetfx30WCFSPLevel();
                    break;

                case WindowsFoundationLibrary.WF:
                    servicePackLevel = GetNetfx30WFSPLevel();
                    break;

                case WindowsFoundationLibrary.WPF:
                    servicePackLevel = GetNetfx30WPFSPLevel();
                    break;

                default:
                    break;
            }

            return servicePackLevel;
        }
        #endregion

        #endregion

        #region GetExactVersion

        #region GetExactVersion(FrameworkVersion frameworkVersion)
        ///<overloads>
        /// Retrieves the exact version number for the specified .NET Framework or
        /// Foundation Library.
        ///</overloads>
        /// <summary>
        /// Retrieves the exact version number for the specified .NET Framework
        /// version.
        /// </summary>
        /// <param name="frameworkVersion">The .NET Framework whose version should be 
        /// retrieved.</param>
        /// <returns>A <see cref="Version">version</see> representing
        /// the exact version number for the specified .NET Framework version.
        /// If the specified .NET Frameowrk version is not found, a 
        /// <see cref="Version"/> is returned that represents a 0.0.0.0 version
        /// number.
        /// </returns>
        public static Version GetExactVersion(FrameworkVersion frameworkVersion)
        {
            Version fxVersion = new Version();

            switch (frameworkVersion)
            {
                case FrameworkVersion.Fx10:
                    fxVersion = GetNetfx10ExactVersion();
                    break;

                case FrameworkVersion.Fx11:
                    fxVersion = GetNetfx11ExactVersion();
                    break;

                case FrameworkVersion.Fx20:
                    fxVersion = GetNetfx20ExactVersion();
                    break;

                case FrameworkVersion.Fx30:
                    fxVersion = GetNetfxExactVersion(Netfx30RegKeyName, NetfxStandardVersionRegValueName);
                    break;

                case FrameworkVersion.Fx35:
                    fxVersion = GetNetfxExactVersion(Netfx35RegKeyName, NetfxStandardVersionRegValueName);
                    break;

                case FrameworkVersion.Fx40:
                    fxVersion = GetNetfxExactVersion(Netfx40RegKeyNameFull, NetfxStandardVersionRegValueName);
                    break;

                default:
                    break;
            }

            return fxVersion;
        }
        #endregion

        #region GetExactVersion(WindowsFoundationLibrary foundationLibrary)
        /// <summary>
        /// Retrieves the exact version number for the specified .NET Framework
        /// Foundation Library.
        /// </summary>
        /// <param name="foundationLibrary">The Foundation Library whose version
        /// should be retrieved.</param>
        /// <returns>A <see cref="Version">version</see> representing
        /// the exact version number for the specified .NET Framework Foundation
        /// Library. If the specified .NET Frameowrk Foundation Library is not
        /// found, a <see cref="Version"/> is returned that represents a 
        /// 0.0.0.0 version number.
        /// </returns>
        public static Version GetExactVersion(WindowsFoundationLibrary foundationLibrary)
        {
            Version fxVersion = new Version();

            switch (foundationLibrary)
            {
                case WindowsFoundationLibrary.CardSpace:
                    fxVersion = GetNetfx30CardSpaceExactVersion();
                    break;

                case WindowsFoundationLibrary.WCF:
                    fxVersion = GetNetfx30WCFExactVersion();
                    break;

                case WindowsFoundationLibrary.WF:
                    fxVersion = GetNetfx30WFExactVersion();
                    break;

                case WindowsFoundationLibrary.WPF:
                    fxVersion = GetNetfx30WPFExactVersion();
                    break;

                default:
                    break;
            }

            return fxVersion;
        }
        #endregion

        #endregion

        #endregion

        #endregion
    }
}
