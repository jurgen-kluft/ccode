set PACKAGE_REPO=J:\Dev.C++.Packages\PACKAGE_REPO
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\bin\Release\MSBuild.XCode.dll %PACKAGE_REPO%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\References\Ionic.Zip.dll %PACKAGE_REPO%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\*.* %PACKAGE_REPO%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\templates\*.* %PACKAGE_REPO%\com\virtuos\xcode\publish\templates /R /Y /Q

set PACKAGE_REPO=J:\Dev.C#.Packages\PACKAGE_REPO                                                    
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\bin\Release\MSBuild.XCode.dll %PACKAGE_REPO%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\tasks\MSBuild.XCode\MSBuild.XCode\References\Ionic.Zip.dll %PACKAGE_REPO%\com\virtuos\xcode\publish /R /Y /Q
xcopy source\main\resources\*.* %PACKAGE_REPO%\com\virtuos\xcode\publish /R /Y /Q                    
xcopy source\main\resources\templates\*.* %PACKAGE_REPO%\com\virtuos\xcode\publish\templates /R /Y /Q