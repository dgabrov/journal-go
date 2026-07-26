- create new controller POST,  post_update_attachment.go with entry point /updateAttachment
- multipart
- it has two parts with a string, attachmentID and title (not sure if one part or two parts, please assess and decide)
- other part is called file and contains a file or not. 

- see if the user is logged in and owns the journal that has the attachment (you already have the functions for that)
- if file is present, then update the files on the server by using the already existent methods in @internal/controller/proc.go - isValidImageFile,detectImageType, getImageExtension, getImageContentType, createThumbnail and I believe you can externalize something more from @proc_upload_files.go for creating the regular image.
- if not present, do not update the file
- then update the title 
- return NewSuccess()