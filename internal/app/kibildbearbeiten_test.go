package app_test

// **Ein vorhandenes Bild bearbeiten** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass bearbeitet wird, was in der Karte steht.** Das Issue trennt
// Generierung und Bearbeitung, und die Bearbeitung braucht eine Vorlage: das
// Bild aus dem Feld, nicht seine Vorschau und nicht irgendein anderes.
//
// **Und dass sie ohne Vorlage nicht stattfindet.** Ein leeres Bildfeld hat
// nichts zu bearbeiten, und das sagt die Naht, statt einen Auftrag
// hinauszuschicken, der Geld kostet.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitVorhandenemBild legt ein Bild an der Karte und in der Ablage an.
func mitVorhandenemBild(t *testing.T, welt bildwelt) bildwelt {
	t.Helper()

	wurzel := t.TempDir()

	ort := filepath.Join(wurzel, "2")
	if fehler := os.MkdirAll(ort, 0o700); fehler != nil {
		t.Fatalf("das Verzeichnis ließ sich nicht anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(filepath.Join(ort, "kunst1.png"),
		welt.bild, 0o600); fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	welt.karten.zeile.Values = append(welt.karten.zeile.Values,
		domain.Value{Field: "kunst", Text: "kunst1.png"})

	welt.anwendung.MitBildablage(service.NeueBildablage(wurzel))

	return welt
}

// bearbeiten lässt das vorhandene Bild verändern.
func bearbeiten(t *testing.T, welt bildwelt) app.KIBildvorschlagAntwort {
	t.Helper()

	return welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gemma4:8b", "mach es dunkler",
		domain.KIBildgroesseHoch, domain.KIBildgueteHoch, true, "")
}

// **Die Vorlage kommt aus dem Feld und geht an den Anbieter.**
func TestDieVorlageKommtAusDemFeld(t *testing.T) {
	t.Parallel()

	welt := mitVorhandenemBild(t, einebildwelt(t))

	antwort := bearbeiten(t, welt)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(welt.laufzeitbild.gesehen) != 1 {
		t.Fatalf("es gingen %d Aufträge hinaus", len(welt.laufzeitbild.gesehen))
	}

	wunsch := welt.laufzeitbild.gesehen[0]
	if len(wunsch.Quelle) != len(welt.bild) {
		t.Fatalf("die Quelle wiegt %d Bytes statt %d",
			len(wunsch.Quelle), len(welt.bild))
	}
}

// **Ohne Bild im Feld gibt es nichts zu bearbeiten.**
func TestOhneBildImFeldGibtEsNichtsZuBearbeiten(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	antwort := bearbeiten(t, welt)
	if antwort.Fehler == nil {
		t.Fatal("die Bearbeitung ohne Vorlage kam durch")
	}

	if len(welt.laufzeitbild.gesehen) != 0 {
		t.Errorf("es ging trotzdem ein Auftrag hinaus: %+v",
			welt.laufzeitbild.gesehen)
	}
}

// **Und ohne lesbare Datei auch nicht.**
//
// Der Fall entsteht, wenn jemand das Bildverzeichnis aufräumt: im Feld steht
// ein Name, auf der Platte liegt nichts.
func TestOhneLesbareDateiGibtEsNichtsZuBearbeiten(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.karten.zeile.Values = append(welt.karten.zeile.Values,
		domain.Value{Field: "kunst", Text: "gibtesnicht.png"})

	if antwort := bearbeiten(t, welt); antwort.Fehler == nil {
		t.Fatal("die fehlende Datei kam durch")
	}
}

// **Ohne Ablage gibt es keinen Ort, an dem die Vorlage läge.**
func TestOhneAblageGibtEsKeineVorlage(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.karten.zeile.Values = append(welt.karten.zeile.Values,
		domain.Value{Field: "kunst", Text: "kunst1.png"})

	welt.anwendung.MitBildablage(service.NeueBildablage(""))

	if antwort := bearbeiten(t, welt); antwort.Fehler == nil {
		t.Fatal("die Bearbeitung ohne Ablageort kam durch")
	}
}

