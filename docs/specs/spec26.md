similar with postEditJournalItem

implement as follows:

- create controller POST /addJournal
- it has payload data.JournalUpdateData
- in case of success, returns Success

- check the user logged in etc. 
- the title in journal data should be filled out - otherwise error
- add journal in the database with the logged in user, and created_dt to time.Now()
