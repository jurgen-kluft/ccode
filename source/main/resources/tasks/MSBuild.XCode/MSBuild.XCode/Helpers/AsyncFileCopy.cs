using System;
using System.IO;
using System.Security.Cryptography;
using System.Text;
using System.Threading;

namespace MSBuild.XCode.Helpers
{
    internal class AsyncUnbufferedCopy
    {
        //file names
        private string _inputfile;
        private string _outputfile;

        //show write progress
        private bool _reportprogress;

        //cursor position
        private int _origRow;
        private int _origCol;

        //number of chunks to copy
        private int _numchunks;

        //track read state and read failed state
        private bool _readfailed;

        //synchronization object
        private readonly object Locker1 = new object();

        //buffer size
        public int CopyBufferSize;
        private long _infilesize;

        //buffer read
        public byte[] Buffer1;
        private int _bytesRead1;

        //buffer overlap
        public byte[] Buffer2;
        private bool _buffer2Dirty;
        private int _bytesRead2;

        //buffer write
        public byte[] Buffer3;

        //total bytes read
        private long _totalbytesread;
        private long _totalbyteswritten;

        //file streams
        private FileStream _infile;
        private FileStream _outfile;

        //secret sauce for unbuffered IO
        const FileOptions FileFlagNoBuffering = (FileOptions)0x20000000;

        public AsyncUnbufferedCopy()
        {
            ProgressFormatStr = "[....]";
            _reportprogress = true;
            _origRow = -1;
            _origCol = -1;
        }

        public string ProgressFormatStr
        {
            get;
            set;
        }

        private void AsyncReadFile()
        {
            //open input file
            try
            {
                _infile = new FileStream(_inputfile, FileMode.Open, FileAccess.Read, FileShare.Read, CopyBufferSize, FileFlagNoBuffering);
            }
            catch (Exception)
            {
                throw;
            }

            int steps = (int)((_infilesize + (CopyBufferSize - 1)) / CopyBufferSize);
            ProgressTracker.Instance.Add(steps);
            int step = 0;

            //if we have data read it
            while (_totalbytesread < _infilesize)
            {
                _bytesRead1 = _infile.Read(Buffer1, 0, CopyBufferSize);
                Monitor.Enter(Locker1);
                try
                {
                    while (_buffer2Dirty) Monitor.Wait(Locker1);
                    Buffer.BlockCopy(Buffer1, 0, Buffer2, 0, _bytesRead1);
                    _buffer2Dirty = true;
                    _bytesRead2 = _bytesRead1;
                    _totalbytesread = _totalbytesread + _bytesRead1;
                    Monitor.PulseAll(Locker1);
                    ProgressTracker.Instance.Next();
                    ++step;
                }
                catch (Exception)
                {
                    _readfailed = true;
                    for (int i = step; i < steps; ++i)
                    {
                        ProgressTracker.Instance.Next();
                    }
                    throw;
                }
                finally
                {
                    Monitor.Exit(Locker1);
                }
            }
            //clean up open handle
            _infile.Close();
            _infile.Dispose();
        }

        private void AsyncWriteFile()
        {
            //open output file set length to prevent growth and file fragmentation and close it.
            //We do this to prevent file fragmentation and make the write as fast as possible.
            string _outputfile_intermediate = _outputfile+".im!";
            try
            {
                if (File.Exists(_outputfile))
                    File.Delete(_outputfile);

                _outfile = new FileStream(_outputfile_intermediate, FileMode.Create, FileAccess.Write, FileShare.None, 8, FileOptions.WriteThrough);

                //set file size to minimum of one buffer to cut down on fragmentation
                _outfile.SetLength(_infilesize > CopyBufferSize ? _infilesize : CopyBufferSize);

                _outfile.Close();
                _outfile.Dispose();
            }
            catch (Exception)
            {
                throw;
            }

            //open file for write unbuffered
            try
            {
                _outfile = new FileStream(_outputfile_intermediate, FileMode.Open, FileAccess.Write, FileShare.None, 8, FileOptions.WriteThrough | FileFlagNoBuffering);
            }
            catch (Exception)
            {
                throw;
            }

            var pctinc = 0.0;
            var progress = pctinc;

            //progress stuff
            if (_reportprogress)
            {
                pctinc = 100.00 / _numchunks;
            }
            while ((_totalbyteswritten < _infilesize) && !_readfailed)
            {
                lock (Locker1)
                {
                    while (!_buffer2Dirty) Monitor.Wait(Locker1);
                    Buffer.BlockCopy(Buffer2, 0, Buffer3, 0, _bytesRead2);
                    _buffer2Dirty = false;
                    _totalbyteswritten = _totalbyteswritten + CopyBufferSize;
                    Monitor.PulseAll(Locker1);

                    if (_reportprogress && (_origCol!=-1 && _origRow!=-1))
                    {
                        if (progress < 101 - pctinc)
                        {
                            progress = progress + pctinc;
                        }
                    }
                }
                try
                {
                    _outfile.Write(Buffer3, 0, CopyBufferSize);
                }
                catch (Exception)
                {
                    throw;
                }
            }

            //close the file handle that was using unbuffered and write through
            _outfile.Close();
            _outfile.Dispose();

            try
            {
                _outfile = new FileStream(_outputfile_intermediate, FileMode.Open, FileAccess.Write, FileShare.None, 8, FileOptions.WriteThrough);
                _outfile.SetLength(_infilesize);
                _outfile.Close();
                _outfile.Dispose();
                File.Move(_outputfile_intermediate, _outputfile);
            }
            catch (Exception)
            {
                throw;
            }
        }

