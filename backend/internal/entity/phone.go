package entity

// NormalizePhone normalizes phone to +7XXXXXXXXXX format.
// Handles: +7..., 8..., 7..., 10 digits without code.
// Returns empty string if format is not recognized.
func NormalizePhone(phone string) string {
	digits := make([]byte, 0, len(phone))
	for i := 0; i < len(phone); i++ {
		if phone[i] >= '0' && phone[i] <= '9' {
			digits = append(digits, phone[i])
		}
	}
	s := string(digits)

	switch {
	case len(s) == 11 && (s[0] == '7' || s[0] == '8'):
		return "+7" + s[1:]
	case len(s) == 10:
		return "+7" + s
	default:
		return ""
	}
}
