package msepaket_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stilmodell öffnet das Prüfstück und gibt sein Modell.
func stilmodell(t *testing.T) msepaket.Stil {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "probe-stil.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	return stil
}

// **Die Kartengröße steht im Kopf — samt Auflösung.**
//
// # Warum die Auflösung zählt
//
// `card dpi: 150` macht die **physische** Größe berechenbar: 375 ÷ 150 = 2,5
// Zoll = 63,5 mm. Das ist die Angabe, deren Fehlen
// [0012](../../docs/entscheidungen/0012-aufloesung-als-zielbreite.md) zur
// Zielbreite gezwungen hat — sie fehlt nicht, sie stand nur nicht in unserem
// Leser.
func TestDieKartengroesseKommtAusDemKopf(t *testing.T) {
	stil := stilmodell(t)

	if stil.Breite != 375 || stil.Hoehe != 523 || stil.DPI != 150 {
		t.Errorf("%d × %d bei %d dpi, erwartet 375 × 523 bei 150", stil.Breite, stil.Hoehe, stil.DPI)
	}

	if stil.Spiel != "probe" {
		t.Errorf("Spiel %q", stil.Spiel)
	}

	// 375 Bildpunkte bei 150 dpi sind 2,5 Zoll — 63,5 mm.
	if breite := stil.BreiteMM(); breite < 63.4 || breite > 63.6 {
		t.Errorf("%.2f mm breit, erwartet 63,50", breite)
	}

	if hoehe := stil.HoeheMM(); hoehe < 88.5 || hoehe > 88.7 {
		t.Errorf("%.2f mm hoch, erwartet 88,60", hoehe)
	}
}

// **Ein Feldstil nennt Stelle und Größe — und jede Zahl darf ein Skript sein.**
//
// Der Leser gibt sie deshalb **roh** heraus, so wie sie dasteht. Was `{ … }`
// bedeutet, weiß `msescript`, und wer es auswertet, ist `service`: ein
// Paketleser, der Skripte auswertete, hätte zwei Aufgaben.
func TestDerFeldstilNenntStelleUndGroesse(t *testing.T) {
	stil := stilmodell(t)

	name, gefunden := stil.Feld("name")
	if !gefunden {
		t.Fatalf("kein Feldstil %q in %v", "name", stil.Felder)
	}

	// **`right` statt `width`**, und das ist der Fall, an dem ein Leser
	// scheitert, der `width` voraussetzt.
	if name.Breite != "" || name.Rechts != `{ 341 - 20 }` {
		t.Errorf("Breite %q, Rechts %q", name.Breite, name.Rechts)
	}

	if name.Links != `{ if card.farbe == "rot" then 32 else 50 }` {
		t.Errorf("Links %q", name.Links)
	}

	if name.Oben != "30" || name.Hoehe != "23" {
		t.Errorf("Oben %q, Hoehe %q", name.Oben, name.Hoehe)
	}

	if name.ZIndex != 1 || name.Ausrichtung != "bottom shrink-overflow" {
		t.Errorf("ZIndex %d, Ausrichtung %q", name.ZIndex, name.Ausrichtung)
	}

	if name.Schrift.Name != "Helvetica" || name.Schrift.Groesse != "16" ||
		name.Schrift.Farbe != "black" {
		t.Errorf("Schrift %+v", name.Schrift)
	}
}

// **Die Zeichenreihenfolge ist `z index`, nicht die Reihenfolge im Text.**
//
// Im Prüfstück steht `rand` zuletzt und trägt 4; `card color` steht zuerst und
// trägt 0. Wer die Textreihenfolge nähme, zeichnete den Rand unter die Karte.
func TestDieFelderKommenNachZIndex(t *testing.T) {
	stil := stilmodell(t)

	if len(stil.Felder) != 3 {
		t.Fatalf("%d Felder, erwartet 3", len(stil.Felder))
	}

	reihe := []string{stil.Felder[0].Feld, stil.Felder[1].Feld, stil.Felder[2].Feld}
	if reihe[0] != "card color" || reihe[1] != "name" || reihe[2] != "rand" {
		t.Errorf("Reihenfolge %v", reihe)
	}
}

// Was ein Feld sonst noch trägt: Maske und Darstellungsart.
func TestEinFeldstilNenntMaskeUndDarstellung(t *testing.T) {
	stil := stilmodell(t)

	rand, _ := stil.Feld("rand")
	if rand.Maske != "rand_maske.png" {
		t.Errorf("Maske %q", rand.Maske)
	}

	farbe, _ := stil.Feld("card color")
	if farbe.Darstellung != "image" || farbe.Bild != "{ probe_hintergrund() }" {
		t.Errorf("Darstellung %q, Bild %q", farbe.Darstellung, farbe.Bild)
	}
}

