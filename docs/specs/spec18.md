similar with postEditJournalItem

- add controller POST /postSearchUsers
- payload is SearchData
- do all the due dilligence with getToken and server.GetUserIdFromToken
- create server.Server method called SearchUsers that take the string parameter search from SearchData and searches for users
- users come ordered by login
- the currently logged in (the one that owns the token) user shall not be part of the results
- return User[] Server method returns User[] that will be passed to the controller and that will return it back to the calling party 