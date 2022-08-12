package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/jmoiron/sqlx"
	"github.com/layerhub-io/api/layerhub"
	"github.com/layerhub-io/api/testhelpers/docker"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func TestMySQL_PutUser(t *testing.T) {
	now := layerhub.Now()
	testscases := []struct {
		name         string
		newUser      layerhub.User
		updateEmail  string
		expectedUser layerhub.User
	}{
		{
			name: "new user",
			newUser: layerhub.User{
				ID:        "user_1",
				FirstName: "Jhon",
				LastName:  "Doe",
				Email:     "jhon.doe@mail.com",
				ApiToken:  "user_1_token",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedUser: layerhub.User{
				ID:        "user_1",
				FirstName: "Jhon",
				LastName:  "Doe",
				Email:     "jhon.doe@mail.com",
				ApiToken:  "user_1_token",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "update email",
			newUser: layerhub.User{
				ID:        "user_1",
				FirstName: "Jhon",
				LastName:  "Doe",
				Email:     "jhon.doe@mail.com",
				ApiToken:  "user_1_token",
				CreatedAt: now,
				UpdatedAt: now,
			},
			updateEmail: "jhon.doe.123@mail.com",
			expectedUser: layerhub.User{
				ID:        "user_1",
				FirstName: "Jhon",
				LastName:  "Doe",
				Email:     "jhon.doe.123@mail.com",
				ApiToken:  "user_1_token",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM users")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutUser(context.TODO(), &tc.newUser)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateEmail != "" {
				tc.newUser.Email = tc.updateEmail
				err := db.PutUser(context.TODO(), &tc.newUser)
				if err != nil {
					t.Fatal(err)
				}
			}

			users, err := db.FindUsers(context.TODO(), &layerhub.Filter{ID: tc.newUser.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(users) == 0 {
				t.Fatal("user not found")
			}

			got := users[0]
			if tc.expectedUser != got {
				t.Errorf("mismatched users:\ngot: %v\n want: %v", got, tc.expectedUser)
			}
		})
	}
}

func TestMySQL_FindUsers(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name          string
		query         *layerhub.Filter
		currentUsers  []layerhub.User
		expectedUsers []layerhub.User
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "user_2"},
			currentUsers: []layerhub.User{
				{
					ID:        "user_1",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedUsers: []layerhub.User{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				Email: "jhon.doe@mail.com",
			},
			currentUsers: []layerhub.User{
				{
					ID:        "user_1",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "user_2",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe2@mail.com",
					ApiToken:  "user_2_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedUsers: []layerhub.User{
				{
					ID:        "user_1",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:  "multiple results",
			query: &layerhub.Filter{},
			currentUsers: []layerhub.User{
				{
					ID:        "user_1",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "user_2",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedUsers: []layerhub.User{
				{
					ID:        "user_1",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "user_2",
					FirstName: "Jhon",
					LastName:  "Doe",
					Email:     "jhon.doe@mail.com",
					ApiToken:  "user_1_token",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	//Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM users")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentUsers) > 0 {
				for _, u := range tc.currentUsers {
					err := db.PutUser(context.TODO(), &u)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			users, err := db.FindUsers(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(users, tc.expectedUsers) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", users, tc.expectedUsers)
			}
		})
	}
}

func TestMySQL_BatchCreateFonts(t *testing.T) {
	testcases := []struct {
		name          string
		newFonts      []layerhub.Font
		expectedFonts []layerhub.Font
	}{
		{
			name: "one font",
			newFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
			expectedFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
		},
		{
			name: "multiple fonts",
			newFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_3",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
			expectedFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_3",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM fonts")
			if err != nil {
				t.Fatal(err)
			}

			err = db.BatchCreateFonts(context.TODO(), tc.newFonts)
			if err != nil {
				t.Fatal(err)
			}

			fonts, err := db.FindFonts(context.TODO(), nil)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(fonts, tc.expectedFonts) {
				t.Fatalf("mismatched fonts:\ngot: %v\nwant: %v", fonts, tc.expectedFonts)
			}
		})
	}
}

func TestMySQL_PutFont(t *testing.T) {
	testscases := []struct {
		name         string
		newFont      layerhub.Font
		updateURL    string
		expectedFont layerhub.Font
	}{
		{
			name: "new font",
			newFont: layerhub.Font{
				ID:             "font_1",
				FullName:       "Fake Font Regular",
				Family:         "Fake Font",
				Style:          "Regular",
				PostscriptName: "FakeFontRegular",
			},
			expectedFont: layerhub.Font{
				ID:             "font_1",
				FullName:       "Fake Font Regular",
				Family:         "Fake Font",
				Style:          "Regular",
				PostscriptName: "FakeFontRegular",
			},
		},
		{
			name: "update URL",
			newFont: layerhub.Font{
				ID:             "font_1",
				FullName:       "Fake Font Regular",
				Family:         "Fake Font",
				Style:          "Regular",
				PostscriptName: "FakeFontRegular",
			},
			updateURL: "cloudfront.com/layerhub/fakefont.ttf",
			expectedFont: layerhub.Font{
				ID:             "font_1",
				FullName:       "Fake Font Regular",
				Family:         "Fake Font",
				Style:          "Regular",
				PostscriptName: "FakeFontRegular",
				URL:            "cloudfront.com/layerhub/fakefont.ttf",
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM fonts")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutFont(context.TODO(), &tc.newFont)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateURL != "" {
				tc.newFont.URL = tc.updateURL
				err := db.PutFont(context.TODO(), &tc.newFont)
				if err != nil {
					t.Fatal(err)
				}
			}

			fonts, err := db.FindFonts(context.TODO(), &layerhub.Filter{ID: tc.newFont.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(fonts) == 0 {
				t.Fatal("font not found")
			}

			got := fonts[0]
			if tc.expectedFont != got {
				t.Errorf("mismatched fonts:\ngot: %v\n want: %v", got, tc.expectedFont)
			}
		})
	}
}

func TestMySQL_FindFonts(t *testing.T) {
	testcases := []struct {
		name          string
		query         *layerhub.Filter
		currentFonts  []layerhub.Font
		expectedFonts []layerhub.Font
		enabledFonts  []string
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "font_2"},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
			expectedFonts: []layerhub.Font{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				PostscriptName: "FakeFontRegular",
			},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			expectedFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
		},
		{
			name:  "multiple results",
			query: &layerhub.Filter{},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			expectedFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
		},
		{
			name:  "search enabled fonts",
			query: &layerhub.Filter{UserID: "user_1", FontEnabled: ptr.Bool(true)},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			enabledFonts: []string{"font_1"},
			expectedFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
		},
		{
			name:  "search not enabled fonts",
			query: &layerhub.Filter{UserID: "user_1", FontEnabled: ptr.Bool(false)},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			enabledFonts: []string{},
			expectedFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM fonts")
			if err != nil {
				t.Fatal(err)
			}

			_, err = sqlDB(db).Exec("DELETE FROM enabled_fonts")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentFonts) > 0 {
				for _, f := range tc.currentFonts {
					err := db.PutFont(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			if len(tc.enabledFonts) > 0 {
				enabledFonts := make([]*layerhub.EnabledFont, len(tc.enabledFonts))
				for i, fid := range tc.enabledFonts {
					enabledFonts[i] = layerhub.NewEnabledFont(tc.query.UserID, fid)
				}
				err := db.BatchCreateEnabledFonts(context.TODO(), enabledFonts)
				if err != nil {
					t.Fatal(err)
				}
			}

			fonts, err := db.FindFonts(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(fonts, tc.expectedFonts) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", fonts, tc.expectedFonts)
			}
		})
	}
}

func TestMySQL_CountFonts(t *testing.T) {
	testcases := []struct {
		name          string
		query         *layerhub.Filter
		enabledFonts  []string
		currentFonts  []layerhub.Font
		expectedCount int
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "font_2"},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
			expectedCount: 0,
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				PostscriptName: "FakeFontRegular",
			},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple results",
			query: &layerhub.Filter{},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			expectedCount: 2,
		},
		{
			name:  "search enabled fonts",
			query: &layerhub.Filter{UserID: "user_1", FontEnabled: ptr.Bool(true)},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			enabledFonts:  []string{"font_1"},
			expectedCount: 1,
		},
		{
			name:  "search not enabled fonts",
			query: &layerhub.Filter{UserID: "user_1", FontEnabled: ptr.Bool(false)},
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
				{
					ID:             "font_2",
					FullName:       "Fake Font 2 Regular",
					Family:         "Fake Font 2",
					Style:          "Regular",
					PostscriptName: "FakeFont2Regular",
				},
			},
			enabledFonts:  []string{},
			expectedCount: 2,
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM fonts")
			if err != nil {
				t.Fatal(err)
			}

			_, err = sqlDB(db).Exec("DELETE FROM enabled_fonts")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentFonts) > 0 {
				for _, f := range tc.currentFonts {
					err := db.PutFont(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			if len(tc.enabledFonts) > 0 {
				enabledFonts := make([]*layerhub.EnabledFont, len(tc.enabledFonts))
				for i, fid := range tc.enabledFonts {
					enabledFonts[i] = layerhub.NewEnabledFont(tc.query.UserID, fid)
				}
				err := db.BatchCreateEnabledFonts(context.TODO(), enabledFonts)
				if err != nil {
					t.Fatal(err)
				}
			}

			count, err := db.CountFonts(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if count != tc.expectedCount {
				t.Fatalf("mismatched count result:\ngot: %v\nwant: %v", count, tc.expectedCount)
			}
		})
	}
}

func TestMySQL_DeleteFont(t *testing.T) {
	testcases := []struct {
		name         string
		deleteID     string
		currentFonts []layerhub.Font
	}{
		{
			name:     "font found",
			deleteID: "font_1",
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
		},
		{
			name:     "font not found",
			deleteID: "font_2",
			currentFonts: []layerhub.Font{
				{
					ID:             "font_1",
					FullName:       "Fake Font Regular",
					Family:         "Fake Font",
					Style:          "Regular",
					PostscriptName: "FakeFontRegular",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM fonts")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentFonts) > 0 {
				for _, f := range tc.currentFonts {
					err := db.PutFont(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err = db.DeleteFont(context.TODO(), tc.deleteID)
			if err != nil {
				t.Fatal(err)
			}

			fonts, err := db.FindFonts(context.TODO(), &layerhub.Filter{ID: tc.deleteID})
			if err != nil {
				t.Fatal(err)
			}

			if len(fonts) != 0 {
				t.Fatalf("font not deleted:\ngot: %v", fonts[0])
			}
		})
	}
}

func TestMySQL_PutTemplate(t *testing.T) {
	now := layerhub.Now()
	testscases := []struct {
		name             string
		newTemplate      layerhub.Template
		updateName       string
		expectedTemplate layerhub.Template
	}{
		{
			name: "new template",
			newTemplate: layerhub.Template{
				ID:   "template_1",
				Name: "Fake design",
				Frame: layerhub.Frame{
					ID:         "template_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Metadata: layerhub.Metadata{
					ID:      "template_1",
					License: "MIT",
				},
				Tags:      []string{"awesome", "free"},
				Colors:    []string{"white", "blue"},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedTemplate: layerhub.Template{
				ID:   "template_1",
				Name: "Fake design",
				Frame: layerhub.Frame{
					ID:         "template_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Metadata: layerhub.Metadata{
					ID:      "template_1",
					License: "MIT",
				},
				Tags:      []string{"awesome", "free"},
				Colors:    []string{"white", "blue"},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "update name",
			newTemplate: layerhub.Template{
				ID:   "template_1",
				Name: "Fake design",
				Frame: layerhub.Frame{
					ID:         "template_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Metadata: layerhub.Metadata{
					ID:      "template_1",
					License: "MIT",
				},
				Tags:      []string{},
				Colors:    []string{},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			updateName: "Updated fake design",
			expectedTemplate: layerhub.Template{
				ID:   "template_1",
				Name: "Updated fake design",
				Frame: layerhub.Frame{
					ID:         "template_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Metadata: layerhub.Metadata{
					ID:      "template_1",
					License: "MIT",
				},
				Tags:      []string{},
				Colors:    []string{},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM templates")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutTemplate(context.TODO(), &tc.newTemplate)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateName != "" {
				tc.newTemplate.Name = tc.updateName
				err := db.PutTemplate(context.TODO(), &tc.newTemplate)
				if err != nil {
					t.Fatal(err)
				}
			}

			templates, err := db.FindTemplates(context.TODO(), &layerhub.Filter{ID: tc.newTemplate.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(templates) == 0 {
				t.Fatal("template not found")
			}

			got := templates[0]
			if !reflect.DeepEqual(tc.expectedTemplate, got) {
				fmt.Println(got.Colors, tc.expectedTemplate.Colors)
				t.Errorf("mismatched templates:\ngot: %v\n want: %v", got, tc.expectedTemplate)
			}
		})
	}
}

func TestMySQL_FindTemplates(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name              string
		query             *layerhub.Filter
		currentTemplates  []layerhub.Template
		expectedTemplates []layerhub.Template
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "template_2"},
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_1",
						License: "MIT",
					},
					Tags:      []string{},
					Colors:    []string{},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedTemplates: []layerhub.Template{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "template_2",
			},
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_1",
						License: "MIT",
					},
					Tags:      []string{},
					Colors:    []string{},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "template_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "template_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_2",
						License: "MIT",
					},
					Tags:      []string{},
					Colors:    []string{},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedTemplates: []layerhub.Template{
				{
					ID:   "template_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "template_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_2",
						License: "MIT",
					},
					Tags:      []string{},
					Colors:    []string{},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_1",
						License: "MIT",
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "template_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "template_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_2",
						License: "MIT",
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_1",
						License: "MIT",
					},
					Tags:      []string{},
					Colors:    []string{},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "template_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "template_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Metadata: layerhub.Metadata{
						ID:      "template_2",
						License: "MIT",
					},
					Tags:      []string{},
					Colors:    []string{},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM templates")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentTemplates) > 0 {
				for _, f := range tc.currentTemplates {
					err := db.PutTemplate(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			templates, err := db.FindTemplates(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(templates, tc.expectedTemplates) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", templates, tc.expectedTemplates)
			}
		})
	}
}

func TestMySQL_CountTemplates(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name             string
		query            *layerhub.Filter
		currentTemplates []layerhub.Template
		expectedCount    int
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "template_2"},
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 0,
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "template_2",
			},
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "template_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "template_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "template_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "template_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "template_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 2,
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM templates")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentTemplates) > 0 {
				for _, f := range tc.currentTemplates {
					err := db.PutTemplate(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			count, err := db.CountTemplates(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if count != tc.expectedCount {
				t.Fatalf("mismatched count result:\ngot: %v\nwant: %v", count, tc.expectedCount)
			}
		})
	}
}

func TestMySQL_DeleteTemplate(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name             string
		deleteID         string
		currentTemplates []layerhub.Template
	}{
		{
			name:     "template found",
			deleteID: "template_1",
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:     "template_1",
						Width:  420,
						Height: 420,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:     "template not found",
			deleteID: "template_2",
			currentTemplates: []layerhub.Template{
				{
					ID:   "template_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:     "template_1",
						Width:  420,
						Height: 420,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM templates")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentTemplates) > 0 {
				for _, f := range tc.currentTemplates {
					err := db.PutTemplate(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err = db.DeleteTemplate(context.TODO(), tc.deleteID)
			if err != nil {
				t.Fatal(err)
			}

			templates, err := db.FindTemplates(context.TODO(), &layerhub.Filter{ID: tc.deleteID})
			if err != nil {
				t.Fatal(err)
			}

			if len(templates) != 0 {
				t.Fatalf("template not deleted:\ngot: %v", templates[0])
			}
		})
	}
}

func TestMySQL_PutProject(t *testing.T) {
	now := layerhub.Now()
	testscases := []struct {
		name            string
		newProject      layerhub.Project
		updateName      string
		expectedProject layerhub.Project
	}{
		{
			name: "new project",
			newProject: layerhub.Project{
				ID:   "project_1",
				Name: "Fake design",
				Frame: layerhub.Frame{
					ID:         "project_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedProject: layerhub.Project{
				ID:   "project_1",
				Name: "Fake design",
				Frame: layerhub.Frame{
					ID:         "project_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "update name",
			newProject: layerhub.Project{
				ID:   "project_1",
				Name: "Fake design",
				Frame: layerhub.Frame{
					ID:         "project_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			updateName: "Updated fake design",
			expectedProject: layerhub.Project{
				ID:   "project_1",
				Name: "Updated fake design",
				Frame: layerhub.Frame{
					ID:         "project_1",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePrivate,
				},
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM projects")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutProject(context.TODO(), &tc.newProject)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateName != "" {
				tc.newProject.Name = tc.updateName
				err := db.PutProject(context.TODO(), &tc.newProject)
				if err != nil {
					t.Fatal(err)
				}
			}

			projects, err := db.FindProjects(context.TODO(), &layerhub.Filter{ID: tc.newProject.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(projects) == 0 {
				t.Fatal("project not found")
			}

			got := projects[0]
			if !reflect.DeepEqual(tc.expectedProject, got) {
				t.Errorf("mismatched projects:\ngot: %v\n want: %v", got, tc.expectedProject)
			}
		})
	}
}

func TestMySQL_FindProjects(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name             string
		query            *layerhub.Filter
		currentProjects  []layerhub.Project
		expectedProjects []layerhub.Project
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "project_2"},
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedProjects: []layerhub.Project{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "project_2",
			},
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "project_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "project_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedProjects: []layerhub.Project{
				{
					ID:   "project_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "project_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "project_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "project_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "project_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "project_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM projects")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentProjects) > 0 {
				for _, f := range tc.currentProjects {
					err := db.PutProject(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			projects, err := db.FindProjects(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(projects, tc.expectedProjects) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", projects, tc.expectedProjects)
			}
		})
	}
}

func TestMySQL_CountProjects(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name            string
		query           *layerhub.Filter
		currentProjects []layerhub.Project
		expectedCount   int
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "project_2"},
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 0,
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "project_2",
			},
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "project_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "project_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:         "project_1",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:   "project_2",
					Name: "Fake design 2",
					Frame: layerhub.Frame{
						ID:         "project_2",
						Width:      420,
						Height:     420,
						Visibility: layerhub.FramePrivate,
					},
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 2,
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM projects")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentProjects) > 0 {
				for _, f := range tc.currentProjects {
					err := db.PutProject(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			count, err := db.CountProjects(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if count != tc.expectedCount {
				t.Fatalf("mismatched count result:\ngot: %v\nwant: %v", count, tc.expectedCount)
			}
		})
	}
}

func TestMySQL_DeleteProject(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name            string
		deleteID        string
		currentProjects []layerhub.Project
	}{
		{
			name:     "project found",
			deleteID: "project_1",
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:     "project_1",
						Width:  420,
						Height: 420,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:     "project not found",
			deleteID: "project_2",
			currentProjects: []layerhub.Project{
				{
					ID:   "project_1",
					Name: "Fake design",
					Frame: layerhub.Frame{
						ID:     "project_1",
						Width:  420,
						Height: 420,
					},
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM projects")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentProjects) > 0 {
				for _, f := range tc.currentProjects {
					err := db.PutProject(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err = db.DeleteProject(context.TODO(), tc.deleteID)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := db.FindProjects(context.TODO(), &layerhub.Filter{ID: tc.deleteID})
			if err != nil {
				t.Fatal(err)
			}

			if len(projects) != 0 {
				t.Fatalf("project not deleted:\ngot: %v", projects[0])
			}
		})
	}
}

func TestMySQL_PutUpload(t *testing.T) {
	now := layerhub.Now()
	testscases := []struct {
		name           string
		newUpload      layerhub.Upload
		updateURL      string
		expectedUpload layerhub.Upload
	}{
		{
			name: "new upload",
			newUpload: layerhub.Upload{
				ID:        "upload_1",
				Name:      "Fake image",
				UserID:    "user_1",
				URL:       "cloudfront.com/uploads/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedUpload: layerhub.Upload{
				ID:        "upload_1",
				Name:      "Fake image",
				UserID:    "user_1",
				URL:       "cloudfront.com/uploads/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "update url",
			newUpload: layerhub.Upload{
				ID:        "upload_1",
				Name:      "Fake image",
				UserID:    "user_1",
				URL:       "cloudfront.com/uploads/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			updateURL: "cloudfront.com/uploads/6.png",
			expectedUpload: layerhub.Upload{
				ID:        "upload_1",
				Name:      "Fake image",
				UserID:    "user_1",
				URL:       "cloudfront.com/uploads/6.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM uploads")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutUpload(context.TODO(), &tc.newUpload)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateURL != "" {
				tc.newUpload.URL = tc.updateURL
				err := db.PutUpload(context.TODO(), &tc.newUpload)
				if err != nil {
					t.Fatal(err)
				}
			}

			uploads, err := db.FindUploads(context.TODO(), &layerhub.Filter{ID: tc.newUpload.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(uploads) == 0 {
				t.Fatal("upload not found")
			}

			got := uploads[0]
			if tc.expectedUpload != got {
				t.Errorf("mismatched uploads:\ngot: %v\n want: %v", got, tc.expectedUpload)
			}
		})
	}
}

func TestMySQL_FindUploads(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name            string
		query           *layerhub.Filter
		currentUploads  []layerhub.Upload
		expectedUploads []layerhub.Upload
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "upload_2"},
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedUploads: []layerhub.Upload{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "upload_2",
			},
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "upload_2",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedUploads: []layerhub.Upload{
				{
					ID:        "upload_2",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "upload_2",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "upload_2",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM uploads")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentUploads) > 0 {
				for _, f := range tc.currentUploads {
					err := db.PutUpload(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			uploads, err := db.FindUploads(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(uploads, tc.expectedUploads) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", uploads, tc.expectedUploads)
			}
		})
	}
}

func TestMySQL_CountUploads(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name           string
		query          *layerhub.Filter
		currentUploads []layerhub.Upload
		expectedCount  int
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "upload_2"},
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 0,
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "upload_2",
			},
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "upload_2",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "upload_2",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 2,
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM uploads")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentUploads) > 0 {
				for _, f := range tc.currentUploads {
					err := db.PutUpload(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			count, err := db.CountUploads(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if count != tc.expectedCount {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", count, tc.expectedCount)
			}
		})
	}
}

func TestMySQL_DeleteUploads(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name           string
		deleteID       string
		currentUploads []layerhub.Upload
	}{
		{
			name:     "upload found",
			deleteID: "upload_1",
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:     "upload not found",
			deleteID: "upload_2",
			currentUploads: []layerhub.Upload{
				{
					ID:        "upload_1",
					Name:      "Fake image",
					UserID:    "user_1",
					URL:       "cloudfront.com/uploads/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM projects")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentUploads) > 0 {
				for _, f := range tc.currentUploads {
					err := db.PutUpload(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err = db.DeleteUpload(context.TODO(), tc.deleteID)
			if err != nil {
				t.Fatal(err)
			}

			uploads, err := db.FindUploads(context.TODO(), &layerhub.Filter{ID: tc.deleteID})
			if err != nil {
				t.Fatal(err)
			}

			if len(uploads) != 0 {
				t.Fatalf("upload not deleted:\ngot: %v", uploads[0])
			}
		})
	}
}

func TestMySQL_PutComponent(t *testing.T) {
	now := layerhub.Now()
	testscases := []struct {
		name              string
		newComponent      layerhub.Component
		updateName        string
		expectedComponent layerhub.Component
	}{
		{
			name: "new component",
			newComponent: layerhub.Component{
				ID:        "component_1",
				Name:      "Fake design",
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedComponent: layerhub.Component{
				ID:        "component_1",
				Name:      "Fake design",
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "update name",
			newComponent: layerhub.Component{
				ID:        "component_1",
				Name:      "Fake design",
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			updateName: "Updated fake design",
			expectedComponent: layerhub.Component{
				ID:        "component_1",
				Name:      "Updated fake design",
				Preview:   "cloudfront.com/preview/1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM components")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutComponent(context.TODO(), &tc.newComponent)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateName != "" {
				tc.newComponent.Name = tc.updateName
				err := db.PutComponent(context.TODO(), &tc.newComponent)
				if err != nil {
					t.Fatal(err)
				}
			}

			components, err := db.FindComponents(context.TODO(), &layerhub.Filter{ID: tc.newComponent.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(components) == 0 {
				t.Fatal("component not found")
			}

			got := components[0]
			if !reflect.DeepEqual(tc.expectedComponent, got) {
				t.Errorf("mismatched components:\ngot: %v\n want: %v", got, tc.expectedComponent)
			}
		})
	}
}

func TestMySQL_FindComponents(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name               string
		query              *layerhub.Filter
		currentComponents  []layerhub.Component
		expectedComponents []layerhub.Component
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "component_2"},
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedComponents: []layerhub.Component{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "component_2",
			},
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "component_2",
					Name:      "Fake design 2",
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedComponents: []layerhub.Component{
				{
					ID:        "component_2",
					Name:      "Fake design 2",
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "component_2",
					Name:      "Fake design 2",
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "component_2",
					Name:      "Fake design 2",
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM components")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentComponents) > 0 {
				for _, f := range tc.currentComponents {
					err := db.PutComponent(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			components, err := db.FindComponents(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(components, tc.expectedComponents) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", components, tc.expectedComponents)
			}
		})
	}
}

func TestMySQL_CountComponents(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name              string
		query             *layerhub.Filter
		currentComponents []layerhub.Component
		expectedCount     int
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "component_2"},
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 0,
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "component_2",
			},
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "component_2",
					Name:      "Fake design 2",
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "component_2",
					Name:      "Fake design 2",
					Preview:   "cloudfront.com/preview/2.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedCount: 2,
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM components")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentComponents) > 0 {
				for _, f := range tc.currentComponents {
					err := db.PutComponent(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			count, err := db.CountComponents(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if count != tc.expectedCount {
				t.Fatalf("mismatched count result:\ngot: %v\nwant: %v", count, tc.expectedCount)
			}
		})
	}
}

func TestMySQL_DeleteComponent(t *testing.T) {
	now := layerhub.Now()
	testcases := []struct {
		name              string
		deleteID          string
		currentComponents []layerhub.Component
	}{
		{
			name:     "component found",
			deleteID: "component_1",
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name:     "component not found",
			deleteID: "component_2",
			currentComponents: []layerhub.Component{
				{
					ID:        "component_1",
					Name:      "Fake design",
					Preview:   "cloudfront.com/preview/1.png",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM components")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentComponents) > 0 {
				for _, f := range tc.currentComponents {
					err := db.PutComponent(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err = db.DeleteComponent(context.TODO(), tc.deleteID)
			if err != nil {
				t.Fatal(err)
			}

			components, err := db.FindComponents(context.TODO(), &layerhub.Filter{ID: tc.deleteID})
			if err != nil {
				t.Fatal(err)
			}

			if len(components) != 0 {
				t.Fatalf("component not deleted:\ngot: %v", components[0])
			}
		})
	}
}

func TestMySQL_PutFrame(t *testing.T) {
	testscases := []struct {
		name          string
		newFrame      layerhub.Frame
		updateHeight  float64
		expectedFrame layerhub.Frame
	}{
		{
			name: "new frame",
			newFrame: layerhub.Frame{
				ID:         "frame_1",
				Name:       "Default",
				Width:      420,
				Height:     420,
				Visibility: layerhub.FramePublic,
				Unit:       layerhub.Pixels,
				Preview:    "cloudfront.com/frames/1.png",
			},
			expectedFrame: layerhub.Frame{
				ID:         "frame_1",
				Name:       "Default",
				Width:      420,
				Height:     420,
				Visibility: layerhub.FramePublic,
				Unit:       layerhub.Pixels,
				Preview:    "cloudfront.com/frames/1.png",
			},
		},
		{
			name: "update width",
			newFrame: layerhub.Frame{
				ID:         "frame_1",
				Name:       "Default",
				Width:      420,
				Height:     420,
				Visibility: layerhub.FramePublic,
				Unit:       layerhub.Pixels,
				Preview:    "cloudfront.com/frames/1.png",
			},
			updateHeight: 600,
			expectedFrame: layerhub.Frame{
				ID:         "frame_1",
				Name:       "Default",
				Width:      420,
				Height:     600,
				Visibility: layerhub.FramePublic,
				Unit:       layerhub.Pixels,
				Preview:    "cloudfront.com/frames/1.png",
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM frames")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutFrame(context.TODO(), &tc.newFrame)
			if err != nil {
				t.Fatal(err)
			}

			if tc.updateHeight != 0 {
				tc.newFrame.Height = tc.updateHeight
				err := db.PutFrame(context.TODO(), &tc.newFrame)
				if err != nil {
					t.Fatal(err)
				}
			}

			frames, err := db.FindFrames(context.TODO(), &layerhub.Filter{ID: tc.newFrame.ID, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if len(frames) == 0 {
				t.Fatal("frame not found")
			}

			got := frames[0]
			if tc.expectedFrame != got {
				t.Errorf("mismatched frames:\ngot: %v\n want: %v", got, tc.expectedFrame)
			}
		})
	}
}

func TestMySQL_FindFrames(t *testing.T) {
	testcases := []struct {
		name           string
		query          *layerhub.Filter
		currentFrames  []layerhub.Frame
		expectedFrames []layerhub.Frame
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "frame_2"},
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
			expectedFrames: []layerhub.Frame{},
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "frame_2",
			},
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
				{
					ID:         "frame_2",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
			expectedFrames: []layerhub.Frame{
				{
					ID:         "frame_2",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
				{
					ID:         "frame_2",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
			expectedFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
				{
					ID:         "frame_2",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM frames")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentFrames) > 0 {
				for _, f := range tc.currentFrames {
					err := db.PutFrame(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			frames, err := db.FindFrames(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(frames, tc.expectedFrames) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", frames, tc.expectedFrames)
			}
		})
	}
}

func TestMySQL_CountFrames(t *testing.T) {
	testcases := []struct {
		name          string
		query         *layerhub.Filter
		currentFrames []layerhub.Frame
		expectedCount int
	}{
		{
			name:  "empty result",
			query: &layerhub.Filter{ID: "frame_2"},
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
			expectedCount: 0,
		},
		{
			name: "one result",
			query: &layerhub.Filter{
				ID: "frame_2",
			},
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
				{
					ID:         "frame_2",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple result",
			query: &layerhub.Filter{},
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
				{
					ID:         "frame_2",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
			expectedCount: 2,
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM frames")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentFrames) > 0 {
				for _, f := range tc.currentFrames {
					err := db.PutFrame(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			count, err := db.CountFrames(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if count != tc.expectedCount {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", count, tc.expectedCount)
			}
		})
	}
}

func TestMySQL_DeleteFrame(t *testing.T) {
	testcases := []struct {
		name          string
		deleteID      string
		currentFrames []layerhub.Frame
	}{
		{
			name:     "frame found",
			deleteID: "frame_1",
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
		},
		{
			name:     "frame not found",
			deleteID: "frame_2",
			currentFrames: []layerhub.Frame{
				{
					ID:         "frame_1",
					Name:       "Default",
					Width:      420,
					Height:     420,
					Visibility: layerhub.FramePublic,
					Unit:       layerhub.Pixels,
					Preview:    "cloudfront.com/frames/1.png",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM projects")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentFrames) > 0 {
				for _, f := range tc.currentFrames {
					err := db.PutFrame(context.TODO(), &f)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			err = db.DeleteFrame(context.TODO(), tc.deleteID)
			if err != nil {
				t.Fatal(err)
			}

			projects, err := db.FindFrames(context.TODO(), &layerhub.Filter{ID: tc.deleteID})
			if err != nil {
				t.Fatal(err)
			}

			if len(projects) != 0 {
				t.Fatalf("frame not deleted:\ngot: %v", projects[0])
			}
		})
	}
}

func TestMySQL_FindEnabledFonts(t *testing.T) {
	testcases := []struct {
		name                 string
		query                string
		currentEnabledFonts  []*layerhub.EnabledFont
		expectedEnabledFonts []layerhub.EnabledFont
	}{
		{
			name:  "empty result",
			query: "user_2",
			currentEnabledFonts: []*layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
			},
			expectedEnabledFonts: []layerhub.EnabledFont{},
		},
		{
			name:  "one result",
			query: "user_1",
			currentEnabledFonts: []*layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
				{
					ID:     "enabled_font_2",
					UserID: "user_2",
					FontID: "font_2",
				},
			},
			expectedEnabledFonts: []layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
			},
		},
		{
			name:  "multiple result",
			query: "user_1",
			currentEnabledFonts: []*layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
				{
					ID:     "enabled_font_2",
					UserID: "user_1",
					FontID: "font_2",
				},
			},
			expectedEnabledFonts: []layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
				{
					ID:     "enabled_font_2",
					UserID: "user_1",
					FontID: "font_2",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM enabled_fonts")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentEnabledFonts) > 0 {
				err := db.BatchCreateEnabledFonts(context.TODO(), tc.currentEnabledFonts)
				if err != nil {
					t.Fatal(err)
				}
			}

			enabledFonts, err := db.FindEnabledFonts(context.TODO(), tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(enabledFonts, tc.expectedEnabledFonts) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", enabledFonts, tc.expectedEnabledFonts)
			}
		})
	}
}

func TestMySQL_BatchDeleteEnabledFonts(t *testing.T) {
	testcases := []struct {
		name                 string
		deleteIDs            []string
		currentEnabledFonts  []*layerhub.EnabledFont
		expectedEnabledFonts []layerhub.EnabledFont
	}{
		{
			name:      "enabled fonts found",
			deleteIDs: []string{"enabled_font_1", "enabled_font_2"},
			currentEnabledFonts: []*layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
				{
					ID:     "enabled_font_2",
					UserID: "user_1",
					FontID: "font_2",
				},
			},
			expectedEnabledFonts: []layerhub.EnabledFont{},
		},
		{
			name:      "enabled fonts not found",
			deleteIDs: []string{"enabled_font_3"},
			currentEnabledFonts: []*layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
				{
					ID:     "enabled_font_2",
					UserID: "user_1",
					FontID: "font_2",
				},
			},
			expectedEnabledFonts: []layerhub.EnabledFont{
				{
					ID:     "enabled_font_1",
					UserID: "user_1",
					FontID: "font_1",
				},
				{
					ID:     "enabled_font_2",
					UserID: "user_1",
					FontID: "font_2",
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM enabled_fonts")
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.currentEnabledFonts) > 0 {
				err := db.BatchCreateEnabledFonts(context.TODO(), tc.currentEnabledFonts)
				if err != nil {
					t.Fatal(err)
				}
			}

			err = db.BatchDeleteEnabledFonts(context.TODO(), tc.deleteIDs)
			if err != nil {
				t.Fatal(err)
			}

			fonts, err := db.FindEnabledFonts(context.TODO(), "user_1")
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(fonts, tc.expectedEnabledFonts) {
				t.Fatalf("mismatched find result:\ngot: %v\nwant: %v", fonts, tc.expectedEnabledFonts)
			}
		})
	}
}

func TestMySQL_PutSubscriptionPlan(t *testing.T) {
	testscases := []struct {
		name                     string
		newSubscriptionPlan      layerhub.SubscriptionPlan
		updateBillings           []*layerhub.Billing
		expectedSubscriptionPlan layerhub.SubscriptionPlan
	}{
		{
			name: "new user",
			newSubscriptionPlan: layerhub.SubscriptionPlan{
				ID:                  "subscription_plan_1",
				Name:                "Basic Plan",
				Description:         "Basic Plan",
				Provider:            layerhub.Paypal,
				AutoBillOutstanding: true,
				SetupFee:            "1",
				Billing: []*layerhub.Billing{
					{
						ID:       "billing_1",
						Interval: "MONTH",
						Price:    "10",
					},
					{
						ID:       "billing_2",
						Interval: "YEAR",
						Price:    "100",
					},
				},
			},
			expectedSubscriptionPlan: layerhub.SubscriptionPlan{
				ID:                  "subscription_plan_1",
				Name:                "Basic Plan",
				Description:         "Basic Plan",
				Provider:            layerhub.Paypal,
				AutoBillOutstanding: true,
				SetupFee:            "1",
				Billing: []*layerhub.Billing{
					{
						ID:                 "billing_1",
						Interval:           "MONTH",
						Price:              "10",
						SubscriptionPlanID: "subscription_plan_1",
					},
					{
						ID:                 "billing_2",
						Interval:           "YEAR",
						Price:              "100",
						SubscriptionPlanID: "subscription_plan_1",
					},
				},
			},
		},
		{
			name: "update billings",
			newSubscriptionPlan: layerhub.SubscriptionPlan{
				ID:                  "subscription_plan_1",
				Name:                "Basic Plan",
				Description:         "Basic Plan",
				Provider:            layerhub.Paypal,
				AutoBillOutstanding: true,
				SetupFee:            "1",
				Billing: []*layerhub.Billing{
					{
						ID:                 "billing_1",
						Interval:           "MONTH",
						Price:              "10",
						SubscriptionPlanID: "subscription_plan_1",
					},
					{
						ID:                 "billing_2",
						Interval:           "YEAR",
						Price:              "100",
						SubscriptionPlanID: "subscription_plan_1",
					},
				},
			},
			updateBillings: []*layerhub.Billing{
				{
					ID:                 "billing_1",
					Interval:           "MONTH",
					Price:              "10",
					SubscriptionPlanID: "subscription_plan_1",
				},
			},
			expectedSubscriptionPlan: layerhub.SubscriptionPlan{
				ID:                  "subscription_plan_1",
				Name:                "Basic Plan",
				Description:         "Basic Plan",
				Provider:            layerhub.Paypal,
				AutoBillOutstanding: true,
				SetupFee:            "1",
				Billing: []*layerhub.Billing{
					{
						ID:                 "billing_1",
						Interval:           "MONTH",
						Price:              "10",
						SubscriptionPlanID: "subscription_plan_1",
					},
				},
			},
		},
	}

	// Share container for speed
	cleanup, dsn := prepareTestContainer(t)
	defer cleanup()

	initDB(t, dsn)

	db, err := New(&Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testscases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlDB(db).Exec("DELETE FROM subscription_plans")
			if err != nil {
				t.Fatal(err)
			}

			err = db.PutSubscriptionPlan(context.TODO(), &tc.newSubscriptionPlan)
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.updateBillings) > 0 {
				tc.newSubscriptionPlan.Billing = tc.updateBillings
				err := db.PutSubscriptionPlan(context.TODO(), &tc.newSubscriptionPlan)
				if err != nil {
					t.Fatal(err)
				}
			}

			plans, err := db.FindSubscriptionPlans(context.TODO())
			if err != nil {
				t.Fatal(err)
			}
			if len(plans) == 0 {
				t.Fatal("user not found")
			}

			got := plans[0]
			if reflect.DeepEqual(plans, tc.expectedSubscriptionPlan) {
				t.Errorf("mismatched users:\ngot: %v\n want: %v", got, tc.expectedSubscriptionPlan)
			}
		})
	}
}

func initDB(t *testing.T, dsn string) {
	m, err := migrate.New("file://../migrations", fmt.Sprintf("mysql://%s", dsn))
	if err != nil {
		t.Fatalf("could not migrate: %s", err)
	}

	if err := m.Down(); err != nil {
		if err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
			t.Fatalf("could not migrate: %s", err)
		}
	}

	if err := m.Up(); err != nil {
		t.Fatalf("could not migrate: %s", err)
	}
}

func sqlDB(db layerhub.DB) *sqlx.DB {
	return db.(*MySQLDB).db
}

type cfg struct {
	docker.ServiceHostPort
	DSN string
}

var _ docker.ServiceConfig = &cfg{}

func prepareTestContainer(t *testing.T) (func(), string) {
	if url := os.Getenv("TEST_MYSQL_DSN"); url != "" {
		return func() {}, url
	}

	runner, err := docker.NewServiceRunner(docker.RunOptions{
		ImageRepo:     "mysql",
		ImageTag:      "latest",
		ContainerName: "mysql-test",
		Ports:         []string{"3306/tcp"},
		Env: []string{
			"MYSQL_ROOT_PASSWORD=dev",
			"MYSQL_PASSWORD=dev",
			"MYSQL_USER=dev",
			"MYSQL_DATABASE=layerhub",
		},
	})
	if err != nil {
		t.Fatalf("could not start local MySQL: %s", err)
	}

	svc, err := runner.StartService(context.Background(), connect)
	if err != nil {
		t.Fatalf("could not start local MySQL: %s", err)
	}

	return svc.Cleanup, svc.Config.(*cfg).DSN
}

func connect(ctx context.Context, host string, port int) (docker.ServiceConfig, error) {
	hostIP := docker.NewServiceHostPort(host, port)
	dsn := fmt.Sprintf("dev:dev@tcp(%s)/layerhub?parseTime=true", hostIP.Address())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &cfg{
		ServiceHostPort: *hostIP,
		DSN:             dsn,
	}, nil
}
