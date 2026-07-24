similar with postEditJournalItem

implement the following:

- create controller POST /removeReading
- it has payload data.RemoveReadingRequest
- JournalID is the journal id for whom the users will have reading privileges removed 
- UserIDs is the list of users

- check the journal belongs to the logged in user. I believe you have already a function for that in the dao
- then  ensure logged in UserID is not one of the UserIDs that is passed, meaning, if it is there, do not return error, just filter it out
- also filter the UserIDs list for duplicates. Do not error if they are present, just filter out the duplicates so that when you trigger the database operations you have a clean list
- then you remove the user_journal table entries for this journal_id that correspond to the UserIDs; filter for relation_cd != 'owner' to ensure ownership is not removed by any chance

If you have multiple steps to do in relation to the database - and you do - please prefer creating multiple functions in dao rather than bundling a bunch of steps in one function
