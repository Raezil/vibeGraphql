**vibeGraphql** is a minimalistic GraphQL library for Go that supports **queries**, **mutations**, and **subscriptions** with a clean and intuitive API. 
It was vibe coded using ChatGPT o3 model.

## ✨ Features

- 🔍 **Query resolvers** for fetching data  
- 🛠️ **Mutation resolvers** for updating data  
- 📡 **Subscription resolvers** for real-time updates  
- 🧵 Thread-safe in-memory data handling  
- 🔌 Simple HTTP handler integration (`/graphql` and `/subscriptions`)  

---

## 🚀 Getting Started

### 1. Install

```bash
go get github.com/Raezil/vibeGraphql
```

### 2. Define Your Schema and Resolvers

```go
graphql.RegisterQueryResolver("user", userResolver)
graphql.RegisterMutationResolver("updateUser", updateUserResolver)
graphql.RegisterSubscriptionResolver("userSubscription", userSubscriptionResolver)
```

### 3. Start HTTP Server

```go
http.HandleFunc("/graphql", graphql.GraphqlHandler)
http.HandleFunc("/subscriptions", graphql.SubscriptionHandler)

log.Fatal(http.ListenAndServe(":8080", nil))
```

---

## 🧪 Full Example

Here is a full example using `vibeGraphql`:

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	graphql "github.com/Raezil/vibeGraphql"
)

// User represents a sample user.
type User struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Friends []*User `json:"friends,omitempty"`
}

var (
	userStore = map[string]*User{
		"123": {ID: "123", Name: "John Doe", Age: 30},
		"456": {ID: "456", Name: "Jane Smith", Age: 25},
	}
	mu sync.Mutex
)

func userResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id argument missing or not a string")
	}
	mu.Lock()
	defer mu.Unlock()
	user, exists := userStore[id]
	if !exists {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	return user, nil
}

func usersResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	idsRaw, ok := args["ids"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("ids argument missing or not an array")
	}
	ids := make([]string, len(idsRaw))
	for i, v := range idsRaw {
		idStr, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		ids[i] = idStr
	}
	mu.Lock()
	defer mu.Unlock()
	var users []*User
	for _, id := range ids {
		if user, exists := userStore[id]; exists {
			users = append(users, user)
		}
	}
	return users, nil
}

func updateUserResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id argument missing or not a string")
	}
	newName, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name argument missing or not a string")
	}
	newAge, ok := args["age"].(int)
	if !ok {
		return nil, fmt.Errorf("age argument missing or not an int")
	}
	mu.Lock()
	defer mu.Unlock()
	user, exists := userStore[id]
	if !exists {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	user.Name = newName
	user.Age = newAge
	return user, nil
}

func userSubscriptionResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	ch := make(chan interface{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			user := userStore["123"]
			mu.Unlock()
			ch <- user
		}
	}()
	return ch, nil
}

// UploadFileResolver is the mutation resolver that accepts a file upload.
// uploadFilesResolver handles multiple file uploads.
// It expects the "files" argument to be an array of file objects,
// where each file object is a map with "filename" and "data" keys.
// uploadFilesResolver handles multiple file uploads.
// It expects the "files" argument to be an array of file objects.
func uploadFilesResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	rawFiles, ok := args["files"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("files argument not found or invalid")
	}
	targetDir := "./tmp"
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", targetDir, err)
	}
	var results []string
	for idx, raw := range rawFiles {
		fileData, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("file at index %d is invalid", idx)
		}
		filename, ok := fileData["filename"].(string)
		if !ok {
			return nil, fmt.Errorf("filename not provided for file at index %d", idx)
		}
		data, ok := fileData["data"].([]byte)
		if !ok {
			return nil, fmt.Errorf("file data not provided for file %q", filename)
		}
		filepath := fmt.Sprintf("%s/%s", targetDir, filename)
		if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to save file %q: %v", filename, err)
		}
		log.Printf("uploadFilesResolver: Received file %q with %d bytes", filename, len(data))
		results = append(results, fmt.Sprintf("Uploaded file %q (%d bytes)", filename, len(data)))
	}
	return results, nil
}

func main() {
	// Register resolvers.
	graphql.RegisterQueryResolver("user", userResolver)
	graphql.RegisterQueryResolver("users", usersResolver)
	graphql.RegisterMutationResolver("updateUser", updateUserResolver)
	// Register the file upload resolver.
	graphql.RegisterMutationResolver("uploadFiles", uploadFilesResolver)
	graphql.RegisterSubscriptionResolver("userSubscription", userSubscriptionResolver)

	// Use the GraphqlUploadHandler for /graphql to support file uploads.
	http.HandleFunc("/graphql", graphql.GraphqlUploadHandler)
	http.HandleFunc("/subscriptions", graphql.SubscriptionHandler)

	fmt.Println("GraphQL server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 💬 Contributing

We welcome contributions! Feel free to open issues, feature requests or submit PRs.


---
