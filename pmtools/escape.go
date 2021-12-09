/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */


package pmtools

import (
    "strings"
)

func EscapeJSON(value string) string {
    return strings.ReplaceAll(value, `"`, `\"`)
}

func UnescapeJSON(value string) string {
    return strings.ReplaceAll(value, `\"`, `"`)
}
//EOF
