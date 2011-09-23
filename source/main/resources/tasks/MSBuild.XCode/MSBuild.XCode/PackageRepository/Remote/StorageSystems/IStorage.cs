using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace xstorage_system
{
    public interface IStorage
    {
        /// <summary>
        /// This works with prefixes like:
        ///   - fs:\\cnshasap2\\Hg_Repo\PACKAGE_REPO
        ///   - http:\\fileserver.net
        ///   - p4:\\{Perforce connection string}
        /// </summary>
        /// <param name="connectionURL"></param>
        void connect(string connectionURL);

        bool holds(string storage_key);

        bool submit(string sourceURL, out string storage_key);
        bool retrieve(string storage_key, string destinationURL);
    }
}
