- create new controller POST,  update_titles.go with entry point /updateTitles
- it contains a JSON payload that can be parsed to a data.Titles


- check user is logged in
- those ids in the TitleHolder instances belong to the user. 
- Please check that they are owned by the logged in user (must be 'owner' in relation_cd inside user_journal database)
- For 'owner' see there is a constant in @internal/data/app.go, use that one

after that, you update the provided entries in the attachment table with the updated titles

