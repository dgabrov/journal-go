- create endpoint DELETE /job
- has a repeated query parameter called "id", so multiple ids for the jobs will be passed

Proceed as follows
- if I pass multiple times same id, please ensure you consolidate the slice in one that has only distinct values
- see all the jobs with the provided ids exist, otherwise error
- see all the jobs belong to the logged in user 

for each jobid:
- first delete the file in jobFolder which is called <<jobid>>.zip and then delete the record from database
- arrived at this level, do not err out: even if error occurs, continue with the loop and ultimately return success