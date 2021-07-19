# `go-fragment`

This module is intended for working with unstructured and structured fragments. This can for instance be useful when marshaling JSON, or passing a fragment of fields that are queried to a function in order to only resolve those fields. A fragment can be created from a fragment expression, e.g. `{ fieldA, fieldB { fieldX } }`, or by manually adding fields by creating a new structured or unstructured fragment.

> :warning: **This module is a work in progress, and you may encounter serious bugs.**

## Developers

- Ludvig Aldén [@ludvigalden](https://github.com/ludvigalden)
