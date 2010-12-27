// $Id: XPathAsserter.cs 241 2006-08-19 16:16:56Z joshuaflanagan $
using System;
using NUnit.Framework;
using System.Xml;


namespace MSBuild.Community.Tasks.Tests.Xml
{
    public class XPathAsserter : AbstractAsserter
    {
        string message;
        string actualValue;
        string expectedValue;

        public XPathAsserter(XmlDocument document, string xpath, string expectedValue, string message, params object[] args) : base(message, args)
        {
            XmlNode node = document.SelectSingleNode(xpath);
            if (node == null)
            {
                actualValue = null;
            }
            else
            {
                actualValue = node.Value;
            }
            this.expectedValue = expectedValue;
            this.message = message;
        }

        public override string Message
        {
            get
            {
                base.FailureMessage.AddExpectedLine(this.Expectation);
                base.FailureMessage.DisplayActualValue(this.actualValue);
                return base.FailureMessage.ToString();
            }
        }

        protected virtual string Expectation
        {
            get
            {
                return string.Format("<\"{0}\">", this.expectedValue);
            }
        }

        public override bool Test()
        {
            if (expectedValue != actualValue) return false;
            return base.Test();
        }

    }

}
