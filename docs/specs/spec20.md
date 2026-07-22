similar with postEditJournalItem

implement the following:

- create controller /postRemoveItems
- please do all the tricks with getToken and get user by token to established the user is properly logged in
- it has payload data.IdsHolder
- returns Success if all good or the error in case of error

- The ids in the payload are ids associated with the journal_item
- Check first they all belong to the logged in user, if not, err
- Then delete them

