/*
Package keyval implements a human readable, streamable, hierarchical data format based on the .ini format.

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

return EOFIncomplete if something left there

tcp is needed, because it is not self-correcting and always valid, and the order matters, too

it is well defined

recommend/switch to use with bufio?

after error, all calls should return the same error


Specification


UTF-8

- Keyval documents or streams contain UTF-8 encoded text.
- Any sequence of UTF-8 characters is valid keyval data. If no structuring characters in it, then it is a single
  key without a value.
- EOFIncomplete when escaped before EOF.


Whitespace

Whitespace in this description has a limited meaning: only ' ' or '\t' are called "whitespace". The '\n' is
handled differently in most cases than the two whitespace characters. The handling of '\r' is undefined. Other
common or uncommon whitespace characters, like the vertical tab, are handled as any other character.


Entry

An entry can have a key, a value and a comment. Keys, values and comments can be empty. Empty means zero number
of characters. Keys can stand of multiple parts, separated by '.'. Key parts of an entry are prepended by the
parts of the section, that the entry belongs to. An entry has at least a non-empty key, value or comment.


Key

- Anything is a key, that is before a new line, and is not a value, a section declaration or a comment.
- A key doesn't need to start in a new line.
- A key is trimmed from leading and trailing whitespace.
- Keys spanning multiple lines are trimmed at only the start of the first line and the end of last line.
- Keys can be separated by '.', defining hierarchical structure of keys.
- Whitespace around '.' separators is trimmed.
- Only '\', '.', ':', '[', ';', '#', '\n', ' ' and '\t' can be escaped in a key.


Value

- Anything is a value, that is after a ':' and before a new line, EOF, another value, section declaration or a
  comment.
- A value doesn't need to start in a new line.
- A value is trimmed from leading and trailing whitespace.
- Values spanning multiple lines are trimmed only at the start of the first and the end of last line.
- Only '\', ':', '[', ';', '#', '\n', ' ' and '\t' can be escaped in a value.


Section

- Anything declares a section between '[' and a ']', when '[' is not inside a comment or a section declaration.
- A '[' inside a section declaration is part of the declaration.
- A section declaration doesn't need to start in a new line.
- A section declaration is trimmed from leading and trailing whitespace.
- A section declaration can span multiple lines.
- Sections spanning multiple lines are trimmed only at the start of the first line and the end of last line.
- Section declarations, just like keys, can be separated by '.', defining hierarchical structure of sections.
- Whitespace around '.' separators is trimmed.
- All keys and values following a section declaration belong to the declared section, until the next section
  declaration. The section is applied to the keys as a prefix.
- An empty section declaration, '[]', discards the current section.
- A section without keys and values, gives an entry with the section as the key and an empty value.
- An incomplete section declaration before EOF gives an error distinct from EOF.
- There are no comments inside a section declaration.
- Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.


Comment

- Anything is a comment between a '#' or a ';' and a new line, when '#' or ';' is not inside a section
  declaration or a comment.
- A comment doesn't need to start in a new line.
- A comment is trimmed from leading and trailing whitespace.
- One or more '#' or ';' are ignored at the beginning of a comment, even if separated by whitespace.
- A '#' or a ';' in a comment is part of the comment.
- A comment can span multiple lines if it is not broken by a section, a key or a value.
- In multiline comments, lines not starting with a '#' or a ';' are ignored.
- An empty comment line starting with '#' or a ';', whitespace not counted, is part of the comment, if it is
  between two non-empty comment lines.
- A standalone, empty comment discards the current comment for the following entries.
- A comment belongs to all following entries until the next comment.
- A comment closed by EOF gives an entry without a key and a value.
- Only '\', '\n', ' ' and '\t' can be escaped in a comment.
*/
package keyval
