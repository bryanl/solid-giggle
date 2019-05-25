package kep

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	source := filepath.Join("testdata", "20180415-crds-to-ga.md")
	f, err := os.Open(source)
	require.NoError(t, err)

	defer f.Close()

	kep, err := Read(f)
	require.NoError(t, err)

	lastUpdated := time.Date(2018, time.April, 24, 0, 0, 0, 0, time.UTC)

	expected := &KEP{
		Title:     "Graduate CustomResourceDefinitions to GA",
		Authors:   []string{"@author1", "@author2"},
		OwningSIG: "sig-kep",
		ParticipatingSIGs: []string{
			"sig-alpha", "sig-beta",
		},
		Reviewers:    []string{"@reviewer1", "@reviewer2"},
		Approvers:    []string{"@approver1", "@approver2"},
		Editor:       "TBD",
		CreationDate: time.Date(2018, time.April, 15, 0, 0, 0, 0, time.UTC),
		LastUpdated:  &lastUpdated,
		Status:       "provisional",
		SeeAlso: []Link{
			{Text: "link 1", URL: "http://example.com/1"},
			{Text: "link 2", URL: "http://example.com/2"},
		},
		Content: "# content\n\ncontent",
	}

	assert.Equal(t, expected, kep)
}

func Test_extractUsers(t *testing.T) {
	makeUserObject := func(name string) map[string]interface{} {
		return map[string]interface{}{
			"name": name,
		}
	}

	cases := []struct {
		name     string
		users    []interface{}
		isErr    bool
		expected []string
	}{
		{
			name:     "as a string",
			users:    []interface{}{"user1", "user2"},
			expected: []string{"user1", "user2"},
		},
		{
			name:     "as objects",
			users:    []interface{}{makeUserObject("user1"), makeUserObject("user2")},
			expected: []string{"user1", "user2"},
		},
		{
			name:     "mixed",
			users:    []interface{}{makeUserObject("user1"), "user2"},
			expected: []string{"user1", "user2"},
		},
		{
			name:  "unexpected input",
			users: []interface{}{1},
			isErr: true,
		},
		{
			name:  "invalid object",
			users: []interface{}{map[string]interface{}{}},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractUsers(tc.users)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
