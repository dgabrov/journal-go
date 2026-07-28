@internal/controller/post_upload_files.go

- after successfully adding the files
- before exiting with that list of ids
- dispatch a go function that will iterate through the ids that you are about to return and do the following for each.

- open file in "regular" folder for reading
- load picture and get picture dimensions
- scale the image down so that the biggest dimension will be the one in the FilesConfig Dimension attribute
- keep the file type (jpeg, png no issue)
- create new file in "small" folder with converted file, same <<id>>.dat as in regular, but in a different folder
- in case of error just log the error but do not err out
- do not scale up picture, only scale down!!!

One note: I would create go function that iterates through ids and then call for each id another function, so that in case of error in second function will return to the first where you log it, so that you do not provide the same logging functionality for errors like you would do if you used only one function. Tell me if I am not clear. Tell me as well if I am clear. 

