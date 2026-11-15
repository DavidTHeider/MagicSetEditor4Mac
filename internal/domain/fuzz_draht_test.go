package domain_test

// **Was über den Draht und durch einen Messenger kommt** — QA-06
// (#1737),
// Flächen 42, 43, 45 und 47 aus `docs/FUZZZIELE.md`.
//
// # Warum diese vier
//
// **Sie liest ein Fremder zuerst.** Ein Einladungscode kommt per Nachricht,
// eine Einlösung und ihre Antwort über die Paarungssitzung, eine
// tailcat-Einladung und eine Adresse von Hand oder aus einer Einladung — und
// keines davon hat vorher ein Mensch gesehen, der wüsste, was er prüft.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse — auf **beiden** Wegen
//   - **angenommen und nicht wieder herstellbar**: was zerlegt wurde, muss
//     sich so kodieren lassen, dass dasselbe herauskommt (Einlösung, Antwort,
//     Einladung, Adresse)
//   - **ein angenommener Wert außerhalb seiner Menge**: ein Ergebnis, das es
//     nicht gibt, eine Fassung über der heutigen, eine leere Adresse, ein
//     Port außerhalb 1–65535
//   - **eine Einlösung über der Grenze**, die trotzdem gelesen wird
//
//	go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeEinloesung -fuzztime 2m

import (
	"errors"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Einlösung lässt sich wieder so kodieren, wie sie gelesen wurde.**
func FuzzZerlegeEinloesung(f *testing.F) {
	for _, saat := range []string{
		`{"fassung":1,"nonce":"` + strings.Repeat("ab", 16) + `","adresse":"192.0.2.1:7000","name":"Bert"}`,
		`{"fassung":1,"nonce":"x"}`, `{"fassung":2,"nonce":"x"}`, `{"fassung":0}`,
		`{"fassung":1,"nonce":"x","mehr":1}`, `null`, `[]`, ``, `{"fassung":"1"}`,
		`{"fassung":1,"nonce":"x","name":"` + strings.Repeat("ä", 3000) + `"}`,
	} {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		einloesung, fehler := domain.ZerlegeEinloesung(roh)
		nochmal, zweiter := domain.ZerlegeEinloesung(roh)

		if (fehler == nil) != (zweiter == nil) || !reflect.DeepEqual(einloesung, nochmal) {
			t.Fatalf("dieselbe Einlösung zweimal verschieden: %v / %v", fehler, zweiter)
		}

		if fehler != nil {
			if fehler.Error() != zweiter.Error() {
				t.Errorf("zwei Meldungen: %q gegen %q", fehler, zweiter)
			}

			return
		}

		if len(roh) > 4<<10 {
			t.Errorf("%d Byte über der Grenze und trotzdem gelesen", len(roh))
		}

		if einloesung.Fassung < 1 || einloesung.Fassung > domain.Einloesefassung {
			t.Errorf("angenommen mit Fassung %d", einloesung.Fassung)
		}

		kodiert, kodierfehler := einloesung.Kodiere()
		if kodierfehler != nil {
			t.Fatalf("angenommen und nicht kodierbar: %v", kodierfehler)
		}

		zurueck, rueckfehler := domain.ZerlegeEinloesung(kodiert)
		if rueckfehler != nil || !reflect.DeepEqual(zurueck, einloesung) {
			t.Errorf("Zerlegen∘Kodieren ist nicht die Identität: %+v gegen %+v (%v)",
				zurueck, einloesung, rueckfehler)
		}
	})
}

// **Eine Antwort trägt ein Ergebnis aus der Menge — und kommt zurück, wie sie
// ging.**
func FuzzZerlegeEinloeseantwort(f *testing.F) {
	for _, saat := range []string{
		`{"fassung":1,"ergebnis":"aufgenommen"}`, `{"fassung":1,"ergebnis":"fremd"}`,
		`{"fassung":1,"ergebnis":"erfunden"}`, `{"fassung":9,"ergebnis":"fassung"}`,
		`{"fassung":1}`, `{}`, `null`, ``,
	} {
		f.Add([]byte(saat))
	}

	bekannt := map[domain.Einloeseergebnis]bool{
		domain.EinloesungAufgenommen: true, domain.EinloesungFremd: true,
		domain.EinloesungAbgelaufen: true, domain.EinloesungVerbraucht: true,
		domain.EinloesungGesperrt: true, domain.EinloesungGedrosselt: true,
		domain.EinloesungZuNeu: true, domain.EinloesungGescheitert: true,
		domain.EinloesungUnbeantwortet: true,
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		antwort, fehler := domain.ZerlegeEinloeseantwort(roh)
		nochmal, zweiter := domain.ZerlegeEinloeseantwort(roh)

		if (fehler == nil) != (zweiter == nil) || antwort != nochmal {
			t.Fatalf("dieselbe Antwort zweimal verschieden: %v / %v", fehler, zweiter)
		}

		if fehler != nil {
			return
		}

		if !bekannt[antwort.Ergebnis] {
			t.Errorf("angenommen mit einem Ergebnis, das es nicht gibt: %q", antwort.Ergebnis)
		}

		if zurueck, rueckfehler := domain.ZerlegeEinloeseantwort(antwort.Kodiere()); rueckfehler != nil || zurueck != antwort {
			t.Errorf("Zerlegen∘Kodieren ist nicht die Identität: %+v gegen %+v (%v)", zurueck, antwort, rueckfehler)
		}
	})
}

