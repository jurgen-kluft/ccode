// $Id: XmlUpdateTest.cs 281 2006-12-12 05:24:25Z joshuaflanagan $

using System;
using System.IO;
using System.Text;
using NUnit.Framework;

namespace MSBuild.Community.Tasks.Tests
{
    /// <summary>
    /// Summary description for XmlUpdateTest
    /// </summary>
    [TestFixture]
    public class XmlUpdateTest
    {
        string prjRootPath;
        string testFile;

        public XmlUpdateTest()
        {
            prjRootPath = TaskUtility.getProjectRootDirectory(true);
            testFile = Path.Combine(prjRootPath, @"Source\Test\Subversion.proj");
        }

        [SetUp]
        public void Setup()
        {
            string sourceFile = Path.Combine(prjRootPath, @"Source\Subversion.proj");
            File.Copy(sourceFile, testFile, true);
        }

        [TearDown]
        public void Cleanup()
        {
            File.Delete(testFile);
        }

        [Test(Description="Update an XML file with XPath navigation")]
        public void XmlUpdateExecute()
        {
            XmlUpdate task = new XmlUpdate();
            task.BuildEngine = new MockBuild();
            
            task.Prefix = "n";
            task.Namespace = "http://schemas.microsoft.com/developer/msbuild/2003";
            task.XmlFileName = testFile;
            task.XPath = "/n:Project/n:PropertyGroup/n:LastUpdate";
            task.Value = DateTime.Now.ToLongDateString();
            
            Assert.IsTrue(task.Execute(), "Execute Failed");

        }
    }
}
