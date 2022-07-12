# Errors

a go package for "custom errors".

This is mainly used as a reference and could easily be copied to your project if you'd like to use it. 
But it also works as a package if you for some reason do not want to copy the file and create a dependency of it. 

## How to use

Copy the error.go file content and modify the struct to match your environment (recommended).

the errors can easily be used in your projects to simply standardize the error messages.

```go

type user struct {
	username string
}

// handler.go
func Login(args ...any) (*user, error) {
	// parse args and validate them (your way)...
if err != nil {
		return nil, errors.New(errors.Op("handler.login"), err)
	}
	return user, nil
}

// auth.go
func Login(username, password string) (*user, error) {
	if username == "admin" && password == "admin" {
		return &user{username: "admin"}, nil
	}
	return nil, errors.New(errors.Op("auth.login"), errors.NotFound, fmt.Errorf("invalid username or password"))
}

// main.go
func main() {
	user, err := handler.Login("admin", "password")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(user)
}
// Output:
// handler.login:
//     auth.login: invalid username or password
```