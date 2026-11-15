// **Die zweite, nur lesende Verbindung** — Werkstück 16 von Phase 4.12
// (#1351),
// nach [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// **Ein Composition Root** (§2): hier wird entschieden, **welche** Verbindung
// die Repositories bekommen. Ob jemand nur lesen darf, entscheidet
// `internal/service`.

package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// leserverbindungRichten nimmt die zweite Verbindung, wenn hier nur gelesen
// werden darf — und gibt sie wieder her, wenn nicht.
//
// # Wann sie steht
//
// **Genau dann, wenn die Rolle im gezeigten Profil nicht bis zum Schreiben
// reicht** — `leser` oder gar keine. Der Eigner ist `verwalter`, ein
// `mitarbeiter` schreibt; wer weniger hat, soll nichts ändern.
//
// # Warum sie **nach** der Migration entsteht
//
// **Weil `Migriere` unbedingt eine Transaktion beginnt** — auch wenn nichts zu
// tun ist. Gemessen am 2026-09-06: auf einer Verbindung mit `query_only`
// endet schon das Beginnen mit „attempt to write a readonly database". Der
// Bestand wird deshalb gewöhnlich geöffnet, migriert, und **erst dann** kommt
// die zweite dazu.
//
// # Und warum ein Fehlschlag den Start nicht aufhält
//
// **Weil er ihn nicht aufhalten darf, ohne mehr kaputtzumachen.** Wer die
// Rolle nicht lesen kann, hat einen Bestand, der klemmt — und eine Anwendung,
// die deswegen gar nicht startet, tauscht einen kleinen Mangel gegen einen
// großen (§8, dieselbe Abwägung wie beim Paketkatalog).
//
// **Was das kostet, gehört gesagt**: in diesem Fall wird geschrieben, wo
// vielleicht nicht geschrieben werden sollte. Deshalb steht der Fehler im
// Protokoll und nicht nur im Nichts.
func (a *App) leserverbindungRichten(ctx context.Context) {
	// **Ohne eigene Verbindung gibt es hier nichts zu richten** — warum,
	// steht an [App.eigeneVerbindung].
	//
	// **Ohne diese Zeile stürzte `mse profil <id>` ab**: `verdrahten` baut die
	// Leute über [App.db], und die war `nil`. Gefunden vom Flächenlauf zu
	// Werkstück 16.
	if !a.eigeneVerbindung() {
		return
	}

	// **Die Revision zuerst, die Rolle danach** — ID-03.AK-02. So ist die
	// Rolle, nach der gerichtet wird, mindestens so neu wie die vermerkte
	// Revision; bewegt sich dazwischen etwas, richtet der nächste Aufruf
	// noch einmal ([App.verbindungNachziehen]).
	//
	// **Vermerkt wird nur, was gelang.** Wer die Rolle nicht lesen oder die
	// zweite Verbindung nicht öffnen konnte, hat nicht gerichtet — der
	// nächste Aufruf soll es noch einmal versuchen, statt den alten Stand
	// für aktuell zu halten.
	revision, revisionsfehler := storage.NeueNutzerablage(a.db).Revision(ctx)
	if a.verbindungRichten(ctx) && revisionsfehler == nil {
		a.gerichtet.Store(revision)
	}
}

// eigeneVerbindung sagt, ob diese Anwendung ihre Verbindung selbst geöffnet
// hat — und damit auch die zweite richten darf.
//
// **Die Flächen ohne Fenster bauen über `NeueApp("")`**, und ihre Verbindung
// gehört dem Composition Root daneben (`cmd/mse`, `cmd/mse-mcp`,
// `cmd/mse-rest`). Wer dort die Leseart bestimmt, tut es beim Öffnen, nicht
// beim Wechseln — ein Befehl endet danach ohnehin. Eine Nutzerverwaltung,
// die die Rolle wüsste, haben die meisten von ihnen nicht.
//
// # Warum nicht mehr „[App.db] ist leer"
//
// **Weil das nicht mehr stimmt.** Hier stand genau diese Frage, und sie traf
// zu, solange keine Fläche ohne Fenster das Feld setzte. Heute tun es zehn
// — die Partiefläche für die Leihstelle, Regel-, Netz-, Peer- und
// Umschlagfläche und die Nähte zu Markt, Anbietern, KI, Preisen und Suche —,
// denn sie bauen ihre Ablagen daraus. Sobald sich die Revision der Rechte
// bewegt hatte, fragte der nächste Ruf nach der Rolle und fiel über einen
// Nullzeiger: `mse partie-anlegen` nach `mse profil <id>`, gefunden vom
// Flächenlauf zu #1819.
//
// **Der Pfad ist das Merkmal**, denn nur [App.Starten] öffnet eine
// Verbindung, und es öffnet sie unter [App.datenbankpfad] — dort, wo auch
// die zweite aufginge. Ohne Pfad gibt es diesen Ort nicht.
func (a *App) eigeneVerbindung() bool {
	return a.db != nil && a.datenbankpfad != ""
}

