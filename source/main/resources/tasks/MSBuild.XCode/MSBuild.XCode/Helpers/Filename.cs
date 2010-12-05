using System;
using System.Collections.Generic;
using System.Diagnostics;

namespace MSBuild.XCode.Helpers
{
    /// <summary>
    ///  Represents a filename
    /// </summary>
    /// <remarks>
    ///  Purpose of this class is to make a filename easy and safe to use.
    ///  Note: Memory usage has been preferred over performance.
    /// </remarks>
    /// <interface>
    /// 
    /// e.g. 
    ///             C:\Documents\Music\Beatles\HeyJude.mp3
    ///             \\cnshaw235\Documents\Music\Beatles\HeyJude.mp3
    /// 
    /// Properties
    ///     Extension                                                       (.mp3)
    ///     Device                                                          (e.g. C: or \\cnshaw235)
    ///     DeviceName                                                      (e.g. C or cnshaw235)
    ///     Path                                                            (e.g. Documents\Music\Beatles)
    ///     FullPath                                                        (e.g. C:\Documents\Music\Beatles or \\cnshaw235\Documents\Music\Beatles)
    ///     ShortName                                                       (HeyJude)
    ///     Name                                                            (HeyJude.mp3)
    ///     Relative                                                        (Documents\Music\Beatles\HeyJude.mp3)
    ///     Full                                                            (C:\Documents\Music\Beatles\HeyJude.mp3)
    ///     Levels                                                          (3)
    ///     
    /// Boolean
    ///     HasDevice
    ///     IsNetworkDevice
    ///     HasPath
    ///     HasExtension
    ///     
    /// Methods
    ///     Clear()
    ///     
    ///     ChangeDevice(device)
    ///     ChangePath(directory)
    ///     ChangeAbsolutePath(device+directory)
    ///     ChangeRelative(directory+name+extension)
    ///     ChangeFull(device+directory+name+extension)
    ///     ChangeShortName(name)
    ///     ChangeName(name+extension)
    ///     ChangeExtension(extension)
    /// 
    ///     MakeAbsolute()                                                  (working directory)
    ///     MakeAbsolute(device + directory)
    ///     MakeRelative()                                                  (working directory)
    ///     MakeRelative(device + directory)
    /// 
    ///     LevelUp()                                                       (C:\Documents\Music\HeyJude.mp3)
    ///     LevelDown(folder)                                               (C:\Documents\Music\Beatles\folder\HeyJude.mp3)
    /// 
    /// </interface>
    [Serializable]
    public struct xFilename
    {
        #region Internal Statics

        private const bool sIgnoreCase = true;

        private const char sSlash = '\\';
        private const string sSlashStr = "\\";
        private const string sDoubleSlash = "\\\\";
        private const char sDot = '.';
        private const char sSemi = ':';
        private const string sDotStr = ".";
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

        static internal bool ContainsChars(string ioString, string inChars)
        {
            foreach (char c in ioString)
            {
                if (inChars.IndexOf(c) != -1)
                    return true;
            }
            return false;
        }

        #endregion
        #region Public Statics

        public static readonly xFilename Empty = new xFilename(string.Empty, string.Empty.GetHashCode());

        #endregion Public Statics
        #region Fields

        private int mHashCode;
        private string mFull;

        #endregion Fields
        #region Constructors

        public xFilename(string full)
        {
            mFull = full.Replace('/', '\\');
            mHashCode = -1;
            ChangeFull(mFull);
        }

        private xFilename(string full, int hashcode)
        {
            mFull = full;
            mHashCode = hashcode;
        }

        public xFilename(xFilename inRHS)
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

        public bool IsAbsolute
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                sParseDevice(mFull, out deviceName, out isNetworkDevice);

