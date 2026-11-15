package app

// **Ein Einstellungsverwalter, gebaut an einer Stelle** — OPS-02.AK-04
// (#1719):
// „alle Flächen verwenden denselben Pfadvertrag".
//
// Hier standen sechs gleichlautende Zeilen — Fenster, Kommandozeile,
// Dienst, MCP und drei Nebenwege —, und die Ordnerprüfung hätte an jeder
// einzeln sitzen müssen. Vergessen an einer, hätte eine Fläche einen Ordner
// angenommen, den die anderen abweisen. `einstellungsverwalter_test.go`
// hält, dass es eine bleibt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einstellungsverwalterUeber baut den Verwalter über einem Bestand — mit
// der Naht zum Menschen, der Frage an die Platte und dem Riegel der
// Netzwege.
//
// **Der Riegel kommt aus demselben Bestand** (#2000): wohin diese Anwendung
// sendet, setzt nur, wer senden darf, und ein Token darf das nie. Weil jede
// Fläche hier vorbeikommt, gilt er an allen vier.
func einstellungsverwalterUeber(db *sql.DB) *service.Einstellungsverwalter {
	return service.NeuerEinstellungsverwalter(storage.NeueEinstellungen(db)).
		MitNutzerwerten(storage.NeueNutzerablage(db)).
		MitOrdnerpruefung(pruefeOrdnerFuer).
		MitVerschluesselungsstand(verschluesselungsstandUeber(db)).
		MitRiegel(AutorisierungAus(db))
}

// verschluesselungsstandUeber fragt die Datei des Bestands, ob sie
// verschlüsselt ist — für jede Fläche dieselbe Frage.
func verschluesselungsstandUeber(db *sql.DB) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		pfad, fehler := storage.Bestandsdatei(ctx, db)
		if fehler != nil {
			return false, fehler
		}

		return storage.IstVerschluesselt(pfad)
	}
}

// MessungEinrichten richtet die Spur nach den Einstellungen des Geräts ein
// (OPS-05, #1722) —
// mit dem gezeigten Profil, denn ohne es liest der Bestand keine.
//
// **Ein Lesefehler lässt die Vorgabe stehen** — die Spur bleibt aus — und
// kommt zurück: die Wurzel entscheidet, ob sie deswegen nicht startet.
func MessungEinrichten(ctx context.Context, db *sql.DB) error {
	gezeigt, fehler := storage.MitGezeigtemProfil(ctx, db)
	if fehler == nil {
		var werte map[string]string

		werte, fehler = einstellungsverwalterUeber(db).Alle(gezeigt)
		service.SpurEinrichten(domain.MessregelAus(werte))
	}

	return fehler //nolint:wrapcheck // die Wurzel reicht ihn weiter, wie er ist
}

// ZwischenspeicherPflegen hält die Grenze auf der Platte ein — beim Start
// eines Servers, wie [App] es beim Start des Fensters tut (Nachlauf über
// Phase 6.4, N9: ein `mse-rest`, das Wochen lief, räumte nie).
//
// **Ein Lesefehler lässt die Vorgabe gelten** und kommt zurück, wie bei
// [MessungEinrichten].
func ZwischenspeicherPflegen(ctx context.Context, db *sql.DB) error {
	gezeigt, fehler := storage.MitGezeigtemProfil(ctx, db)

	var werte map[string]string
	if fehler == nil {
		werte, fehler = einstellungsverwalterUeber(db).Alle(gezeigt)
	}

	pflegeZwischenspeicher(domain.ZwischenspeichergrenzeAus(werte[domain.ZwischenspeicherGrenze]))

	return fehler //nolint:wrapcheck // die Wurzel reicht ihn weiter, wie er ist
}

// GrenzenUeber gibt die Grenzen, die ein Dienst über diesem Bestand anwendet:
// die des Maschinenhaushalts, wie die Einstellungen des gezeigten Profils sie
// wollen — für Dienst und Werkzeugserver beim Start, für die Kommandozeile
// bei jeder Frage (QA-08.AK-03).
//
// **Ein Lesefehler lässt die Vorgabe gelten** — mit Guthaben — und kommt
// zurück, wie bei [MessungEinrichten] und [ZwischenspeicherPflegen] — und
// steht deshalb neben ihnen (`grenzen.go` sagt, was daraus wird).
func GrenzenUeber(ctx context.Context, db *sql.DB) (domain.Haushaltsgrenzen, error) {
	grenzen := service.Maschinenhaushalt().Grenzen()

	gezeigt, fehler := storage.MitGezeigtemProfil(ctx, db)

	var werte map[string]string
	if fehler == nil {
		werte, fehler = einstellungsverwalterUeber(db).Alle(gezeigt)
	}

	if fehler != nil {
		return grenzen, fehler //nolint:wrapcheck // die Wurzel reicht ihn weiter, wie er ist
	}

	return grenzen.NachEinstellungen(werte), nil
}

// messungEinrichten richtet die Spur des Fensters nach seinen Einstellungen
// — ein Lesefehler lässt sie, wie sie ist.
func (a *App) messungEinrichten(ctx context.Context) {
	if werte, fehler := a.einstellungen.Alle(ctx); fehler == nil {
		service.SpurEinrichten(domain.MessregelAus(werte))
	}
}
