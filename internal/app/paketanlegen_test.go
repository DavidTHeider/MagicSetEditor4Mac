// **Ein Paket anlegen — über die Fassade und über den Flächenweg**
// (#1350).
//
// **Zwei Wege, zwei Läufe**: `App.PaketAnlegen` ist die Wails-Naht,
// `Paketweg.Anlegen` die für Kommandozeile, MCP und Dienst. Sie liegen an
// verschiedenen Wurzeln (§2), und ein Lauf über den einen sagt über den
// anderen nichts.
package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Flächenweg legt an und der Vollname fällt auf den Namen zurück.**
//
// Der Rückfall steht dort und nicht in jeder Fläche: drei Kopien derselben
// Zeile laufen auseinander.
func TestDerFlaechenwegLegtEinPaketAn(t *testing.T) {
	weg, _ := pflegeweg(t, "unbenutzt.mse-style")

	paket, fehler := weg.Anlegen(probekontext(t), "mse-game", "eigenes", "  ", "")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if paket.Art != "mse-game" || paket.Name != "eigenes" {
		t.Fatalf("angelegt: %+v", paket)
	}

	// **Der Vollname ist der Name** — nicht leer.
	liste, _, lesefehler := weg.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	for _, eines := range liste {
		if eines.Paket.Name == "eigenes" && !strings.Contains(eines.Paket.Pfad, "eigenes.mse-game") {
			t.Errorf("der Pfad ist %q", eines.Paket.Pfad)
		}

		// **Ein angelegtes Paket ist ein eigenes** (DOM-01.AK-01, #1696).
		if eines.Paket.Name == "eigenes" && eines.Herkunftsart != string(domain.PaketEigen) {
			t.Errorf("die Herkunftsart ist %q", eines.Herkunftsart)
		}
	}
}

// **Und ein Stil ohne Spiel wird auch dort abgewiesen.**
func TestDerFlaechenwegWeistEinenStilOhneSpielAb(t *testing.T) {
	weg, _ := pflegeweg(t, "unbenutzt.mse-style")

	if _, fehler := weg.Anlegen(
		probekontext(t), "mse-style", "schlicht", "Schlicht", ""); fehler == nil {
		t.Error("ein Stil ohne Spiel kam durch")
	}
}

// **Die Fassade reicht den Fehler als Umschlag durch** (§8).
func TestDiePaketfassadeReichtIhrenFehlerDurch(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketanlegerprobe{fehler: domain.NewFieldError(
		"name", domain.SchluesselValidierung, "krumm")})

	antwort := anwendung.PaketAnlegen(t.Context(), "mse-game", "!", "", "")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.Fehler.Feld != "name" {
		t.Errorf("Feld %q", antwort.Fehler.Feld)
	}
}

// **Die Fassade gibt das angelegte Paket zurück.**
//
// Ohne diesen Fall sagte der Lauf darüber nur, dass etwas fehlschlägt — und
// das täte auch eine Fassade, die gar nichts kann.
func TestDiePaketfassadeGibtDasAngelegtePaket(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketanlegerprobe{})

	antwort := anwendung.PaketAnlegen(t.Context(), "mse-game", "eigenes", "", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Paket == nil || antwort.Paket.Name != "eigenes" {
		t.Fatalf("Paket: %+v", antwort.Paket)
	}
}

// **Und die Liste der anlegbaren Arten kommt aus der Anwendung.**
//
// Vier von sieben Paketendungen lassen sich aus einem Gerüst anlegen; eine
// Aufzählung im Frontend wäre eine zweite Wahrheit.
func TestDieAnlegbarenArtenNennenVier(t *testing.T) {
	arten := app.Anlegbar()
	if len(arten) != 4 {
		t.Fatalf("%d Arten: %+v", len(arten), arten)
	}

	brauchtSpiel := map[string]bool{
		"mse-game": false, "mse-style": true, "mse-include": false, "mse-export-template": true,
	}

	for _, eine := range arten {
		if will, da := brauchtSpiel[eine.Art]; !da || will != eine.BrauchtSpiel {
			t.Errorf("%s braucht ein Spiel: %v", eine.Art, eine.BrauchtSpiel)
		}
	}
}

// paketanlegerprobe ist ein Paketführer, der beim Anlegen scheitert.
type paketanlegerprobe struct {
	paketfuehrer

	fehler error
}

func (p *paketanlegerprobe) Anlegen(_ context.Context, wunsch service.Neuanlage) (
	domain.Paket, error,
) {
	if p.fehler != nil {
		return domain.Paket{}, p.fehler
	}

	return domain.Paket{Art: wunsch.Art, Name: wunsch.Name, Kurzname: wunsch.Vollname}, nil
}
