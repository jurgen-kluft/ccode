# Internal Predefines of a Compiler 

For Gcc, you can do this:

```terminal
touch test.c
gcc -E -dM test.c > predefines.txt
```