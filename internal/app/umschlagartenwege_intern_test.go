package app

// **Die Fehlerwege der vier weiteren Arten** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// # Warum sie einen eigenen Lauf bekommen
//
// **Weil sie sonst niemand betritt.** Der Abnahmelauf daneben fährt die Wege,
// die glücken; §10 sagt dazu: „wer nur den Erfolgsweg prüft, hat die Hälfte
// ausgelassen." Und jeder dieser Wege endet in einem Satz, den ein Mensch
// liest.

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Karte und eine Vorlage gehen hinaus.**
//
// **Beide über dieselbe Gestalt, die auch REST und die Kommandozeile
// ausgeben** — ein zweiter Serializer liefe mit dem ersten auseinander.
func TestEineKarteUndEineVorlageGehenAlsUmschlagHinaus(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Kartenzeilen(ctx, geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Kartenzeilen: %+v", karten.Fehler)
	}

	if len(karten.Zeilen) == 0 {
		t.Skip("das Saatset führt keine Karte")
	}

	geschnuert := anwendung.UmschlagSchnueren(ctx, "karte",
		karten.Zeilen[0].ID, einUmschlagsziel(t, "karte"))
	if geschnuert.Fehler != nil {
		t.Fatalf("eine Karte schnüren: %+v", geschnuert.Fehler)
	}

	if len(geschnuert.Objekte) != 1 || geschnuert.Objekte[0].Art != "karte" {
		t.Errorf("der Umschlag führt %+v", geschnuert.Objekte)
	}

	// **Und eine Vorlage** — über eine Werkstatt, deren Wurzel dieser Lauf
	// wählt. `Paketdateien` liest aus dem **Paketverzeichnis** dieses
	// Profils und weist alles davor ab; in einem Testzuhause liegt dort
	// nichts, und ein Lauf über die Saat prüfte einen Ort, den die Werkstatt
	// gar nicht ansieht.
	anwendung.paketdateien = werkstattUeber(t, saatpaket(t))

	vorlage := anwendung.UmschlagSchnueren(ctx, "vorlage",
		"probe.mse-game", einUmschlagsziel(t, "vorlage"))
	if vorlage.Fehler != nil {
		t.Fatalf("eine Vorlage schnüren: %+v", vorlage.Fehler)
	}

	if len(vorlage.Objekte) == 0 {
		t.Error("der Umschlag führt keine Datei")
	}
}

// saatpaket legt ein echtes Paket in ein Wegwerfverzeichnis.
//
// **Ein echtes und kein nachgebautes** (§3): `msepaket.Oeffne` liest ein
// Paketformat, und eine zusammengesetzte Zeichenkette prüfte den Leser
// statt den Weg.
func saatpaket(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()

	// **Ein Paket ist ein Verzeichnis** — die Saat legt es entpackt ab, und
	// `msepaket.Oeffne` liest beides.
	quelle := filepath.Join("..", "..", "saat", "probe.mse-game")
	ziel := filepath.Join(wurzel, "probe.mse-game")

	if fehler := os.MkdirAll(ziel, 0o750); fehler != nil {
		t.Fatalf("das Ziel anlegen: %v", fehler)
	}

	// **Über `os.Root` und nicht über zusammengesetzte Pfade.** Ein Name aus
	// einem Verzeichnis ist für `gosec` ein fremder Wert (G703), und das ist
	// keine Schikane: er käme aus einem Paket, das jemand anderes gebaut
	// hat. Die Wurzel hält ihn drin.
	auf, fehler := os.OpenRoot(quelle)
	if fehler != nil {
		t.Fatalf("die Saat öffnen: %v", fehler)
	}

	defer func() { _ = auf.Close() }() //nolint:errcheck // nur gelesen

	hinein, fehler := os.OpenRoot(ziel)
	if fehler != nil {
		t.Fatalf("das Ziel öffnen: %v", fehler)
	}

	defer func() { _ = hinein.Close() }() //nolint:errcheck // nur geschrieben

	eintraege, fehler := os.ReadDir(quelle)
	if fehler != nil {
		t.Fatalf("die Saat lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			continue
		}

		roh, lesefehler := fs.ReadFile(auf.FS(), eintrag.Name())
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), lesefehler)
		}

		datei, schreibfehler := hinein.Create(eintrag.Name())
		if schreibfehler != nil {
			t.Fatalf("%s anlegen: %v", eintrag.Name(), schreibfehler)
		}

		if _, schreibfehler := datei.Write(roh); schreibfehler != nil {
			t.Fatalf("%s schreiben: %v", eintrag.Name(), schreibfehler)
		}

		if schliessfehler := datei.Close(); schliessfehler != nil {
			t.Fatalf("%s schließen: %v", eintrag.Name(), schliessfehler)
		}
	}

	return wurzel
}

