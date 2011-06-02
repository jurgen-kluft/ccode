set PACKAGE_DRIVE=K:\

set PACKAGE_REPO1=%PACKAGE_DRIVE%Dev.C++.Packages\PACKAGE_REPO
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\bin\x86\Release\MSBuild.XCode.dll %PACKAGE_REPO1%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\References\Ionic.Zip.dll %PACKAGE_REPO1%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\*.* %PACKAGE_REPO1%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\templates\*.* %PACKAGE_REPO1%\com\virtuos\xcode\publish\templates /R /Y /Q

set PACKAGE_REPO3=%PACKAGE_DRIVE%Dev.C#.Packages\PACKAGE_REPO
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\bin\x86\Release\MSBuild.XCode.dll %PACKAGE_REPO3%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\References\Ionic.Zip.dll %PACKAGE_REPO3%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\*.* %PACKAGE_REPO3%\com\virtuos\xcode\publish /R /Y /Q                    
xcopy source\main\resources\templates\*.* %PACKAGE_REPO3%\com\virtuos\xcode\publish\templates /R /Y /Q

