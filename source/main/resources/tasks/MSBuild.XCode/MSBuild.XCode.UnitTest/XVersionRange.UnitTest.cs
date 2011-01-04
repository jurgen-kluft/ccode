using System;
using System.Text;
using System.Collections.Generic;
using System.Linq;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using MSBuild.XCode;

namespace MSBuild.XCode.UnitTest
{
    [TestClass]
    public class XVersion_UnitTest
    {
        [TestMethod]
        public void Test_VersionUnbound()
        {
            VersionRange xrange1 = new VersionRange("[1.2,)");
            Assert.IsFalse(xrange1.IsInRange(new ComparableVersion("0.9")));
            Assert.IsFalse(xrange1.IsInRange(new ComparableVersion("1.0")));
            Assert.IsFalse(xrange1.IsInRange(new ComparableVersion("1.1.2")));
            Assert.IsTrue(xrange1.IsInRange(new ComparableVersion("1.2")));
            Assert.IsTrue(xrange1.IsInRange(new ComparableVersion("1.21")));
            Assert.IsTrue(xrange1.IsInRange(new ComparableVersion("2.1")));
        }

        [TestMethod]
        public void Test_UnboundVersionOrVersionUnbound()
        {
            VersionRange xrange1 = new VersionRange("(,1.0],[1.2,)");
            Assert.IsFalse(xrange1.IsInRange(new ComparableVersion("1.1.2")));
            Assert.IsTrue(xrange1.IsInRange(new ComparableVersion("0.9")));
        }

    }
}
