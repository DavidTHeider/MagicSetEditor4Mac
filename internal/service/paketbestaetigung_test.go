// **Die Bestätigung beim Entfernen eines Pakets** —
// #1344.
//
// **Eine eigene Datei**, weil `paketverwalter_test.go` an der Zeilengrenze aus
// §2 stand. Der Schnitt folgt der Frage: dort steht, *was das Verwalten tut*,
// hier, *woran es sich hält* — und der Vergleich ist erst seit diesem Befund
// überhaupt hier.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Entfernen gegen eine überholte Bestätigung schlägt fehl** —
// #1344.
//
// # Der Befund, der diesen Lauf verlangt
//
// `schreiben.paket.entfernen` und `DELETE /pakete/{art}/{name}` verlangen den
// Pfad aus dem Katalog als Bestätigung. Der Ablauf war:
//
//  1. Die Wurzel liest den Katalog und vergleicht den Pfad.
//  2. Danach ruft sie `Entfernen(art, name)`.
//  3. Der Dienst liest den Katalog **erneut** und löscht, was dann unter Art
//     und Name liegt.
//
// **Bestätigung und Wirkung bezogen sich nicht auf denselben Stand.** Wer
// dazwischen dasselbe Paket entfernt und ein anderes unter demselben Namen
// einbindet, ließ den alten Aufruf das neue löschen — und beide Vergleiche
// waren für sich richtig.
//
// Der Vergleich steht jetzt **im Dienst**, hinter derselben einen Lesung, die
// auch löscht.
func TestEntfernenGegenEineUeberholteBestaetigung(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "fremd.mse-style",
		"mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: 2026-02-02\n")

	ziel := t.TempDir()
	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, ziel, ziel)

	if _, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "fremd.mse-style")); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	// **Der Pfad, den der Rufende gesehen hat, ist nicht mehr der im Katalog.**
	fehler := verwalter.Entfernen(t.Context(), "mse-style", "fremd",
		filepath.Join(ziel, "ein-anderes.mse-style"))
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Fehler %v, erwartet war ErrConflict", fehler)
	}

	// **Und nichts ist gelöscht.** Ohne diese Zusicherung wäre der Lauf auch
	// dann grün, wenn der Dienst erst löschte und danach klagte.
	if _, err := os.Stat(filepath.Join(ziel, "fremd.mse-style", "style")); err != nil {
		t.Errorf("das Paket wurde trotzdem entfernt: %v", err)
	}

	if len(katalog.pakete) != 1 {
		t.Errorf("Katalog: %+v", katalog.pakete)
	}
}

// **Ohne Bestätigung wird geprüft, was der Katalog sagt.**
//
// Das Fenster fragt in einem Dialog, die Kommandozeile zeigt ihren Aufruf —
// beide brauchen die Pfadbestätigung nicht ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
// **Ein leerer Pfad heißt deshalb: nicht bestätigt**, und nicht „bestätige den
// leeren Pfad".
func TestEntfernenOhneBestaetigungGehtDurch(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "fremd.mse-style",
		"mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: 2026-02-02\n")

	ziel := t.TempDir()
	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, ziel, ziel)

	if _, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "fremd.mse-style")); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	if err := verwalter.Entfernen(t.Context(), "mse-style", "fremd", ""); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if len(katalog.pakete) != 0 {
		t.Errorf("Katalog: %+v", katalog.pakete)
	}
}
