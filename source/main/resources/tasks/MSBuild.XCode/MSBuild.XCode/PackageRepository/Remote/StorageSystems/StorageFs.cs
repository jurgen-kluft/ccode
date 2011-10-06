using System;
using System.IO;
using System.Security.Cryptography;
using MSBuild.XCode.Helpers;

namespace xstorage_system
{
    public class StorageFs : IStorage
    {
        private string mBasePath;

        public void connect(string connectionURL)
        {
            mBasePath = connectionURL.Replace("fs::", "");
        }

        private string keyToFile(string storage_key)
        {
            string path = string.Empty;

            while (storage_key.Length < 40)
                storage_key = "0" + storage_key;

            const int seperator_char_cnt = 2;
            int cnt = seperator_char_cnt;
            foreach (char c in storage_key)
            {
                if (cnt == 0)
                {
                    path = path + "\\";
                    cnt = seperator_char_cnt;
                }
                --cnt;
                path = path + c;
            }
            return path + ".dat";
        }

        public bool holds(string storage_key)
        {
            try
            {
                string filepath = keyToFile(storage_key);
                return (File.Exists(filepath));
            }
            catch (System.Exception)
            {
            }
            return false;
        }

        public bool submit(string sourceURL, out string storage_key)
        {
            try
            {
                if (File.Exists(sourceURL))
                {
                    FileStream stream = File.OpenRead(sourceURL);
                    SHA1CryptoServiceProvider hash_provider = new SHA1CryptoServiceProvider();
                    byte[] hash = hash_provider.ComputeHash(stream);
                    stream.Close();

                    storage_key = string.Empty;
                    foreach (byte b in hash)
                        storage_key = storage_key + b.ToString("X");

                    string destinationURL = mBasePath + keyToFile(storage_key);
                    string destionationDir = Path.GetDirectoryName(destinationURL);
                    if (!Directory.Exists(destionationDir))
                    {
                        Directory.CreateDirectory(destionationDir);
                    }

                    if (File.Exists(destinationURL))
                    {
                        return true;
                    }
                    else
                    {
                        AsyncUnbufferedCopy xcopy = new AsyncUnbufferedCopy();
                        xcopy.ProgressFormatStr = "Uploading package to storage, progress: {0}%";
                        return CopyFileWithProgress(xcopy, sourceURL, destinationURL);
                    }
                }
                else
                {
                    storage_key = string.Empty;
                }
            }
            catch (System.Exception)
            {
                storage_key = string.Empty;
            }
            return false;
        }

        private static bool CopyFileWithProgress(AsyncUnbufferedCopy xcopy, string src, string dst)
        {
            try
            {
                xcopy.AsyncCopyFileUnbuffered(src, dst, true, false, false, 1 * 1024 * 1024, true);
                return true;
            }
            catch (Exception)
            {
                return false;
            }
            finally
            {
                xcopy = null;
            }
        }

        public bool retrieve(string storage_key, string destinationURL)
        {
            string srcFilename = mBasePath + keyToFile(storage_key);
            string destFilename = destinationURL;
            AsyncUnbufferedCopy xcopy = new AsyncUnbufferedCopy();
            xcopy.ProgressFormatStr = "Downloading package from storage, progress: {0}%";
            return CopyFileWithProgress(xcopy, srcFilename, destFilename);
        }
    }
}
