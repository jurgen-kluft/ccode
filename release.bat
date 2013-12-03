rem set PACKAGE_DRIVE=K:\Dev.C++.Packages\\REMOTE_PACKAGE_REPO\
set PACKAGE_DRIVE=\\cnshasap2\Hg_Repo\PACKAGE_REPO\

set PACKAGE_REPO2=%PACKAGE_DRIVE%
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\bin\x86\Release\MSBuild.XCode.dll %PACKAGE_REPO2%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\References\Ionic.Zip.dll %PACKAGE_REPO2%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\*.* %PACKAGE_REPO2%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\templates\*.* %PACKAGE_REPO2%\com\virtuos\xcode\publish\templates /R /Y /Q

set PACKAGE_REPO4=%PACKAGE_DRIVE%
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\bin\x86\Release\MSBuild.XCode.dll %PACKAGE_REPO4%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\References\Ionic.Zip.dll %PACKAGE_REPO4%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\*.* %PACKAGE_REPO4%\com\virtuos\xcode\publish /R /Y /Q                    
xcopy source\main\resources\templates\*.* %PACKAGE_REPO4%\com\virtuos\xcode\publish\templates /R /Y /Q

del %PACKAGE_DRIVE%\com\virtuos\xcode\publish\sync.rev