// verbindungRichten nimmt die zweite Verbindung oder gibt sie her — und
// sagt, ob das gelang.
func (a *App) verbindungRichten(ctx context.Context) bool {
	nurLesend, fehler := a.nurLesendGefragt(ctx)
	if fehler != nil {
		slog.Default().ErrorContext(ctx,
			"die Rolle im gezeigten Profil ließ sich nicht lesen",
			"grund", fehler.Error())

		return false
	}

	if !nurLesend {
		a.leserverbindungSchliessen()

		return true
	}

	// **Schon da?** Dann bleibt sie — ein zweites Öffnen gäbe eine zweite
	// Verbindung auf dieselbe Datei, und die erste bliebe offen.
	if a.nurLesen != nil {
		return true
	}

	zweite, oeffnungsfehler := storage.OeffnenNurLesend(ctx, a.datenbankpfad)
	if oeffnungsfehler != nil {
		slog.Default().ErrorContext(ctx, "die nur lesende Verbindung ließ sich nicht öffnen",
			"grund", oeffnungsfehler.Error())

		return false
	}

	a.nurLesen = zweite
	a.verdrahten(zweite)

	return true
}

// nurLesendGefragt sagt, ob hier nur gelesen werden darf.
//
// **Ohne Riegel gegen eine fehlende Nutzerverwaltung**: beide Aufrufer haben
// sie — [App.Starten] verdrahtet sie vor der Frage, und [NurLesendGefragt]
// baut die Fläche, die sie setzt. Ein Zweig für einen Fall, den es nicht
// gibt, war eine Lücke in der Abdeckung, die niemand schließen konnte (§3).
func (a *App) nurLesendGefragt(ctx context.Context) (bool, error) {
	rolle, fehler := a.profilwechsel.RolleImGezeigten(ctx)
	if fehler != nil {
		return false, fehler
	}

	return !schreibtMit(rolle), nil
}

// schreibtMit sagt, ob eine Rolle bis zum Schreiben reicht.
//
// # Warum über `Akteur.Darf` und nicht über einen eigenen Vergleich
//
// **Zwei Antworten auf dieselbe Frage waren der Befund** — ID-03, nativer
// Befund F01 (#1702).
// Hier stand „ist die Rolle genau `leser`?", mit der Begründung, wer gar keine
// habe, komme nicht hinein. Er kommt hinein: ein Eigner ohne zweiten Faktor
// lässt jeden herein (`service.Nutzerverwaltung.Zutritt`), und wer so kommt,
// hat dort die Rolle `""`. Das Banner (`Arbeitskontext`) fragte
// `Akteur.Darf(WirkungSchreiben)` und sagte „nur lesen" — die Verbindung
// fragte nach `leser` und blieb schreibend. In 0.5.0 entstanden so Decks,
// Sets und Turniere in einem Profil, in dem niemand eine Freigabe hatte.
//
// **Jetzt fragen beide dieselbe Regel**, und sie steht in `domain`.
func schreibtMit(rolle domain.Rolle) bool {
	return domain.Akteur{Rolle: rolle}.Darf(domain.WirkungSchreiben) == nil
}

// leserverbindungSchliessen gibt die zweite Verbindung her und verdrahtet
// wieder über die schreibende.
//
// **Die Reihenfolge ist umgekehrt zum Öffnen**: erst verdrahten, dann
// schließen. Wer zuerst schlösse, hätte einen Augenblick, in dem die
// Repositories auf einem geschlossenen Griff sitzen.
func (a *App) leserverbindungSchliessen() {
	if a.nurLesen == nil {
		return
	}

	a.verdrahten(a.db)

	if fehler := a.nurLesen.Close(); fehler != nil {
		slog.Default().Error("die nur lesende Verbindung ließ sich nicht schließen",
			"grund", fehler.Error())
	}

	a.nurLesen = nil
}

