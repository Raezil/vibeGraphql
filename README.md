# vibeGraphql

**vibeGraphql** is a minimalistic and powerful GraphQL library for Go that supports **queries**, **mutations**, and **subscriptions** with a clean and intuitive API.

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
	"log"
	"net/http"
	"sync"
	"time"

	"graphql" // import your GraphQL package that includes resolvers and handlers
)

// User represents a sample user.
type User struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Friends []*User `json:"friends,omitempty"`
}

// userStore holds our in-memory users.
var (
	userStore = map[string]*User{
		"123": {ID: "123", Name: "John Doe", Age: 30},
	}
	mu sync.Mutex
)

// userResolver fetches a user by ID.
// Expects an argument "id" of type string.
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

// updateUserResolver updates a user.
// Expects arguments "id" (string), "name" (string), and "age" (int).
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

// userSubscriptionResolver returns a channel that emits the state of a user every 2 seconds.
// In a real application, you would have more robust subscription handling and cancellation.
func userSubscriptionResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	ch := make(chan interface{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				user := userStore["123"]
				mu.Unlock()
				ch <- user
			}
		}
	}()
	return ch, nil
}

func main() {
	// Register resolvers.
	graphql.RegisterQueryResolver("user", userResolver)
	graphql.RegisterMutationResolver("updateUser", updateUserResolver)
	graphql.RegisterSubscriptionResolver("userSubscription", userSubscriptionResolver)

	// Register GraphQL HTTP endpoints.
	http.HandleFunc("/graphql", graphql.GraphqlHandler)
	http.HandleFunc("/subscriptions", graphql.SubscriptionHandler)

	fmt.Println("GraphQL server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 💬 Contributing

We welcome contributions! Feel free to open issues, feature requests or submit PRs.


---
