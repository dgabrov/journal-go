- create endpoint GET /job/download
- has a query parameter called "id" which holds the job id

Proceed as follows
- see the job id exists, if not, error
- see the job id belongs to the logged in user (easy: user_id field in the job table must be logged in user)
- don't treat the two checks above together, I want different error messages depending on the error
- see the job id has status "completed", if not, error
- get the file from jobFolder and serve it as a download
