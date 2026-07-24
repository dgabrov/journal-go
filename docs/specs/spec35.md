- create new controller POST,  remove_files.go with entry point /removeFiles
- it contains a JSON payload that can be parsed to a data.IdsHolder


- check user is logged in
- those ids in payload are attachment_id items. 
- Please check that they are owned by the logged in user (must be 'owner' in relation_cd inside user_journal database)
- For 'owner' see there is a constant in @internal/data/app.go, use that one

after that, you
- delete file with that ID from the "small" folder
- then delete image file with that id from the "regular" folder - both small and regular are in ConfigData
- then delete the entry in attachment table, for that attachment_id

to remind you, the files in both small (thumbnail) and regular folder have structure <<id>>.dat



