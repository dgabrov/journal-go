@docs/db/job.sql

- look in the file @docs/db/job.sql

- create endpoint POST /job
- payload structure is JournalItemHolder from internal/data/app.go at line 99

Implement the endpoint as follows:
- check the journalItemID passed in the payload exists in journal_item table or else error
- check it is accessible by logged in user (relation_cd can be anything, read, owner, anything) or else return error
- check there is at least one attachment associated with it, if not, return error
- get the title associated with the passed journal item

- create an entry in the job table as follows
  - job_id -> v7()
  - name -> journal title + '_' + created_dt from journal_item table formatted '2025-12-12'
  - status 'pending'
  - journal_item_id of course the passed journal_item_id
  - create_dt will be autopopulated with current timestamp
