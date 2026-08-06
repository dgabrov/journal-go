create endpoint PUT /rotate

the structure of the payload is this one

```json
{
  "id": "picture",
  "quotient": 1
}

```

- quotient can be -1, 1. The quotient multiplies by 90 to find the angle in degrees. So -1 means rotate to the left 90 degrees, +1 means rotate to the right with 90 degrees
- any other value for the quotient returns error
- id is the picture id
- picture must exist
- user must be logged in
- picture is the attachment and must be owned by the logged in user - relation_cd 'owner' in the user_journal table
- the above involve at least "ValidateAttachmentsOwnership" function in files.go at line 35

- find the main attachment in the regular folder and rotate it with the specified angle
- then use existent functionality to recreate the thumbnail, there is a func called createThumbnail on proc.go at line 205, please run it in background (go etc)
