@internal/server/server.go

implement func(s Server) GetUserIdFromToken(ctx context.Context, token string) (string, error)

- database is defined in @docs/db/ddl.sql
- search for token in session table
- see expiry_ind is not 'Y' 
- see current date is not past expiry_dt

create two error types in server file

ErrTokenNotFound
ErrTokenExpired

if token not found return ErrTokenNotFound
if token expired return ErrTokenExpired

if all good
- move expiry_dt to now + config data tokenTimeToLive
- return user_id from the database, not the provided_id to be clear.

