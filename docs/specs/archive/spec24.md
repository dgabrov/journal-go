similar with postEditJournalItem

- create controller POST /readingUsers
- it has payload data.StringHolder
- Val in payload is JournalID 

- check the journal belongs to the logged in user (relation_cd 'owner'). I believe you have already a function for that in the dao, if not, my bad
- retrieve the list of users that are in user_journal related to this journal_id where the relation_cd is NOT owner - so the people who can only read
- return the list of users ordered by login

