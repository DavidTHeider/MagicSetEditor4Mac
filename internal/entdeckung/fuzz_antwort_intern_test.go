package entdeckung

// **Was ein Gerät im selben Netz auf die Suche antwortet** — QA-06
// (#1737),
// Fläche 48 aus `docs/FUZZZIELE.md`.
//
// # Warum im Paket
//
// **`deute` ist nicht exportiert**, und eine Antwort mit krummen Feldern kommt
// aus keinem echten Netz — außer von jemandem, der es darauf anlegt. Dieselbe
// Lage wie bei `deutung_intern_test.go`.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **ein angenommener Fund ohne lesbare Kennung, ohne Port oder ohne
//     Wirt** — dorthin ließe sich nicht wählen
//   - **ein angenommener Name, den ein Mensch nicht lesen kann**: ungültiges
//     UTF-8, mehr als 64 Zeichen, ein Steuer- oder Formatzeichen — der Name
//     steht in der Liste „wer ist da", und ein U+202E verdrehte, was dort
//     steht. **Dieselbe Regel wie beim Namen einer Einladung**
//     (`domain.pruefeEinladungsname`)
//
//	go test ./internal/entdeckung -run Fuzz -fuzz FuzzDeute -fuzztime 2m

import (
	"crypto/rand"
	"net"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/mdns"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

func FuzzDeute(f *testing.F) {
	ich, fehler := peer.Erzeugen(rand.Reader)
	if fehler != nil {
		f.Fatal(fehler)
	}

	kennung := kennungsmarke + string(ich.Kennung())

	f.Add(kennung+"\x00"+namensmarke+"Anna", "anna.local.", 41234, []byte{192, 0, 2, 1})
	f.Add(kennung+"\x00"+namensmarke+"A\u202Enna", "anna.local.", 41234, []byte{192, 0, 2, 1})
	f.Add(kennung+"\x00"+namensmarke+"\x07", "", 1, []byte{})
	f.Add(kennungsmarke+"kaputt", "x.local.", 0, []byte{10, 0, 0, 1})
	f.Add("", "", 0, []byte{})

	f.Fuzz(func(t *testing.T, felder, host string, port int, adresse []byte) {
		antwort := &mdns.ServiceEntry{
			Name: "fuzz", Host: host, Port: port, InfoFields: strings.Split(felder, "\x00"),
		}

		if len(adresse) == net.IPv4len {
			antwort.AddrV4 = net.IP(adresse)
		}

		fund, gilt := deute(antwort)
		if !gilt {
			return
		}

		if peer.Fingerabdruck(fund.Kennung) == "" || fund.Port < 1 || fund.Wirt == "" {
			t.Errorf("angenommen ohne Kennung, Port oder Wirt: %+v", fund)
		}

		if fund.Name == "" {
			return
		}

		if !utf8.ValidString(fund.Name) || utf8.RuneCountInString(fund.Name) > 64 {
			t.Errorf("ein Name, der keiner ist: %q", fund.Name)
		}

		for _, zeichen := range fund.Name {
			if unicode.IsControl(zeichen) || unicode.Is(unicode.Cf, zeichen) {
				t.Errorf("angenommen mit %U im Namen %q", zeichen, fund.Name)
			}
		}
	})
}
