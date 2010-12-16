using System;
using System.IO;
using System.Xml;
using System.Text;
using System.Collections.Generic;
using Ionic.Zip;

namespace MSBuild.XCode
{
    public class XPackage
    {
        public XGroup Group { get; set; }
        public string Name { get; set; }
        public string Branch { get; set; }
        public XVersion Version { get; set; }
        public string Platform { get; set; }
        public string Path { get; set; }

        public enum ELocation { Remote, Local, Target }

        public ELocation Location { get; set; }
        public bool Local { get { return Location == ELocation.Local; } set { if (value) Location = ELocation.Local; } }
        public bool Remote { get { return Location == ELocation.Remote; } set { if (value) Location = ELocation.Remote; } }
        public bool Target { get { return Location == ELocation.Target; } set { if (value) Location = ELocation.Target; } }

        public XPom Pom { get; set; }

        public void Extract(string destinationDir)
        {

        }

        public bool LoadPom()
        {
            Pom = null;

            if (File.Exists(Path))
            {
                ZipFile zip = new ZipFile(Path);
                if (zip.Entries.Count > 0)
                {
                    ZipEntry entry = zip[Name + "\\pom.xml"];
                    if (entry != null)
                    {
                        using (MemoryStream stream = new MemoryStream())
                        {
                            entry.Extract(stream);
                            stream.Position = 0;
                            using (StreamReader reader = new StreamReader(stream))
                            {
                                string xml = reader.ReadToEnd();
                                reader.Close();
                                stream.Close();
                                Pom = new XPom();
                                Pom.LoadXml(xml);
                                Pom.PostLoad();
                                return true;
                            }
                        }
                    }
                }
            }
            return false;
        }


    }
}
