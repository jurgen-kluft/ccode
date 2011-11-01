// $Id: FtpUploadTest.cs 281 2006-12-12 05:24:25Z joshuaflanagan $

using System;
using System.IO;
using System.Text;
using NUnit.Framework;

namespace MSBuild.Community.Tasks.Tests
{
    /// <summary>
    /// Summary description for FtpTest
    /// </summary>
    [TestFixture]
    public class FtpUploadTest
    {
        [Test(Description="Upload zip file to localhost"), Explicit]
        public void FtpExecute()
        {
            ZipTest zip = new ZipTest();
            zip.ZipExecute(); 
            
            FtpUpload task = new FtpUpload();
            task.BuildEngine = new MockBuild();
            string testDir = TaskUtility.TestDirectory;
            task.LocalFile = Path.Combine(testDir, ZipTest.ZIP_FILE_NAME);
            task.RemoteUri = "ftp://localhost/MSBuild.Community.Tasks.zip";
            Assert.IsTrue(task.Execute(), "Execute Failed");

        }
    }
}
