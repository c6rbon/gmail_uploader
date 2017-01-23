# Bugs / Known Issues

* 20170122 bhk For some messages, the API is overwriting the Date: header field regardless of the `internalDateSource` field. This is painful in both that it is the wrong date, and that it creates dups if you upload a second time.