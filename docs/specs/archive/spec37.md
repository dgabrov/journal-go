@internal/server/server.go the method at line 374 GetJournalItems please return []CompleteJournalItem instead of []JournalItem
- change please the retrieveJournalItems method to return as well []CompleteJournalItems and do this by implementing left join between the table journal_item and attachment with the link on journal_item_id
- At this time, as you can see, I only want attachment_id and the title, we will see later 
