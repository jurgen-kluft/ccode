using System;
using System.IO;
using System.Text;
using System.Collections.Generic;
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

namespace MSBuild.XCode
{
    public class PackageConstruct : Task
    {
        [Required]
        public string Path { get; set; }
        [Required]
        public string Name { get; set; }
        [Required]
        public string Language { get; set; }
        [Required]
        public string TemplatePath { get; set; }

        public override bool Execute()
        {
            // Check for the availability of package.xml, pom.props and pom.targets
            // If they do not exist then
            //   Initialize the package dir using a name
            //   Copy files from the template directory
            //   Replace patterns (${Name}, ${GUID}, ${Language})
            // Else
            //   Load package.xml
            //   Verify that the directory structure exists
            //   If not then create them
            // Endif
            if (!Path.EndsWith("\\"))
                Path = Path + "\\";

            if (File.Exists(Path + "package.xml"))
            {
                // For C++
                XProject CppProjectTemplate = new XProject();
                CppProjectTemplate.Language = "cpp";
                CppProjectTemplate.Load(TemplatePath + "vcxproj.xml.template");

                // For C#
                //XProject CsProjectTemplate = new XProject();
                //CsProjectTemplate.Language = "cs";
                //CsProjectTemplate.Load(TemplatePath + "csproj.xml.template");

                XPackage package = new XPackage();
                package.Templates.Add(CppProjectTemplate);
                //package.Templates.Add(CsProjectTemplate);
                package.Load(Path + "package.xml");

                // Check directory structure
                foreach (XAttribute xa in package.DirectoryStructure)
                {
                    if (xa.Name == "Folder")
                    {
                        if (!Directory.Exists(Path + xa.Value))
                            Directory.CreateDirectory(Path + xa.Value);
                    }
                }

                // Sync dependencies

                // Generate the projects and solution
                package.GenerateProjects(Path);
                package.GenerateSolution(Path);
            }
            else 
            {
                if (!TemplatePath.EndsWith("\\"))
                    TemplatePath = TemplatePath + "\\";

                string DstPath = Path + Name + "\\";
                Directory.CreateDirectory(DstPath);

                // pom.targets.template ==> pom.targets
                // pom.props.template ==> pom.props
                // package.xml.template ==> package.xml
                FileCopy(TemplatePath + "pom.targets.template", DstPath + "pom.targets");
                FileCopy(TemplatePath + "pom.props.template", DstPath + "pom.props");
                FileCopy(TemplatePath + "package.xml.template", DstPath + "package.xml");
            }
            return true;
        }

        private bool FileCopy(string srcfile, string dstfile)
        {
            string[] lines = File.ReadAllLines(srcfile);

            using (FileStream wfs = new FileStream(dstfile, FileMode.Create, FileAccess.Write))
            {
                using (StreamWriter writer = new StreamWriter(wfs))
                {
                    foreach (string line in lines)
                    {
                        string l = line.Replace("${Name}", Name);
                        l = l.Replace("${Language}", Language);
                        if (l.Contains("${GUID}"))
                        {
                            string uuid = Guid.NewGuid().ToString();
                            l = l.Replace("${GUID}", uuid);
                        }
                        writer.WriteLine(l);
                    }
                    writer.Close();
                    wfs.Close();
                    return true;
                }
            }
        }
    }
}