// **Und ohne lesbare Karte erst recht nicht.**
func TestOhneLesbareKarteGibtEsKeineVorlage(t *testing.T) {
	t.Parallel()

	welt := mitVorhandenemBild(t, einebildwelt(t))
	welt.karten.fehler = domain.MitGrund(domain.ErrNotFound, nil,
		"die Karte gibt es nicht")

	if antwort := bearbeiten(t, welt); antwort.Fehler == nil {
		t.Fatal("die Bearbeitung ohne Karte kam durch")
	}
}

// **Ein JPEG bekommt seine Endung.**
//
// # Warum das zählt
//
// **Weil der Zuschnittweg über eine Datei geht** und `image.Decode` sie am
// Inhalt erkennt — die Endung ist für den Menschen, der sie im
// Zwischenspeicher findet. Sie falsch zu setzen wäre eine kleine Lüge, die
// bleibt.
func TestEinJPEGBekommtSeineEndung(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.laufzeitbild.bild.Inhalt = einProbeJPEG(t)
	welt.laufzeitbild.bild.Medientyp = "image/jpeg"

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if !strings.HasPrefix(vorschlag.Vorschau, "data:image/jpeg;base64,") {
		t.Fatalf("die Vorschau beginnt mit %.40q", vorschlag.Vorschau)
	}

	// **Und der Zuschnitt kommt trotzdem an.**
	if antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		vorschlag.Kennung); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}
}

// **Ohne Profil im Kontext geht nichts.**
//
// **Alle vier Wege**: ein Ruf ohne Profil weiß nicht, wessen Bild gemeint
// wäre — und der Vorrat bindet an genau dieses.
func TestOhneProfilGehtKeinBildweg(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	ohne := t.Context()

	if antwort := welt.anwendung.KIBildVorschlagen(ohne, welt.karte, "kunst",
		"ollama", "gemma4:8b", "ein Blitz", "", "", false, ""); antwort.Fehler == nil {
		t.Error("der Vorschlag gelang ohne Profil")
	}

	if antwort := welt.anwendung.KIBildVerwerfen(ohne, "egal"); antwort.Fehler == nil {
		t.Error("das Verwerfen gelang ohne Profil")
	}

	if antwort := welt.anwendung.KIBildZuschnitt(ohne, "4", "kunst",
		"egal"); antwort.Fehler == nil {
		t.Error("der Zuschnitt gelang ohne Profil")
	}

	if antwort := welt.anwendung.KIBildEinsetzen(ohne, "2", "4", "kunst",
		"egal", app.AuswahlDTO{}); antwort.Fehler == nil {
		t.Error("das Einsetzen gelang ohne Profil")
	}
}

// **Eine unlesbare Setkennung setzt nichts ein.**
func TestEineUnlesbareSetkennungSetztNichtsEin(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "sieben", "4",
		"kunst", "egal", app.AuswahlDTO{}); antwort.Fehler == nil {
		t.Error("die unlesbare Setkennung kam durch")
	}
}

// **Ein Vorrat, den niemand aufräumt, läuft trotzdem ab.**
//
// **Das ist die Frist aus dem Issue**, hier an der Naht: ein abgelaufener
// Vorschlag lässt die Karte unverändert, weil es ihn nicht mehr gibt.
func TestEinAbgelaufenerVorschlagLaesstDieKarteUnveraendert(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	kurz := service.UeblicheBildgrenzen()
	kurz.Frist = time.Nanosecond

	// **Eine Uhr, die geht.** Eine feste Zeit ließe nichts ablaufen: die
	// Frist wird beim Ablegen aus **ihr** gerechnet.
	vorrat, fehler := service.NeuerBildvorrat(t.TempDir(), kurz, time.Now)
	if fehler != nil {
		t.Fatalf("der Vorrat ließ sich nicht bauen: %v", fehler)
	}

	welt.anwendung.MitBildvorrat(vorrat)

	vorher := welt.kartenwerte(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler == nil {
		t.Fatal("der abgelaufene Vorschlag kam durch")
	}

	if nachher := welt.kartenwerte(t); nachher != vorher {
		t.Errorf("die Karte wurde geändert: %q statt %q", nachher, vorher)
	}
}
