// $Id: WebDirectoryDeleteTest.cs 281 2006-12-12 05:24:25Z joshuaflanagan $

using System;
using System.Text;
using Microsoft.Build.Utilities;
using MSBuild.Community.Tasks.IIS;
using NUnit.Framework;

namespace MSBuild.Community.Tasks.Tests.IIS
{
	[TestFixture]
	public class WebDirectoryDeleteTest
	{
		private string mVirtualDirectoryName = "VirDirTest";

		[Test]
		public void WebDirectoryDeleteLocal()
		{
			// Local machine test
            if (!TaskUtility.IsMinimumIISVersionInstalled("localhost", 5, 0))
			{
				Assert.Ignore(@"IIS 5.0 was not found on the machine.  IIS 5.0 is required to run this test.");
			}
			
			WebDirectoryDelete task = new WebDirectoryDelete();
			task.BuildEngine = new MockBuild();
			task.VirtualDirectoryName = mVirtualDirectoryName;
			Assert.IsTrue(task.Execute(), "Execute Failed!");
		}
		
		[Test]
		public void WebDirectoryDeleteRemote()
		{
		    string mServer = "fenway";
            if (!TaskUtility.IsAdminOnRemoteMachine(mServer))
            {
                Assert.Ignore(String.Format("Unable to connect as administrator to {0}", mServer));
            }

			// Remote machine test
            if (!TaskUtility.IsMinimumIISVersionInstalled(mServer, 5, 0))
			{
				Assert.Ignore(@"IIS 5.0 was not found on the machine.  IIS 5.0 is required to run this test.");
			}
			
			WebDirectoryDelete task = new WebDirectoryDelete();
			task.BuildEngine = new MockBuild();
			task.ServerName = mServer;
			task.VirtualDirectoryName = mVirtualDirectoryName;
			// task.Username = mWAMUsername;
			// task.Password = mWAMPassword;
			Assert.IsTrue(task.Execute(), "Execute Failed!");
		}
	}
}
