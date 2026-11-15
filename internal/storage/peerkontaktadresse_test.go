package storage_test

// **Die Adresse an der Kontaktablage** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine eigene Datei**, weil `peerkontakt_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, **wer** ein
// Kontakt ist — hier, **wo** er erreichbar ist.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine eingetragene Adresse kommt wieder heraus — auch aus der Liste.**
//
// **Beide Wege und nicht einer**: `Lesen` und `Liste` haben je eine eigene
// Abfrage, und eine vergessene Spalte fiele nur an einem von beiden auf.
func TestEineAdresseKommtAusBeidenWegenZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, ctx, "bkennung-eins", "Der andere")

	if fehler := ablage.AdresseSetzen(
		ctx, "bkennung-eins", "192.168.1.7:41234"); fehler != nil {
		t.Fatalf("Adresse: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "bkennung-eins")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Adresse != "192.168.1.7:41234" {
		t.Errorf("Lesen gibt %q", gelesen.Adresse)
	}

	liste, fehler := ablage.Liste(ctx)
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(liste) != 1 || liste[0].Adresse != "192.168.1.7:41234" {
		t.Fatalf("die Liste gibt %+v", liste)
	}
}

// **Ohne Eintrag ist sie leer und nicht etwa unbekannt.**
func TestOhneEintragIstDieAdresseLeer(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, ctx, "bkennung-zwei", "Der andere")

	gelesen, fehler := ablage.Lesen(ctx, "bkennung-zwei")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Adresse != "" {
		t.Fatalf("die Adresse ist %q", gelesen.Adresse)
	}
}

// **Eine leere Adresse löscht die alte.**
//
// **Und das ist eine Änderung und kein Auslassen** — „ich kenne keine mehr"
// ist eine Aussage. Bliebe die alte stehen, ginge der nächste Versuch an
// einen Rechner, der inzwischen jemand anderem gehört.
func TestEineLeereAdresseLoeschtDieAlte(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, ctx, "bkennung-drei", "Der andere")

	for _, adresse := range []string{"beispiel.invalid:443", ""} {
		if fehler := ablage.AdresseSetzen(
			ctx, "bkennung-drei", adresse); fehler != nil {
			t.Fatalf("Adresse %q: %v", adresse, fehler)
		}
	}

	gelesen, fehler := ablage.Lesen(ctx, "bkennung-drei")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Adresse != "" {
		t.Fatalf("die alte Adresse steht noch da: %q", gelesen.Adresse)
	}
}

// **Eine Begegnung rührt die Adresse nicht an.**
//
// **Derselbe Grund wie beim Stand**: was dieses Profil eingetragen hat, darf
// eine hereinkommende Nachricht nicht überschreiben.
func TestEineBegegnungRuehrtDieAdresseNichtAn(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, ctx, "bkennung-vier", "Der andere")

	if fehler := ablage.AdresseSetzen(
		ctx, "bkennung-vier", "beispiel.invalid:443"); fehler != nil {
		t.Fatalf("Adresse: %v", fehler)
	}

	begegne(t, ablage, ctx, "bkennung-vier", "Derselbe unter neuem Namen")

	gelesen, fehler := ablage.Lesen(ctx, "bkennung-vier")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Adresse != "beispiel.invalid:443" {
		t.Fatalf("die Adresse ist %q", gelesen.Adresse)
	}
}

// **Eine krumme Adresse erreicht die Tabelle nicht.**
func TestEineKrummeAdresseErreichtDieTabelleNicht(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, ctx, "bkennung-fuenf", "Der andere")

	if fehler := ablage.AdresseSetzen(
		ctx, "bkennung-fuenf", "frag mal David",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und wen es nicht gibt, der bekommt keine.**
func TestWenEsNichtGibtDerBekommtKeineAdresse(t *testing.T) {
	db := frischeDatenbank(t)

	if fehler := storage.NeueKontaktablage(db).AdresseSetzen(
		probekontext(t), "bkennung-niemand", "beispiel.invalid:443",
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam %v", fehler)
	}
}
