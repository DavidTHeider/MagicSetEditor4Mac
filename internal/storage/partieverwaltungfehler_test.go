package storage_test

// **Die Fehlerwege der Partieverwaltung** — PLAY-03.AK-05
// (#1711).
//
// # Warum eine eigene Datei
//
// **Die Zeilengrenze** (§2), und der Schnitt trägt: `partieverwaltung_test.go`
// beantwortet, was **geschieht**; hier steht, was passiert, wenn es nicht
// geht. Der Zählfehler steht noch einmal woanders — er braucht den
// Stolpertreiber und damit einen internen Lauf.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Profil geht nichts** — der Bestand ist je Profil getrennt, und eine
// Löschung ohne Profilangabe träfe entweder nichts oder das Falsche.
func TestDieVerwaltungBrauchtEinProfil(t *testing.T) {
	ablage := partieablage(t)

	if fehler := ablage.Loeschen(t.Context(), "p1"); fehler == nil {
		t.Error("ohne Profil wurde gelöscht")
	}
}

// **Eine geschlossene Datenbank klagt** — und der Satz sagt, wobei.
//
// **Zwei Wege und nicht einer**: das Löschen schreibt, das Zählen liest, und
// beide haben ihren eigenen Fehlerweg (§10).
func TestDieVerwaltungMeldetDieGeschlosseneDatenbank(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePartieablage(db)
	ctx := probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	// **Geprüft wird der Satz und nicht die Ursache** (§8): was ein
	// geschlossener Bestand meldet, ist die Sache von `database/sql`; was
	// **dieses** Programm zusagt, ist die Auskunft, **wobei** es geschah.
	fehler := ablage.Loeschen(ctx, "p1")
	if fehler == nil || !strings.Contains(fehler.Error(), "die Partie löschen") {
		t.Errorf("das Löschen meldete %v", fehler)
	}

	// **Das Zählen scheitert seit #1815 schon an der Frage nach dem Besitz**
	// — und sagt das. Der eigene Fehlerweg des Zählens steht im
	// Stolperlauf darunter.
	_, fehler = ablage.Offenlegungszahl(ctx, "p1")
	if fehler == nil ||
		!strings.Contains(fehler.Error(), "den Besitz") {
		t.Errorf("das Zählen meldete %v", fehler)
	}
}
