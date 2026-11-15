package service_test

// **Die Notiz zu einem Gegenüber** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum eine eigene Datei
//
// **Weil `kontaktadresse_test.go` an der Zeilengrenze steht** (§2) — und weil
// der Schnitt der Frage folgt: dort steht, **wohin** gewählt wird, hier, was
// man **über** ein Gegenüber denkt.
//
// # Und warum es keine Prüfung gibt
//
// **Weil eine Notiz Freitext für den ist, der sie schreibt.** Sie geht über
// keine Leitung: eine Sichtkarte trägt sie nicht, eine Freigabe erst recht
// nicht. Es gibt also nichts, wogegen sie zu halten wäre — und das steht hier
// als Lauf, statt als Annahme im Kopf.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Notiz kommt an und steht am Kontakt.**
func TestEineNotizStehtDanachAmKontakt(t *testing.T) {
	t.Parallel()

	doppel := einKontaktdoppel("bkennung-eins")
	werk := service.NeuesKontaktwerk(doppel, einePeeruhr)

	for _, notiz := range []string{"tauscht fair", ""} {
		if fehler := werk.NotizSetzen(
			t.Context(), "bkennung-eins", notiz); fehler != nil {
			t.Fatalf("Notiz: %v", fehler)
		}

		liste, fehler := werk.Liste(t.Context())
		if fehler != nil {
			t.Fatalf("Liste: %v", fehler)
		}

		if len(liste) != 1 || liste[0].Notiz != notiz {
			t.Fatalf("die Liste gibt %+v", liste)
		}
	}
}

// **Und was die Ablage klagt, kommt zurück.**
func TestWasDieAblageBeiDerNotizKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	klage := errors.New("die Platte ist voll")

	doppel := einKontaktdoppel("bkennung-eins")
	doppel.scheitert = klage

	if fehler := service.NeuesKontaktwerk(doppel, einePeeruhr).NotizSetzen(
		t.Context(), "bkennung-eins", "hm"); !errors.Is(fehler, klage) {
		t.Errorf("es kam %v", fehler)
	}
}
