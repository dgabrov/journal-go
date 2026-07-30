There is database access as follows:

- @internal/controller/put_rotate.go:138 in getAttachmentContentType method (QueryRowContext call to fetch attachment content_type)

Create entry in Server and move this functionality there
