package client

import "encoding/json"

// ToJsonString is a generic stringer which outputs
// a struct in its equivalent (indented) json representation
// If json marshalling not sucessful returns error
func ToJsonString(d interface{}) (string, error) {
	s, err := json.MarshalIndent(d, "", "	")
	if err != nil {
		return "", err
	}
	return string(s), nil
}