                return isNetworkDevice || !String.IsNullOrEmpty(deviceName);
            }
        }

        public string Extension
        {
            get
            {
                string name;
                string extension;
                sParseName(mFull, out name, out extension);
                return extension;
            }
            set
            {
                ChangeExtension(value);
            }
        }

        public string Device
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                sParseDevice(mFull, out deviceName, out isNetworkDevice);

                return (isNetworkDevice) ? sDoubleSlash + deviceName : deviceName + sSemi;
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
                string name;
                string extension;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);

                if (isNetworkDevice != value)
                {
                    sConstructFull(deviceName, value, path, name, extension, out mFull, out mHashCode);
                }
            }
        }

        public xDirname Path
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                string path;
                int levels;
                string name;
                string extension;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);
                return new xDirname(path);
            }
            set
            {
                if (value.HasDevice)
                    ChangeAbsolutePath(value);
                else
                    ChangePath(value);
            }
        }

        public xDirname AbsolutePath
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                string path;
                int levels;
                string name;
                string extension;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);

                if (deviceName.Length != 0)
                {
                    return new xDirname((path.Length != 0) ? deviceName + sSemiSlash + path : deviceName);
                }

                xDirname dir = new xDirname(path);
                return dir.MakeAbsolute();
            }
            set
            {
                if (value.HasDevice)
                    ChangeAbsolutePath(value);
                else
                    ChangePath(value);
            }
        }

        public string ShortName
        {
            get
            {
                string name;
                string extension;
                sParseName(mFull, out name, out extension);
                return name;
            }
            set
            {
                ChangeShortName(value);
            }
        }

        public string Name
        {
            get
            {
                string name;
                string extension;
                sParseName(mFull, out name, out extension);
                return name + extension;
            }
            set
            {
                ChangeName(value);
            }
        }

        public string Relative
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                string path;
                int levels;
                string name;
                string extension;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);

                return (path.Length != 0) ? (path + sSlash + name + extension) : name + extension;
            }
            set
            {
                ChangeRelative(value);
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

        private int HashCode
        {
            get
            {
                return mHashCode;
            }
        }

        public int Levels
        {
            get
            {
                string deviceName;
                bool isNetworkDevice;
                string path;
                int levels;
                string name;
                string extension;
                sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);
                return levels;
            }
        }

        #endregion
        #region Private Static Methods

        private static void sConstructFull(string deviceName, bool isNetworkDevice, string path, string name, string extension, out string outFull, out int outHashCode)
        {
            string relative = (path.Length != 0) ? (path + sSlash + name + extension) : name + extension;
            string device = isNetworkDevice ? (sDoubleSlash + deviceName) : (deviceName + sSemi);
            outFull = (deviceName.Length != 0) ? (device + sSlashStr + relative) : relative;
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

        private static void sParseName(string inFull, out string outName, out string outExtension)
        {
            int slashIndex = inFull.LastIndexOf(sSlash);
            if (slashIndex == -1)
                slashIndex = 0;
            else
                slashIndex++;

            int dotIndex = inFull.LastIndexOf(sDot);
            if (dotIndex == -1)
            {
                // No extension
                outExtension = string.Empty;
                outName = inFull.Substring(slashIndex, inFull.Length - slashIndex);
            }
            else
            {
                outExtension = inFull.Substring(dotIndex);
                if (dotIndex != slashIndex)
                    outName = inFull.Substring(slashIndex, dotIndex - slashIndex);
                else
                    outName = string.Empty;
            }
        }

        private static void sParseRelative(string inFullWithoutDevice, out string outPath, out int outLevels, out string outName, out string outExtension)
        {
            // InRelative can be any of these:
            //   - HeyJude.mp3
            //   - Documents\Music\Beatles\HeyJude.mp3

            // Path
            int lastSlashIndex = inFullWithoutDevice.LastIndexOf(sSlash);
            if (lastSlashIndex >= 0)
            {
                outPath = inFullWithoutDevice.Substring(0, lastSlashIndex);
                outName = inFullWithoutDevice.Substring(lastSlashIndex + 1);
            }
            else
            {
                outPath = string.Empty;
                outName = inFullWithoutDevice;
            }

            // Clean up the Path and count the number of folders (levels)
            string[] folders = outPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);
            outLevels = 0;
            outPath = string.Empty;
            for (int i = 0; i < folders.Length; i++)
            {
                outPath = outPath + (outPath.Length == 0 ? folders[i] : (sSlash + folders[i]));
                ++outLevels;
            }

            sParseName(outName, out outName, out outExtension);
        }

        private static void sParseFull(string inFull, out string outDeviceName, out bool outIsNetworkDevice, out string outPath, out int outLevels, out string outName, out string outExtension)
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
            sParseRelative(outPath, out outPath, out outLevels, out outName, out outExtension);
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
            string name;
            string extension;
            sParseFull(mFull, out _deviceName, out _isNetworkDevice, out path, out levels, out name, out extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public void ChangePath(string inPath)
        {
            string path;
            int levels;
            sParsePath(inPath, out path, out levels);

            string deviceName;
            bool isNetworkDevice;
            string _path;
            string name;
            string extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out _path, out levels, out name, out extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public void ChangeAbsolutePath(string deviceAndDirectory)
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
            string name;
            string extension;
            sParseFull(mFull, out _deviceName, out _isNetworkDevice, out _path, out levels, out name, out extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public void ChangeAbsolutePath(xDirname dirName)
        {
            ChangeAbsolutePath(dirName.ToString());
        }

        public void ChangeRelative(string relative)
        {
            string path;
            int levels;
            string name;
            string extension;
            sParseRelative(relative, out path, out levels, out name, out extension);

            string deviceName;
            bool isNetworkDevice;
            string _path;
            int _levels;
            string _name;
            string _extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out _path, out _levels, out _name, out _extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public void ChangeFull(string full)
        {
            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string name;
            string extension;
            sParseFull(full, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public void ChangeShortName(string inName)
        {
            string name = RemoveChars(inName, sIllegalNameChars);

            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string _name;
            string extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out _name, out extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public void ChangeName(string inName)
        {
            string name = RemoveChars(inName, sIllegalNameChars);
            string extension;
            sParseName(name, out name, out extension);

            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string _name;
            string _extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out _name, out _extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, extension, out mFull, out mHashCode);
        }

        public xFilename ChangedName(string inName)
        {
            xFilename f = new xFilename(this);
            f.ChangeName(inName);
            return f;
        }

        public void ChangeExtension(string inExtension)
        {
            string ext = RemoveChars(inExtension, sIllegalNameChars);
            if (ext.StartsWith(sDotStr) && ext.Length > 1)
            {

            }
            else if (ext.Length > 0)
            {
                ext = sDot + ext;
            }
            else
            {
                // Empty extension
                ext = string.Empty;
            }

            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string name;
            string extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);
            sConstructFull(deviceName, isNetworkDevice, path, name, ext, out mFull, out mHashCode);
        }

        public xFilename ChangedExtension(string inExtension)
        {
            xFilename f = new xFilename(this);
            f.ChangeExtension(inExtension);
            return f;
        }

        public xFilename PushedExtension(string ext)
        {
            // ext = .doc
            // C:\Documents\Music\Beatles\HeyJude.mp3 --> C:\Documents\Music\Beatles\HeyJude.mp3.doc
            // Will promote the extension to be part of the current name and set @param ext as the real extension.
            xFilename f = new xFilename(this);
            if (ext.Length != 0)
            {
                if (ext[0] != sDot)
                    ext = sDot + ext;
                f.Full = f.Full + ext;
            }
            return f;
        }

        public void PushExtension(string ext)
        {
            // ext = .doc
            // C:\Documents\Music\Beatles\HeyJude.mp3 --> C:\Documents\Music\Beatles\HeyJude.mp3.doc
            // Will promote the extension to be part of the current name and set @param ext as the real extension.
            if (ext.Length != 0)
            {
                if (ext[0] != sDot)
                    ext = sDot + ext;
                mFull = mFull + ext;
                mHashCode = mFull.ToLower().GetHashCode();
            }
        }

        public xFilename PoppedExtension()
        {
            xFilename f = new xFilename(this);

            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string name;
            string extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);

            // C:\Documents\Music\Beatles\HeyJude.mp3.doc --> C:\Documents\Music\Beatles\HeyJude.mp3
            // Will promote a possible extension that is part of the current name as the real extension.
            int dotIndex = name.LastIndexOf(sDot);
            if (dotIndex >= 0)
            {
                string ext = name.Substring(dotIndex);
                name = dotIndex > 0 ? name.Substring(0, dotIndex) : string.Empty;
                sConstructFull(deviceName, isNetworkDevice, path, name, ext, out f.mFull, out f.mHashCode);
            }
            return f;
        }

        public void PopExtension()
        {
            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string name;
            string extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);

            // C:\Documents\Music\Beatles\HeyJude.mp3.doc --> C:\Documents\Music\Beatles\HeyJude.mp3
            // Will promote a possible extension that is part of the current name as the real extension.
            int dotIndex = name.LastIndexOf(sDot);
            if (dotIndex >= 0)
            {
                string ext = name.Substring(dotIndex);
                name = dotIndex > 0 ? name.Substring(0, dotIndex) : string.Empty;
                sConstructFull(deviceName, isNetworkDevice, path, name, ext, out mFull, out mHashCode);
            }
        }

        public xFilename MakeAbsolute()
        {
            return MakeAbsolute(Environment.CurrentDirectory);
        }

        public xFilename MakeAbsolute(xDirname dir)
        {
            return MakeAbsolute(dir.ToString());
        }

        public xFilename MakeAbsolute(string absolutePath)
        {
            string deviceName;
            bool isNetworkDevice;
            string path;
            int levels;
            string name;
            string extension;
            sParseFull(mFull, out deviceName, out isNetworkDevice, out path, out levels, out name, out extension);

            if (deviceName.Length != 0)
                return new xFilename(this);

            if (path.Length != 0)
            {
                if (absolutePath.EndsWith(sSlashStr))
                    absolutePath += path;
                else
                    absolutePath += sSlash + path;
            }

            xFilename newFilename = new xFilename(this);
            newFilename.ChangeAbsolutePath(absolutePath);
            return newFilename;
        }

        public xFilename MakeRelative()
        {
            return MakeRelative(Environment.CurrentDirectory);
        }

        public xFilename MakeRelative(xDirname dir)
        {
            return MakeRelative(dir.ToString());
        }

        public xFilename MakeRelative(string absolutePath)
        {
            // IN:      C:\My Media\Documents
            // THIS:    C:\My Media\Documents\Music\Beatles\HeyJude.mp3

            // RESULT:  Music\Beatles\HeyJude.mp3

            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            string thisName;
            string thisExtension;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels, out thisName, out thisExtension);

            if (thisPath.Length == 0)
                return new xFilename(this);

            xFilename newFilename = new xFilename(this);
            newFilename.ChangeAbsolutePath(absolutePath);

            string newDeviceName;
            bool newIsNetworkDevice;
            string newPath;
            int newLevels;
            string newName;
            string newExtension;
            sParseFull(newFilename.mFull, out newDeviceName, out newIsNetworkDevice, out newPath, out newLevels, out newName, out newExtension);

            bool sameDevice = true;
            if (String.Compare(thisDeviceName, newDeviceName, sIgnoreCase) != 0)
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

            sConstructFull(newDeviceName, newIsNetworkDevice, newPath, newName, newExtension, out newFilename.mFull, out newFilename.mHashCode);
            return newFilename;
        }

        public void LevelUp()
        {
            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            string thisName;
            string thisExtension;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels, out thisName, out thisExtension);

            if (thisPath.Length == 0)
                return;

            string[] folders = thisPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

            thisPath = string.Empty;
            for (int i = 0; i < (folders.Length - 1); i++)
            {
                thisPath = thisPath + (thisPath.Length == 0 ? folders[i] : (sSlash + folders[i]));
            }
            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, thisName, thisExtension, out mFull, out mHashCode);
        }

        public xFilename LeveledUp()
        {
            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            string thisName;
            string thisExtension;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels, out thisName, out thisExtension);

            if (thisPath.Length == 0)
                return new xFilename(this);

            string[] folders = thisPath.Split(new char[] { sSlash }, StringSplitOptions.RemoveEmptyEntries);

            thisPath = string.Empty;
            for (int i = 0; i < (folders.Length - 1); i++)
            {
                thisPath = thisPath + (thisPath.Length == 0 ? folders[i] : (sSlash + folders[i]));
            }

            xFilename f = new xFilename();
            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, thisName, thisExtension, out f.mFull, out f.mHashCode);
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
            string thisName;
            string thisExtension;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels, out thisName, out thisExtension);

            thisPath = thisPath + (thisPath.Length == 0 ? folder : (sSlash + folder));

            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, thisName, thisExtension, out mFull, out mHashCode);
        }

        public xFilename LeveledDown(string folder)
        {
            if (folder.Length == 0)
                return new xFilename(this);

            string thisDeviceName;
            bool thisIsNetworkDevice;
            string thisPath;
            int thisLevels;
            string thisName;
            string thisExtension;
            sParseFull(mFull, out thisDeviceName, out thisIsNetworkDevice, out thisPath, out thisLevels, out thisName, out thisExtension);

            thisPath = thisPath + (thisPath.Length == 0 ? folder : (sSlash + folder));

            xFilename f = new xFilename();
            sConstructFull(thisDeviceName, thisIsNetworkDevice, thisPath, thisName, thisExtension, out f.mFull, out f.mHashCode);
            return f;
        }


        #endregion
        #region Object Methods

        public override int GetHashCode()
        {
            return HashCode;
        }

        public override bool Equals(object o)
        {
            if (o is xFilename)
            {
                xFilename other = (xFilename)o;
                return (String.Compare(Full, other.Full, sIgnoreCase) == 0);
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

        public static bool operator ==(xFilename inLHS, xFilename inRHS)
        {
            return string.Compare(inLHS.Full, inRHS.Full, sIgnoreCase) == 0;
        }

        public static bool operator !=(xFilename inLHS, xFilename inRHS)
        {
            return string.Compare(inLHS.Full, inRHS.Full, sIgnoreCase) != 0;
        }

        public static implicit operator string(xFilename f)
        {
            return f.Full;
        }

        #endregion
        #region UnitTest

        public static bool UnitTest()
        {
            xFilename test0 = new xFilename(@"C:\Temp\Test\Movie");
            Debug.Assert(test0.Name == "Movie");
            Debug.Assert(test0.Extension == "");
            xFilename test1 = new xFilename(@"C:\Temp\Test\Movie.avi");
            Debug.Assert(test1.Name == "Movie.avi");
            xFilename test2 = new xFilename(@"\\cnshaw235\Temp\Test\Movie.avi");
            Debug.Assert(test2.Name == "Movie.avi");
            xFilename test3 = new xFilename(@"C:\Movie.avi");
            Debug.Assert(test3.Name == "Movie.avi");
            xFilename test4 = new xFilename(@"Temp\Test\Movie.avi");
            Debug.Assert(test4.Name == "Movie.avi");
            xFilename test5 = new xFilename(@"\\cnshaw235\Movie.avi");
            Debug.Assert(test5.Name == "Movie.avi");
            xFilename test6 = new xFilename(@"Movie.avi");
            Debug.Assert(test6.Name == "Movie.avi");
            xFilename test7 = new xFilename(@"C:\Temp\\Test\Movie.avi");
            Debug.Assert(test7.Name == "Movie.avi");

            test1 = test1.MakeRelative(@"C:\Temp");
            test2 = test2.MakeRelative(@"\\cnshaw235\Temp");

            Debug.Assert(test1.GetHashCode() == 0x5960ef95);
            Debug.Assert(test2.GetHashCode() == 0x5960ef95);

            test1 = test1.MakeAbsolute(@"C:\Temp");
            test2 = test2.MakeAbsolute(@"\\cnshaw235\Temp");

            test3 = test3.LeveledDown("Temp");
            test4 = test4.LeveledUp();

            return true;
        }

        #endregion
    }

    /// <summary>
    /// Our own string comparer since we compare filenames in LowerCase due
    /// to possible case differences.
    /// </summary>
    public class FilenameInsensitiveComparer : IEqualityComparer<xFilename>
    {
        public bool Equals(xFilename x, xFilename y)
        {
            return (x == y);
        }

        public int GetHashCode(xFilename obj)
        {
            return obj.GetHashCode();
        }
    }
}