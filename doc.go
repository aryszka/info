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

Specification


UTF-8

- Keyval documents or streams are UTF-8 encoded.
- Any sequence of UTF-8 characters is valid. If no structuring characters in it, then it is a single key without
  a value.


Whitespace

Whitespaces in this description have a limited meaning: only ' ' or '\t' are called "whitespace". The '\n' is
handled differently in most cases than the two whitespace characters. The handling of '\r' is undefined. Other
common or uncommon whitespace characters, like the vertical tab, are handled as any other character.


Entry

An entry can have a key, a value and a comment. Keys, values and comments can be empty. Empty means zero number
of characters.


Section

- '.' by convention.


Comment

- Anything is a comment between a '#' or a ';' and a new line.
- A comment doesn't need to start in a new line.
- A comment is trimmed from leading and tail whitespace.
- One or more '#' or ';' are ignored at the beginning of a comment, even if separated by whitespace.
- A comment can span multiple lines if it is not broken by a section, a key or a value.
- In multiline comments, lines not starting with a '#' or a ';' are ignored.
- An empty comment line starting with '#' or a ';', whitespace not counted, is part of the comment, if it is
  between two non-empty comment lines.
- A comment closed by EOF gives an entry without a key and a value.
- A comment belongs to all following entries until the next comment.
- Only '\' and '\n' can be escaped in a comment.
*/
package keyval
