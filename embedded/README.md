# embedded

Note: Whenever you run `go run {package name}.go` in the root of your project, the embedded files will be regenerated.

In file `embedded.go`, the function `WriteEmbedded()` will scan the `embedded` directory and for each file found, it will generate a .cpp file containing the file's content as a byte array, this allows embedding static files directly into a C/C++ application.

Two variables are exposed for each embedded file:
- `unsigned char <filename>[]`: A pointer to the byte array containing the file's content.
- `unsigned int <filename>_len`: The length of the byte array in bytes.
