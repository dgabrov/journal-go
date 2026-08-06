in the attachment table I already added two fields

width int not null default 0
height int not null default 0

Please, when uploading pictures, there is a method processValidFile line 100 in internal/controller/post_upload_files.go

please make that function calculate width and height and add them as well to the database alongside with the other attachment information. 

also when you serve the attachment data please include those two fields as well, I believe in CompleteJournalItem type

