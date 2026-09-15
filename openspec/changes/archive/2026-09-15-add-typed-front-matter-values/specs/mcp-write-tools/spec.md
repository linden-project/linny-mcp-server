## ADDED Requirements

### Requirement: Front-matter values preserve the caller's type

`set_front_matter` SHALL write a value as the YAML node type corresponding to the
type the caller supplied: string, boolean, integer, float, null, or a sequence of
scalars. No value SHALL be rendered through generic string formatting. A JSON number
with no fractional part and within int64 range SHALL be written as an integer,
otherwise as a float. Types SHALL be taken from the caller and SHALL NOT be inferred
from a string's contents.

#### Scenario: A number stays a number

- **WHEN** `set_front_matter` sets a key to the JSON number `3`
- **THEN** the front matter contains an unquoted integer `3`, not the string `'3'`

#### Scenario: A string is never re-typed

- **WHEN** `set_front_matter` sets a key to the JSON string `"2026-09-14"`
- **THEN** the front matter contains a string value, not a YAML timestamp

#### Scenario: Null is representable

- **WHEN** `set_front_matter` sets a key to JSON `null`
- **THEN** the key is present with a null value and is not removed

### Requirement: Sequences are first-class front-matter values

`set_front_matter` SHALL accept a list of scalars and write it as a YAML sequence.
The resulting term membership SHALL include every element of the sequence.

#### Scenario: Setting a taxonomy key to several terms

- **WHEN** `set_front_matter` sets a declared taxonomy key to `["acme", "globex"]`
- **THEN** the front matter contains a two-element sequence and the returned
  membership includes both terms

#### Scenario: No stringified list ever reaches the corpus

- **WHEN** any list value is written
- **THEN** the file contains no value of the form `'[…]'` produced by formatting a
  list as a string

### Requirement: Nested structures are refused

`set_front_matter` SHALL refuse a value that is a mapping, or a list containing
anything other than scalars, and SHALL leave the document unchanged. Linny front
matter is flat.

#### Scenario: Nested mapping refused

- **WHEN** `set_front_matter` sets a key to a JSON object
- **THEN** the write is refused and the document is unchanged

### Requirement: Writes that the indexer would ignore are refused

When the key is a taxonomy declared by the notebook, `set_front_matter` SHALL refuse
a value that is neither a string nor a sequence of strings, because such a value
produces no term membership. The refusal SHALL name the taxonomy. Keys that are not
declared taxonomies SHALL accept any supported scalar type.

#### Scenario: Number on a taxonomy key is refused

- **WHEN** `set_front_matter` sets a declared taxonomy key to the number `3`
- **THEN** the write is refused, the message names the taxonomy, and the document is
  unchanged

#### Scenario: Number on an ordinary property is accepted

- **WHEN** `set_front_matter` sets a key that is not a declared taxonomy to `3`
- **THEN** the write succeeds and the value is an integer

### Requirement: add_term and remove_term manage taxonomy membership

The server SHALL expose `add_term` and `remove_term`, each taking a slug, a taxonomy,
and a term. Both SHALL be idempotent. `add_term` SHALL create the key as a sequence
when absent, promote an existing scalar value to a sequence before appending, and
make no change when the term is already a member. `remove_term` SHALL drop the term,
and SHALL remove the key entirely when the last term is removed. Membership
comparison SHALL use the indexer's normalized term form while the caller's spelling
is what gets written.

#### Scenario: Adding to an absent key

- **WHEN** `add_term` targets a document with no such taxonomy key
- **THEN** the key is created as a sequence containing the term

#### Scenario: Adding to a scalar value promotes it

- **WHEN** `add_term` targets a document whose taxonomy key holds a single scalar
- **THEN** the value becomes a sequence containing the original value and the new term

#### Scenario: Adding an existing term changes nothing

- **WHEN** `add_term` adds a term already present, differing only in case or spacing
- **THEN** the document is unchanged and the result reports success

#### Scenario: Removing the last term removes the key

- **WHEN** `remove_term` removes the only term of a taxonomy key
- **THEN** the key is absent from the front matter and no empty sequence remains

#### Scenario: Removing an absent term changes nothing

- **WHEN** `remove_term` removes a term that is not present
- **THEN** the document is unchanged and the result reports success

### Requirement: Newly coined terms are reported

A term with no declared term config SHALL still be written, and the result SHALL list
it under `new_terms` so that coining a term is visible to the caller.

#### Scenario: Undeclared term is written and reported

- **WHEN** `add_term` adds a term that has no `L2-CONF-TAX-<tax>-TRM-<term>.yml`
- **THEN** the write succeeds and the result lists the term under `new_terms`

#### Scenario: Declared term is not reported as new

- **WHEN** `add_term` adds a term that has a declared term config
- **THEN** `new_terms` is empty

### Requirement: Typed edits remain surgical

Writing a typed or sequence value SHALL preserve the order of untouched front-matter
keys and any comments, and SHALL leave the document body unchanged.

#### Scenario: Key order and comments survive a sequence write

- **WHEN** a sequence value is written to a document whose front matter carries
  several keys and a comment
- **THEN** the untouched keys keep their original order and the comment is retained
