/**
 * Sam here: used AI to make these tests I was way too lazy
 */

package main

import (
	"reflect"
	"testing"
	"time"

	"pantree/api/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Helper function to create a UUID for testing
func newUUID(t *testing.T, uuidStr string) pgtype.UUID {
	var uuid pgtype.UUID
	err := uuid.Scan(uuidStr)
	if err != nil {
		t.Fatalf("Failed to create UUID: %v", err)
	}
	return uuid
}

// Helper function to create a Timestamp for testing
func newTimestamp(t *testing.T, timeStr string) pgtype.Timestamp {
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t.Fatalf("Failed to create Timestamp: %v", err)
	}
	var timestamp pgtype.Timestamp
	err = timestamp.Scan(parsedTime)
	if err != nil {
		t.Fatalf("Failed to set Timestamp: %v", err)
	}
	return timestamp
}

func Test_mergeSyncStates(t *testing.T) {
	// Setup standard User (Context only, logic doesn't depend heavily on this yet)
	dummyUser := db.User{
		Email: "test@test.com",
		Name:  "Test User",
	}

	// Setup common IDs
	id1 := newUUID(t, "11111111-1111-1111-1111-111111111111")
	id2 := newUUID(t, "22222222-2222-2222-2222-222222222222")
	id3 := newUUID(t, "33333333-3333-3333-3333-333333333333")

	// Setup Timestamps
	timeOld := newTimestamp(t, "2023-01-01T10:00:00Z")
	timeNew := newTimestamp(t, "2023-01-02T10:00:00Z")

	tests := []struct {
		name     string
		remote   SyncState
		local    SyncState
		expected SyncOperations
	}{
		{
			name: "Item Added Locally (Exists in Local, Not Remote)",
			remote: SyncState{
				Items: []db.Useritem{},
				User:  dummyUser,
			},
			local: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeNew},
				},
				User: dummyUser,
			},
			expected: SyncOperations{
				ItemsToAdd:    []db.Useritem{{ID: id1, LastModified: timeNew}},
				ItemsToUpdate: []db.Useritem{},
				ItemsToDelete: []db.Useritem{},
				UserToUpdate:  &dummyUser,
			},
		},
		{
			name: "Item Deleted Locally (Exists in Remote, Not Local)",
			remote: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeNew},
				},
				User: dummyUser,
			},
			local: SyncState{
				Items: []db.Useritem{}, // Empty local
				User:  dummyUser,
			},
			expected: SyncOperations{
				ItemsToAdd:    []db.Useritem{},
				ItemsToUpdate: []db.Useritem{},
				ItemsToDelete: []db.Useritem{{ID: id1, LastModified: timeNew}},
				UserToUpdate:  &dummyUser,
			},
		},
		{
			name: "Item Updated Locally (Timestamp Local > Remote)",
			remote: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeOld, Quantity: decimal.NewFromInt(1)},
				},
				User: dummyUser,
			},
			local: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeNew, Quantity: decimal.NewFromInt(5)}, // Newer
				},
				User: dummyUser,
			},
			expected: SyncOperations{
				ItemsToAdd:    []db.Useritem{},
				ItemsToUpdate: []db.Useritem{{ID: id1, LastModified: timeNew, Quantity: decimal.NewFromInt(5)}},
				ItemsToDelete: []db.Useritem{},
				UserToUpdate:  &dummyUser,
			},
		},
		{
			name: "No Changes (Timestamps Equal)",
			remote: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeNew},
				},
				User: dummyUser,
			},
			local: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeNew},
				},
				User: dummyUser,
			},
			expected: SyncOperations{
				ItemsToAdd:    []db.Useritem{},
				ItemsToUpdate: []db.Useritem{},
				ItemsToDelete: []db.Useritem{},
				UserToUpdate:  &dummyUser,
			},
		},
		{
			name: "Complex Mix (1 Add, 1 Delete, 1 Update)",
			remote: SyncState{
				Items: []db.Useritem{
					{ID: id2, LastModified: timeOld}, // Will be updated
					{ID: id3, LastModified: timeOld}, // Will be deleted (missing in local)
				},
				User: dummyUser,
			},
			local: SyncState{
				Items: []db.Useritem{
					{ID: id1, LastModified: timeNew}, // New item
					{ID: id2, LastModified: timeNew}, // Update to existing
				},
				User: dummyUser,
			},
			expected: SyncOperations{
				ItemsToAdd:    []db.Useritem{{ID: id1, LastModified: timeNew}},
				ItemsToUpdate: []db.Useritem{{ID: id2, LastModified: timeNew}},
				ItemsToDelete: []db.Useritem{{ID: id3, LastModified: timeOld}},
				UserToUpdate:  &dummyUser,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := _mergeSyncStates(tt.remote, tt.local)

			// 1. Verify Additions
			if len(got.ItemsToAdd) != len(tt.expected.ItemsToAdd) {
				t.Errorf("ItemsToAdd length mismatch: got %v, want %v", len(got.ItemsToAdd), len(tt.expected.ItemsToAdd))
			} else if len(got.ItemsToAdd) > 0 {
				// Simple check on ID of first element for brevity
				if got.ItemsToAdd[0].ID != tt.expected.ItemsToAdd[0].ID {
					t.Errorf("ItemsToAdd mismatch ID: got %v, want %v", got.ItemsToAdd[0].ID, tt.expected.ItemsToAdd[0].ID)
				}
			}

			// 2. Verify Deletions
			if len(got.ItemsToDelete) != len(tt.expected.ItemsToDelete) {
				t.Errorf("ItemsToDelete length mismatch: got %v, want %v", len(got.ItemsToDelete), len(tt.expected.ItemsToDelete))
			} else if len(got.ItemsToDelete) > 0 {
				if got.ItemsToDelete[0].ID != tt.expected.ItemsToDelete[0].ID {
					t.Errorf("ItemsToDelete mismatch ID: got %v, want %v", got.ItemsToDelete[0].ID, tt.expected.ItemsToDelete[0].ID)
				}
			}

			// 3. Verify Updates
			if len(got.ItemsToUpdate) != len(tt.expected.ItemsToUpdate) {
				t.Errorf("ItemsToUpdate length mismatch: got %v, want %v", len(got.ItemsToUpdate), len(tt.expected.ItemsToUpdate))
			} else if len(got.ItemsToUpdate) > 0 {
				if got.ItemsToUpdate[0].ID != tt.expected.ItemsToUpdate[0].ID {
					t.Errorf("ItemsToUpdate mismatch ID: got %v, want %v", got.ItemsToUpdate[0].ID, tt.expected.ItemsToUpdate[0].ID)
				}
			}

			// 4. Verify User
			if !reflect.DeepEqual(got.UserToUpdate, tt.expected.UserToUpdate) {
				t.Errorf("UserToUpdate mismatch")
			}
		})
	}
}
