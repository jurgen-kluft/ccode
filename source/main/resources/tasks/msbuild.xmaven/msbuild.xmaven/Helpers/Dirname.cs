using System;

namespace msbuild.xmaven.helpers
{
    /// <summary>
    /// Represents a directory
    /// <remarks>
    /// Purpose of this class is to make a directory easy to use and safe.
    /// Note: Memory usage has been prevered over performance.
    /// </remarks>
    /// <interface>
    /// 
    /// e.g. 
    ///             C:\Documents\Music\Beatles
    ///             \\cnshaw235\Documents\Music\Beatles
    /// 
    /// Properties
    ///     Device                                                          (e.g. C:\)
    ///     DeviceName                                                      (e.g. C)
    ///     Relative                                                        (Documents\Music\Beatles)
    ///     Full                                                            (C:\Documents\Music\Beatles)
    ///     Levels                                                          (3)
    ///     
    /// Boolean
    ///     HasDevice
    ///     IsNetworkDevice
    ///     HasPath
    ///     
    /// Methods
    ///     Clear()
    ///     
    ///     ChangeDevice(device)
    ///     ChangePath(directory)
    ///     ChangeFull(device+directory)
    /// 
    ///     MakeAbsolute()                                                  (working directory)
    ///     MakeAbsolute(device + directory)
    ///     MakeRelative()                                                  (working directory)
    ///     MakeRelative(device + directory)
    /// 
    ///     LevelUp()                                                       (C:\Documents\Music)
    ///     LevelDown(folder)                                               (C:\Documents\Music\Beatles\folder)
    /// 
    /// </interface>
    /// </summary>
    [Serializable]
    public struct xDirname
    {
        #region Internal Statics

        private const bool sIgnoreCase = true;

        private const char sSlash = '\\';
        private const string sSlashStr = "\\";
        private const string sDoubleSlash = "\\\\";
        private const char sSemi = ':';
        private const string sSemiSlash = ":\\";
        private const string sIllegalNameChars = "/\\:*?<>|";

        static internal string RemoveChars(string ioString, string inChars)
        {
            int cc = 0;
            int nn = 0;
            Char[] str = ioString.ToCharArray();
            for (int i = 0; i < ioString.Length; ++i)
            {
                Char c = str[nn];
                if (inChars.IndexOf(c) >= 0)
                {
                    nn++;
                }
                else
                {
                    if (cc < nn) str[cc] = c;
                    ++cc;
                    ++nn;
                }
            }

            // If nothing removed just return the incoming string
            if (cc == nn)
                return ioString;

            return new string(str, 0, cc);
        }

        static internal bool ContainsChars(string inString, string inChars)
        {
            foreach (Char c in inString)
            {
                if (inChars.IndexOf(c) != -1)
                    return true;
            }
            return false;
        }

        #endregion
        #region Public Statics

        public static readonly xDirname Empty = new xDirname(string.Empty, string.Empty.GetHashCode());

        #endregion Public Statics
        #region Fields

        private int mHashCode;
        private string mFull;

        #endregion
        #region Constructors

        public xDirname(string inRHS)
        {
            mFull = inRHS.Replace('/', '\\');
            mHashCode = mFull.GetHashCode();
            ChangeFull(inRHS);
        }

        private xDirname(string inRHS, int hashcode)
        {
            mFull = inRHS;
            mHashCode = hashcode;
            ChangeFull(inRHS);
        }

        public xDirname(xDirname inRHS)
        {
            mFull = inRHS.mFull;
            mHashCode = inRHS.mHashCode;
        }

        #endregion
        #region Properties

        public bool IsEmpty
        {
            get
            {
                return mFull.Length == 0;
            }
        }

        public string Device
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                sParseDevice(mFull, out deviceName, out isNetworkDevice);