// Ein Feld, das es nicht gibt, wird gemeldet und nicht erfunden.
func TestEinUnbekanntesFeldGibtEsNicht(t *testing.T) {
	if _, gefunden := stilmodell(t).Feld("gibtsnicht"); gefunden {
		t.Error("ein Feld, das nicht dasteht, wurde gefunden")
	}
}

// **Ohne `card dpi` gibt es keine physische Größe.**
//
// 17 von 190 Stilen eines bezogenen Packs führen keine — und dann ist die
// Antwort 0 und nicht eine erfundene Zahl. Genau das ist die Lage, die
// [0012](../../docs/entscheidungen/0012-aufloesung-als-zielbreite.md)
// beschreibt: sie gilt weiter, sie gilt nur nicht **überall**.
func TestOhneAufloesungGibtEsKeineGroesseInMillimetern(t *testing.T) {
	stil := skriptstil(t)

	if stil.DPI != 0 || stil.BreiteMM() != 0 || stil.HoeheMM() != 0 {
		t.Errorf("%d dpi, %.2f × %.2f mm", stil.DPI, stil.BreiteMM(), stil.HoeheMM())
	}
}

// **Ein unlesbarer `z index` ist ein Befund, kein Abbruch** (§8).
func TestEinUnlesbarerZIndexIstEinBefund(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "skriptblock.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer", len(befunde))
	}

	// Und das Feld ist trotzdem da — mit 0 als Reihenfolge.
	rand, gefunden := stil.Feld("rand")
	if !gefunden || rand.ZIndex != 0 {
		t.Errorf("rand: %+v", rand)
	}
}

// Was ein Feldstil sonst noch führt: `bottom`, und Unbekanntes fällt weg.
func TestEinFeldstilNenntAuchDieUnterkante(t *testing.T) {
	name, _ := skriptstil(t).Feld("name")

	if name.Unten != "43" {
		t.Errorf("Unten %q", name.Unten)
	}
}

// skriptstil gibt das Modell des Skript-Prüfstücks.
func skriptstil(t *testing.T) msepaket.Stil {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "skriptblock.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()

	return stil
}

// Ein Kopfeintrag, den es nicht gibt, ist leer und keine Ausrede.
func TestEinFehlenderKopfeintragIstLeer(t *testing.T) {
	if spiel := skriptstil(t).Spiel; spiel != "probe" {
		t.Errorf("Spiel %q", spiel)
	}
}

// **Ein Stil nennt sein eigenes Paket**, und ohne diese Angabe ist ein
// einfacher Dateiname nicht auflösbar.
//
// `mask: border_mask.png` steht 173-mal in den Stilen von `M15-Magic-Pack`,
// `mask: /magic-genevensis-00-main.mse-style/pride_mask.png` fünfmal. Die erste
// Form meint eine Datei **im eigenen Paket** — wer den Namen ohne diese Angabe
// weiterreicht, sucht sie in der Paketwurzel und findet sie nie.
func TestEinStilNenntSeinEigenesPaket(t *testing.T) {
	if paket := stilmodell(t).Paket; paket != "probe-stil.mse-style" {
		t.Errorf("Paket %q, erwartet %q", paket, "probe-stil.mse-style")
	}
}

// **`default:` ist das Bild, wenn das Feld keines trägt** — und `symbol font:`
// ein eigener Unterblock neben `font:`.
//
// Beide stehen in `magic-m15.mse-style`: `image:` führt
// `default: {default_image(card.card_color)}`, `casting cost` führt eine
// Symbolschrift mit Name, Größe und eigener Ausrichtung. Gezählt über die 190
// Stile des Packs: 71-mal `default:`, 455-mal `symbol font:`.
func TestEinFeldstilNenntStandardbildUndSymbolschrift(t *testing.T) {
	stil := stilmodell(t)

	farbe, _ := stil.Feld("card color")
	if farbe.Standardbild != "ersatz.png" {
		t.Errorf("Standardbild %q", farbe.Standardbild)
	}

	name, _ := stil.Feld("name")
	if name.Symbolschrift.Name != "probe-symbole" ||
		name.Symbolschrift.Groesse != "15" ||
		name.Symbolschrift.Ausrichtung != "middle right" {
		t.Errorf("Symbolschrift %+v", name.Symbolschrift)
	}

	// **Und sie ist nicht die Textschrift.** Beide Unterblöcke tragen `name:`
	// und `size:`; wer sie in dasselbe Feld liest, zeichnet den Text in einer
	// Schrift, die es auf dem Rechner nicht gibt.
	if name.Schrift.Name != "Helvetica" || name.Schrift.Groesse != "16" {
		t.Errorf("Schrift %+v", name.Schrift)
	}
}
