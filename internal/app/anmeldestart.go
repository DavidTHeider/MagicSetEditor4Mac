// **Die Anmeldung verdrahten** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Wo Kennung und Geheimnis liegen
//
// **Dieses Projekt liefert keine Zugangsdaten mit** (Vermerk 0066). Jeder
// Anbieter verlangt eine registrierte Anwendung, GitHub dazu ein Geheimnis;
// ein mitgeliefertes wäre keines — es stünde in einem öffentlichen
// Repositorium und in jedem ausgelieferten Bündel.
//
// **Die Kennung ist kein Geheimnis** und steht deshalb im Bestand: sie geht in
// jeder Anmeldeadresse durch den Browser. **Das Geheimnis liegt im
// Schlüsselbund**, aus demselben Grund wie jeder andere Zugang
// ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)): die
// Bestandsdatei ist die, die man weitergibt.

package app

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Anbieterdienst ist der Name, unter dem die Geheimnisse im Bund liegen.
const Anbieterdienst = "MagicSetEditor4Mac-Anmeldung"

// AnbieterKennungsschluessel gibt den Einstellungsschlüssel einer Kennung.
//
// **Im Bestand und nicht im Bund**: eine `client_id` ist öffentlich, und ein
// Schlüsselbunddialog für einen Wert, der in jeder Adresse steht, wäre eine
// Rückfrage ohne Gegenwert.
//
// # Warum das jetzt durchreicht statt selbst zu rechnen
//
// **Weil hier eine zweite Wahrheit stand.** Der Schlüssel wurde hier gebaut
// und in `internal/domain` nicht geführt — `PruefeEinstellung` wies ihn ab,
// und damit ließ sich die Kennung an **keiner** Fläche setzen. Vier Anbieter,
// ein fertiger OAuth-Lauf, und kein Weg hinein; gefunden vom Auftraggeber am
// 2026-09-06 an der laufenden Anwendung.
//
// **Der Name bleibt**, weil er exportiert ist und Aufrufer hat.
func AnbieterKennungsschluessel(anbieter string) string {
	return domain.AnbieterKennung(anbieter)
}

// zugangsquelleUeber baut die Quelle für Kennung und Geheimnis.
func zugangsquelleUeber(db *sql.DB, bund *schluesselbund.Bund) Zugangsquelle {
	einstellungen := storage.NeueEinstellungen(db)

	return func(ctx context.Context, anbieter string) (Anbieterzugang, error) {
		alle, fehler := einstellungen.Alle(ctx)
		if fehler != nil {
			return Anbieterzugang{}, fehler
		}

		kennung := alle[AnbieterKennungsschluessel(anbieter)]
		if kennung == "" {
			return Anbieterzugang{}, fmt.Errorf(
				"%w: für %s ist keine Kennung eingetragen — dieses Programm "+
					"liefert keine mit, wer sich anmelden will, registriert seine "+
					"eigene Anwendung beim Anbieter",
				domain.ErrValidation, anbieter)
		}

		// **Ein fehlendes Geheimnis ist hier kein Fehler.** Google nennt es
		// optional, Microsoft verbietet es — `anmeldung.Beginne` entscheidet
		// je Anbieter, und diese Stelle soll die Entscheidung nicht vorwegnehmen.
		geheimnis, _ := bund.Lies(ctx, Anbieterdienst, anbieter) //nolint:errcheck // siehe oben

		return Anbieterzugang{Kennung: kennung, Geheimnis: geheimnis}, nil
	}
}

// browserOeffnen macht eine Adresse im Standardbrowser auf.
//
// **`open` und nicht `runtime.BrowserOpenURL`**: jener bräuchte einen echten
// Wails-Kontext, und die drei Flächen ohne Fenster haben keinen. Ein Weg für
// alle vier ist einer weniger, der auseinanderlaufen kann.
func browserOeffnen(ctx context.Context, adresse string) error {
	// #nosec G204 -- die Adresse baut `anmeldung.Beginne` aus festen Endpunkten
	if err := exec.CommandContext(ctx, "open", adresse).Run(); err != nil {
		return fmt.Errorf("den Browser öffnen: %w", err)
	}

	return nil
}

// anmeldungVerdrahten hängt die Anmeldung an eine Anwendung.
//
// **Beide Arten, sich anzumelden, hängen hier** — über einen Anbieter
// (Werkstück 4) und unter einer eigenen Kennung (Werkstück 6). Zwei
// Verdrahtungen wären zwei Ablagen über demselben Bestand, und die
// Zeilengrenze aus §2 ist kein Grund, eine Naht zu zerreißen, die
// zusammengehört.
func anmeldungVerdrahten(anwendung *App, db *sql.DB) {
	ablage := storage.NeueNutzerablage(db)

	// **Der zweite Faktor hängt hier**, weil beide Wege ihn brauchen: die
	// Anmeldung unter einer eigenen Kennung und der Profilwechsel
	// (Werkstück 7, #1351). **Einer für beide** — zwei Prüfer wären zwei
	// Stände desselben Nutzers.
	// **Über `bundOderAnmeldebund` und nicht über den Anmeldebund
	// unmittelbar**: sonst schriebe jeder Lauf, der die Anwendung startet, in
	// den **echten** Schlüsselbund dieses Rechners (§3).
	// **Die Bestandskennung wird einmal gelesen und nicht je Ruf** — aber
	// **erst beim ersten Gebrauch**, seit dem 2026-09-07.
	//
	// Hier stand ein Ruf an dieser Stelle, dessen Fehler verworfen wurde.
	// `verdrahten` läuft **vor** der Migration; bei einem frischen Bestand gab
	// es die Tafel `bestandskennung` noch nicht, die Kennung blieb für die
	// ganze Sitzung leer, und ein damals eingerichteter zweiter Faktor lag im
	// Schlüsselbund unter `.1` statt unter `<kennung>.1` — beim nächsten Start
	// unauffindbar. Nachlauf zum Zweitblick über die Phasen 4.0 bis 4.13
	// (Befund 2).
	faktor := service.NeuerZweiterfaktor(
		ablage,
		faktorgeheimnisse{
			bund:    anwendung.bundOderAnmeldebund(),
			kennung: bestandskennung(db),
		},
		anwendung.sensorOderRechner())

	anwendung.MitZweitemFaktor(faktor)
	anwendung.sitzungsablage = ablage
	anwendung.MitKontoarten(service.NeueKontoanmeldung(ablage).MitFaktor(faktor))

	anwendung.MitAnmeldung(NeueAnmeldenaht(
		service.NeueAnmeldung(ablage),
		zugangsquelleUeber(db, schluesselbund.Anmeldebund()),
		browserOeffnen,
		anmeldung.NeuerRuferMit(netz.Proxyauswahl(anwendung.Proxyquelle())),
	), ablage)
}
