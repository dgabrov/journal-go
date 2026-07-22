@internal/server/server.go

- private access methods should be put in dao.go in same package
- please implement get login response. 

top level in LoginResponse:
User     - the logged in user
Journals slice of complete journal - those are all the journals the logged in user can either read or owns which means he can read or do whatever

how you assemble the values:

- left join user -> user_journal -> journal
- load the data
- Owner will be false if logged in user can only read, or true if owner
- User is the user who owns the journal, can be somebody else, I need this to show the name in the list