// werkstattUeber baut eine Paketwerkstatt über einer gewählten Wurzel.
func werkstattUeber(t *testing.T, wurzel string) Paketdateiquelle {
	t.Helper()

	return &Werkstattnaht{werkstatt: service.NeuePaketwerkstatt(wurzel)}
}

// **Jeder Weg, der nicht geht, sagt es.**
func TestJederUmschlagswegDerNichtGehtSagtEs(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	for name, fall := range map[string]struct{ art, kennung string }{
		"eine Karte, die es nicht gibt":   {"karte", "999999"},
		"ein Deck, das es nicht gibt":     {"deck", "999999"},
		"eine Vorlage, die es nicht gibt": {"vorlage", "gibtesnicht"},
		"eine Sicht ohne Zweck":           {"sammlungssicht", "bgegenueber"},
		"eine Sicht ohne Empfänger":       {"sammlungssicht", "/handel"},
		"eine Sicht zu niemandem":         {"sammlungssicht", "bwer/handel"},
		"eine Kennung, die keine ist":     {"set", "keine Zahl"},
	} {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.UmschlagSchnueren(ctx, fall.art, fall.kennung,
				einUmschlagsziel(t, "fehl"))
			if antwort.Fehler == nil {
				t.Fatal("der Weg kam durch")
			}

			if antwort.Objekte == nil {
				t.Error("die Liste ist nil")
			}
		})
	}
}

// **Und jeder Weg der Annahme.**
func TestJederAnnahmewegDerNichtGehtSagtEs(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	// **Ein Umschlag, den es gibt, aber mit der falschen Art darin.**
	angelegt := anwendung.DeckAnlegen(ctx, "Reisedeck", "magic", "")
	if angelegt.Fehler != nil {
		t.Fatalf("Deck anlegen: %+v", angelegt.Fehler)
	}

	ziel := einUmschlagsziel(t, "deckdrin")

	if antwort := anwendung.UmschlagSchnueren(
		ctx, "deck", angelegt.Deck.ID, ziel); antwort.Fehler != nil {
		t.Fatalf("schnüren: %+v", antwort.Fehler)
	}

	// **Eine Sicht darin gibt es nicht** — und das ist eine Auskunft.
	if sicht := anwendung.UmschlagSicht(ctx, ziel); sicht.Fehler == nil {
		t.Error("ein Deck kam als Sicht durch")
	} else if sicht.Zeilen == nil {
		t.Error("die Zeilen sind nil")
	}

	// **Einen Umschlag, den es nicht gibt, kann niemand lesen.**
	fehlt := filepath.Join(t.TempDir(), "gibtesnicht.msepackage")

	if antwort := anwendung.UmschlagDeck(ctx, fehlt); antwort.Fehler == nil {
		t.Error("ein fehlender Umschlag kam durch")
	}

	if antwort := anwendung.UmschlagSicht(ctx, fehlt); antwort.Fehler == nil {
		t.Error("ein fehlender Umschlag kam durch")
	}

	// **Und eine Datei, die im Manifest steht und nicht im Verzeichnis
	// liegt.** Sie entsteht, wenn jemand den Umschlag von Hand anfasst —
	// der Weg dahinter meldet sie, statt ein leeres Deck anzulegen.
	//
	// **Über den Helfer und nicht über den ganzen Weg**: `austausch.Lies`
	// prüft die Prüfsummen, also käme ein so verletzter Umschlag gar nicht
	// erst durch. Der Zweig ist trotzdem die Stelle, an der es laut würde.
	gelesen := geleseneSendung{ort: t.TempDir()}
	gelesen.Verzeichnis = t.TempDir()
	gelesen.Manifest.Objekte = []austausch.Objekt{{Art: austausch.ArtDeck}}

	if _, fehler := lesenAusSendung(
		gelesen, austausch.ArtDeck, Deckdatei); fehler == nil {
		t.Error("die fehlende Datei kam durch")
	}
}
