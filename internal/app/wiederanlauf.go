package app

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// StartmarkeSetzen beginnt die Sitzung des Fensters: sie liest, welche
// frühere nicht sauber endete, legt die eigene Marke in `ordner` und richtet
// die Absturzspur ein (F04, #2343). **Nur das Fenster ruft es** (`main.go`)
// — Kommandozeile und Dienste sind kurz oder haben ihr eigenes Protokoll.
//
// Ein Abbruch steht danach im Protokoll und über [App.Wiederanlauf] im
// Fenster.
func (a *App) StartmarkeSetzen(ordner string) error {
	beginn, err := sitzungBeginnen(ordner, os.Getpid(), time.Now(), prozessLebt, setzeAbsturzausgabe)
	if err != nil {
		return err
	}

	a.lauf.startmarke = beginn

	for _, abbruch := range beginn.abbrueche {
		art := "ohne Absturzspur"
		if abbruch.MitSpur {
			art = "mit Absturzspur"
		}

		slog.Warn("die letzte Sitzung endete nicht sauber",
			"stand", abbruch.Gestartet.Format(time.RFC3339), "art", art)
	}

	return nil
}

// WiederanlaufAntwort nennt die früheren Sitzungen, die nicht sauber endeten.
type WiederanlaufAntwort struct {
	Abbrueche []Abbruch `json:"abbrueche"`
	Fehler    *Fehler   `json:"fehler"`
}

// Wiederanlauf sagt dem Fenster, ob die letzte Sitzung nicht sauber endete
// — **eine leere Liste**, wenn sie es tat oder keine Marke gelegt wurde.
func (a *App) Wiederanlauf(ctx context.Context) WiederanlaufAntwort {
	ctx = a.imProfil(ctx)

	// **Ein abgebrochener Ruf bekommt seinen Abbruch** und keine Liste, die
	// niemand mehr liest.
	if err := ctx.Err(); err != nil {
		return WiederanlaufAntwort{Abbrueche: []Abbruch{}, Fehler: Uebersetze(ctx, err)}
	}

	abbrueche := a.lauf.startmarke.abbrueche
	if abbrueche == nil {
		abbrueche = []Abbruch{}
	}

	return WiederanlaufAntwort{Abbrueche: abbrueche}
}
