package domain_test

// **Was ein Gegenüber am Tisch schickt** — QA-06
// (#1737),
// Flächen 49, 50, 51, 55 und 56 aus `docs/FUZZZIELE.md`.
//
// # Warum diese fünf
//
// **Eine Partie läuft über eine Leitung zu einem Fremden**, und jede dieser
// Prüfungen ist die erste, die eine Nachricht von dort trifft: der
// Kontrollumschlag, die Nachricht selbst, der Stand der Folgen, der Wunsch
// eines Gastes, wieder einzusteigen, und ein Satz am Tisch.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **eine Kontrolle, die nicht wiederkommt, wie sie ging**, oder eine
//     angenommene Art, die nicht erwartet war
//   - **ein Stand, der zurückspringt** oder eine Folge auslässt: nach „neu"
//     steht genau diese Folge, sonst bleibt der Stand
//   - **eine Wiederaufnahme, die weiterspielt**, obwohl der Gast weiter ist als
//     der Wirt, oder die ab einer Folge nachreicht, die der Wirt nicht hat
//   - **ein angenommener Satz mit Steuerzeichen oder einer
//     Richtungsüberschreibung** — er verdreht, was die anderen lesen
//
//	go test ./internal/domain -run Fuzz -fuzz FuzzEmpfaengerstand -fuzztime 2m

import (
	"encoding/json"
	"reflect"
	"testing"
	"unicode"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Kontrolle kommt wieder, wie sie ging — und nur, wenn sie erwartet
// war.**
func FuzzLiesKontrolle(f *testing.F) {
	for _, saat := range []string{
		`{"art":"rueckkehr","inhalt":{}}`, `{"art":"abzug","inhalt":{"art":"weiter"}}`,
		`{"art":"erfunden","inhalt":{}}`, `{"art":"rueckkehr"}`, `{"art":"rueckkehr","inhalt":"text"}`,
		`null`, `[]`, ``,
	} {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		erwartet := []domain.Kontrollart{domain.KontrolleRueckkehr, domain.KontrolleAbzug}

		art, inhalt, fehler := domain.LiesKontrolle(roh, "gast", erwartet...)
		zweiteArt, zweiterInhalt, zweiter := domain.LiesKontrolle(roh, "gast", erwartet...)

		if (fehler == nil) != (zweiter == nil) || art != zweiteArt || string(inhalt) != string(zweiterInhalt) {
			t.Fatalf("dieselbe Kontrolle zweimal verschieden: %v / %v", fehler, zweiter)
		}

		if fehler != nil {
			return
		}

		if art != domain.KontrolleRueckkehr && art != domain.KontrolleAbzug {
			t.Errorf("angenommen mit einer Art, die nicht erwartet war: %q", art)
		}

		var nutzlast any
		if json.Unmarshal(inhalt, &nutzlast) != nil {
			t.Fatalf("angenommen mit einem Inhalt, der kein JSON ist: %q", inhalt)
		}

		verpackt, verpackfehler := domain.VerpackeKontrolle(art, nutzlast)
		if verpackfehler != nil {
			t.Fatalf("angenommen und nicht wieder verpackbar: %v", verpackfehler)
		}

		zurueck, zurueckInhalt, rueckfehler := domain.LiesKontrolle(verpackt, "gast", erwartet...)

		var vorher, nachher any
		_ = json.Unmarshal(inhalt, &vorher)         //nolint:errcheck // oben gelesen
		_ = json.Unmarshal(zurueckInhalt, &nachher) //nolint:errcheck // dito

		if rueckfehler != nil || zurueck != art || !reflect.DeepEqual(vorher, nachher) {
			t.Errorf("Lesen∘Verpacken ist nicht die Identität: %q/%s gegen %q/%s (%v)",
				art, inhalt, zurueck, zurueckInhalt, rueckfehler)
		}
	})
}

// **Eine Partienachricht wird geprüft, bevor sie wirkt — und immer gleich.**
func FuzzPartienachricht(f *testing.F) {
	for _, saat := range []string{
		`{"partie":"p1","epoche":1,"folge":1,"art":"ereignis","inhalt":{},"kette":""}`,
		`{"partie":"p1","epoche":0,"folge":1,"art":"ereignis"}`,
		`{"partie":"p1","epoche":1,"art":"zeiger","inhalt":{}}`,
		`{"partie":"p1","epoche":1,"folge":7,"art":"zeiger"}`, `{}`, `null`, ``,
	} {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		var nachricht domain.Partienachricht
		if json.Unmarshal(roh, &nachricht) != nil {
			return
		}

		fehler, zweiter := nachricht.Pruefe(), nachricht.Pruefe()
		if (fehler == nil) != (zweiter == nil) {
			t.Fatalf("dieselbe Nachricht zweimal verschieden: %v / %v", fehler, zweiter)
		}

		if fehler == nil && nachricht.Epoche == 0 {
			t.Errorf("angenommen ohne Epoche: %+v", nachricht)
		}
	})
}

