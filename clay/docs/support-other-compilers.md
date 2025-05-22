# Clay

This could actually work on all OSes, with many different compilers, the only hurdle is the dependency tracking.
Once that piece is in place, we could rewrite Clay to become more modular and be able to support any compiler.

That would mean we could:

- Work on Windows, Linux, and MacOS and compile using
  - Windows OS:
    - MSVC
    - Clang
    - GCC
  - Linux OS:
    - Clang
    - GCC
  - MacOS:
    - Clang
    - GCC

This would make all the other generators obsolete, as they would be able to use the same codebase and just change the compiler.

