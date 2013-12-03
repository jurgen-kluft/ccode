// $Id: SvnVersionTest.cs 301 2007-02-17 22:26:16Z joshuaflanagan $

using System.IO;
using MSBuild.Community.Tasks.Subversion;
using NUnit.Framework;

namespace MSBuild.Community.Tasks.Tests.Subversion
{
    /// <summary>
    /// Summary description for SvnVersionTest
    /// </summary>
    [TestFixture]
    public class SvnVersionTest
    {
        [Test(Description="Test SVN Version of project source directory")]
        public void SvnVersionExecute()
        {
            SvnVersion task = new SvnVersion();
            task.BuildEngine = new MockBuild();

            string prjRootPath = TaskUtility.getProjectRootDirectory(true);
            task.LocalPath = Path.Combine(prjRootPath, @"Source");

            Assert.IsTrue(task.Execute(), "Execute Failed");

            Assert.IsTrue(task.Revision > 0, "Invalid Revision Number");
        }
    }
}