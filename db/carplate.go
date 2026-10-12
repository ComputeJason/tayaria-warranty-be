package db

import "strings"

// normalizeCarPlate strips all whitespace (leading, trailing, and internal) and uppercases a
// car plate, so stored values and every search/lookup compare consistently regardless of how
// a customer or dealer happened to type the spacing (e.g. "abc 1234", " ABC1234", "A B C 1234"
// all normalize to "ABC1234"). Applied on every write (CreateWarranty, CreateClaim) and every
// read that looks a car plate up (GetWarrantiesByCarPlate, GetValidWarrantyByCarPlate,
// GetAllValidWarrantiesForCarPlate), so it's enforced centrally in one place rather than
// depending on every caller/handler to remember to normalize first.
func normalizeCarPlate(carPlate string) string {
	return strings.ToUpper(strings.Join(strings.Fields(carPlate), ""))
}
