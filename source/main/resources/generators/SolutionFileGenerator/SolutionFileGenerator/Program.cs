using System;
using System.IO;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace SolutionFileGenerator
{
    class Program
    {
        static void Main(string[] args)
        {
            {
                List<string> projects = new List<string>();
                projects.Add(@"i:\HgDev.Modules\xmaven\source\main\resources\tasks\msbuild.xmaven\msbuild.xmaven\msbuild.xmaven.csproj");
                projects.Add(@"i:\HgDev.Modules\xmaven\source\main\resources\tasks\msbuild.xmaven\msbuild.xmaven.test\xmaven_test.csproj");

                SolutionGenerator gen = new SolutionGenerator(SolutionGenerator.EVersion.VS2010, SolutionGenerator.ELanguage.CS);

                int projectCount = gen.Execute("I:\\Test_CS.sln", projects);
            }

            {
                List<string> projects = new List<string>();
                projects.Add(@"i:\HgDev.Modules\xbase\source\main\cpp\xbase.vcxproj");
                projects.Add(@"i:\HgDev.Modules\xbase\source\test\cpp\xbase_test.vcxproj");

                SolutionGenerator gen = new SolutionGenerator(SolutionGenerator.EVersion.VS2010, SolutionGenerator.ELanguage.CPP);

                int projectCount = gen.Execute("I:\\Test_CPP.sln", projects);
            }

        }
    }
}
