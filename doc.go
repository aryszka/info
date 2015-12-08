/*
Package keyval implements a human readable, streamable, hierarchical data format based on the .ini format.


TODO:
- clarify the case of multi-part, empty keys, in the reader/writer/buffer
- actually, clarify the whole handling of the empty keys and values
- read and write of the empty keys and values should be the same
- it is not really a buffer, because reading from it doesn't empty it
- use the value of the last entry


Based on the INI file format, but (tries to be) well defined and support hierarchical data structures.


Streaming only possible where the underlying network protocol makes sure that the packets all arrive, are intact
and the order of receiving is the same of sending.


The functions in the package are not synchronized.


The possible evaluation methods of the same keys.


Specification


UTF-8

- Keyval documents or streams contain UTF-8 encoded text.
- Any sequence of UTF-8 characters is valid keyval data. If no structuring characters in it, then it is a single
  key without a value.
- EOFIncomplete when escaped before EOF.


Whitespace

Whitespace in this description has a limited meaning: only ' ' or '\t' are called "whitespace". The '\n' is
handled differently in most cases than the two whitespace characters. The handling of '\r' is undefined. Other
common or uncommon whitespace characters, like the vertical tab, are handled as any other character. When '\n'
is escaped, it means that there is '\' in front of the actual newline character, and this document may call it
'\n' by accident only out of being accustomed to it.


Entry

An entry can have a key, a value and a comment. Keys, values and comments can be empty. Empty means zero number
of characters. Keys can stand of multiple parts, separated by '.'. Key parts of an entry are prepended by the
parts of the section, that the entry belongs to. An entry has at least a non-empty key, value or comment. Values
don't have types. They are just a bunch of characters, and it's up to the utilizing application to decide what
to do with it.


Key

- Anything is a key, that is before a new line, and is not a value, a section declaration or a comment.
- A key doesn't need to start in a new line.
- A key is trimmed from leading and trailing whitespace.
- Keys spanning multiple lines are trimmed at only the start of the first line and the end of last line.
- Keys can be separated by '.', defining hierarchical structure of keys.
- Whitespace around '.' separators is trimmed.
- Inside a key, '\', '.', '=', '[', '#', '\n' can be escaped. At the key boundaries, ' ' and '\t' can be
  escaped, causing the escaped character be part of the key.


Value

- Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
  comment.
- A value doesn't need to start in a new line.
- A value is trimmed from leading and trailing whitespace.
- Values spanning multiple lines are trimmed only at the start of the first and the end of last line.
- Inside a value, '\', '=', '[', '#', '\n' can be escaped. At the value boundaries, ' ' and '\t' can be escaped,
  causing the escaped character be part of the value.


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
- Inside a section declaration, '\', '.', ']' can be escaped. At the section declaration boundaries, '\n', ' '
  and '\t' can be escaped, causing the escaped character be part of the section declaration.


Comment

- Anything is a comment between a '#' and a new line, when '#' is not inside a section
  declaration or a comment.
- A comment doesn't need to start in a new line.
- A comment is trimmed from leading and trailing whitespace.
- One or more '#' are ignored at the beginning of a comment, even if separated by whitespace.
- A '#' in a comment is part of the comment.
- A comment can span multiple lines if it is not broken by a section, a key or a value.
- In multiline comments, lines not starting with a '#' are ignored.
- An empty comment line starting with '#', whitespace not counted, is part of the comment, if it is
  between two non-empty comment lines.
- A standalone, empty comment discards the current comment for the following entries.
- A comment belongs to all following entries until the next comment.
- A comment closed by EOF gives an entry without a key and a value.
- Inside a comment, '\', '\n' can be escaped. At the comment boundaries, ' ' and '\t' can be escaped, causing
  the escaped character be part of the section declaration.
*/
package keyval
