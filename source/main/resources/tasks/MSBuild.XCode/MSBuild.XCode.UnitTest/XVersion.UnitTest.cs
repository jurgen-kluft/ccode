using System;
using System.Text;
using System.Collections.Generic;
using System.Linq;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using MSBuild.XCode;

namespace MSBuild.XCode.UnitTest
{
    [TestClass]
    public class XVersionRange_UnitTest
    {
        [TestMethod]
        public void TestParser()
        {
            ComparableVersion xversion1 = new ComparableVersion("1.2.23.0");
            string[] version_components1 = xversion1.ToStrings();
            Assert.AreEqual(3, version_components1.Length);
            Assert.AreEqual(version_components1[0], "1");
            Assert.AreEqual(version_components1[1], "2");
            Assert.AreEqual(version_components1[2], "23");

            ComparableVersion xversion2 = new ComparableVersion("1.0.0.0");
            string[] version_components2 = xversion2.ToStrings();
            Assert.AreEqual(1, version_components2.Length);
            Assert.AreEqual(version_components2[0], "1");
        }

    }
}
