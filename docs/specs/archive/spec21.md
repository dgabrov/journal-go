similar with postEditJournalItem

implement the following:

- create controller /postRemoveJournals
- please do all the tricks with getToken and get user by token to established the user is properly logged in
- it has payload data.IdsHolder
- returns Success if all good or the error in case of error

- The ids in the payload are ids associated with journals - journal_id
- Check first they all belong to the logged in user, if not, err
- Then delete them. I don't care if there are child records (journal_item) associated with the journals: in this case it will err, and I want to get the error back. This functionality is done usually when somebody added a journal by mistake. 

