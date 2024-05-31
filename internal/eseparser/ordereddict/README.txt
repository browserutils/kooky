# Ordered Dict implementation for go.

There are a number of OrderedDict implementations out there and this
is one is less capable than a more generic implementation:

- We do not support key deletion
- Replacing a key is O(n) with the number of keys

Therefore this implementation is only suitable for dicts with few keys
that are not generally replaced or deleted.

The main benefit of this implementation is that it maintains key order
when serializing/unserializing from JSON and it is very memory
efficient.