// **Jede Adresse kommt durch eine Einladung unverstümmelt zurück**, und was
// keine Einladung ist, wird mit einem Satz abgewiesen.
func FuzzZerlegeEinladung(f *testing.F) {
	for _, saat := range []string{
		"mse4mac:1:abc", "mse4mac:2:abc", "mse4mac:0:abc", "mse4mac:1:", "mse4mac:x:abc",
		"mse4mac:1:a:b:c", "anders:1:abc", "", ":", "mse4mac:+1:abc",
	} {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, text string) {
		adresse, fehler := domain.ZerlegeEinladung(domain.Einladung(text))

		if fehler == nil && adresse == "" {
			t.Errorf("%q angenommen mit leerer Adresse", text)
		}

		if fehler != nil && !errors.Is(fehler, domain.ErrValidation) && !errors.Is(fehler, domain.ErrEinladungZuNeu) {
			t.Errorf("%q: eine Ablehnung ohne Klasse: %v", text, fehler)
		}

		// **Die Gegenrichtung für jeden Text als Adresse.**
		if text == "" {
			return
		}

		zurueck, rueckfehler := domain.ZerlegeEinladung(domain.NeueEinladung(text))
		if rueckfehler != nil || zurueck != text {
			t.Errorf("die Adresse %q kam als %q zurück (%v)", text, zurueck, rueckfehler)
		}
	})
}

// **Ein Einladungscode wird gelesen oder benannt abgewiesen — nie halb.**
func FuzzZerlegeKontakteinladung(f *testing.F) {
	for _, saat := range []string{
		"mse4mac-kontakt:1:e30.AA", "mse4mac-kontakt:2:e30.AA", "mse4mac-kontakt:1:e30",
		"mse4mac-kontakt:1:!!.!!", "mse4mac-kontakt:1:.", "  mse4mac-kontakt:1:e30.AA  ", "", "x:y:z",
	} {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, text string) {
		einladung, unterschrieben, fehler := domain.ZerlegeKontakteinladung(text)
		zweite, zweitesUnterschrieben, zweiter := domain.ZerlegeKontakteinladung(text)

		if (fehler == nil) != (zweiter == nil) || !reflect.DeepEqual(einladung, zweite) ||
			!reflect.DeepEqual(unterschrieben, zweitesUnterschrieben) {
			t.Fatalf("derselbe Code zweimal verschieden: %v / %v", fehler, zweiter)
		}

		if fehler != nil {
			if !errors.Is(fehler, domain.ErrValidation) && !errors.Is(fehler, domain.ErrEinladungZuNeu) {
				t.Errorf("eine Ablehnung ohne Klasse: %v", fehler)
			}

			return
		}

		// **Angenommen heißt hier nur: lesbar.** Gelten muss sie erst nach der
		// Unterschrift — aber ohne Daten und ohne Unterschrift gibt es nichts
		// zu prüfen, und dann darf es auch nicht „gelesen" heißen.
		if len(unterschrieben.Daten) == 0 || len(unterschrieben.Unterschrift) == 0 {
			t.Errorf("gelesen ohne Daten oder ohne Unterschrift: %+v", unterschrieben)
		}
	})
}

// **Eine Adresse zerfällt in Wirt und Port, und beide setzen sie wieder
// zusammen.** Zerlegt wird die **Horchform** (`:7000` und Port 0 sind dort
// erlaubt); was `PruefeAdresse` zum Wählen annimmt — weder leer noch eine
// Einladung —, hat einen Wirt und einen Port ab eins.
func FuzzZerlegeAdresse(f *testing.F) {
	for _, saat := range []string{
		"192.0.2.1:7000", "[2001:db8::1]:7000", "2001:db8::1:7000", "wirt:0", "wirt:65536",
		"wirt:1", ":1", "wirt", "", "wirt mit leerraum:1", "a/b:1", strings.Repeat("a", 300) + ":1",
		"mse4mac:1:abc",
	} {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, adresse string) {
		wirt, port, fehler := domain.ZerlegeAdresse(adresse)

		if domain.PruefeAdresse(adresse) == nil && adresse != "" &&
			!domain.IstEinladung(adresse) && (fehler != nil || wirt == "" || port < 1) {
			t.Errorf("%q gilt zum Wählen, zerlegt aber zu %q:%d (%v)", adresse, wirt, port, fehler)
		}

		if fehler != nil {
			return
		}

		if port < 0 || port > 65535 || len(adresse) > 259 {
			t.Errorf("%q angenommen als Wirt %q, Port %d", adresse, wirt, port)
		}

		zusammen := net.JoinHostPort(wirt, strconv.Itoa(port))
		zurueckWirt, zurueckPort, rueckfehler := domain.ZerlegeAdresse(zusammen)

		if rueckfehler != nil || zurueckWirt != wirt || zurueckPort != port {
			t.Errorf("%q → %q:%d → %q gibt %q:%d (%v)", adresse, wirt, port, zusammen, zurueckWirt, zurueckPort, rueckfehler)
		}
	})
}
