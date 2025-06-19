package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// User struct matches JSONPlaceholder API structure
type User struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Address Address `json:"address"`
	Company Company `json:"company"`
}

type Address struct {
	Street  string `json:"street"`
	Suite   string `json:"suite"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
}

type Company struct {
	Name        string `json:"name"`
	CatchPhrase string `json:"catchPhrase"`
}

func main() {
	// Fetch users from JSONPlaceholder API
	users, err := fetchUsers("https://jsonplaceholder.typicode.com/users")
	if err != nil {
		fmt.Printf("Error fetching users: %v\n", err)
		return
	}

	// Channel for filtered users
	filteredUsers := make(chan User)
	var wg sync.WaitGroup

	// Process users in parallel
	for _, user := range users {
		wg.Add(1)
		go func(u User) {
			defer wg.Done()

			// Display user information
			displayUser(u)

			// Check if catch phrase contains "task-force" (case-insensitive)
			if strings.Contains(strings.ToLower(u.Company.CatchPhrase), "task-force") {
				filteredUsers <- u
			}
		}(user)
	}

	go func() {
		wg.Wait()
		close(filteredUsers)
	}()

	// Collect filtered users
	var usersToPersist []User
	for user := range filteredUsers {
		usersToPersist = append(usersToPersist, user)
	}

	// Persist filtered users to YAML file
	if len(usersToPersist) > 0 {
		err = persistUsersYaml(usersToPersist, "filtered_users.yaml")
		if err != nil {
			fmt.Printf("Error persisting users: %v\n", err)
		} else {
			fmt.Printf("\nSuccessfully persisted %d users to filtered_users.yaml\n", len(usersToPersist))
		}
	} else {
		fmt.Println("\nNo users matched the filter criteria")
	}
}

func fetchUsers(url string) ([]User, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func displayUser(user User) {
	fmt.Println("Name:", user.Name)
	fmt.Println("Email:", user.Email)
	fmt.Printf("Address: %s, %s, %s, %s\n",
		user.Address.Street,
		user.Address.Suite,
		user.Address.City,
		user.Address.Zipcode)
	fmt.Printf("Company: %s, %s\n\n",
		user.Company.Name,
		user.Company.CatchPhrase)
}

func persistUsersYaml(users []User, filename string) error {
	data, err := yaml.Marshal(users)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
