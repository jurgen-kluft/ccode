# TODO

- Specific input to Clay
  - build config, e.g. dev-debug, prod-release, dev-release-test, etc..
  - build target, e.g. linux(amd64) windows(amd64), darwin(arm64), arduino(esp32)
  - project name

Notes:
- denv.Project has a list of denv.Config, however I wonder if this is really needed at this level.
- clay.Project has 'Config *Config', but this just contains 2 variables:
	- Config denv.BuildConfig
	- Target denv.BuildTarget
  We should remove Config and just embed these variables into Project, furthermore we should perhaps
  rename Project to ProjectBuildConfig or something that makes it more clear what it is.

# Clay TODO

- Load the package.json from the 'target/' folder
- denv.Project -> generate multiple ProjectBuildConfig objects
- For every ProjectBuildConfig glob all the source files, filtered by the BuildTarget (OS, Arch filtering)
- 
