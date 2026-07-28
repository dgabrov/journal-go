add endpoint POST /updateAttachmentTitles
it gets a payload that parses to an array of Attachment

in each item there is an attachment 
please check the user is logged in and is the owner (relation_cd 'owner') of all the attachments or return error
please update the title by for the attachment, by id for each of the provided attachments


the ownership - you already have functionality. 

Create of course separate file for this new endpoint and follow the structure from other endpoints, for example post_remove_reading.go etc

