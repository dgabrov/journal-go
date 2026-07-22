@internal/server/server.go

the implementation in the service you do as follows:
- on the top level of server each entry point will open tx
- create private function associated with any operation and implement there the database access

func (s Server) GetUserByProvidedId(id string) (string, error) {
return "", nil
}

the above: return user_id from user table if provided_id is found; if not, return empty string but do not return error for that but for database errors

