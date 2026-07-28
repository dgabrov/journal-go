similar with postEditJournalItem

- create controller POST /addReadingUsers
- it has payload data.JournalUsersRequest

- it adds user_journal entries with 'read' for each of the users passed in the request in UserIDs

- check the journal belongs to the logged in user (relation_cd 'owner'). I believe you have already a function for that in the dao, if not, my bad
- filter out the logged in user from UserIds if added by mistake
- eliminate duplicates from UserIDs before operating in the database 
- I believe some functionality already exists
- first read the entries in user_journal for this journal_id and enter only the userIDs that are not already there
