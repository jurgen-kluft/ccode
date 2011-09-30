using System;
using System.Text;
using System.IO;
using System.Net;
using System.Net.Sockets;
using System.Diagnostics;
using System.Security.Cryptography;

namespace xstorage_system
{
    public class StorageFtp : IStorage
    {
        private string mFtpServerIp;
        private int mFtpServerPort;
        private string mUserName;
        private string mPassword;
        private string mBasePath;
        private Ftp.Connection mFtp;

        public void connect(string connectionURL)
        {
            mFtpServerIp = "127.0.0.1";
            mFtpServerPort = 14147;
            mUserName = "admin";
            mPassword = "p1";
            mBasePath = ".storage\\";

            // ftp::ip=127.0.0.1,port=14147,username=admin,password=p1,base=.storage\
            connectionURL = connectionURL.Replace("ftp::", "");
            string[] parts = connectionURL.Split(new char[] { ',' }, StringSplitOptions.RemoveEmptyEntries);
            foreach(string part in parts)
            {
                if (part.StartsWith("ip=", StringComparison.InvariantCultureIgnoreCase))
                {
                    mFtpServerIp = part.Remove(0, "ip=".Length);
                    mFtpServerIp.Trim();
                }
                else if (part.StartsWith("port=", StringComparison.InvariantCultureIgnoreCase))
                {
                    Int32.TryParse(part.Remove(0, "port=".Length), out mFtpServerPort);
                }
                else if (part.StartsWith("username=", StringComparison.InvariantCultureIgnoreCase))
                {
                    mUserName = part.Remove(0, "username=".Length);
                    mUserName.Trim();
                }
                else if (part.StartsWith("password=", StringComparison.InvariantCultureIgnoreCase))
                {
                    mPassword = part.Remove(0, "password=".Length);
                    mPassword.Trim();
                }
                else if (part.StartsWith("base=", StringComparison.InvariantCultureIgnoreCase))
                {
                    mBasePath = part.Remove(0, "base=".Length);
                    mBasePath.Trim();
                }
            }

            mFtp = new Ftp.Connection(mFtpServerIp, mUserName, mPassword, 30*60*60, mFtpServerPort);
            mFtp.Login();

            if (!mFtp.DirectoryExists(mBasePath))
                mFtp.CreateDirectory(mBasePath);
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
                string filepath = mBasePath + keyToFile(storage_key);
                return (mFtp.FileExists(filepath));
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
                    if (!mFtp.DirectoryExists(destionationDir))
                    {
                        if (!mFtp.CreateDirectory(destionationDir))
                            return false;
                    }

                    if (mFtp.FileExists(destinationURL))
                    {
                        return true;
                    }
                    else
                    {
                        mFtp.Upload(sourceURL, destinationURL, false);
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

        public bool retrieve(string storage_key, string destinationURL)
        {
            string srcFilename = mBasePath + keyToFile(storage_key);
            string destFilename = destinationURL;

            return mFtp.Download(srcFilename, destFilename);
        }
    }

namespace Ftp
{
	public class Connection
	{
		public class FtpException : Exception
		{
			public FtpException(string message) : base(message){}
			public FtpException(string message, Exception innerException) : base(message,innerException){}
		}

		private static int BUFFER_SIZE = 512;
		private static Encoding ASCII = Encoding.ASCII;

		private bool verboseDebugging = false;

		// defaults
		private string server = "127.0.0.1";
		private string remotePath = ".";
		private string username = "anonymous";
		private string password = "anonymous@anonymous.net";
		private string message = null;
		private string result = null;

		private int port = 21;
		private int bytes = 0;
		private int resultCode = 0;

		private bool loggedin = false;
		private bool binMode = false;

		private Byte[] buffer = new Byte[BUFFER_SIZE];
		private Socket clientSocket = null;

		private int timeoutSeconds = 10;

		/// <summary>
		/// Default constructor
		/// </summary>
		public Connection()
		{
		}
		/// <summary>
		/// 
		/// </summary>
		/// <param name="server"></param>
		/// <param name="username"></param>
		/// <param name="password"></param>
		public Connection(string server, string username, string password)
		{
			this.server = server;
			this.username = username;
			this.password = password;
		}
		/// <summary>
		/// 
		/// </summary>
		/// <param name="server"></param>
		/// <param name="username"></param>
		/// <param name="password"></param>
		/// <param name="timeoutSeconds"></param>
		/// <param name="port"></param>
		public Connection(string server, string username, string password, int timeoutSeconds, int port)
		{
			this.server = server;
			this.username = username;
			this.password = password;
			this.timeoutSeconds = timeoutSeconds;
			this.port = port;
		}

		/// <summary>
		/// Display all communications to the debug log
		/// </summary>
		public bool VerboseDebugging
		{
			get
			{
				return this.verboseDebugging;
			}
			set
			{
				this.verboseDebugging = value;
			}
		}
		/// <summary>
		/// Remote server port. Typically TCP 21
		/// </summary>
		public int Port
		{
			get
			{
				return this.port;
			}
			set
			{
				this.port = value;
			}
		}
		/// <summary>
		/// Timeout waiting for a response from server, in seconds.
		/// </summary>
		public int Timeout
		{
			get
			{
				return this.timeoutSeconds;
			}
			set
			{
				this.timeoutSeconds = value;
			}
		}
		/// <summary>
		/// Gets and Sets the name of the FTP server.
		/// </summary>
		/// <returns></returns>
		public string Server
		{
			get
			{
				return this.server;
			}
			set
			{
				this.server = value;
			}
		}
		/// <summary>
		/// Gets and Sets the port number.
		/// </summary>
		/// <returns></returns>
		public int RemotePort
		{
			get
			{
				return this.port;
			}
			set
			{
				this.port = value;
			}
		}
		/// <summary>
		/// GetS and Sets the remote directory.
		/// </summary>
		public string RemotePath
		{
			get
			{
				return this.remotePath;
			}
			set
			{
				this.remotePath = value;
			}

		}
		/// <summary>
		/// Gets and Sets the username.
		/// </summary>
		public string Username
		{
			get
			{
				return this.username;
			}
			set
			{
				this.username = value;
			}
		}
		/// <summary>
		/// Gets and Set the password.
		/// </summary>
		public string Password
		{
			get
			{
				return this.password;
			}
			set
			{
				this.password = value;
			}
		}

		/// <summary>
		/// If the value of mode is true, set binary mode for downloads, else, Ascii mode.
		/// </summary>
		public bool BinaryMode
		{
			get
			{
				return this.binMode;
			}
			set
			{
				if ( this.binMode == value ) return;

				if ( value )
					sendCommand("TYPE I");

				else
					sendCommand("TYPE A");

				if ( this.resultCode != 200 ) throw new FtpException(result.Substring(4));
			}
		}
		/// <summary>
		/// Login to the remote server.
		/// </summary>
		public void Login()
		{
			if ( this.loggedin ) this.Close();

			Debug.WriteLine("Opening connection to " + this.server, "Ftp.Connection" );

			IPAddress addr = null;
			IPEndPoint ep = null;

			try
			{
				this.clientSocket = new Socket( AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.Tcp );
				//addr = Dns.Resolve(this.server).AddressList[0];
                addr = Dns.GetHostEntry(this.server).AddressList[0];
				ep = new IPEndPoint( addr, this.port );
				this.clientSocket.Connect(ep);
			}
			catch(Exception ex)
			{
				// doubtful
				if ( this.clientSocket != null && this.clientSocket.Connected ) this.clientSocket.Close();

				throw new FtpException("Couldn't connect to remote server",ex);
			}

			this.readResponse();

			if(this.resultCode != 220)
			{
				this.Close();
				throw new FtpException(this.result.Substring(4));
			}

			this.sendCommand( "USER " + username );

			if( !(this.resultCode == 331 || this.resultCode == 230) )
			{
				this.cleanup();
				throw new FtpException(this.result.Substring(4));
			}

			if( this.resultCode != 230 )
			{
				this.sendCommand( "PASS " + password );

				if( !(this.resultCode == 230 || this.resultCode == 202) )
				{
					this.cleanup();
					throw new FtpException(this.result.Substring(4));
				}
			}

			this.loggedin = true;

			Debug.WriteLine( "Connected to " + this.server, "Ftp.Connection" );

			this.ChangeDir(this.remotePath, true);
		}
		
		/// <summary>
		/// Close the FTP connection.
		/// </summary>
		public void Close()
		{
			Debug.WriteLine("Closing connection to " + this.server, "Ftp.Connection" );

			if( this.clientSocket != null )
			{
				this.sendCommand("QUIT");
			}

			this.cleanup();
		}

		/// <summary>
		/// Return a string array containing the remote directory's file list.
		/// </summary>
		/// <returns></returns>
		public string[] GetFileList()
		{
            string[] list;
			this.GetFileList("*.*", out list);
            return list;
		}

        public bool DirectoryExists(string filepath)
        {
            string[] files;
            GetFileList(filepath + "*.*", out files);
            return files.Length >= 1;
        }

        public bool FileExists(string filepath)
        {
            string[] files;
            GetFileList(Path.GetDirectoryName(filepath) + "*" + Path.GetExtension(filepath), out files);
            string filea = Path.GetFileNameWithoutExtension(filepath);
            foreach(string file in files)
            {
                string fileb = Path.GetFileNameWithoutExtension(file);
                if (String.Compare(filea, fileb, true) == 0)
                {
                       return true;
                }
            }
            return false;
        }

		/// <summary>
		/// Return a string array containing the remote directory's file list.
		/// </summary>
		/// <param name="mask"></param>
		/// <returns></returns>
        public bool GetFileList(string mask, out string[] list)
        {
            if (!this.loggedin) this.Login();

            Socket cSocket = createDataSocket();

            this.sendCommand("NLST " + mask);

            if (this.resultCode == 550)
            {
                cSocket.Close(); 
                list = new string[] { };
                return true;
            }

            if (!(this.resultCode == 150 || this.resultCode == 125))
            {
                Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                list = new string[] { };
                return false;
            }

            this.message = "";

            DateTime timeout = DateTime.Now.AddSeconds(this.timeoutSeconds);

            while (timeout > DateTime.Now)
            {
                int bytes = cSocket.Receive(buffer, buffer.Length, 0);
                this.message += ASCII.GetString(buffer, 0, bytes);

                if (bytes < this.buffer.Length) break;
            }

            string[] msg = this.message.Replace("\r", "").Split('\n');

            cSocket.Close();

            if (this.message.IndexOf("No such file or directory") != -1)
                msg = new string[] { };

            this.readResponse();

            if (this.resultCode != 226)
                msg = new string[] { };

            list = msg;
            return true;
        }
		
		/// <summary>
		/// Return the size of a file.
		/// </summary>
		/// <param name="fileName"></param>
		/// <returns></returns>
		public long GetFileSize(string fileName)
		{
			if ( !this.loggedin ) this.Login();

			this.sendCommand("SIZE " + fileName);
			long size=0;

			if ( this.resultCode == 213 )
				size = long.Parse(this.result.Substring(4));

			else
				throw new FtpException(this.result.Substring(4));

			return size;
		}
	
		
		/// <summary>
		/// Download a file to the Assembly's local directory,
		/// keeping the same file name.
		/// </summary>
		/// <param name="remFileName"></param>
		public void Download(string remFileName)
		{
			this.Download(remFileName,"",false);
		}

		/// <summary>
		/// Download a remote file to the Assembly's local directory,
		/// keeping the same file name, and set the resume flag.
		/// </summary>
		/// <param name="remFileName"></param>
		/// <param name="resume"></param>
		public void Download(string remFileName,Boolean resume)
		{
			this.Download(remFileName,"",resume);
		}
		
		/// <summary>
		/// Download a remote file to a local file name which can include
		/// a path. The local file name will be created or overwritten,
		/// but the path must exist.
		/// </summary>
		/// <param name="remFileName"></param>
		/// <param name="locFileName"></param>
		public bool Download(string remFileName,string locFileName)
		{
			return this.Download(remFileName,locFileName,false);
		}

		/// <summary>
		/// Download a remote file to a local file name which can include
		/// a path, and set the resume flag. The local file name will be
		/// created or overwritten, but the path must exist.
		/// </summary>
		/// <param name="remFileName"></param>
		/// <param name="locFileName"></param>
		/// <param name="resume"></param>
        public bool Download(string remFileName, string locFileName, Boolean resume)
        {
            if (!this.loggedin) this.Login();

            this.BinaryMode = true;

            Debug.WriteLine("Downloading file " + remFileName + " from " + server + "/" + remotePath, "Ftp.Connection");

            if (locFileName.Equals(""))
            {
                locFileName = remFileName;
            }

            FileStream output = null;
            try
            {
                output = new FileStream(locFileName, FileMode.OpenOrCreate);
            }
            catch (SystemException)
            {
                Debug.WriteLine("Exception: failed to open/create file for writing'" + locFileName + "'", "Ftp.Connection");
                return false;
            }

            Socket cSocket = createDataSocket();

            long offset = 0;

            if (resume)
            {
                offset = output.Length;

                if (offset > 0)
                {
                    this.sendCommand("REST " + offset);
                    if (this.resultCode != 350)
                    {
                        //Server doesn't support resuming
                        offset = 0;
                        Debug.WriteLine("Resuming not supported:" + result.Substring(4), "Ftp.Connection");
                    }
                    else
                    {
                        Debug.WriteLine("Resuming at offset " + offset, "Ftp.Connection");
                        output.Seek(offset, SeekOrigin.Begin);
                    }
                }
            }

            this.sendCommand("RETR " + remFileName);

            if (this.resultCode != 150 && this.resultCode != 125)
            {
                Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                return false;
            }

            Int64 totalBytes = GetFileSize(remFileName);
            Int64 transferedBytes = 0;

            // Set the size of the file here
            try
            {
                output.SetLength(totalBytes);
            }
            catch (SystemException)
            {
                Debug.WriteLine("Exception: failed to download file '" + remFileName + "' to '" + locFileName + "'", "Ftp.Connection");
                output.Close();
                cSocket.Close();
                return false;
            }

            DateTime timeout = DateTime.Now.AddSeconds(this.timeoutSeconds);

            int cl = Console.CursorLeft;
            int ct = Console.CursorTop;

            while (timeout > DateTime.Now)
            {
                this.bytes = cSocket.Receive(buffer, buffer.Length, 0);
                output.Write(this.buffer, 0, this.bytes);
                if (this.bytes <= 0)
                    break;

                transferedBytes += this.bytes;
                if (totalBytes > 0)
                {
                    Console.SetCursorPosition(cl, ct);
                    Console.Write("{0}%", (transferedBytes * 100) / totalBytes);
                }
            }

            output.Close();

            if (cSocket.Connected)
                cSocket.Close();

            this.readResponse();

            if (this.resultCode != 226 && this.resultCode != 250)
            {
                Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                return false;
            }

            return true;
        }

		/// <summary>
		/// Upload a file.
		/// </summary>
		/// <param name="fileName"></param>
        public void Upload(string filepath)
		{
			this.Upload(filepath, Path.GetFileName(filepath), false);
		}

		/// <summary>
		/// Upload a file and set the resume flag.
		/// </summary>
		/// <param name="fileName"></param>
		/// <param name="resume"></param>
		public bool Upload(string src_filepath, string dst_filepath, bool resume)
		{
			if ( !this.loggedin ) this.Login();

			Socket cSocket = null ;
			long offset = 0;

			if ( resume )
			{
				try
				{
					this.BinaryMode = true;
                    offset = GetFileSize(Path.GetFileName(dst_filepath));
				}
				catch(Exception)
				{
					// file not exist
					offset = 0;
				}
			}

			// open stream to read file
            FileStream input = new FileStream(src_filepath, FileMode.Open, FileAccess.Read, FileShare.Read);

			if ( resume && input.Length < offset )
			{
				// different file size
                Debug.WriteLine("Overwriting " + dst_filepath, "Ftp.Connection");
				offset = 0;
			}
			else if ( resume && input.Length == offset )
			{
				// file done
				input.Close();
                Debug.WriteLine("Skipping completed " + dst_filepath + " - turn resume off to not detect.", "Ftp.Connection");
				return false;
			}

			// don't create until we know that we need it
			cSocket = this.createDataSocket();

			if ( offset > 0 )
			{
				this.sendCommand( "REST " + offset );
				if ( this.resultCode != 350 )
				{
					Debug.WriteLine("Resuming not supported", "Ftp.Connection");
					offset = 0;
				}
			}

            if (this.ChangeDir("..", true))
            {
                string dst_directory = Path.GetDirectoryName(dst_filepath);
                if (this.CreateDirectory(dst_directory))
                {
                    if (this.ChangeDir(Path.GetDirectoryName(dst_filepath), false))
                    {
                        this.sendCommand("STOR " + Path.GetFileName(dst_filepath));

                        if (this.resultCode != 125 && this.resultCode != 150)
                        {
                            input.Close();
                            cSocket.Close();
                            Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                            return false;
                        }

                        if (offset != 0)
                        {
                            Debug.WriteLine("Resuming at offset " + offset, "Ftp.Connection");
                            input.Seek(offset, SeekOrigin.Begin);
                        }

                        Debug.WriteLine("Uploading file '" + src_filepath + "' to '" + dst_filepath + "'", "Ftp.Connection");

                        int cl = Console.CursorLeft;
                        int ct = Console.CursorTop;

                        Int64 totalBytes = input.Length;
                        Int64 transferedBytes = 0;
                        while ((bytes = input.Read(buffer, 0, buffer.Length)) > 0)
                        {
                            Console.SetCursorPosition(cl, ct);
                            Console.Write("{0}%", (transferedBytes * 100) / totalBytes);
                            cSocket.Send(buffer, bytes, 0);
                            transferedBytes += bytes;
                        }

                        input.Close();

                        if (cSocket.Connected)
                        {
                            cSocket.Close();
                        }

                        this.readResponse();

                        if (this.resultCode != 226 && this.resultCode != 250)
                        {
                            Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                            return false;
                        }
                    }
                    else
                    {
                        Debug.WriteLine("Error: unable to change directory to '" + dst_directory + "'", "Ftp.Connection");
                        return false;
                    }
                }
                else
                {
                    Debug.WriteLine("Error: unable to create directory '" + dst_directory + "'", "Ftp.Connection");
                    return false;
                }
            }
            else
            {
                Debug.WriteLine("Error: unable to change directory to '/'", "Ftp.Connection");
                return false;
            }
            return true;
		}
		
		/// <summary>
		/// Upload a directory and its file contents
		/// </summary>
		/// <param name="path"></param>
		/// <param name="recurse">Whether to recurse sub directories</param>
		public void UploadDirectory(string path, bool recurse)
		{
			this.UploadDirectory(path,recurse,"*.*");
		}
		
		/// <summary>
		/// Upload a directory and its file contents
		/// </summary>
		/// <param name="path"></param>
		/// <param name="recurse">Whether to recurse sub directories</param>
		/// <param name="mask">Only upload files of the given mask - everything is '*.*'</param>
		public void UploadDirectory(string path, bool recurse, string mask)
		{
			string[] dirs = path.Replace("/",@"\").Split('\\');
			string rootDir = dirs[ dirs.Length - 1 ];

			// make the root dir if it does not exist
            string[] filelist;
            this.GetFileList(rootDir, out filelist);
			if ( filelist.Length < 1 ) 
                this.CreateDirectory(rootDir);

			this.ChangeDir(rootDir, true);

			foreach ( string file in Directory.GetFiles(path,mask) )
			{
				this.Upload(file, Path.GetFileName(file), true);
			}
			if ( recurse )
			{
				foreach ( string directory in Directory.GetDirectories(path) )
				{
					this.UploadDirectory(directory,recurse,mask);
				}
			}

			this.ChangeDir("..", true);
		}

		/// <summary>
		/// Delete a file from the remote FTP server.
		/// </summary>
		/// <param name="fileName"></param>
		public void DeleteFile(string fileName)
		{
			if ( !this.loggedin ) this.Login();

			this.sendCommand( "DELE " + fileName );

			if ( this.resultCode != 250 ) throw new FtpException(this.result.Substring(4));

			Debug.WriteLine( "Deleted file " + fileName, "Ftp.Connection" );
		}

		/// <summary>
		/// Rename a file on the remote FTP server.
		/// </summary>
		/// <param name="oldFileName"></param>
		/// <param name="newFileName"></param>
		/// <param name="overwrite">setting to false will throw exception if it exists</param>
		public void RenameFile(string oldFileName,string newFileName, bool overwrite)
		{
			if ( !this.loggedin ) this.Login();

			this.sendCommand( "RNFR " + oldFileName );

			if ( this.resultCode != 350 ) throw new FtpException(this.result.Substring(4));

            string[] filelist;
            this.GetFileList(newFileName, out filelist);

			if ( !overwrite && filelist.Length > 0 ) throw new FtpException("File already exists");

			this.sendCommand( "RNTO " + newFileName );

			if ( this.resultCode != 250 ) throw new FtpException(this.result.Substring(4));

			Debug.WriteLine( "Renamed file " + oldFileName + " to " + newFileName, "Ftp.Connection" );
		}
		
		/// <summary>
		/// Create a directory on the remote FTP server.
		/// </summary>
		/// <param name="dirName"></param>
		public bool CreateDirectory(string dirName)
		{
			if ( !this.loggedin ) this.Login();
            if (String.IsNullOrEmpty(dirName))
                return true;

			this.sendCommand( "MKD " + dirName );

            if (this.resultCode == 550)
            {
                Debug.WriteLine("Directory already exists" + dirName, "Ftp.Connection"); 
                return true;
            }

            if (this.resultCode != 250 && this.resultCode != 257)
            {
                Debug.WriteLine("Exception " + this.result.Substring(4), "Ftp.Connection");
                return false;
            }

			Debug.WriteLine( "Created directory " + dirName, "Ftp.Connection" );
            return true;
		}

		/// <summary>
		/// Delete a directory on the remote FTP server.
		/// </summary>
		/// <param name="dirName"></param>
		public void RemoveDir(string dirName)
		{
			if ( !this.loggedin ) this.Login();

			this.sendCommand( "RMD " + dirName );

			if ( this.resultCode != 250 ) throw new FtpException(this.result.Substring(4));

			Debug.WriteLine( "Removed directory " + dirName, "Ftp.Connection" );
		}

		/// <summary>
		/// Change the current working directory on the remote FTP server.
		/// </summary>
		/// <param name="dirName"></param>
		public bool ChangeDir(string dirName, bool absolute)
		{
			if( dirName == null || dirName.Equals(".") || dirName.Length == 0 )
			{
				return false;
			}

			if ( !this.loggedin ) this.Login();

            if (absolute)
            {
                // Move up to root
                while (true)
                {
                    this.sendCommand("CWD ..");
                    if (this.resultCode != 257 && this.resultCode!=250) throw new FtpException(result.Substring(4));
                    string path = this.message.Split('"')[1];
                    if (String.IsNullOrEmpty(path) || path=="/")
                        break;
                }
            }

            if (dirName == ".." && String.IsNullOrEmpty(dirName))
            {
                Debug.WriteLine("Current directory is '/'", "Ftp.Connection");
                return true;
            }

			this.sendCommand("CWD " + dirName );
            if (this.resultCode != 250)
            {
                Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                return false;
            }

			this.sendCommand( "PWD" );
            if (this.resultCode != 257)
            {
                Debug.WriteLine("Exception: " + result.Substring(4), "Ftp.Connection");
                return false;
            }

			// gonna have to do better than this....
			this.remotePath = this.message.Split('"')[1];

			Debug.WriteLine( "Current directory is '" + this.remotePath + "'", "Ftp.Connection" );
            return true;
		}

		/// <summary>
		/// 
		/// </summary>
		private void readResponse()
		{
			this.message = "";
			this.result = this.readLine();

			if ( this.result.Length > 3 )
				this.resultCode = int.Parse( this.result.Substring(0,3) );
			else
				this.result = null;
		}

		/// <summary>
		/// 
		/// </summary>
		/// <returns></returns>
		private string readLine()
		{
			while(true)
			{
				this.bytes = clientSocket.Receive( this.buffer, this.buffer.Length, 0 );
				this.message += ASCII.GetString( this.buffer, 0, this.bytes );

				if ( this.bytes < this.buffer.Length )
				{
					break;
				}
			}

			string[] msg = this.message.Split('\n');

			if ( msg.Length > 2 )
				this.message = msg[ msg.Length - 2 ];

			else
				this.message = msg[0];


			if ( this.message.Length > 4 && !this.message.Substring(3,1).Equals(" ") ) 
                return this.readLine();

			if ( this.verboseDebugging )
			{
				for(int i = 0; i < msg.Length - 1; i++)
				{
					Debug.Write( msg[i], "Ftp.Connection" );
				}
			}

			return message;
		}

		/// <summary>
		/// 
		/// </summary>
		/// <param name="command"></param>
		private void sendCommand(String command)
		{
			if ( this.verboseDebugging ) Debug.WriteLine(command,"Ftp.Connection");

			Byte[] cmdBytes = Encoding.ASCII.GetBytes( ( command + "\r\n" ).ToCharArray() );
			clientSocket.Send( cmdBytes, cmdBytes.Length, 0);
			this.readResponse();
		}

		/// <summary>
		/// when doing data transfers, we need to open another socket for it.
		/// </summary>
		/// <returns>Connected socket</returns>
		private Socket createDataSocket()
		{
			this.sendCommand("PASV");

			if ( this.resultCode != 227 ) throw new FtpException(this.result.Substring(4));

			int index1 = this.result.IndexOf('(');
			int index2 = this.result.IndexOf(')');

			string ipData = this.result.Substring(index1+1,index2-index1-1);

			int[] parts = new int[6];

			int len = ipData.Length;
			int partCount = 0;
			string buf="";

			for (int i = 0; i < len && partCount <= 6; i++)
			{
				char ch = char.Parse( ipData.Substring(i,1) );

				if ( char.IsDigit(ch) )
					buf+=ch;

				else if (ch != ',')
					throw new FtpException("Malformed PASV result: " + result);

				if ( ch == ',' || i+1 == len )
				{
					try
					{
						parts[partCount++] = int.Parse(buf);
						buf = "";
					}
					catch (Exception ex)
					{
						throw new FtpException("Malformed PASV result (not supported?): " + this.result, ex);
					}
				}
			}

			string ipAddress = parts[0] + "."+ parts[1]+ "." + parts[2] + "." + parts[3];

			int port = (parts[4] << 8) + parts[5];

			Socket socket = null;
			IPEndPoint ep = null;

			try
			{
				socket = new Socket(AddressFamily.InterNetwork,SocketType.Stream,ProtocolType.Tcp);
                IPAddress ip = Dns.GetHostEntry(ipAddress).AddressList[0];
				ep = new IPEndPoint(ip, port);
				socket.Connect(ep);
			}
			catch(Exception ex)
			{
				// doubtful....
				if ( socket != null && socket.Connected )
                    socket.Close();

				throw new FtpException("Can't connect to remote server", ex);
			}

			return socket;
		}
		
		/// <summary>
		/// Always release those sockets.
		/// </summary>
		private void cleanup()
		{
			if ( this.clientSocket!=null )
			{
				this.clientSocket.Close();
				this.clientSocket = null;
			}
			this.loggedin = false;
		}

		/// <summary>
		/// Destructor
		/// </summary>
		~Connection()
		{
			this.cleanup();
		}


		/**************************************************************************************************************/
		#region Async methods (auto generated)

		private delegate void LoginCallback();
		public System.IAsyncResult BeginLogin(  System.AsyncCallback callback )
		{
			LoginCallback ftpCallback = new LoginCallback( this.Login);
			return ftpCallback.BeginInvoke(callback, null);
		}
		private delegate void CloseCallback();
		public System.IAsyncResult BeginClose(  System.AsyncCallback callback )
		{
			CloseCallback ftpCallback = new CloseCallback( this.Close);
			return ftpCallback.BeginInvoke(callback, null);
		}
		private delegate String[] GetFileListCallback();
		public System.IAsyncResult BeginGetFileList(  System.AsyncCallback callback )
		{
			GetFileListCallback ftpCallback = new GetFileListCallback( this.GetFileList);
			return ftpCallback.BeginInvoke(callback, null);
		}
		private delegate bool GetFileListMaskCallback(String mask, out String[] list);
        public System.IAsyncResult BeginGetFileList(String mask, out String[] list, System.AsyncCallback callback)
		{
			GetFileListMaskCallback ftpCallback = new GetFileListMaskCallback(this.GetFileList);
			return ftpCallback.BeginInvoke(mask, out list, callback, null);
		}
		private delegate Int64 GetFileSizeCallback(String fileName);
		public System.IAsyncResult BeginGetFileSize( String fileName, System.AsyncCallback callback )
		{
			GetFileSizeCallback ftpCallback = new GetFileSizeCallback(this.GetFileSize);
			return ftpCallback.BeginInvoke(fileName, callback, null);
		}
		private delegate void DownloadCallback(String remFileName);
		public System.IAsyncResult BeginDownload( String remFileName, System.AsyncCallback callback )
		{
			DownloadCallback ftpCallback = new DownloadCallback(this.Download);
			return ftpCallback.BeginInvoke(remFileName, callback, null);
		}
		private delegate void DownloadFileNameResumeCallback(String remFileName,Boolean resume);
		public System.IAsyncResult BeginDownload( String remFileName,Boolean resume, System.AsyncCallback callback )
		{
			DownloadFileNameResumeCallback ftpCallback = new DownloadFileNameResumeCallback(this.Download);
			return ftpCallback.BeginInvoke(remFileName, resume, callback, null);
		}
		private delegate bool DownloadFileNameFileNameCallback(String remFileName,String locFileName);
		public System.IAsyncResult BeginDownload( String remFileName,String locFileName, System.AsyncCallback callback )
		{
			DownloadFileNameFileNameCallback ftpCallback = new DownloadFileNameFileNameCallback(this.Download);
			return ftpCallback.BeginInvoke(remFileName, locFileName, callback, null);
		}
		private delegate bool DownloadFileNameFileNameResumeCallback(String remFileName,String locFileName,Boolean resume);
		public System.IAsyncResult BeginDownload( String remFileName,String locFileName,Boolean resume, System.AsyncCallback callback )
		{
			DownloadFileNameFileNameResumeCallback ftpCallback = new DownloadFileNameFileNameResumeCallback(this.Download);
			return ftpCallback.BeginInvoke(remFileName, locFileName, resume, callback, null);
		}
		private delegate void UploadCallback(String fileName);
		public System.IAsyncResult BeginUpload( String fileName, System.AsyncCallback callback )
		{
			UploadCallback ftpCallback = new UploadCallback(this.Upload);
			return ftpCallback.BeginInvoke(fileName, callback, null);
		}
        private delegate bool UploadFileNameResumeCallback(String src_filepath, String dst_filepath, Boolean resume);
		public System.IAsyncResult BeginUpload( String src_filepath, String dst_filepath, Boolean resume, System.AsyncCallback callback )
		{
			UploadFileNameResumeCallback ftpCallback = new UploadFileNameResumeCallback(this.Upload);
            return ftpCallback.BeginInvoke(src_filepath, dst_filepath, resume, callback, null);
		}
		private delegate void UploadDirectoryCallback(String path,Boolean recurse);
		public System.IAsyncResult BeginUploadDirectory( String path,Boolean recurse, System.AsyncCallback callback )
		{
			UploadDirectoryCallback ftpCallback = new UploadDirectoryCallback(this.UploadDirectory);
			return ftpCallback.BeginInvoke(path, recurse, callback, null);
		}
		private delegate void UploadDirectoryPathRecurseMaskCallback(String path,Boolean recurse,String mask);
		public System.IAsyncResult BeginUploadDirectory( String path,Boolean recurse,String mask, System.AsyncCallback callback )
		{
			UploadDirectoryPathRecurseMaskCallback ftpCallback = new UploadDirectoryPathRecurseMaskCallback(this.UploadDirectory);
			return ftpCallback.BeginInvoke(path, recurse, mask, callback, null);
		}
		private delegate void DeleteFileCallback(String fileName);
		public System.IAsyncResult BeginDeleteFile( String fileName, System.AsyncCallback callback )
		{
			DeleteFileCallback ftpCallback = new DeleteFileCallback(this.DeleteFile);
			return ftpCallback.BeginInvoke(fileName, callback, null);
		}
		private delegate void RenameFileCallback(String oldFileName,String newFileName,Boolean overwrite);
		public System.IAsyncResult BeginRenameFile( String oldFileName,String newFileName,Boolean overwrite, System.AsyncCallback callback )
		{
			RenameFileCallback ftpCallback = new RenameFileCallback(this.RenameFile);
			return ftpCallback.BeginInvoke(oldFileName, newFileName, overwrite, callback, null);
		}
		private delegate bool MakeDirCallback(String dirName);
		public System.IAsyncResult BeginMakeDir( String dirName, System.AsyncCallback callback )
		{
            MakeDirCallback ftpCallback = new MakeDirCallback(this.CreateDirectory);
			return ftpCallback.BeginInvoke(dirName, callback, null);
		}
		private delegate void RemoveDirCallback(String dirName);
		public System.IAsyncResult BeginRemoveDir( String dirName, System.AsyncCallback callback )
		{
			RemoveDirCallback ftpCallback = new RemoveDirCallback(this.RemoveDir);
			return ftpCallback.BeginInvoke(dirName, callback, null);
		}
		private delegate bool ChangeDirCallback(String dirName, Boolean absolute);
        public System.IAsyncResult BeginChangeDir(String dirName, Boolean absolute, System.AsyncCallback callback)
		{
			ChangeDirCallback ftpCallback = new ChangeDirCallback(this.ChangeDir);
			return ftpCallback.BeginInvoke(dirName, absolute, callback, null);
		}

		#endregion
	}
}
}
