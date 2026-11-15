package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Massenschreiber ist, was `app` für eine Massenänderung braucht.
type Massenschreiber interface {
	// FelderVorschau sagt, was der Wert auf den Karten hieße — ohne zu
	// schreiben (Werkstück U6, #1417).
	FelderVorschau(ctx context.Context, setID int64, karten []int64, feld, wert string) (
		domain.Massenvorschau, error)

	// FelderSetzenMitErwartung schreibt nur, wenn jede Karte noch auf dem
	// Altwert steht, den die Vorschau gezeigt hat.
	FelderSetzenMitErwartung(ctx context.Context, setID int64, erwartet []domain.Feldvorwert, feld, wert string) (
		vorgang string, fehler error)
}

// MassenAntwort trägt die betroffenen Karten — oder keine.
type MassenAntwort struct {
	// Betroffen nennt die Karten, an denen die Änderung gescheitert ist.
	//
	// **Sie ist nicht die Auswahl.** Geschrieben wurde nichts; diese hier sind
	// die, die daran schuld sind, und nur sie gehören hervorgehoben.
	Betroffen []string `json:"betroffen"`

	// Vorgang benennt diesen Schritt im Verlauf
	// (#1393).
	//
	// **Damit die Rücknahme trifft, was gemeint war.** `Zuruecknehmen` nimmt
	// sonst den jüngsten noch geltenden Schritt des Sets — am Fenster genau
	// richtig, über MCP, Kommandozeile und REST womöglich die Änderung eines
	// zweiten Prozesses.
	//
	// Bei leerer Auswahl und im Fehlerfall ist sie leer: es hat kein Vorgang
	// stattgefunden.
	Vorgang string `json:"vorgang"`

	Fehler *Fehler `json:"fehler"`
}

// MitMassenschreiber setzt die Naht zur Massenänderung.
func (a *App) MitMassenschreiber(schreiber Massenschreiber) { a.massenschreiber = schreiber }
