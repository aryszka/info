/*
Package KeyVal implements a human readable, text based data format.

trim line
read key until unescaped =
trim key
read value until unescaped \n
trim value
problem: escaped whitespace
solve: define states

keyLeadWhitespace
key
key or tailWhitespace
valueLeadWhitespace
value
value or tailWhitespace
escaped

define valid transitions
extend with sections
comments
[] discards a section
cmd
parse hierarchy or flat
dot to meaningful characters
no parsing of types: that's the major mistake of most common file formats
shall it be thread safe?

states:
whitespace
comment-whitespace
comment
comment-or-whitespace
section-whitespace
section
section-or-whitespace
key
key-or-whitespace
value-whitespace
value
value-or-whitespace

unicode, with restricted whitespace

return EOFRemainder if something left there

tcp is needed, because it is not self-correcting and always valid, and the order matters, too
*/
package keyval
