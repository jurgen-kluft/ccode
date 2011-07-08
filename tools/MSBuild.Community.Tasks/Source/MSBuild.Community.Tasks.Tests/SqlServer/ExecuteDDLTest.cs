// $Id: ExecuteDDLTest.cs 281 2006-12-12 05:24:25Z joshuaflanagan $
using System;
using System.IO;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;
using NUnit.Framework;
using MSBuild.Community.Tasks.SqlServer;

namespace MSBuild.Community.Tasks.Tests.SqlServer
{
	[TestFixture]
	public class ExecuteDDLTest
	{
		private string _dbFilename;
		private string _logFilename;
		private ExecuteDDL _ddl;
		private MockBuild _engine;
		
		private TaskItem WriteSqlFile(string sql)
		{
			string name = Path.GetTempFileName();
			File.WriteAllText(name, sql);
			return new TaskItem(name);
		}
		
		[TestFixtureSetUp]
		public void FixtureSetup()
		{
			MockBuild buildEngine = new MockBuild();
			TaskUtility.makeTestDirectory(buildEngine);
			ServiceQuery squery = new ServiceQuery();
			squery.BuildEngine = buildEngine;
			squery.ServiceName = "MSSQLSERVER";
            if (!squery.Execute())
            {
                Assert.Ignore("ServiceQuery for SqlServer failed.");
            }
            if (!squery.Exists){
                Assert.Ignore("MS SqlServer is not installed");
            }
            if (!squery.Status.Equals("Running", StringComparison.InvariantCultureIgnoreCase))
            {
                Assert.Ignore("MS SqlServer is not running");
            }
		}
		
		[TestFixtureTearDown]
		public void FixtureTeardown()
		{
			if (!String.IsNullOrEmpty(_dbFilename) && File.Exists(_dbFilename))
			{
				Setup();
				_ddl.Files = new ITaskItem[] { WriteSqlFile("DROP DATABASE ExecuteDDLTest; ") };
				_ddl.Execute();
			}
		}
		
		[SetUp]
		public void Setup()
		{
			_engine = new MockBuild();
			_ddl = new ExecuteDDL();
			_ddl.BuildEngine = _engine;
			_ddl.ConnectionString = "Server=localhost;Integrated Security=true";
		}
		
		[Test]
		public void ExecuteDDL()
		{
			_dbFilename = Path.Combine(TaskUtility.TestDirectory, "ExecuteDDLTest.mdf");
			_logFilename = Path.Combine(TaskUtility.TestDirectory, "ExecuteDDLTest.ldf");
			TaskItem sqlFile = WriteSqlFile(String.Format(@"CREATE DATABASE ExecuteDDLTest ON ( NAME = 'ExecuteDDLTest', FILENAME = '{0}' )
				LOG ON ( NAME = 'ExecuteDDLTest_log', FILENAME = '{1}' );", _dbFilename, _logFilename));
			_ddl.Files = new ITaskItem[] { sqlFile };
			Assert.IsTrue(_ddl.Execute(), "ExecuteDDL Create database failed.");
			Assert.IsTrue(File.Exists(_dbFilename));
		}
	}
}
