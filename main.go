package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	rts "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

var ketoClient *Keto

// Initialize a Keto client with values from environment variables. If environment variables
// are not provided, will default the values for read/write endpoints.
func init() {
	ctx := context.Background()
	ketoReadAddr := os.Getenv("KETO_READ_ADDR")
	ketoWriteAddr := os.Getenv("KETO_WRITE_ADDR")

	if ketoReadAddr == "" {
		ketoReadAddr = "localhost:4466"
	}
	if ketoWriteAddr == "" {
		ketoWriteAddr = "localhost:4467"
	}

	cfg := KetoConfig{
		ReadGPRCAddress:  ketoReadAddr,
		WriteGRPCAddress: ketoWriteAddr,
	}

	keto, err := NewKetoClient(ctx, &cfg)
	if err != nil {
		panic(fmt.Sprintf("Unable to create a Keto client: %v", err))
	}

	ketoClient = keto
}

func main() {
	ctx := context.Background()

	// Our actors who will be accessing our document
	alice := NewSubject("user", "alice", "member")
	bob := NewSubject("user", "bob", "member")
	charlie := NewSubject("user", "charlie", "member")
	eve := NewSubject("user", "eve", "member")

	// An admin role subject that we will assign access to the document object
	admin := NewSubject("role", "admin", "member")

	// The document that the actors will try to access
	documentId := "480158d4-0031-4412-9453-1bb0cdf76104"

	// Run cleanup step before exiting the application at any point
	defer cleanup(ctx, documentId, admin)

	// Allow enough time for the Keto server to be seeded before execution
	time.Sleep(2 * time.Second)

	// Add Alice as an admin
	ok, err := ketoClient.CreateRelation(ctx, "role", "admin", "member", alice)
	if !assertResult(ok, true, err, "Adding Alice as an admin") {
		return
	}

	// Add Bob as an editor
	ok, err = ketoClient.CreateRelation(ctx, "role", "editor", "member", bob)
	if !assertResult(ok, true, err, "Adding Bob as an editor") {
		return
	}

	// Add Charlie as a viewer
	ok, err = ketoClient.CreateRelation(ctx, "role", "viewer", "member", charlie)
	if !assertResult(ok, true, err, "Adding Charlie as a viewer") {
		return
	}

	// =========================	Admin access to document 	=============================================

	// Check whether Alice has admin access (should be false at first)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "admin", alice)
	if !assertResult(ok, false, err, "Checking that Alice does not have admin access to document (before adding admin privileges)") {
		return
	}

	ok, err = ketoClient.CreateRelation(ctx, "document", documentId, "admin", admin)
	if !assertResult(ok, true, err, "Granting admin access to document") {
		return
	}

	// =========================	Checking Alice's access to document 	=============================================

	// Check whether Alice has admin access (should be true)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "admin", alice)
	if !assertResult(ok, true, err, "Checking that Alice has admin access to document (after adding admin privileges)") {
		return
	}

	// Check whether Alice has edit access (should be true)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "edit", alice)
	if !assertResult(ok, true, err, "Checking that Alice has edit access to document") {
		return
	}

	// Check whether Alice has view access (should be true)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "view", alice)
	if !assertResult(ok, true, err, "Checking that Alice has view access to document") {
		return
	}

	// =========================	Checking Bob's access to document 	=============================================

	// Check whether Bob has admin access (should be false)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "admin", bob)
	if !assertResult(ok, false, err, "Checking that Bob does not have admin access to document") {
		return
	}

	// Check whether Bob has edit access (should be true)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "edit", bob)
	if !assertResult(ok, true, err, "Checking that Bob has edit access to document") {
		return
	}

	// Check whether Bob has view access (should be true)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "view", bob)
	if !assertResult(ok, true, err, "Checking that Bob has view access to document") {
		return
	}

	// =========================	Checking Charlie's access to document 	=============================================

	// Check whether Charlie has admin access (should be false)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "admin", charlie)
	if !assertResult(ok, false, err, "Checking that Charlie does not have admin access to document") {
		return
	}

	// Check whether Charlie has edit access (should be false)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "edit", charlie)
	if !assertResult(ok, false, err, "Checking that Charlie does not have edit access to document") {
		return
	}

	// Check whether Charlie has view access (should be true)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "view", charlie)
	if !assertResult(ok, true, err, "Checking that Charlie has view access to document") {
		return
	}

	// =========================	Checking Eve's access to document 	=============================================

	// Check whether Eve has admin access (should be false)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "admin", eve)
	if !assertResult(ok, false, err, "Checking that Eve does not have admin access to document") {
		return
	}

	// Check whether Eve has edit access (should be false)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "edit", eve)
	if !assertResult(ok, false, err, "Checking that Eve does not have edit access to document") {
		return
	}

	// Check whether Eve has view access (should be false)
	ok, err = ketoClient.CheckObjectPermission(ctx, "document", documentId, "view", eve)
	if !assertResult(ok, false, err, "Checking that Eve does not have view access to document") {
		return
	}
}

func cleanup(ctx context.Context, documentId string, admin *rts.Subject) {
	// Cleanup step: Removing admin access to the document so we can repeat our test
	ok, err := ketoClient.DeleteRelation(ctx, "document", documentId, "admin", admin)
	assertResult(ok, true, err, "Removing admin access to document")
}

func assertResult(result bool, expected bool, err error, context string) bool {
	log.Printf("Evaluating test context: %s\n", context)
	if err != nil {
		log.Println(fmt.Sprintf("FAIL: An unexpected error occurred with the request. Error: %s", err.Error()))
		return false
	}
	if result != expected {
		log.Println(fmt.Sprintf("FAIL: Expected %t, Received: %t.", expected, result))
		return false
	}
	log.Println("PASS")
	return true
}
