package domain_test

// **Worüber im Handel unterschrieben wird** — QA-06
// (#1737),
// Fläche 58 aus `docs/FUZZZIELE.md`.
//
// # Warum
//
// **Eine Unterschrift gilt für eine Form, nicht für eine Nachricht.** Bilden
// zwei verschiedene Umschläge dieselbe Form, gilt die Unterschrift des einen
// für den anderen — der Kopf von `Unterschriftsform` sagt genau das, und der
// Trenner soll es verhindern.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic
//   - **zwei gültige Umschläge mit derselben Form**: ein Nullbyte in einem
//     Feld verschiebt die Grenze zum nächsten, und die Form bleibt gleich
//   - zweimal derselbe Umschlag, zwei Urteile
//
//	go test ./internal/domain -run Fuzz -fuzz FuzzHandelsumschlag -fuzztime 2m

import (
	"bytes"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func FuzzHandelsumschlag(f *testing.F) {
	f.Add(1, "angebot", "n-1", "v-1", 1, "abdruck", int64(1), "bafzaa")
	f.Add(1, "angebot", "n-1\x00v", "-1", 1, "abdruck", int64(1), "bafzaa")
	f.Add(0, "", "", "", 0, "", int64(0), "")

	f.Fuzz(func(t *testing.T, fassung int, art, nachricht, vorgang string,
		basis int, abdruck string, folge int64, absender string,
	) {
		umschlag := domain.Handelsumschlag{
			Fassung: fassung, Art: domain.Handelsnachrichtenart(art), NachrichtID: nachricht,
			Vorgang: vorgang, Basisrevision: basis, Abdruck: abdruck, Folge: folge, Absender: absender,
		}

		befunde := umschlag.Validate()
		if len(befunde) != len(umschlag.Validate()) {
			t.Fatal("derselbe Umschlag, zwei Urteile")
		}

		if len(befunde) > 0 {
			return
		}

		// **Steht ein Nullbyte in der Nachrichtenkennung, wandert der Rest in
		// den Vorgang** — und die Form bleibt dieselbe.
		if stelle := bytes.IndexByte([]byte(nachricht), 0); stelle >= 0 {
			verschoben := umschlag
			verschoben.NachrichtID = nachricht[:stelle]
			verschoben.Vorgang = nachricht[stelle+1:] + "\x00" + vorgang

			if len(verschoben.Validate()) == 0 &&
				bytes.Equal(verschoben.Unterschriftsform(), umschlag.Unterschriftsform()) {
				t.Errorf("zwei gültige Umschläge, eine Form: %q/%q und %q/%q",
					umschlag.NachrichtID, umschlag.Vorgang, verschoben.NachrichtID, verschoben.Vorgang)
			}
		}

		for _, feld := range []string{nachricht, vorgang, abdruck, absender, art} {
			if bytes.IndexByte([]byte(feld), 0) >= 0 {
				t.Errorf("gültig mit einem Nullbyte im Feld %q — dem Trenner der Form", feld)
			}
		}
	})
}
