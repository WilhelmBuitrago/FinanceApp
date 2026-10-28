package domain

import "strings"

/* The product lives in Colombia: one currency, spelled one way. The
   account route used to store whatever string arrived — "cop", "cop ",
   "COP" reached the database as three different monies, and anything
   at all ("pesos") was accepted as a currency. Normalization maps the
   input onto a known catalog; the default is COP. */

// supportedCurrencies is the catalog of money codes the app stores.
// Colombia is the product's home; the rest keep the door open without
// pretending any string is a currency.
var supportedCurrencies = map[string]struct{}{
	"COP": {}, "USD": {}, "EUR": {}, "MXN": {},
	"PEN": {}, "BRL": {}, "ARS": {}, "CLP": {},
	"GBP": {}, "CAD": {}, "CHF": {}, "AUD": {},
}

// DefaultCurrency is the money the product runs in.
const DefaultCurrency = "COP"

// NormalizeCurrency trims and uppercases the code, defaults an empty
// value to COP, and reports whether the result is a known currency.
func NormalizeCurrency(raw string) (string, bool) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if code == "" {
		return DefaultCurrency, true
	}
	if _, ok := supportedCurrencies[code]; !ok {
		return "", false
	}
	return code, true
}
