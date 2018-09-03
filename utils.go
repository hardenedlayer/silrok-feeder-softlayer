package srfsoftlayer

import (
	"encoding/json"
	"fmt"
)

// Inspect print out given data to standard output.
func Inspect(desc string, data interface{}) {
	fmt.Printf("{\n	\"desc\": \"%v\",\n\"datatype\": \"%T\",\n\"data\": %v }\n\n", desc, data, ToJSON(data))
}

// ToJSON returns json marshalled string of given data.
func ToJSON(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}
