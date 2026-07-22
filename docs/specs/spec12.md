- @internal/controller/proc.go
- implement getToken as follows:

- see first if there is authorization header and if so, return from there, either the value if it does not start with "bearer " (note the space) or whatever comes after "bearer " if it does
- if authorization not found, please search for cookie called jou12 (use existent constant)
- if not found at all, return "", error -> for error create new type error called ErrNoAuth in proc.go
