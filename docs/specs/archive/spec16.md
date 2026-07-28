@internal/server/server.go

implement UpdateJournalItem

- check journal item id, journal id are filled out
- journal id must belong to the provided userId meaning it is in a journal that is linked to the user with a user_journal that has relation_cd of value 'owner'
- update updated_dt to time.Now()
