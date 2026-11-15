package mtgapi_test

// **Eine Kartenantwort aus fremder Hand** — Werkstück A19 von Phase 5.3
// (#1550).
//
// # Warum
//
// **`mtgapi` liest, was ein anderer geschrieben hat.** Eine Antwort von
// magicthegathering.io ist nicht weniger fremd als eine Datei von der Platte;
// sie kommt nur schneller und öfter. §16 verlangt „Pflichtfelder streng,
// Zusatzfelder tolerant" — und was bei einer Gestalt geschieht, die weder das
// eine noch das andere ist, sagt erst ein Fuzzer.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **zweimal dieselbe Eingabe, zwei Ergebnisse** — auf dem Erfolgs- **und**
//     auf dem Ablehnungsweg. Der Nachlauf über Phase 4.2
//     (#1289)
//     hat an vier Zielen gefunden, dass nur der Erfolgsweg geprüft war — und
//     die eingebaute Prüfung der Ablehnung legte sofort einen echten Fehler
//     frei
//   - **Karten und ein Fehler zugleich**, oder ein Fehler ohne beides
//   - **ein Befund ohne Ort oder ohne Meldung** — ein Befund, der nicht sagt,
//     wo er herkommt, ist keiner (§8)
//
//	go test ./internal/mtgapi -run Fuzz -fuzz FuzzLiesKarten -fuzztime 2m

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
)

// mtgapisaat sind die Gestalten, mit denen der Fuzzer beginnt.
//
// **Aus der Form, die der Dienst wirklich hat** — dazu die Fälle, die weh
// tun: fehlende Pflichtfelder, falscher Typ, extreme Zahl, leere Liste,
// `null`.
var mtgapisaat = []string{
	`{}`,
	`{"cards":[]}`,
	`{"cards":null}`,
	`{"cards":[{}]}`,
	`{"cards":[{"name":"A","set":"PRB","id":"i1"}]}`,
	`{"cards":[{"name":"A","set":"PRB","id":"i1","cmc":1e309}]}`,
	`{"cards":[{"name":123,"set":"PRB","id":"i1"}]}`,
	`{"cards":[{"name":"Ω 🜂","set":"PRB","id":"i1","text":"a\nb"}]}`,
	`{"cards":[{"name":"A","set":"PRB","id":"i1","colors":"Red"}]}`,
	`{"cards":[{"name":"A","set":"PRB","id":"i1","legalities":[{"format":"Standard"}]}]}`,
	`[]`,
	`null`,
	``,
}

// **Der Leser bricht nicht ab, und jeder Befund sagt, wo er herkommt.**
func FuzzLiesKarten(f *testing.F) {
	for _, saat := range mtgapisaat {
		f.Add([]byte(saat))
	}

	// **Und die eingecheckten Beispiele dazu** — sie sind die Gestalt, die
	// dieses Paket wirklich liest, und der beste Ausgangspunkt für den
	// Mutator.
	for _, datei := range []string{"karte.json", "kreatur.json"} {
		// **Der Pfad kommt aus einer festen Liste in dieser Datei** und
		// nicht von außen — `gosec` sieht nur die Veränderliche.
		roh, fehler := os.ReadFile(filepath.Join("testdaten", datei)) //nolint:gosec // eingecheckte Beispiele
		if fehler != nil {
			f.Fatalf("%s lesen: %v", datei, fehler)
		}

		f.Add(roh)
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		karten, befunde, fehler := mtgapi.LiesKarten(roh)

		if fehler != nil {
			if karten != nil {
				t.Errorf("abgelehnt **und** Karten zurück: %v", fehler)
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("der Fehler trägt keine Klasse: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein.**
			nochmal, _, zweiter := mtgapi.LiesKarten(roh)
			if zweiter == nil || nochmal != nil {
				t.Errorf("zweimal dieselbe Eingabe: erst %v, dann %v",
					fehler, zweiter)
			}

			return
		}

		for _, befund := range befunde {
			// **`Ort` darf leer sein und heißt dann »die ganze Datei«**
			// ([domain.Befund]); `Datei` und `Meldung` dürfen es nicht: ein
			// Befund ohne Herkunft ist keiner, und einer ohne Satz sagt
			// nichts (§8).
			if befund.Datei == "" || befund.Meldung == "" {
				t.Errorf("ein Befund ohne Datei oder Meldung: %+v", befund)
			}
		}

		// **Und zweimal dasselbe auf dem Erfolgsweg.**
		nochmal, _, zweiter := mtgapi.LiesKarten(bytes.Clone(roh))
		if zweiter != nil || len(nochmal) != len(karten) {
			t.Errorf("zweimal dieselbe Eingabe: %d Karten, dann %d (%v)",
				len(karten), len(nochmal), zweiter)
		}
	})
}
