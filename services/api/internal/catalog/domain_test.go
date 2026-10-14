package catalog

import (
	"errors"
	"testing"
)

func TestValidateInput(t *testing.T) {
	valid := BookInput{Title: "Toán 6", Grade: 6, SubjectCode: "math", RightsOwner: "Publisher", RightsLicense: "Permission-123", RightsSource: "Publisher", Visibility: "public"}
	global := Actor{UserID: "1", Global: true}
	if err := valid.Validate(global); err != nil {
		t.Fatalf("valid input: %v", err)
	}
	if err := valid.Validate(Actor{UserID: "1", SchoolID: "school"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("public school scope: %v", err)
	}
	valid.Visibility = "school"
	if err := valid.Validate(Actor{UserID: "1", SchoolID: "school"}); err != nil {
		t.Fatal(err)
	}
	valid.RightsSource = " "
	if err := valid.Validate(global); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}

func TestStatuses(t *testing.T) {
	pairs := map[string][2]string{"submit": {"draft", "review"}, "approve": {"review", "approved"}, "publish": {"approved", "published"}, "archive": {"published", "archived"}}
	for name, pair := range pairs {
		from, to, err := NextStatus(name)
		if err != nil || from != pair[0] || to != pair[1] {
			t.Fatalf("%s unexpected: %s %s %v", name, from, to, err)
		}
	}
	if _, _, err := NextStatus("skip"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unexpected allowed status")
	}
}