// **Der Stand der Folgen springt nie zurück und lässt nichts aus**, gleich in
// welcher Reihenfolge Zustellungen kommen.
func FuzzEmpfaengerstand(f *testing.F) {
	f.Add([]byte{1, 1, 1, 2, 1, 3})
	f.Add([]byte{1, 2, 1, 1, 2, 1, 2, 3, 1, 9})
	f.Add([]byte{255, 255, 0, 0, 1, 1})

	f.Fuzz(func(t *testing.T, roh []byte) {
		stand := domain.NeuerEmpfaengerstand(1)
		epoche, folge := uint64(1), uint64(0)

		for stelle := 0; stelle+1 < len(roh); stelle += 2 {
			// **Kleine Zahlen und die Ränder**: ein Byte über 250 heißt der
			// größte Wert — dort laufen Rechnungen über.
			neueEpoche, neueFolge := zahl(roh[stelle]), zahl(roh[stelle+1])

			urteil := stand.Nimm(neueEpoche, neueFolge)

			if urteil == domain.FolgeNeu {
				if neueEpoche < epoche || (neueEpoche == epoche && neueFolge != folge+1) ||
					(neueEpoche > epoche && neueFolge != 1) {
					t.Fatalf("(%d,%d) nach (%d,%d) als neu angenommen", neueEpoche, neueFolge, epoche, folge)
				}

				epoche, folge = neueEpoche, neueFolge
			}

			if stand.Epoche() != epoche {
				t.Fatalf("der Stand steht in Epoche %d, erwartet %d", stand.Epoche(), epoche)
			}
		}
	})
}

func zahl(b byte) uint64 {
	if b > 250 {
		return ^uint64(0) - uint64(255-b)
	}

	return uint64(b)
}

// **Wer weiter ist als der Wirt, spielt nicht weiter, und nachgereicht wird
// nur, was der Wirt hat.**
func FuzzWiederaufnahme(f *testing.F) {
	f.Add(uint64(1), uint64(5), "k5", uint64(1), uint64(9), "k9", uint64(1))
	f.Add(uint64(1), uint64(9), "k9", uint64(1), uint64(9), "k9", uint64(3))
	f.Add(uint64(1), uint64(9), "anders", uint64(1), uint64(9), "k9", uint64(1))
	f.Add(uint64(1), uint64(12), "", uint64(1), uint64(9), "k9", uint64(1))
	f.Add(uint64(2), uint64(1), "", uint64(1), uint64(9), "k9", uint64(5))

	f.Fuzz(func(t *testing.T, gastEpoche, gastFolge uint64, gastKette string,
		wirtEpoche, wirtFolge uint64, wirtKette string, aelteste uint64,
	) {
		wunsch := domain.Wiederaufnahmewunsch{Partie: "p1", Epoche: gastEpoche, Folge: gastFolge, Kette: gastKette}
		stand := domain.Wirtsstand{Epoche: wirtEpoche, Folge: wirtFolge, Kette: wirtKette, AelteteFolge: aelteste}

		weg := domain.Wiederaufnahme(wunsch, stand)
		if !reflect.DeepEqual(weg, domain.Wiederaufnahme(wunsch, stand)) {
			t.Fatal("derselbe Wunsch zweimal verschieden")
		}

		if gastFolge > wirtFolge && weg.Art != domain.WegAbbruch {
			t.Errorf("der Gast steht bei %d, der Wirt bei %d — und es geht weiter (%s)", gastFolge, wirtFolge, weg.Art)
		}

		if weg.Art == domain.WegEreignisse && (weg.AbFolge < 1 || weg.AbFolge > wirtFolge || weg.AbFolge < aelteste) {
			t.Errorf("nachgereicht ab %d, der Wirt hat %d bis %d", weg.AbFolge, aelteste, wirtFolge)
		}
	})
}

// **Ein angenommener Satz trägt kein Steuerzeichen und keine
// Richtungsüberschreibung.**
func FuzzChat(f *testing.F) {
	for _, saat := range []string{"Hallo", "", "  ", "a\u202Eb", "a\u0007b", "👨\u200D👩", "zwei\nZeilen"} {
		f.Add("anna", saat)
	}

	f.Fuzz(func(t *testing.T, von, text string) {
		satz := domain.Chat{Von: von, Text: text}
		if satz.Pruefe() != nil {
			return
		}

		for _, zeichen := range text {
			if unicode.Is(unicode.Cc, zeichen) || richtungsueberschreibung(zeichen) {
				t.Fatalf("angenommen mit %U in %q", zeichen, text)
			}
		}
	})
}

// richtungsueberschreibung sind die Zeichen, die die Leserichtung umstellen —
// U+202A bis U+202E und U+2066 bis U+2069.
func richtungsueberschreibung(zeichen rune) bool {
	return (zeichen >= 0x202A && zeichen <= 0x202E) || (zeichen >= 0x2066 && zeichen <= 0x2069)
}
