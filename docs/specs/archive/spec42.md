@internal/controller/put_rotate.go line 82

it crashes with "invalid image format"

What I believe
- surely when you save it, you need to save it with a correct extension according to the image type
- not sure if when you load it you also have to have right extension

bottom line, add the following steps wherever needed
- instead of loading the image "in place", copy it in temporary folder with the right extension (figure it out either probing the file or from the table attachment for that id, where image type is saved in conent type). I believe you have already functionality to look in the file, if so, please do and forget about database lookup
- load it 
- do the rotation to some file also with the right extension, also in temp folder, prefix the name with something to avoid collision
- copy file back in regular folder with .dat extension
- take over from there with the process thumbnail or however that is called
