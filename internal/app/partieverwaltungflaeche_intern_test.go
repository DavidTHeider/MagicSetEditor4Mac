package app

// **Die Verwaltung an der kopflosen Fläche** — PLAY-03.AK-05
// (#1711).
//
// # Warum das hier geprüft wird und nicht nur in `internal/flaechen`
//
// **Weil dort die Abdeckung dieses Pakets nicht zählt** — derselbe Grund wie
// bei `partieflaechenweg_intern_test.go`: die Flächenläufe fahren gebaute
// Programme. Was `mse`, MCP und der Dienst **gemeinsam** nehmen, steht hier
// einmal und wird hier gemessen.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der ganze Weg über die kopflose Fläche: ausführen, einlesen, kopieren,
// löschen.**
func TestDieKopfloseFlaecheVerwaltetEinePartie(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	deckID := einDeckImBestand(t, db)
	flaeche := PartieflaecheUeber(db)

	angelegt := flaeche.PartieAusWuenschen(ctx, "magic", "", []Sitzwunsch{
		{Kennung: "sitz-1", Name: "David", DeckID: deckID},
		{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
	})
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	partieID := angelegt.Partie.ID

	// **Der Text und die Datei tragen dasselbe** — zwei Gestalten, eine
	// Bedeutung (§16).
	alsText := flaeche.PartieAusfuhrText(ctx, partieID)
	if alsText.Fehler != nil {
		t.Fatalf("Ausfuhr als Text: %+v", alsText.Fehler)
	}

	ort := filepath.Join(t.TempDir(), "partie.msepartie")

	if antwort := flaeche.PartieAusfuhrSchreiben(
		ctx, partieID, ort); antwort.Fehler != nil {
		t.Fatalf("Ausfuhr in eine Datei: %+v", antwort.Fehler)
	}

	roh, fehler := os.ReadFile(ort) //nolint:gosec // der Ort kommt aus dem Lauf
	if fehler != nil {
		t.Fatalf("die Datei lesen: %v", fehler)
	}

	if string(roh) != alsText.Inhalt {
		t.Errorf("Datei und Text sind verschieden:\n%.80q\n%.80q",
			roh, alsText.Inhalt)
	}

	// **Und beide Wege zurück legen je eine neue Partie an.**
	ausText := flaeche.PartieEinfuhrText(ctx, alsText.Inhalt)
	ausDatei := flaeche.PartieEinfuhrLesen(ctx, ort)

	switch {
	case ausText.Fehler != nil:
		t.Fatalf("Einfuhr aus Text: %+v", ausText.Fehler)
	case ausDatei.Fehler != nil:
		t.Fatalf("Einfuhr aus Datei: %+v", ausDatei.Fehler)
	case ausText.Kennung == ausDatei.Kennung:
		t.Error("beide Einfuhren gaben dieselbe Kennung")
	case ausDatei.Pfad != ort:
		t.Errorf("die Einfuhr aus der Datei nennt den Ort %q", ausDatei.Pfad)
	}

	kopie := flaeche.PartieKopieren(ctx, partieID)
	if kopie.Fehler != nil || kopie.Kennung == partieID {
		t.Fatalf("kopieren: %+v", kopie)
	}

	if geloescht := flaeche.PartieLoeschen(
		ctx, partieID); geloescht.Fehler != nil {
		t.Fatalf("löschen: %+v", geloescht.Fehler)
	}

	// **Und danach ist sie weg** — die Gegenprobe.
	if nochmal := flaeche.PartieLoeschen(ctx, partieID); nochmal.Fehler == nil {
		t.Error("die Partie ließ sich zweimal löschen")
	}
}

// **Eine Datei, die nicht taugt, legt nichts an** — und der Satz sagt, woran
// es liegt.
func TestDieKopfloseFlaecheWeistEineFremdeDateiAb(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	antwort := PartieflaecheUeber(db).PartieEinfuhrText(ctx, `{"fassung":99}`)
	if antwort.Fehler == nil {
		t.Fatal("die fremde Fassung ging durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Fassung 99") {
		t.Errorf("die Meldung lautet %q", antwort.Fehler.Meldung)
	}
}
