similar with /addJournal

Implement as follows:

- create controller POST /updateJournal
- it has payload data.JournalUpdateData
- in case of success, returns Success

- check the user logged in etc. 
- the title in journal data should be filled out - otherwise error
- check journal with provided id is in the database and belongs to the logged in user, with relation cd 'owner'
- then update it
- don't touch the created_dt as this is for creation only