        public int AsyncCopyFileUnbuffered(string inputfile, string outputfile, bool overwrite, bool movefile, bool checksum, int buffersize, bool reportprogress)
        {
            //report write progress
            _reportprogress = reportprogress;

            //set file name globals
            _inputfile = inputfile;
            _outputfile = outputfile;

            //setup single buffer size, remember this will be x3.
            CopyBufferSize = buffersize;

            //buffer read
            Buffer1 = new byte[CopyBufferSize];

            //buffer overlap
            Buffer2 = new byte[CopyBufferSize];

            //buffer write
            Buffer3 = new byte[CopyBufferSize];

            //clear all flags and handles
            _totalbytesread = 0;
            _totalbyteswritten = 0;
            _bytesRead1 = 0;
            _buffer2Dirty = false;

            //if the overwrite flag is set to false check to see if the file is there.
            if (File.Exists(outputfile) && !overwrite)
            {
                return 0;
            }

            //create the directory if it doesn't exist
            if (!Directory.Exists(outputfile))
            {
                try
                {
                    Directory.CreateDirectory(Path.GetDirectoryName(outputfile));
                }
                catch (Exception e)
                {
                    Loggy.Error(String.Format("Error: Create Directory Failed!, " + e.Message));
                    throw;
                }
            }

            //get input file size for later use
            var inputFileInfo = new FileInfo(_inputfile);
            _infilesize = inputFileInfo.Length;

            //get number of buffer sized chunks used to correctly display percent complete.
            if (_infilesize < CopyBufferSize)
                _numchunks = 1;
            else
                _numchunks = (int)(_infilesize / (long)CopyBufferSize);

            //leave a blank line for the progress indicator
            if (_reportprogress)
            {
                Loggy.Info(String.Format(ProgressFormatStr, "[....]"));
            }

            //create read thread and start it.
            var readfile = new Thread(AsyncReadFile) { Name = "ReadThread", IsBackground = true };
            readfile.Start();

            //create write thread and start it.
            var writefile = new Thread(AsyncWriteFile) { Name = "WriteThread", IsBackground = true };
            writefile.Start();

            //wait for threads to finish
            readfile.Join();
            writefile.Join();

            if (movefile && File.Exists(inputfile) && File.Exists(outputfile))
            {
                try
                {
                    File.Delete(inputfile);
                }
                catch (IOException ioex)
                {
                    Loggy.Error(String.Format("Error: File in use or locked!, " + ioex.Message));
                }
                catch (Exception ex)
                {
                    Loggy.Error(String.Format("Error: File failed to delete!, " + ex.Message));
                }
            }
            return 1;
        }

        public string GetSha1HashFromFile(string filename)
        {
            var fs = new FileStream(filename, FileMode.Open, FileAccess.Read, FileShare.Read, CopyBufferSize);
            SHA1 hash = new SHA1CryptoServiceProvider();
            byte[] retVal = hash.ComputeHash(fs);
            fs.Close();

            var sb = new StringBuilder();
            for (var i = 0; i < retVal.Length; i++)
                sb.Append(retVal[i].ToString("x2"));
            return sb.ToString();
        }
    }
}