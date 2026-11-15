package peer_test

// **Die Nachrichtenhülle** — Werkstück N1 von Phase 5.4
// (#1454),
// zweiter Schnitt.
//
// # Wogegen hier geprüft wird
//
// **Dass eine Unterschrift den Kopf mitmeint.** Wer nur den Rumpf
// unterschriebe, ließe jede Nachricht in ein anderes Protokoll, eine andere
// Sitzung oder unter einen anderen Absender verschieben, ohne dass die
// Unterschrift bricht.
//
// **Und dass zwei verschiedene Köpfe nie dieselbe Unterschrift bekommen.** Ein
// Kopf, dessen Felder einfach aneinandergehängt würden, hat diese Eigenschaft
// **nicht**: `Sitzung="ab" Art="c"` und `Sitzung="a" Art="bc"` ergäben
// dieselbe Zeichenkette. Genau das prüft der Lauf über die verschobene Grenze.
//
// **Und dass zu Großes und Unbekanntes vor der Fachausführung endet.** Das
// Issue verlangt es wörtlich: „unzulässige/zu große Typen vor Fachausführung
// abweisen".

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineHuelle gibt eine vollständige Hülle.
func eineHuelle(anders func(*peer.Huelle)) peer.Huelle {
	huelle := peer.Huelle{
		Protokoll: peer.ProtokollPaarung,
		Art:       "hallo",
		ID:        "n-1",
		Sitzung:   "s-1",
		Absender:  "bwer",
		Folge:     1,
		Rumpf:     json.RawMessage(`{"was":"wert"}`),
	}

	if anders != nil {
		anders(&huelle)
	}

	return huelle
}

// **Eine Hülle geht hinaus und kommt als dieselbe zurück.**
func TestEineHuelleGehtHinausUndKommtZurueck(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	roh, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
	}))
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	zurueck, fehler := peer.Auspacke(roh)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	if zurueck.Art != "hallo" || zurueck.Sitzung != "s-1" || zurueck.Folge != 1 {
		t.Fatalf("zurück kam %+v", zurueck)
	}

	// **Der Rumpf kommt Byte für Byte zurück** — unterschrieben sind die
	// Bytes und nicht ihre Bedeutung; ein zweites Kodieren wäre eine zweite
	// Gestalt.
	if string(zurueck.Rumpf) != `{"was":"wert"}` {
		t.Errorf("der Rumpf ist %q", zurueck.Rumpf)
	}
}

// **Was am Kopf geändert wird, bricht die Unterschrift.**
//
// **Jedes Feld einzeln** — ein Lauf über nur eines ließe die übrigen
// ungeprüft, und genau eines davon wäre dann verschiebbar.
func TestJedeAenderungAmKopfBrichtDieUnterschrift(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	faelle := map[string]func(*peer.Huelle){
		"anderes Protokoll": func(h *peer.Huelle) { h.Protokoll = peer.ProtokollSpiel },
		"andere Art":        func(h *peer.Huelle) { h.Art = "tschuess" },
		"andere ID":         func(h *peer.Huelle) { h.ID = "n-2" },
		"andere Sitzung":    func(h *peer.Huelle) { h.Sitzung = "s-2" },
		"andere Folge":      func(h *peer.Huelle) { h.Folge = 2 },
		"anderer Rumpf":     func(h *peer.Huelle) { h.Rumpf = json.RawMessage(`{"was":"anders"}`) },
	}

	for name, aendere := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			roh, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
				h.Absender = string(ich.Kennung())
			}))
			if fehler != nil {
				t.Fatalf("Verpacke: %v", fehler)
			}

			zurueck, fehler := peer.Auspacke(roh)
			if fehler != nil {
				t.Fatalf("Auspacke: %v", fehler)
			}

			aendere(&zurueck)

			if fehler := zurueck.Pruefe(); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("die geänderte Hülle kam durch: %v", fehler)
			}
		})
	}
}

// **Zwei Köpfe, die aneinandergehängt gleich hießen, haben zwei
// Unterschriften.**
//
// # Der Fund
//
// **Ohne Längen vor den Feldern wären sie eine.** `Art="ab"`, `ID="c"` und
// `Art="a"`, `ID="bc"` ergäben denselben Text — und wer eine Nachricht
// abfängt, verschöbe die Grenze und behielte die Unterschrift.
//
// **Und die zwei Felder grenzen in der Unterschriftsform aneinander.** Der
// erste Anlauf nahm `Sitzung` und `Art`; zwischen ihnen steht die ID, ihre
// Verkettungen waren ohnehin verschieden, und die Gegenprobe blieb grün — ein
// Lauf, der seinen eigenen Namen nicht einlöst (§10).
func TestZweiKoepfeMitVerschobenerGrenzeHabenZweiUnterschriften(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	erste, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
		h.Art = "ab"
		h.ID = "c"
	}))
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	zweite, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
		h.Art = "a"
		h.ID = "bc"
	}))
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	auf1, fehler := peer.Auspacke(erste)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	auf2, fehler := peer.Auspacke(zweite)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	if auf1.Unterschrift == auf2.Unterschrift {
		t.Fatal("zwei Köpfe teilen eine Unterschrift")
	}

	// **Und die getauschte Unterschrift gilt an keiner von beiden.**
	auf1.Unterschrift = auf2.Unterschrift

	if fehler := auf1.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die getauschte Unterschrift kam durch: %v", fehler)
	}
}

// **Ein fremder Absender prüft nicht.**
//
// **Die Kennung im Kopf ist der Schlüssel** — wer sie austauscht, tauscht den
// Prüfschlüssel mit, und dann bricht die Unterschrift.
func TestEinFremderAbsenderPrueftNicht(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	andere, fehler := peer.Erzeugen(&festerZufall{stand: 42})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	roh, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
	}))
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	zurueck, fehler := peer.Auspacke(roh)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	zurueck.Absender = string(andere.Kennung())

	if fehler := zurueck.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Absender kam durch: %v", fehler)
	}
}

// **Und wer unter fremdem Namen verpackt, wird beim Verpacken abgewiesen.**
//
// **Nicht erst beim Prüfen**: eine Nachricht, die unterschrieben hinausgeht
// und nirgends gilt, ist ein Fehler, der beim Absender entsteht.
func TestWerUnterFremdemNamenVerpacktWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = "bfremd"
	})); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Name kam durch: %v", fehler)
	}
}

// **Eine unterwegs geänderte Nachricht kommt nicht durch [peer.Auspacke].**
//
// # Warum das ein eigener Lauf ist
//
// **Weil ein Empfänger nicht [peer.Huelle.Pruefe] ruft, sondern
// [peer.Auspacke].** Die Läufe darüber greifen in die schon ausgepackte
// Struktur; dieser nimmt den Weg, den eine Nachricht wirklich geht — Bytes
// von der Leitung, und der Fälscher hat nur sie in der Hand.
func TestEineVerfaelschteNachrichtKommtNichtDurchAuspacke(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	roh, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
	}))
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	// **Gleich lang und gültiges JSON** — die Fälschung soll am Zerteiler
	// vorbeikommen und erst an der Unterschrift enden. Sonst maße der Lauf
	// wieder etwas anderes als seinen Namen (§10).
	gefaelscht := bytes.Replace(roh, []byte(`"wert"`), []byte(`"Wert"`), 1)

	if bytes.Equal(gefaelscht, roh) {
		t.Fatal("die Fälschung hat nichts geändert")
	}

	if _, fehler := peer.Auspacke(gefaelscht); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die gefälschte Nachricht kam durch: %v", fehler)
	}
}
