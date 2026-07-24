package app

import (
	"os"
	"strconv"
	"strings"
)

// OS-currency localization — specs/savings-ledger-mvp-design.md. Offline-first:
// a checked-in, dated FX table, never a network call. The localized figure is
// a second imputation layer over the USD estimate, so USD is always carried
// alongside as the auditable source of truth, and any miss falls back to USD
// rather than fabricating a number.

const savingsFXAsOf = "2026-07"

// usdToRate: USD → currency multiplier, as of savingsFXAsOf.
var usdToRate = map[string]float64{
	"USD": 1.0,
	"EUR": 0.92,
	"GBP": 0.79,
	"INR": 83.0,
	"JPY": 157.0,
	"CAD": 1.37,
	"AUD": 1.52,
	"CNY": 7.20,
	"BRL": 5.40,
	"SGD": 1.35,
}

type currencyFormat struct {
	Symbol   string
	Decimals int
	Indian   bool // lakh/crore grouping
}

var currencyFormats = map[string]currencyFormat{
	"USD": {"$", 2, false},
	"EUR": {"€", 2, false},
	"GBP": {"£", 2, false},
	"INR": {"₹", 2, true},
	"JPY": {"¥", 0, false},
	"CAD": {"CA$", 2, false},
	"AUD": {"A$", 2, false},
	"CNY": {"CN¥", 2, false},
	"BRL": {"R$", 2, false},
	"SGD": {"S$", 2, false},
}

// regionToCurrency maps an OS locale region to a currency.
var regionToCurrency = map[string]string{
	"US": "USD", "GB": "GBP", "IN": "INR", "JP": "JPY", "CA": "CAD",
	"AU": "AUD", "CN": "CNY", "BR": "BRL", "SG": "SGD",
	"DE": "EUR", "FR": "EUR", "ES": "EUR", "IT": "EUR", "NL": "EUR", "IE": "EUR",
}

type localizedAmount struct {
	Currency string
	Local    string // formatted local string, e.g. "₹9,900.00"
	USD      float64
	Rate     float64
	FellBack bool // detected a non-USD currency but had no rate; used USD
}

// detectDisplayCurrency resolves the display currency from JINI_CURRENCY, then
// the OS locale (LC_MONETARY / LC_ALL / LANG), defaulting to USD.
func detectDisplayCurrency() string {
	if override := strings.ToUpper(strings.TrimSpace(os.Getenv("JINI_CURRENCY"))); override != "" {
		if _, ok := usdToRate[override]; ok {
			return override
		}
	}
	for _, key := range []string{"LC_MONETARY", "LC_ALL", "LANG"} {
		region := regionFromLocale(os.Getenv(key))
		if region == "" {
			continue
		}
		if cur, ok := regionToCurrency[region]; ok {
			return cur
		}
	}
	return "USD"
}

// regionFromLocale extracts the 2-letter region from a locale such as
// "en_IN.UTF-8" → "IN".
func regionFromLocale(locale string) string {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return ""
	}
	if i := strings.IndexAny(locale, ".@"); i >= 0 {
		locale = locale[:i]
	}
	parts := strings.Split(locale, "_")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToUpper(parts[1])
}

// fxRate returns the USD→currency rate. JINI_FX_RATE overrides the table for a
// non-USD detected currency (testing / correcting a stale rate).
func fxRate(currency string) (float64, bool) {
	if currency != "USD" {
		if raw := strings.TrimSpace(os.Getenv("JINI_FX_RATE")); raw != "" {
			if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 {
				return v, true
			}
		}
	}
	rate, ok := usdToRate[currency]
	return rate, ok
}

// localize converts a USD amount to the OS display currency. On any miss it
// falls back to USD and flags it, so a figure is never fabricated.
func localize(usd float64) localizedAmount {
	currency := detectDisplayCurrency()
	if currency == "USD" {
		return localizedAmount{Currency: "USD", Local: formatMoney("USD", usd), USD: usd, Rate: 1.0}
	}
	rate, ok := fxRate(currency)
	if !ok {
		return localizedAmount{Currency: "USD", Local: formatMoney("USD", usd), USD: usd, Rate: 1.0, FellBack: true}
	}
	return localizedAmount{Currency: currency, Local: formatMoney(currency, usd*rate), USD: usd, Rate: rate}
}

// formatMoney renders an amount with the currency's symbol, decimals, and
// digit grouping (Indian lakh/crore where applicable).
func formatMoney(currency string, amount float64) string {
	f, ok := currencyFormats[currency]
	if !ok {
		f = currencyFormat{Symbol: currency + " ", Decimals: 2}
	}
	neg := amount < 0
	if neg {
		amount = -amount
	}
	s := strconv.FormatFloat(amount, 'f', f.Decimals, 64)
	intPart, fracPart := s, ""
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		intPart, fracPart = s[:dot], s[dot:]
	}
	out := f.Symbol + groupDigits(intPart, f.Indian) + fracPart
	if neg {
		out = "-" + out
	}
	return out
}

func groupDigits(intPart string, indian bool) string {
	n := len(intPart)
	if n <= 3 {
		return intPart
	}
	if indian {
		head, tail := intPart[:n-3], intPart[n-3:]
		var groups []string
		for len(head) > 2 {
			groups = append([]string{head[len(head)-2:]}, groups...)
			head = head[:len(head)-2]
		}
		groups = append([]string{head}, groups...)
		return strings.Join(groups, ",") + "," + tail
	}
	var b strings.Builder
	pre := n % 3
	if pre > 0 {
		b.WriteString(intPart[:pre])
	}
	for i := pre; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(intPart[i : i+3])
	}
	return b.String()
}
