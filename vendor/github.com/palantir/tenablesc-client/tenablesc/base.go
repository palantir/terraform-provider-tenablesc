package tenablesc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Common structures used throughout the Tenable API.

// BaseInfo is used in the API to refer to related objects;
//   For input, only the ID is needed; responses will include name and description.
type BaseInfo struct {
	ID          ProbablyString `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
}

// UserInfo is used inb the API to refer to users.
//   For input, only the ID is needed; responses will include remaining fields.
type UserInfo struct {
	ID        ProbablyString `json:"id,omitempty"`
	Username  string         `json:"username,omitempty"`
	Firstname string         `json:"firstname,omitempty"`
	Lastname  string         `json:"lastname,omitempty"`
}

// FakeBool helps us wrap the tenable API's stringy booleans so users
// don't have to think about them quite as much.
type FakeBool string

const (
	FakeTrue  FakeBool = "true"
	FakeFalse FakeBool = "false"
)

func (f FakeBool) AsBool() bool {
	return f == FakeTrue
}

func ToFakeBool(b bool) FakeBool {
	if b {
		return FakeTrue
	}
	return FakeFalse
}

// ProbablyString is used for most ID fields in the API;
//  the SC API generally returns positive IDs as strings, but
//  -1 and sometimes 0 are returned numerically.
//  Since the API _accepts_ strings in those locations always, we only need to handle
//  the output from calls.
type ProbablyString string

func (p *ProbablyString) UnmarshalJSON(data []byte) error {
	//try as a string
	var str string
	//slightly weird control flow but improves nesting
	if err := json.Unmarshal(data, &str); err == nil {
		*p = ProbablyString(str)
		return nil
	}

	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return fmt.Errorf("failed to unmarshal '%s' as either a string or an int: %w", string(data), err)
	}

	*p = ProbablyString(strconv.Itoa(i))
	return nil

}

// UnixEpochStringTime would ideally be time.Time, but partial-unmarshalling
//  every single place we use a timestamp to do that conversion would be somewhat excruciating.
//  So this is the compromise.
type UnixEpochStringTime string

func (s UnixEpochStringTime) ToDateTime() (time.Time, error) {
	i, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
	}
	return time.Unix(i, 0), nil
}
