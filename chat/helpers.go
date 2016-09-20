// Various helper functions

package chat


func contains(arr []string, s string) bool {
    for _, a := range arr {
        if a == s {
            return true
        }
    }
    return false
}
