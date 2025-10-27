# Interpolation

Basic interpolation is written `$(FOO)` and just fetches the value associated with `FOO` from the environment structure. If `FOO` is bound to multiple values, they are joined together with spaces.

## Interpolation Options

It includes a number of interpolation shortcuts to build strings from the environment. For example, to construct a list of include paths from a environment variable `CPPPATH`, you can say `$(CPPPATH:p-I)`.

Interpolation Syntax

|  Syntax               |  Effect
| --------------------- | ------------------------------------------------------------
|`$(VAR:f)`             |    Convert to forward slashes (`/`)
|`$(VAR:b)`             |    Convert to backward slashes (`\`)
|`$(VAR:n)`             |    Convert to native path slashes for host platform
|`$(VAR:u)`             |    Convert to upper case
|`$(VAR:l)`             |    Convert to lower case
|`$(VAR:B)`             |    *filenames*: Only keep the base part of a filename (w/o extension)
|`$(VAR:F)`             |    *filenames*: Only keep the filename (w/o dir)
|`$(VAR:D)`             |    *filenames*: Only keep the directory
|`$(VAR:p<prefix>)`     |    Prefix all values with the string `<prefix>`
|`$(VAR:s<suffix>)`     |    Suffix all values with the string `<suffix>`
|`$(VAR:d<delimiter>)`  |    Delimits all values with the string `<delimiter>`
|`$(VAR:t<delimiter>)`  |    Trims all values with the string `<delimiter>`
|`$(VAR:T<delimiters>)` |    Trims all values with any character from `<delimiters>`
|`$(VAR:P<prefix>)`     |    Prefix all values with `<prefix>` unless it is already there
|`$(VAR:S<suffix>)`     |    Suffix all values with `<suffix>` unless it is already there
|`$(VAR:j<sep>)`        |    Join all values with `<sep>` as a separator rather than space
|`$(VAR:i<index>)`      |    Select the item at the (one-based) `index`

These interpolation options can be combined arbitrarily by tacking on several options. If an option parameter contains a colon the colon must be escaped with a backslash or it will be taken as the start of the next interpolation option.

## Interpolation Examples

Assume there is an environment with the following bindings:

|     Key               |  Value
| --------------------- | ------------------------------------------------------------
|   `FOO`               |   `{ "String" }`
|   `BAR`               |   `{ "A", "B", "C" }`
|   `BOB`               |   `{ "trtHellotrt" }`

Then interpolating the following strings will give the associated result:

|   Expression          |   Resulting String
| --------------------- | ------------------------------------------------------------
|`$(FOO)`               |`String`
|`$(FOO:u)`             |`STRING`
|`$(FOO:l)`             |`string`
|`$(FOO:p__)`           |`__String`
|`$(FOO:p__:s__)`       |`__String__`
|`$(BAR)`               |`A B C`
|`$(BAR:u)`             |`A B C`
|`$(BAR:l)`             |`a b c`
|`$(BAR:p__)`           |`__A __B __C`
|`$(BAR:p__:s__:j!)`    |`__A__!__B__!__C__`
|`$(BAR:d__)`           |`__A__`, `__B__`, `__C__`
|`$(BAR:p\::s!)`        |`:A! :B! :C!`
|`$(BAR:SC)`            |`AC BC C`
|`$(BOB:ttrt)`          |`Hello`
|`$(BOB:Ttr)`           |`Hello`
|`$(BAR:i1)`            |`B`

## Nested Interpolation

Nested interpolation is possible, but should be used with care as it can be hard to debug and understand. Here's an example of how the generic C toolchain inserts compiler options dependening on what variant is currently active:

`$(CCOPTS_$(CURRENT_VARIANT:u))`

This works because the inner expansion will evalate `CURRENT_VARIANT` first (say, it has the value `debug`). That value is then converted to upper-case and spliced into the former which yields a new expression `$(CCOPTS_DEBUG)` which is then expanded in turn.

Used with care this is a powerful way of letting users customize variables per configuration and then glue everything together with a simple template.

One other feature is that each key is associated with an array of values, so you can use interpolation to combine values from multiple variables. For example, since `FOO` and `BAR` are  arrays, you will generate combinations with the following:

`$(FOO) $(BAR)` with `FOO = {"Hello", "Happy"}` and `BAR = {"Earth", "Mars", 'Jupiter"}` will yield `{"Hello Earth", "Hello Mars", "Hello Jupiter", "Happy Earth", "Happy Mars", "Happy Jupiter"}`.
