@internal/controller/post_login.go processAuth at line around 68 please implement:

- post against the url the data.Login
- if not good, you get http error, 400+ in this case, get all the response payload and put it in an error object that you return
- if good, you get 200, parse to data.Authentication and return
- 