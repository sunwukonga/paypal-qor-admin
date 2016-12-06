package utils

import "reflect"

// Started this and then abandoned it when I realized that it was, although possible, not particularly useful.
func contains(s []interface{}, e interface) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
