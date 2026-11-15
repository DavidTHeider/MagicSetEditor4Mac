// **Das Profil, auf das eine Abfrage klemmt** — Werkstück 2 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Warum es im Kontext reist und nicht in jeder Signatur
//
// **Weil es sonst 94 Signaturen wären.** Der Kontext fließt ohnehin durch
// jede Abfrage dieses Speichers; das Profil daran zu hängen ändert keinen
// Aufrufer, der es nicht angeht.
//
// **Und weil ein Vergessen laut sein muss.** Ein Kontextwert, den eine
// Abfrage stillschweigend als Null liest, gäbe null Zeilen — die Anwendung
// sähe aus, als wären alle Daten weg, und niemand wüsste warum. [ProfilAus]
// gibt deshalb einen **Fehler**, wenn keines gesetzt ist.
//
// # Was das nicht ist
//
// **Kein `ROW LEVEL SECURITY`.** Es wirkt in dieser Anwendung, nicht in der
// Datei: wer sie mit einem fremden Werkzeug öffnet, sieht alles. Vermerk 0066
// nennt das beim Namen, und die Oberfläche darf nichts anderes behaupten.

package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// profilschluessel ist der Schlüssel im Kontext.
//
// **Ein eigener Typ**, damit kein fremdes Paket ihn versehentlich trifft —
// die Form, die `context` dafür vorsieht.
type profilschluessel struct{}

// MitProfil hängt das aktive Profil an einen Kontext.
func MitProfil(ctx context.Context, profilID int64) context.Context {
	return context.WithValue(ctx, profilschluessel{}, profilID)
}

// ProfilAus liest es wieder heraus.
//
// **Ein Fehler und keine Null.** Eine Abfrage mit `profil_id = 0` fände
// nichts, und das sähe aus wie ein leerer Bestand statt wie ein Fehler im
// Aufruf (§8).
func ProfilAus(ctx context.Context) (int64, error) {
	profilID, da := ctx.Value(profilschluessel{}).(int64)
	if !da || profilID == 0 {
		return 0, fmt.Errorf(
			"%w: dieser Aufruf nennt kein Profil — ohne eines lässt sich nicht "+
				"sagen, wessen Daten gemeint sind", domain.ErrValidation)
	}

	return profilID, nil
}

// MitGezeigtemProfil hängt an einen Kontext das Profil, das der Bestand als
// gezeigtes führt.
//
// **Für die Flächen ohne Fenster.** Kommandozeile, MCP und REST haben je einen
// Wurzelkontext, den sie selbst bauen — anders als das Fenster, dem Wails
// seinen hereinreicht. Sie statten ihn einmal aus, und alles darunter erbt ihn.
//
// **Ein Fehler hier ist ein Fehler**, kein Protokolleintrag: eine Fläche ohne
// Profil kann nichts lesen und nichts schreiben, und ein Start, der das
// verschweigt, meldet gleich darauf einen Bestandsfehler ohne Ursache.
func MitGezeigtemProfil(ctx context.Context, db *sql.DB) (context.Context, error) {
	profilID, fehler := NeueNutzerablage(db).Profilzeiger(ctx)
	if fehler != nil {
		return ctx, fehler
	}

	return MitProfil(ctx, profilID), nil
}
