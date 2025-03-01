package cmd

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"db_cli/internal/user"
)

func RunCLI(userService *user.UserService) {
	serverCmd := flag.Bool("server", false, "Run HTTP server (Gin) for CRUD operations")
	createCmd := flag.Bool("create", false, "Create a new user")
	readCmd := flag.Bool("read", false, "Read user by id")
	updateCmd := flag.Bool("update", false, "Update user by id")
	deleteCmd := flag.Bool("delete", false, "Delete user by id")
	listCmd := flag.Bool("list", false, "List all users")

	userID := flag.String("id", "", "User ID (uint)")
	userName := flag.String("name", "", "User name")
	userStatus := flag.String("status", "", "User status")

	flag.Parse()

	switch {
	case *serverCmd:
		return

	case *createCmd:
		if *userName == "" {
			fmt.Println("Usage: --create --name=<NAME> [--status=<STATUS>]")
			return
		}
		newUser, err := userService.CreateUser(*userName, *userStatus)
		if err != nil {
			log.Fatalf("Failed to create user: %v\n", err)
		}
		fmt.Printf("Created user: %+v\n", newUser)

	case *readCmd:
		if *userID == "" {
			fmt.Println("Usage: --read --id=<USER_ID>")
			return
		}
		id, err := strconv.Atoi(*userID)
		if err != nil {
			log.Fatalf("Invalid user ID: %v\n", err)
		}
		u, err := userService.GetUser(uint(id))
		if err != nil {
			log.Fatalf("Failed to read user: %v\n", err)
		}
		fmt.Printf("User: %+v\n", u)

	case *updateCmd:
		if *userID == "" || *userName == "" {
			fmt.Println("Usage: --update --id=<USER_ID> --name=<NAME> [--status=<STATUS>]")
			return
		}
		id, err := strconv.Atoi(*userID)
		if err != nil {
			log.Fatalf("Invalid user ID: %v\n", err)
		}
		u, err := userService.UpdateUser(uint(id), *userName, *userStatus)
		if err != nil {
			log.Fatalf("Failed to update user: %v\n", err)
		}
		fmt.Printf("Updated user: %+v\n", u)

	case *deleteCmd:
		if *userID == "" {
			fmt.Println("Usage: --delete --id=<USER_ID>")
			return
		}
		id, err := strconv.Atoi(*userID)
		if err != nil {
			log.Fatalf("Invalid user ID: %v\n", err)
		}
		if err := userService.DeleteUser(uint(id)); err != nil {
			log.Fatalf("Failed to delete user: %v\n", err)
		}
		fmt.Println("User deleted.")

	case *listCmd:
		users, err := userService.GetAllUsers()
		if err != nil {
			log.Fatalf("Failed to list users: %v\n", err)
		}
		for _, u := range users {
			fmt.Printf("%+v\n", u)
		}

	default:
		fmt.Println("Usage:")
		fmt.Println("  --server                           Run server mode")
		fmt.Println("  --create --name=<NAME> [--status=<STATUS>]")
		fmt.Println("  --read   --id=<USER_ID>")
		fmt.Println("  --update --id=<USER_ID> --name=<NAME> [--status=<STATUS>]")
		fmt.Println("  --delete --id=<USER_ID>")
		fmt.Println("  --list                             List all users")
	}
}
