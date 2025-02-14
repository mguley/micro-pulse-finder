package agent

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// ChromeAgent generates User-Agent strings simulating Chrome browser.
type ChromeAgent struct {
	versions []string
	os       []string
}

// NewChromeAgent creates a new instance of ChromeAgent.
func NewChromeAgent() *ChromeAgent {
	return &ChromeAgent{
		versions: []string{
			"126.0.6478.114", "126.0.6478.62", "126.0.6478.61",
			"126.0.6478.56", "124.0.6367.243", "124.0.6367.233",
			"124.0.6367.230", "124.0.6367.221", "124.0.6367.208",
			"124.0.6367.201", "124.0.6367.118", "123.0.6358.132",
			"123.0.6358.121", "122.0.6345.98", "122.0.6345.67",
		},
		os: []string{
			"Windows NT 10.0; Win64; x64",
			"Macintosh; Intel Mac OS X 10_15_7",
			"X11; Linux x86_64", "Windows NT 6.1; Win64; x64",
			"Macintosh; Intel Mac OS X 10_14_6",
		},
	}
}

// Generate creates a random User-Agent string for Chrome browsers.
func (a *ChromeAgent) Generate() (userAgent string) {
	var (
		v = a.versions[a.rand(len(a.versions))]
		o = a.os[a.rand(len(a.os))]
	)
	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", o, v)
}

// rand generates random number.
func (a *ChromeAgent) rand(number int) (result int) {
	var (
		value *big.Int
		err   error
	)

	if value, err = rand.Int(rand.Reader, big.NewInt(int64(number))); err != nil {
		return number
	}
	return int(value.Int64())
}
