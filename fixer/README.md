
# Prepare a 3rdparty library

- Copy all header files to source/main/include/{name of library}
- Open terminal, go to the include/{name of library} directory and run: `find . -type f -name "*.c" -delete`
- Copy all cpp files to source/main/cpp/{name of library}
- Open terminal and go to the cpp/{name of library} directory and run: `find . -type f -name "*.h" -delete`

# Open the name-of-package.go in the root of the repository and change

- the import of ccode from `github.com/jurgen-kluft/ccode` to `github.com/jurgen-kluft/ccode/fixer`
  - modify `Generate(pkg)` into `Generate(pkg, true, true, false, true)` to do a dry run with verbose 
    output and only fix include directives
- Open terminal and run `go run name-of-package.go` to see the output of the fixer and check if the include directives are correct. If not, fix the include directives in the cpp files and run the fixer again until all include directives are correct. The output of the fixer will show you which files have been fixed and which files still have incorrect include directives.
