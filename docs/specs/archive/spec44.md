@internal/server/server.go

The following methods have following things that I want fixed
1. access to sql scripts should be done from private method that is called from this public method. The private methods should reside in dao.go same package. 
2. I do not want two different sql hits in a method: create two private methods and each one is servicing a query
3. For the following three methods, the first sql hit is very similar in each method - check ownership of some item, please consolidate and avoid code duplication

line  95 GetUserIdFromToken
line 141 UpdateJournalItem
line 177 AddJournalItem

The following methods have the issue with sql access from public method: create private method, in dao.go and call that method from public method which holds the transaction
line 429 CreateUserByProvidedID
line 447 ValidateAttachmentAccess
line 473 GetAttachmentContentType
line 497 ValidateJournalItemOwnership
line 525 CreateAttachment
line 545 ValidateAttachmentsOwnership
line 585 DeleteAttachments
line 627 UpdateAttachmentTitles
line 648 UpdateAttachmentTitle
line 663 UpdateAttachmentContentType
