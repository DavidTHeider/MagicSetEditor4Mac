// **Die Bilddateinaht ohne Fenster** —
// #1291.
package app_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bildquelle schreibt ein winziges PNG und gibt seinen Pfad.
func bildquelle(t *testing.T) string {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, 400, 300))
	bild.Set(0, 0, color.RGBA{R: 255, A: 255})

	var roh bytes.Buffer
	if err := png.Encode(&roh, bild); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	pfad := filepath.Join(t.TempDir(), "quelle.png")
	if err := os.WriteFile(pfad, roh.Bytes(), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// kunstkarte legt ein Set mit Bildfeld an und gibt Naht, Set und Karte.
//
// **`probe-kunst` aus der Saat**: `probe-stil` hat keine Fläche für ein Bild,
// und ein Lauf dagegen prüfte den Fehlerweg statt der Fähigkeit (§11).
func kunstkarte(t *testing.T) (naht *app.Bilddatei, setID, karteID int64) {
	t.Helper()

	// **Ein eigenes Zuhause mit der Saat darin** (§3, §11). Die Naht sucht
	// die Pakete unter `HOME`; ohne diese Zeilen läse der Lauf den Bestand
	// dieses Rechners, und ohne Stil gäbe es keine Fläche zum Zuschneiden.
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	quelle, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	ziel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete")
	if kopierfehler := os.CopyFS(ziel, os.DirFS(quelle)); kopierfehler != nil {
		t.Fatalf("Saat kopieren: %v", kopierfehler)
	}

	db := offeneProbedatenbank(t)

	kennung, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Kunstset", Game: "probe", Style: "probe-kunst",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Eine"}}}},
	}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), kennung)
	if lesefehler != nil || len(karten) == 0 {
		t.Fatalf("Karten: %v (%d)", lesefehler, len(karten))
	}

	naht, nahtfehler := app.Bilddateinaht(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Bilddateinaht: %v", nahtfehler)
	}

	return naht, kennung, karten[0].ID
}

// **Ein Bild von der Platte landet im Feld — auf die Fläche des Stils zugeschnitten.**
func TestDieBilddateinahtSetztEinBildEin(t *testing.T) {
	naht, setID, karteID := kunstkarte(t)

	gesetzt, fehler := naht.Einsetzen(probekontext(t), setID, karteID, "kunst", bildquelle(t))
	if fehler != nil {
		t.Fatalf("einsetzen: %v", fehler)
	}

	if gesetzt.QuelleBreite != 400 || gesetzt.QuelleHoehe != 300 {
		t.Errorf("die Quellmaße stimmen nicht: %+v", gesetzt)
	}

	// **315 × 230 stehen in `saat/probe-kunst.mse-style`.** Ohne sie wäre nur
	// eine Datei abgelegt und nicht zugeschnitten worden.
	if gesetzt.ZielBreite != 315 || gesetzt.ZielHoehe != 230 {
		t.Errorf("die Zielmaße stimmen nicht: %+v", gesetzt)
	}

	if gesetzt.Feld != "kunst" {
		t.Errorf("das Feld ist %q", gesetzt.Feld)
	}
}

// **Eine Datei, die kein Bild ist, wird abgewiesen — bevor etwas geschrieben wird.**
func TestDieBilddateinahtWeistEineNichtbilddateiAb(t *testing.T) {
	naht, setID, karteID := kunstkarte(t)

	kein := filepath.Join(t.TempDir(), "kein.png")
	if err := os.WriteFile(kein, []byte("hallo"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, fehler := naht.Einsetzen(probekontext(t), setID, karteID, "kunst", kein); fehler == nil {
		t.Error("eine Textdatei ging als Bild durch")
	}
}

// **Ein Feld, das der Stil nicht führt, ebenso.**
func TestDieBilddateinahtWeistEinFremdesFeldAb(t *testing.T) {
	naht, setID, karteID := kunstkarte(t)

	_, fehler := naht.Einsetzen(probekontext(t), setID, karteID, "gibtsnicht", bildquelle(t))
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler: %v, erwartet ErrValidation", fehler)
	}
}

// **Und eine Setkennung, die nicht zur Karte gehört.**
//
// Die Karte gegen ihr eigenes Set zu halten ginge immer auf — ein Wächter, der
// nicht scheitern kann, ist keiner (#808, §10).
func TestDieBilddateinahtWeistEinFremdesSetAb(t *testing.T) {
	naht, _, karteID := kunstkarte(t)

	if _, fehler := naht.Einsetzen(probekontext(t), 4711, karteID, "kunst", bildquelle(t)); fehler == nil {
		t.Error("eine fremde Setkennung ging durch")
	}
}

// **Ohne Benutzerverzeichnis kommt die Naht nicht zustande.**
//
// **Alle drei Fehler auf einmal**: Schreibnaht, Zeichennaht und
// Bilderverzeichnis scheitern an derselben Ursache, und alle drei zu melden
// ist die bessere Auskunft als der erste davon.
func TestOhneBenutzerverzeichnisGibtEsKeineBilddateinaht(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.Bilddateinaht(nil, domain.UrheberKommandozeile); err == nil {
		t.Error("kein Fehler")
	}
}
