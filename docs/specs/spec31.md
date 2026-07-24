Implement this:

- regarding what you just did with the files, you added these comments, will write through them:

- extract the files
- you can use the tempFolder defined in the files section in the config for the upload and actually reduce the needed memory
- the file location will eventually be in the regularFolder, and will have name <<guid>>.dat where guid is a v7()
- check the file is jpeg or png (separate method)
- if not, delete it
- if yes, leave it there, assemble an entry for attachment table and persist

- do not return NewSuccess but a list of the generated ids of the type IdsHolder
