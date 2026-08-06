Server type

- add a public function called ProcessJob
- takes parameter jobID which is string

Proceeds as follows:
- searches for jobID in the job table, if not present, exits with error
- if status value in the job record table is not 'pending', logs the status, an error message and exits, but exits with success
- get journal_item_id from record
- load attachment records from attachment tables for this journal_item_id
- if no attachment records, change job status field for the record to 'error' , log something and exits with success

otherwise:
- create guid with v7()
- get temporary folder for golang
- create folder called job_<<guid>> where guid is the value generated above. 
- copy all the files from regular folder for the above attachment_id values to this temp folder. But do not copy as is, keep the file name, but change the extension from ".dat" to whatever extension should have based on the file type. For this purpose there is a function getExtensionFromContentType in internal/controller/get_file.go, feel free to take it out from there to some util package and make it public, your choice.
- zip contents of the folders to a zip called <<jobID>>.zip and place it in the job folder that is defined in ConfigData / Files / JobFolder
- ultimately delete the whole directory job_<<guid>>

notes:
- the temp folder and the job folder might be on different mounts, keep this in mind, you can surely copy files in between, however might not be able to move them
- all is implemented in golang, do not invoke system tools 
- don't assume this will run on linux, might run on windows, although this will not be the case, but still. 
