@internal/data/app.go

**step 1**
Fill out the Job struct at line 103 with all the fields in the job table, nicely renamed to golang friendly

**step 2**
- create endpoint GET /job
- it returns the list of jobs that belong to the logged in user
- it returns an array of Job struct mentioned at **step 1**

**step 3**
- do not return nil, return empty array if no items