// NurLesendGefragt sagt, ob über dieser Verbindung nur gelesen werden darf.
//
// **Exportiert, weil es vier Composition Roots gibt** (§2) — dasselbe
// Argument wie bei `SicherungVorMigration` nebenan: die Frage gehört an eine
// Stelle, und was eine Fläche vergisst, ist keine Zusage.
//
// # Über die vorhandene Wurzel und nicht über eine zweite
//
// **Der erste Anlauf baute hier seine eigene Nutzerverwaltung** — und der
// Wächter über die Wurzeln strich es an: sie kam ohne `MitFaktor`. Eine
// Wurzel, die eine Naht vergisst, ist der Fall, für den es ihn gibt, und die
// Ausnahme wäre zu breit gewesen: sie schaltete ihn für die ganze
// Nutzerverwaltung ab.
//
// **`NutzerflaecheUeber` ist die Wurzel, die es schon gibt** — dieselbe, die
// jede kopflose Fläche für Nutzer und Profile nimmt.
func NurLesendGefragt(ctx context.Context, db *sql.DB) (bool, error) {
	return NutzerflaecheUeber(db).app.nurLesendGefragt(ctx)
}

// OhneLeserwechsel weist einen Wechsel ab, nach dem nur noch gelesen werden
// dürfte.
//
// # Wofür es das gibt
//
// **Für die beiden Flächen, die ihre Verbindung nicht wechseln können** —
// Dienst und MCP-Server halten eine über ihre ganze Laufzeit, und jede Naht
// hat sie sich beim Aufbau abgeholt. Sie mitten im Betrieb auf eine zweite
// umzuhängen hieße, sie alle neu zu bauen.
//
// **Das ist wörtlich der Grund, aus dem dort schon das Zurückspielen fehlt**
// und aus dem die Verschlüsselung dort nicht geschaltet wird
// ([0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)).
//
// **Was es kostet, gehört gesagt**: wer über MCP oder den Dienst in ein Profil
// wechseln will, in dem er nur `leser` ist, bekommt eine Absage — und tut es
// am Fenster oder mit `mse profil`.
//
// **Beim Start greift es nicht**: steht der Zeiger schon dort, öffnet der
// Prozess seine Verbindung gleich nur lesend. Der Schutz gilt also überall;
// nur der **Wechsel** dorthin geht an zwei Flächen nicht.
func (f *Nutzerflaeche) OhneLeserwechsel() *Nutzerflaeche {
	f.app.ohneLeserwechsel = true

	return f
}

// leserwechselAbgewiesen sagt, ob dieser Wechsel abzulehnen ist.
//
// **Gefragt wird nach der Rolle im ZIEL und nicht im gezeigten Profil** — der
// Wechsel hat noch nicht stattgefunden, und was danach gälte, hängt am Ziel.
// Der erste Anlauf fragte das gezeigte und hätte damit die Lage von vorher
// beurteilt.
func (a *App) leserwechselAbgewiesen(ctx context.Context, ziel int64) error {
	if !a.ohneLeserwechsel || a.freigaben == nil {
		return nil
	}

	werArbeitet, fehler := a.arbeitenderNutzer(ctx)
	if fehler != nil {
		return fehler
	}

	profil, suchfehler := a.profilMitKennung(ctx, ziel)
	if suchfehler != nil {
		return suchfehler
	}

	rolle, rollenfehler := a.freigaben.Rolle(ctx, werArbeitet, profil)
	if rollenfehler != nil {
		return rollenfehler
	}

	if schreibtMit(rolle) {
		return nil
	}

	return fmt.Errorf(
		"%w: in diesem Profil dürftest du nur lesen, und diese Fläche kann "+
			"nicht auf eine nur lesende Verbindung umschalten — sie hält eine "+
			"über ihre ganze Laufzeit. Am Fenster oder mit »mse profil« geht es",
		domain.ErrErlaubnis)
}
