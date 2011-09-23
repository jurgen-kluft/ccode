using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace xstorage_system
{
    public class StorageSystem
    {
        private IStorage mStorage;

        public StorageSystem()
        {
            mStorage = null;
        }

        public void connect(string connectionURL)
        {
            mStorage = new StorageFs();
            mStorage.connect(connectionURL);
        }

        public bool holds(string key)
        {
            return mStorage.holds(key);
        }

        public bool submit(string sourceURL, out string key)
        {
            return mStorage.submit(sourceURL, out key);
        }

        public bool retrieve(string key, string destinationURL)
        {
            return mStorage.retrieve(key, destinationURL);
        }
    }
}
