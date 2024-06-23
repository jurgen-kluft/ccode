# There are seven “core” automatic variables:

- $@: The filename representing the target.
- $%: The filename element of an archive member specification.
- $<: The filename of the first prerequisite.
- $?: The names of all prerequisites that are newer than the target, separated by spaces.
- $^: The filenames of all the prerequisites, separated by spaces. This list has duplicate filenames removed since for most uses, such as compiling, copying, etc., duplicates are not wanted.
- $+: Similar to $^, this is the names of all the prerequisites separated by spaces, except that $+ includes duplicates. This variable was created for specific situations such as arguments to linkers where duplicate values have meaning.
- $*: The stem of the target filename. A stem is typically a filename without its suffix. Its use outside of pattern rules is discouraged.


For example, consider the following declaration:

all: library.cpp main.cpp
In this case:

$@ evaluates to all
$< evaluates to library.cpp
$^ evaluates to library.cpp main.cpp
