namespace xstorage_system
{
    public class StorageSystem
    {
        private IStorage mStorage;

        public StorageSystem()
        {
            mStorage = null;
        }

        public bool connect(string connectionURL)
        {
            if (connectionURL.StartsWith("storage::"))
            {
                mStorage = new StorageFs();
                mStorage.connect(connectionURL);
            }
            else if (connectionURL.StartsWith("ftp::"))
            {
                mStorage = new StorageFtp();
                mStorage.connect(connectionURL);
            }
            else
            {
                return false;
            }

            return true;
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
