# Fork-in-regex implementation

Authors: Alina Lozovskaia, Egor Molchan

The approach is to use a tokenizer to parse an MD file (CommonMark) into an RDX collection (an array of tokens).

To run the tokenizer:
```
 go build && ./fork-in-regex test-file.md
```

To run the tests:

```
go test
```
