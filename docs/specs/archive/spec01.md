- program expects - mandatory - environment variable called CONFIG_FILE it contains a full path of the configuration file to be loaded upon startup
- please populate ConfigData in @internal/data/data.go with all the values in the file at this moment
- for inner types, please create new struct type alongside ConfigData
- please write function that is called from start.go Start that gets the variable and loads the file and returns either pointer to ConfigData or error
- using slog, inside that function, please log the values in the config data one by one and add some comments, except password for the database where you say only whether it is filled out or not.

- then add dependency to mariadb drivers and load them (prompt for running go get)
- using the db provided in config data, write a function called in Start that returns pointer to db and error
- in this function check if you can establish connectivity (do some select 1 or anything)
1