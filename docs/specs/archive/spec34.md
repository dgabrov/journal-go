- create new controller get_file, with GET
- query parameters
  1. id - this is the attachment_id
  2. small - if this is merely present, you return the small file, the thumbnail; if it is not present, then you return the regular file


this will be retrieved from a web page, in an img src= tag. 

behavior:
- if small id is requested and thumbnail does not exist, then return a placeholder image 150x150 that should be preloaded in the golang executable with go:embed
- the placeholder image is @internal/controller/placeholder.png please go embed it, you can add teh code in proc.go 
- if regular id is placed and file is not there, then return error
