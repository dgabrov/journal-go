I added a bool parameter to serveRegularFile at line 109 in @internal/controller/get_file.go

this is triggered whether a query parameter called download exist in the query parameter list

if small parameter exists, this has precedence over everything. 

Now I want as follows:
- if download is present, rather than serving the regular image as is, you serve it to be downloaded
- the file name how it is stored now in the regular folder is <<id>>.dat. Please look to see the content type, and the disposition fileName shall have the extension corresponding to the content type (jpeg, png)
- if download is not present, no change compared with what is now. 

