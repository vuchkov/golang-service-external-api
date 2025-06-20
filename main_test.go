package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

// Test data
var mockUsers = []User{
	{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Address: Address{
			Street:  "123 Main St",
			Suite:   "Apt 4",
			City:    "Testville",
			Zipcode: "12345",
		},
		Company: Company{
			Name:        "Test Corp",
			CatchPhrase: "Task-force oriented solutions",
		},
	},
	{
		ID:    2,
		Name:  "Another User",
		Email: "another@example.com",
		Address: Address{
			Street:  "456 Oak Ave",
			Suite:   "Suite 100",
			City:    "Sample City",
			Zipcode: "67890",
		},
		Company: Company{
			Name:        "Example Inc",
			CatchPhrase: "Just another company",
		},
	},
}

// TestFetchUsers tests the fetchUsers function
func TestFetchUsers(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockUsers)
	}))
	defer ts.Close()

	users, err := fetchUsers(ts.URL)
	if err != nil {
		t.Fatalf("fetchUsers() error = %v", err)
	}

	if len(users) != len(mockUsers) {
		t.Errorf("Expected %d users, got %d", len(mockUsers), len(users))
	}
}

// TestDisplayUser tests the displayUser function output
func TestDisplayUser(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displayUser(mockUsers[0])

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := `Name: Test User
Email: test@example.com
Address: 123 Main St, Apt 4, Testville, 12345
Company: Test Corp, Task-force oriented solutions

`
	if output != expected {
		t.Errorf("displayUser() output = %v, want %v", output, expected)
	}
}

// TestPersistUsersYAML tests the YAML persistence function
func TestPersistUsersYAML(t *testing.T) {
	testFile := "test_users.yaml"
	defer os.Remove(testFile)

	err := PersistUsersYAML(mockUsers, testFile)
	if err != nil {
		t.Fatalf("PersistUsersYAML() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("YAML file was not created")
	}

	// Verify file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	if !strings.Contains(string(content), "Test User") {
		t.Errorf("YAML file content is incorrect")
	}
}

// TestFilteringLogic tests the filtering condition
func TestFilteringLogic(t *testing.T) {
	tests := []struct {
		name        string
		catchPhrase string
		want        bool
	}{
		{"Match", "Our task-force is great", true},
		{"No Match", "We are different", false},
		{"Case Insensitive", "TASK-FORCE rocks", true},
		{"Partial Match", "This is a task-force project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Company: Company{
					CatchPhrase: tt.catchPhrase,
				},
			}

			got := strings.Contains(strings.ToLower(user.Company.CatchPhrase), "task-force")
			if got != tt.want {
				t.Errorf("Filter condition for %q = %v, want %v", tt.catchPhrase, got, tt.want)
			}
		})
	}
}

// TestMainWorkflow tests the complete workflow
func TestMainWorkflow(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockUsers)
	}))
	defer ts.Close()

	// Replace the API URL with our test server URL
	oldURL := apiURL
	apiURL = ts.URL
	defer func() { apiURL = oldURL }()

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main (we'll use a helper function to avoid os.Exit issues)
	testMain()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected user data
	if !strings.Contains(output, "Test User") || !strings.Contains(output, "Another User") {
		t.Errorf("Main output doesn't contain expected user data")
	}

	// Verify YAML file was created for filtered users
	testFile := "filtered_users.yaml"
	defer os.Remove(testFile)

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("Filtered users YAML file was not created")
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	// Should only contain the user with "task-force" in catch phrase
	if strings.Contains(string(content), "Another User") {
		t.Errorf("YAML file contains unfiltered user")
	}
}

// testMain is a helper function to test the main workflow without os.Exit
func testMain() {
	// This replicates the main() function but without os.Exit
	users, err := fetchUsers(apiURL)
	if err != nil {
		fmt.Printf("Error fetching users: %v\n", err)
		return
	}

	filteredUsers := make(chan User)
	var wg sync.WaitGroup

	for _, user := range users {
		wg.Add(1)
		go func(u User) {
			defer wg.Done()
			displayUser(u)
			if strings.Contains(strings.ToLower(u.Company.CatchPhrase), "task-force") {
				filteredUsers <- u
			}
		}(user)
	}

	go func() {
		wg.Wait()
		close(filteredUsers)
	}()

	var usersToPersist []User
	for user := range filteredUsers {
		usersToPersist = append(usersToPersist, user)
	}

	if len(usersToPersist) > 0 {
		err = PersistUsersYAML(usersToPersist, "filtered_users.yaml")
		if err != nil {
			fmt.Printf("Error persisting users: %v\n", err)
		}
	}
}

// We need to make apiURL a variable for testing
var apiURL = "https://jsonplaceholder.typicode.com/users"
