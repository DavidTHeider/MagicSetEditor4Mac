package peer_test

// **Was ein Gegenüber in einer Hülle schickt** — QA-06
// (#1737),
// Flächen 34 und 38 aus `docs/FUZZZIELE.md`.
//
// # Warum
//
// **Jede Nachricht zwischen zwei Programmen steckt in einer Hülle**, und
// `Auspacke` ist das Erste, was sie liest: Länge, JSON, Kopf, Unterschrift.
// Was hier durchgeht, gilt als von seinem Absender gesagt.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **eine Hülle über der Höchstlänge, die gelesen wird**
//   - **eine angenommene Hülle, deren Unterschrift nicht zu ihrem Absender
//     passt** — dann hätte jemand im Namen eines anderen gesprochen
//   - **eine verfälschte Hülle, die durchgeht**: ein Byte am Rumpf einer
//     gültigen Hülle, und sie muss abgewiesen werden
//   - **ein Fingerabdruck für eine Kennung, die sich nicht lesen lässt**, oder
//     ein Panic beim Versuch
//
//	go test ./internal/peer -run Fuzz -fuzz FuzzAuspacke -fuzztime 2m

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

func FuzzAuspacke(f *testing.F) {
	ich, fehler := peer.Erzeugen(rand.Reader)
	if fehler != nil {
		f.Fatal(fehler)
	}

	gueltig, fehler := ich.Verpacke(peer.Huelle{
		Protokoll: peer.ProtokollPaarung, Art: "einloesung", ID: "n-1", Sitzung: "s-1",
		Absender: string(ich.Kennung()), Rumpf: []byte(`{"fassung":1}`),
	})
	if fehler != nil {
		f.Fatal(fehler)
	}

	f.Add(gueltig, 0)
	f.Add(gueltig, len(gueltig)/2)
	f.Add([]byte(`{}`), 0)
	f.Add([]byte(`{"protokoll":"unbekannt"}`), 0)
	f.Add([]byte(`null`), 0)
	f.Add(bytes.Repeat([]byte(" "), peer.Hoechstlaenge+1), 0)

	f.Fuzz(func(t *testing.T, roh []byte, stelle int) {
		huelle, fehler := peer.Auspacke(roh)
		zweite, zweiter := peer.Auspacke(roh)

		if (fehler == nil) != (zweiter == nil) || string(huelle.Rumpf) != string(zweite.Rumpf) ||
			huelle.Absender != zweite.Absender {
			t.Fatalf("dieselbe Hülle zweimal verschieden: %v / %v", fehler, zweiter)
		}

		if fehler != nil {
			return
		}

		if len(roh) > peer.Hoechstlaenge {
			t.Errorf("%d Byte über der Höchstlänge und trotzdem gelesen", len(roh))
		}

		if huelle.Pruefe() != nil {
			t.Errorf("angenommen, und die eigene Prüfung weist ab: %v", huelle.Pruefe())
		}

		if peer.Fingerabdruck(peer.Kennung(huelle.Absender)) == "" {
			t.Errorf("angenommen von einem Absender ohne lesbare Kennung: %q", huelle.Absender)
		}

		// **Ein Byte daneben, und sie gilt nicht mehr** — außer das Byte
		// trifft Leerraum zwischen den Feldern, den JSON nicht unterscheidet.
		if len(roh) == 0 {
			return
		}

		verfaelscht := bytes.Clone(roh)
		stelle = ((stelle % len(roh)) + len(roh)) % len(roh)
		verfaelscht[stelle] ^= 0x01

		if andere, ok := peer.Auspacke(verfaelscht); ok == nil &&
			(string(andere.Rumpf) != string(huelle.Rumpf) || andere.Absender != huelle.Absender ||
				andere.Art != huelle.Art || andere.ID != huelle.ID || andere.Folge != huelle.Folge ||
				andere.Sitzung != huelle.Sitzung || andere.Protokoll != huelle.Protokoll) {
			t.Errorf("eine verfälschte Hülle ging mit anderem Inhalt durch (Byte %d)", stelle)
		}
	})
}

// **Ein Fingerabdruck gibt es nur für eine lesbare Kennung — und nie einen
// Panic.**
func FuzzFingerabdruck(f *testing.F) {
	ich, fehler := peer.Erzeugen(rand.Reader)
	if fehler != nil {
		f.Fatal(fehler)
	}

	f.Add(string(ich.Kennung()))
	f.Add("")
	f.Add("b")
	f.Add("bafzaa")

	f.Fuzz(func(t *testing.T, kennung string) {
		abdruck := peer.Fingerabdruck(peer.Kennung(kennung))
		if abdruck != peer.Fingerabdruck(peer.Kennung(kennung)) {
			t.Fatal("dieselbe Kennung, zwei Fingerabdrücke")
		}

		if abdruck == "" {
			return
		}

		// **Gelesen heißt: es gibt einen Schlüssel, der unterschreiben könnte**
		// — eine Unterschrift von irgendwem prüft gegen ihn und scheitert.
		if peer.Pruefe(peer.Kennung(kennung), []byte("daten"), ich.Signiere([]byte("daten"))) == nil &&
			kennung != string(ich.Kennung()) {
			t.Errorf("die Unterschrift eines anderen gilt für %q", kennung)
		}
	})
}
