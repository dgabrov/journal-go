similar with postEditJournalItem

implement the following:

- you create new function process but you can alter its signature no issue
- add controller POST /postLogout
- no payload
- get token, user associated with it, then look for session entry and mark expiry_ind to 'Y'
- for all the flows, the endpoint returns http 200 with Success as response payload (NewSuccess() in @proc.go)
- add needed server.Server and dao methods

- if token does not exist in the payload, log error
- if token already expired log error
- if token invalid (not in database), user invalid, any error you log it, do not return 400 http, only 200
- if all good, log that the operation was successful