                if (deviceName.Length != 0)
                    return (isNetworkDevice) ? sDoubleSlash + deviceName : deviceName + sSemi;
                return string.Empty;
            }
            set
            {
                ChangeDevice(value);
            }
        }

        public string DeviceName
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                sParseDevice(mFull, out deviceName, out isNetworkDevice);
                return deviceName;
            }
            set
            {
                string device = RemoveChars(value, sIllegalNameChars);
                ChangeDevice(device);
            }
        }

        public bool IsNetworkDevice
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                sParseDevice(mFull, out deviceName, out isNetworkDevice);
                return isNetworkDevice;
            }
            set
            {
                string deviceName;
                bool isNetworkDevice;
                string path;
                int levels;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels);

                if (isNetworkDevice != value)
                {
                    sConstructFull(deviceName, value, path, out mFull, out mHashCode);
                }

            }
        }

        public string Path
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                string path;
                int levels;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels);
                return path;
            }
            set
            {
                ChangePath(value);
            }
        }


        public int Levels
        {
            get
            {
                string name;
                int levels;
                sParseNameAndLevels(mFull, out name, out levels);
                return levels;
            }
        }

        public string Name
        {
            get
            {
                string name;
                int levels;
                sParseNameAndLevels(mFull, out name, out levels);
                return name;
            }
        }

        public string Full
        {
            get
            {
                return mFull;
            }
            set
            {
                ChangeFull(value);
            }
        }

        public bool HasDevice
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                sParseDevice(mFull, out deviceName, out isNetworkDevice);
                return deviceName.Length != 0;
            }
        }

        #endregion
        #region Private Static Methods

        private static void sConstructFull(string deviceName, bool isNetworkDevice, string path, out string outFull, out int outHashCode)
        {
            string device = isNetworkDevice ? (sDoubleSlash + deviceName) : (deviceName + sSemi);
            outFull = (deviceName.Length != 0) ? (device + sSlashStr + path) : path;
            outHashCode = outFull.ToLower().GetHashCode();
        }

        private static void sParseDevice(string inFull, out string outDevice, out bool outIsNetworkDevice)
        {
            string device;

            // Device
            bool networkDevice = inFull.StartsWith(sDoubleSlash);
            if (networkDevice)
            {
                int slashIndex = inFull.IndexOf(sSlash, 2);
                device = slashIndex >= 0 ? inFull.Substring(2, slashIndex - 2) : inFull.Substring(2);
            }
            else
            {
                int semiSlashIndex = inFull.IndexOf(sSemiSlash);
                device = semiSlashIndex >= 0 ? inFull.Substring(0, semiSlashIndex) : string.Empty;
            }

            outDevice = RemoveChars(device, sIllegalNameChars);
            if (networkDevice)
            {
                outIsNetworkDevice = true;
            }
            else if (outDevice.Length > 0)
            {
                outIsNetworkDevice = false;
            }
            else
            {
                outIsNetworkDevice = false;
                outDevice = string.Empty;
            }
        }

        private static void sParsePath(string inFullWithoutDevice, out string outPath, out int outLevels)
        {
            // Documents\Music.Collection\Beatles.Album
            // Count levels
            string[] folders = inFullWithoutDevice.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

            outPath = string.Empty;
            outLevels = 0;
            foreach (string folder in folders)
            {
                if (!ContainsChars(folder, sIllegalNameChars))
                {
                    if (outLevels == 0)
                        outPath = folder;
                    else
                        outPath = outPath + sSlash + folder;
                    ++outLevels;
                }
            }
        }

        private static void sParseNameAndLevels(string inFull, out string outName, out int outLevels)
        {
            bool outIsNetworkDevice;
            sParseDevice(inFull, out outName, out outIsNetworkDevice);
            sRemoveDevice(inFull, outIsNetworkDevice, out outName);

            // Documents\Music.Collection\Beatles.Album
            // Count levels
            // Name of folder
            string[] folders = outName.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

            if (folders.Length > 0)
                outName = folders[folders.Length - 1];
            else
                outName = string.Empty;

            outLevels = 0;
            foreach (string folder in folders)
            {
                if (!ContainsChars(folder, sIllegalNameChars))
                {
                    ++outLevels;
                }
            }
        }

        private static void sRemoveDevice(string inPathWithDevice, bool inIsNetworkDevice, out string outPath)
        {
            // Remove device
            outPath = inPathWithDevice;
            if (inIsNetworkDevice)
            {
                int slashIndex = inPathWithDevice.IndexOf(sSlash, 2);
                if (slashIndex >= 0)
                {
                    outPath = inPathWithDevice.Substring(slashIndex + 1);
                }
            }
            else
            {
                int semiSlashIndex = inPathWithDevice.IndexOf(sSemiSlash);
                if (semiSlashIndex >= 0)
                {
                    outPath = inPathWithDevice.Substring(semiSlashIndex + 2);
                }
            }
        }

        private static void sParseFull(string inFull, out string outDeviceName, out bool outIsNetworkDevice, out string outPath, out int outLevels)
        {
            // inFull can be any of these:
            //   - HeyJude.mp3
            //   - Documents\Music\Beatles\HeyJude.mp3
            //   - C:\HeyJude.mp3
            //   - \\cnshaw235\HeyJude.mp3
            //   - C:\Documents\Music\Beatles\HeyJude.mp3
            //   - \\cnshaw235\Documents\Music\Beatles\HeyJude.mp3
            sParseDevice(inFull, out outDeviceName, out outIsNetworkDevice);
            sRemoveDevice(inFull, outIsNetworkDevice, out outPath);
            sParsePath(outPath, out outPath, out outLevels);
        }

        #endregion
        #region Public Methods

        public void Clear()
        {
            mFull = string.Empty;
            mHashCode = mFull.GetHashCode();
        }

        public void ChangeDevice(string inDevice)
        {
            bool isNetworkDevice;
            string deviceName;
            sParseDevice(inDevice, out deviceName, out isNetworkDevice);

            string _deviceName;
            bool _isNetworkDevice;
            string path;
            int levels;
            sParseFull(mFull, out _deviceName, out _isNetworkDevice, out path, out levels);
            sConstructFull(deviceName, isNetworkDevice, path, out mFull, out mHashCode);
        }

        public void ChangePath(string inPath)
        {
            string path;
            int levels;
            sParsePath(inPath, out path, out levels);

            string deviceName;
            bool isNetworkDevice;
            string _path;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out _path, out levels);
            sConstructFull(deviceName, isNetworkDevice, path, out mFull, out mHashCode);
        }

        public void ChangeFull(string deviceAndDirectory)
        {
            // C:\Documents\Music.Collection\Beatles.Album
            // Count levels
            string deviceName;
            bool isNetworkDevice;
            sParseDevice(deviceAndDirectory, out deviceName, out isNetworkDevice);

            string _path;
            sRemoveDevice(deviceAndDirectory, isNetworkDevice, out _path);

            string path;
            int levels;
            sParsePath(_path, out path, out levels);

            string _deviceName;
            bool _isNetworkDevice;
            sParseFull(mFull, out _deviceName, out _isNetworkDevice, out _path, out levels);
            sConstructFull(deviceName, isNetworkDevice, path, out mFull, out mHashCode);
        }

        public static xDirname Add(xDirname deviceAndPath, xDirname path)
        {
            string deviceNameL;
            bool isNetworkDeviceL;
            string pathL;
            int levelsL;
            sParseFull(deviceAndPath.Full, out deviceNameL, out isNetworkDeviceL, out pathL, out levelsL);

            string deviceNameR;
            bool isNetworkDeviceR;
            string pathR;
            int levelsR;
            sParseFull(path.Full, out deviceNameR, out isNetworkDeviceR, out pathR, out levelsR);

            xDirname s = new xDirname();
            sConstructFull(deviceNameL, isNetworkDeviceL, (pathL.Length != 0 && pathR.Length != 0) ? (pathL + sSlash + pathR) : (pathL + pathR), out s.mFull, out s.mHashCode);
            return s;
        }

        public xDirname MakeAbsolute()
        {
            return MakeAbsolute(Environment.CurrentDirectory);
        }

        public xDirname MakeAbsolute(string absolutePath)
        {
            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels);

            if (deviceName.Length != 0)
                return new xDirname(this);

            if (path.Length != 0)
            {
                if (absolutePath.EndsWith(sSlashStr))
                    absolutePath += path;
                else
                    absolutePath += sSlash + path;
            }

            xDirname newDirname = new xDirname(this);
            newDirname.ChangeFull(absolutePath);
            return newDirname;
        }

        public xDirname MakeRelative()
        {
            return MakeRelative(Environment.CurrentDirectory);
        }

        public xDirname MakeRelative(string absolutePath)
        {
            // IN:      C:\My Media\Documents
            // THIS:    C:\My Media\Documents\Music\Beatles

            // RESULT:  Music\Beatles

            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels);

            if (thisPath.Length == 0)
                return new xDirname(this);

            xDirname newDirname = new xDirname(this);
            newDirname.ChangeFull(absolutePath);

            string newDeviceName;
            bool newIsNetworkDevice;
            string newPath;
            int newLevels;
            sParseFull(newDirname.mFull, out newDeviceName, out newIsNetworkDevice, out newPath, out newLevels);

            bool sameDevice = true;
            if (String.Compare(thisDeviceName, newDeviceName, true) != 0)
                sameDevice = false;

            if (newPath.Length == 0)
            {
                newPath = thisPath;

                if (sameDevice)
                {
                    newIsNetworkDevice = false;
                    newDeviceName = string.Empty;
                }
                else
                {
                    // Incoming absolute path, no device and no path.
                    // Restore to our initial state.
                    newDeviceName = thisDeviceName;
                }
            }
            else
            {
                if (sameDevice)
                {
                    string[] folders = thisPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);
                    string[] inFolders = newPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

                    bool samePath = true;
                    for (int i = 0; i < inFolders.Length && samePath; i++)
                        samePath = String.Compare(inFolders[i], folders[i], sIgnoreCase) == 0;

                    newIsNetworkDevice = false;
                    newDeviceName = string.Empty;
                    if (samePath)
                    {
                        string path = string.Empty;
                        for (int i = newLevels; i < thisLevels; i++)
                        {
                            path = path + (path.Length == 0 ? folders[i] : (sSlash + folders[i]));
                        }

                        newPath = path;
                    }
                    else
                    {
                        newPath = thisPath;
                    }
                }
                else
                {
                    newIsNetworkDevice = thisIsNetworkDevice;
                    newDeviceName = thisDeviceName;
                    newPath = thisPath;
                }
            }

            sConstructFull(newDeviceName, newIsNetworkDevice, newPath, out newDirname.mFull, out newDirname.mHashCode);
            return newDirname;
        }

        public void LevelUp()
        {
            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels);

            if (thisPath.Length == 0)
                return;

            string[] folders = thisPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

            thisPath = string.Empty;
            for (int i = 0; i < (folders.Length - 1); i++)
            {
                thisPath = thisPath + (thisPath.Length == 0 ? folders[i] : (sSlash + folders[i]));
            }
            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, out mFull, out mHashCode);
        }

        public xDirname LeveledUp()
        {
            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels);

            if (thisPath.Length == 0)
                return new xDirname(this);

            string[] folders = thisPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

            thisPath = string.Empty;
            for (int i = 0; i < (folders.Length - 1); i++)
            {
                thisPath = thisPath + (thisPath.Length == 0 ? folders[i] : (sSlash + folders[i]));
            }

            xDirname f = new xDirname();
            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, out f.mFull, out f.mHashCode);
            return f;
        }

        public void LevelDown(string folder)
        {
            if (folder.Length == 0)
                return;

            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels);

            thisPath = thisPath + (thisPath.Length == 0 ? folder : (sSlash + folder));

            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, out mFull, out mHashCode);
        }

        public xDirname LeveledDown(string folder)
        {
            if (folder.Length == 0)
                return new xDirname(this);

            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels);

            thisPath = thisPath + (thisPath.Length == 0 ? folder : (sSlash + folder));

            xDirname f = new xDirname();
            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, out f.mFull, out f.mHashCode);
            return f;
        }
        #endregion
        #region Object Methods

        public override int GetHashCode()
        {
            return mHashCode;
        }

        public override bool Equals(object o)
        {
            if (o is xDirname)
            {
                xDirname other = (xDirname)o;
                return string.Compare(Full, other.Full, sIgnoreCase) == 0;
            }
            return false;
        }

        /// <summary>
        /// Will return the full filename
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return Full;
        }

        #endregion
        #region Operators

        public static bool operator ==(xDirname inLHS, xDirname inRHS)
        {
            return string.Compare(inLHS.Full, inRHS.Full, sIgnoreCase) == 0;
        }

        public static bool operator !=(xDirname inLHS, xDirname inRHS)
        {
            return string.Compare(inLHS.Full, inRHS.Full, sIgnoreCase) != 0;
        }

        public static xDirname operator +(xDirname inLHS, xDirname inRHS)
        {
            return xDirname.Add(inLHS, inRHS);
        }

        public static xFilename operator +(xDirname inLHS, xFilename inRHS)
        {
            return inRHS.MakeAbsolute(inLHS);
        }

        public static implicit operator string(xDirname f)
        {
            return f.ToString();
        }

        #endregion
        #region UnitTest

        public static bool UnitTest()
        {
            xDirname test1 = new xDirname(@"C:\Temp\Test\Movies");
            xDirname test2 = new xDirname(@"\\cnshaw235\Temp\Test\Movies");
            xDirname test3 = new xDirname(@"Movies");
            xDirname test4 = new xDirname(@"Test\Movies");

            test2 = test2.MakeAbsolute();
            test3 = test3.MakeAbsolute(@"C:\Temp\Test");
            test4 = test4.MakeAbsolute(@"C:\Temp");

            test2.Device = @"C:\";

            return true;
        }

        #endregion
    }
}
