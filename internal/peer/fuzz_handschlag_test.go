package peer_test

// **Der Handschlag als Folge, mit einer Seite in fremder Hand** — QA-06c
// (#1976),
// Flächen 35, 36 und 37 aus `docs/FUZZZIELE.md`: Angebot, Annahme, Abschluss.
//
// # Was es neben `FuzzAuspacke` (34) zu sehen gibt
//
// **Den Zustand dazwischen.** `FuzzAuspacke` prüft die Hülle allein; hier
// steht die ehrliche Gegenseite mitten in ihrem Handschlag, und an **einer**
// Stelle kommt statt der ehrlichen Nachricht eine beliebige an.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **angenommen wird ein anderer Inhalt als die ehrliche Nachricht** —
//     verglichen wird die gelesene Hülle in kanonischer Schreibung
//     (`gleicherInhalt`, QA-FZ01)
//
// **Hier stand „Byte für Byte", und der erste Lauf zeigte, dass das zu eng
// war**: die ehrliche Nachricht mit einem Leerzeichen dahinter ging durch.
// Das ist kein Fund — die Unterschrift gilt der Form, die aus den gelesenen
// Feldern entsteht ([peer.Huelle] `unterschriftsform`), nicht den Bytes, und
// Leerraum kennt JSON nicht. Ein anderer Inhalt dagegen ändert die Form.
//   - **dasselbe ehrliche Angebot ein zweites Mal** gegen dieselbe Liste der
//     Zufallszahlen angenommen
//   - nach dem Abschluss **zwei Seiten mit verschiedenen Sitzungen**
//
//	go test ./internal/peer -run Fuzz -fuzz FuzzHandschlag -fuzztime 2m -fuzzminimizetime 5s

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// gleicherInhalt sagt, ob zwei Nachrichten dieselbe Hülle tragen — **den
// Inhalt, nicht seine Schreibung** (QA-FZ01, #2091). Die Unterschrift gilt
// den dekodierten Bytes ([peer.Huelle.Pruefe] liest sie mit
// `hex.DecodeString`, das groß wie klein nimmt), und der Rumpf zählt in
// seiner verdichteten Form. Bis zum 2026-10-06 verglich es die Zeichenkette,
// und der Fund `d7e6246289f77b4d` — ein großes `D` — war ein falscher.
func gleicherInhalt(fremd, ehrlich []byte) bool {
	gelesen, fehler := peer.Auspacke(fremd)
	if fehler != nil {
		return false
	}

	richtig, fehler := peer.Auspacke(ehrlich)

	return fehler == nil && reflect.DeepEqual(inhalt(gelesen), inhalt(richtig))
}

// inhalt gibt eine Hülle in ihrer kanonischen Schreibung.
func inhalt(h peer.Huelle) peer.Huelle {
	if roh, fehler := hex.DecodeString(h.Unterschrift); fehler == nil {
		h.Unterschrift = hex.EncodeToString(roh)
	}

	var verdichtet bytes.Buffer
	if json.Compact(&verdichtet, h.Rumpf) == nil {
		h.Rumpf = verdichtet.Bytes()
	}

	return h
}

// liste gibt eine frische Liste der Zufallszahlen mit fester Uhr.
func liste() *peer.Nonceliste {
	jetzt := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

	return peer.NeueNonceliste(func() time.Time { return jetzt }, time.Hour)
}

func wunschUndBereitschaft() (peer.Wunsch, peer.Bereitschaft) {
	return peer.Wunsch{Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1", Name: "A"},
		peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1", Name: "B"}
}

// beide gibt zwei Identitäten mit festem Schlüssel.
func beide(t testing.TB) (ich, du peer.Identitaet) {
	t.Helper()

	ich, fehler := peer.Erzeugen(&festerZufall{stand: 1})
	if fehler != nil {
		t.Fatal(fehler)
	}

	du, fehler = peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatal(fehler)
	}

	return ich, du
}

// saatnachrichten fährt einen ehrlichen Handschlag und gibt seine drei
// Nachrichten — nur als Saat; jeder Lauf rechnet seine eigenen.
func saatnachrichten(t testing.TB) [][]byte {
	t.Helper()

	ich, du := beide(t)
	wunsch, bereit := wunschUndBereitschaft()

	anbahnung, angebot, fehler := ich.Anbieten(wunsch, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, annahme, fehler := du.Entgegennehmen(angebot, bereit, &festerZufall{stand: 200}, liste())
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, abschluss, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return [][]byte{angebot, annahme, abschluss}
}

func FuzzHandschlag(f *testing.F) {
	for schritt, nachricht := range saatnachrichten(f) {
		f.Add(byte(schritt), nachricht)
		f.Add(byte(schritt), nachricht[:len(nachricht)/2])
		f.Add(byte(schritt), append(bytes.Clone(nachricht), ' '))
	}

	f.Fuzz(func(t *testing.T, schritt byte, fremd []byte) {
		ich, du := beide(t)
		wunsch, bereit := wunschUndBereitschaft()

		anbahnung, angebot, fehler := ich.Anbieten(wunsch, &festerZufall{stand: 7})
		if fehler != nil {
			t.Fatal(fehler)
		}

		// **Die ehrliche Gegenseite dieses Laufs** — gegen sie wird
		// verglichen, nicht gegen die Saat.
		entgegennahme, annahme, fehler := du.Entgegennehmen(angebot, bereit, &festerZufall{stand: 200}, liste())
		if fehler != nil {
			t.Fatal(fehler)
		}

		switch schritt % 3 {
		case 0: // das Angebot kommt aus fremder Hand
			eigene := liste()

			_, _, fremdfehler := du.Entgegennehmen(fremd, bereit, &festerZufall{stand: 200}, eigene)
			if fremdfehler == nil && !gleicherInhalt(fremd, angebot) {
				t.Fatalf("ein fremdes Angebot wurde angenommen: %q", fremd)
			}

			if fremdfehler == nil {
				if _, _, nochmal := du.Entgegennehmen(fremd, bereit, &festerZufall{stand: 201}, eigene); nochmal == nil {
					t.Fatal("dasselbe Angebot wurde gegen dieselbe Liste zweimal angenommen")
				}
			}

		case 1: // die Annahme kommt aus fremder Hand
			if _, _, fremdfehler := anbahnung.Schliessen(fremd); fremdfehler == nil && !gleicherInhalt(fremd, annahme) {
				t.Fatalf("eine fremde Annahme wurde geschlossen: %q", fremd)
			}

		default: // der Abschluss kommt aus fremder Hand
			meine, abschluss, schlussfehler := anbahnung.Schliessen(annahme)
			if schlussfehler != nil {
				t.Fatal(schlussfehler)
			}

			deine, abschlussfehler := entgegennahme.Abschliessen(fremd)
			if abschlussfehler == nil && !gleicherInhalt(fremd, abschluss) {
				t.Fatalf("ein fremder Abschluss wurde angenommen: %q", fremd)
			}

			if abschlussfehler == nil && deine.Kennung != meine.Kennung {
				t.Fatalf("zwei Sitzungen: %q und %q", meine.Kennung, deine.Kennung)
			}
		}
	})
}
