package pointer

// UpdatePointerField updates dest with src if src is not nil and (for strings) not empty
func UpdatePointerField[T any](dest *T, src *T) {
    if src == nil {
        return
    }
    if s, ok := any(*src).(string); ok && s == "" {
        return
    }
    *dest = *src
}
