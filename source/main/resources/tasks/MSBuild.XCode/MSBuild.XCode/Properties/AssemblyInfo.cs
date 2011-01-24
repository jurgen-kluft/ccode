using System.Reflection;
using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;

// General Information about an assembly is controlled through the following 
// set of attributes. Change these attribute values to modify the information
// associated with an assembly.
[assembly: AssemblyTitle("MSBuild.XCode")]
[assembly: AssemblyDescription("")]
[assembly: AssemblyConfiguration("")]
[assembly: AssemblyCompany("")]
[assembly: AssemblyProduct("MSBuild XCode")]
[assembly: AssemblyCopyright("Copyright © 2010 Jurgen Kluft")]
[assembly: AssemblyTrademark("")]
[assembly: AssemblyCulture("")]

// Setting ComVisible to false makes the types in this assembly not visible 
// to COM components.  If you need to access a type in this assembly from 
// COM, set the ComVisible attribute to true on that type.
[assembly: ComVisible(false)]

// The following GUID is for the ID of the typelib if this project is exposed to COM
[assembly: Guid("631fb1e8-bc0b-4b10-a1d0-39311a492445")]

// Version information for an assembly consists of the following four values:
//
//      Major Version
//      Minor Version 
//      Build Number
//      Revision
//
// You can specify all the values or you can default the Revision and Build Numbers 
// by using the '*' as shown below:
[assembly: AssemblyVersion("1.0.4.0")]
[assembly: AssemblyFileVersion("1.0.4.0")]

// -----------------------------------------------------------------------------------
// Feature List:
// - Pom.xml validation
// - Package Mode = Shared (default), Local
// - Package Type = Package, Source
// - Integration in Visual Studio (Compile, Install and Deploy)
// - Git support
// 
// -----------------------------------------------------------------------------------
// Coming --> 1.0.4.0 (? January 2011)
// - MsDev 2010 project now saves .filters file
// - C#
// - Fixed bug in task logging
// 
// -----------------------------------------------------------------------------------
// 1.0.3.0 (15th January 2011)
// - Share Repository
// - Solution (.sln) now is correctly handling project configurations, projects can have 
//   different configurations and these are now listed as 'Build=False'
// - Fixed a bug in Local Repository and other Repository logic, should filter files by
//   platform!
// - 
// -----------------------------------------------------------------------------------
// 1.0.2.0 (7th January 2011)
// - PackageRepository: Local, Target, Cache, Remote
// - Instance/Resource for some objects
// - Package dependency tree is now a seperate object
// - Dependency package update is faster, since we now know the version of the target
// - 
// -----------------------------------------------------------------------------------
// 1.0.1.0 (4th January 2010)
// - First release, improved logic of dependency package handling and updating
// - Clean up of the code
// - 
// 
// -----------------------------------------------------------------------------------
// 1.0.0.0 (29th December 2010)
// - First prototype version, features Compile, Install, Deploy, Sync
