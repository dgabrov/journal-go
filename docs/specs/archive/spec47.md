@internal/job/job.go

Implement function StartJob as follows

- there is a loop with a select
- select has three channels
  - the parameter in the function of this type
  - a channel that waits 5 minutes
  - os channel to catch interrupts

- the function loops all the time waiting for one of the three
- in case of interrupt, exits
- if case an integer comes through the channel passed as parameter
  - assemble the list of job ids from the job table that have status 'pending'
  - then drain the channel 
  - then for each jobID, process it, using the Server function ProcessJob
  - if one call returns error, log it and proceed further; if not, log that it was done successfully and add the jobID in the log as well
- in case of timeout comes on the 5 minutes channel
  - assemble the list of job ids from the job table that have status 'pending'
  - then drain the channel pass as parameter, even if this flow was not triggered by something coming from the channel
  - then for each jobID, process it, using the Server function ProcessJob
  - if one call returns error, log it and proceed further; if not, log that it was done successfully and add the jobID in the log as well

then continue looping. 

