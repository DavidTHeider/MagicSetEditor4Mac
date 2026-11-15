package app

import (
	"context"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// imProfil hängt an einen Kontext das Profil, dessen Daten diese Fläche zeigt.
//
// # Warum an jeder Bindung und nicht ein einziges Mal
//
// **Den Kontext gibt Wails her, nicht dieses Paket.** Er entsteht in der
// Laufzeit und erreicht jede Bindung als erstes Argument; es gibt keine Stelle,
// an der er hier gebaut würde und an der man ihn einmal ausstatten könnte.
//
// **Und der Kontext ist auch der richtige Träger.** Eine Fläche mit Anmeldung
// sieht je Anfrage ein anderes Profil — REST bekommt das in Werkstück 4. Ein
// Wert am Prozess könnte das nie, und ein Wert an der Ablage hieße, ihn bei
// jedem Wechsel an zwanzig Stellen nachzuziehen.
//
// **Gehalten wird er atomar**, weil der Wechsel aus einer Bindung kommt und
// die Leser aus jeder anderen: Wails ruft Bindungen aus mehreren Goroutinen,
// und ein blankes `int64` wäre ein Wettlauf, den `-race` findet und ein
// Anwender als falsche Daten sähe. **Null heißt „noch nicht geladen"** —
// `profile.id` ist `AUTOINCREMENT` und beginnt bei 1; jede Abfrage weist eine
// Null ab (§8), statt sie für ein Profil zu halten.
//
// **Ein Wächter zählt nach**, dass keine Bindung die Zeile vergisst —
// `internal/werkzeuge/profilnaht_test.go`. Ohne ihn wäre eine neue Bindung
// still profillos: sie liefe gegen [storage.ProfilAus] auf einen Fehler, und
// der sähe aus wie ein Bestandsfehler.
func (a *App) imProfil(ctx context.Context) context.Context {
	// **Eine Null überschreibt nichts.** Sie heißt „noch nicht geladen" — und
	// ein Kontext, der schon eines trägt, ist dann die bessere Auskunft als
	// eine Zahl, die keine ist. Im Fenster kommt der Fall nicht vor: Wails
	// reicht einen nackten Kontext herein, und [App.Starten] hat den Zeiger
	// vorher gelesen. Er kommt vor, wo eine Fläche ihren Wurzelkontext selbst
	// ausstattet — [storage.MitGezeigtemProfil] — und in Tests, die keine
	// Anwendung starten.
	// **Lesend gehalten, solange die Kennung gelesen wird** — Befund 1 des
	// Zweitblicks über die Phasen 4.0 bis 4.13. Ein Wechsel, der gerade
	// umbaut, hält hier an; wer schon durch ist, arbeitet mit einer Kennung,
	// zu der die Nähte passen.
	a.verbindungNachziehen(ctx)

	// **Und ein abgebrochener Ruf wartet nicht weiter** (Zweitblick über A
	// und B, Befund 4): ein Wechsel hält die Sperre exklusiv, und ohne
	// Kontext wartete der Ruf, bis er fertig war. Er geht dann ohne Profil
	// weiter — der Abbruch steht in seinem Kontext, und das erste SQL sagt
	// es.
	if !sperreMitKontext(ctx, a.umschaltung.TryRLock, a.umschaltung.RLock, a.umschaltung.RUnlock) {
		return MitNeuemVorgang(a.imProfilOhneSperre(ctx))
	}

	defer a.umschaltung.RUnlock()

	// **Und hier entsteht die Vorgangskennung** — Werkstück O9 von Phase 5.2
	// (#1433). Diese Zeile steht in **jeder** Bindung des Fensters, und ein
	// Wächter hält das (`internal/werkzeuge/profilnaht_test.go`): damit ist
	// sie die Grenze, an der ein Ruf aus dem Fenster ankommt. Die zweite
	// Hälfte ist [App.kontext] — für die Bindungen, die keinen Kontext
	// bekommen.
	ausgestattet := MitNeuemVorgang(a.imProfilOhneSperre(ctx))

	// **Und auf der ausführlichen Stufe steht jeder Ruf** — Befund 5 des
	// Zweitblicks über Phase 6.4: sie hatte keine einzige Zeile, und die
	// Diagnose sagte trotzdem „ausführlich an". Genannt wird die Bindung —
	// eine von endlich vielen, kein Inhalt —, und die Kennung setzt der
	// schwärzende Schreiber. **Gefragt wird vorher**, damit der Name des
	// Aufrufers auf der üblichen Stufe nichts kostet.
	if slog.Default().Enabled(ausgestattet, slog.LevelDebug) {
		if stelle := aufrufer(); !imTakt(stelle) {
			slog.DebugContext(ausgestattet, "Ruf des Fensters", "stelle", stelle)
		}
	}

	return ausgestattet
}

// gezeigtesProfil gibt das Profil, in dem dieser Ruf arbeitet.
//
// # Warum aus dem Kontext und nicht aus [App.profil]
//
// **Weil das Feld nur das Fenster hat.** Kommandozeile, MCP und REST bauen
// ihren Wurzelkontext selbst und stellen ihr Profil dort hinein
// ([storage.MitGezeigtemProfil]); ihr Feld bleibt null. Wer eine Regel am
// Feld festmacht, hat sie an **einer** von vier Flächen — und §16 verlangt
// überall dieselbe Antwort. Nachgemessen an `PartieAusWuenschen`: mit dem
// Feld saß dort jeder beim Wirt, gleich woran sein Sitz hing.
//
// # Und warum sie hier steht
//
// **Weil `internal/schichten` den Import von `internal/storage` auf eine
// Aufzählung beschränkt** (§2, Composition Roots). Die Aufzählung wird nicht
// länger, nur weil ein weiterer Weg das Profil braucht — sie steht dort, wo
// die Profilnaht ohnehin liegt.
func gezeigtesProfil(ctx context.Context) (int64, error) {
	return storage.ProfilAus(ctx)
}

// imProfilOhneSperre hängt das Profil an, ohne die Umschaltsperre zu nehmen.
//
// **Für den Wechsel selbst** (Befund 1 des Zweitblicks über die Phasen 4.0 bis
// 4.13): er hält sie schon schreibend, und ein Ruf an [App.imProfil] klemmte
// dort gegen sich selbst — ein `RWMutex` ist nicht wiedereintrittsfähig.
//
// **Und eine verschlossene Sitzung bekommt keines** — ID-02 (#1701):
// gesperrt oder abgemeldet erreicht keine Abfrage den Bestand eines Profils,
// gleich, was das Fenster fragt. Die Oberfläche zeigt dann den
// Sperrbildschirm; diese Zeile hält, was er verspricht.
func (a *App) imProfilOhneSperre(ctx context.Context) context.Context {
	if a.sitzungVerschlossen() {
		return ctx
	}

	return a.imProfilTrotzSitzung(ctx)
}

// imProfilTrotzSitzung hängt das Profil an, auch wenn die Sitzung
// verschlossen ist — **für die zwei Schritte, die sie öffnen**: ihren Beginn
// und den Nutzerwechsel beim Anmelden. Beide brauchen den Zeiger, der im
// Profil steht.
func (a *App) imProfilTrotzSitzung(ctx context.Context) context.Context {
	profilID := a.profil.Load()
	if profilID == 0 {
		return ctx
	}

	return storage.MitProfil(ctx, profilID)
}

// profilLaden liest den Zeiger aus dem Bestand und merkt ihn sich.
//
// **Ein Fehler hält den Start nicht auf.** Der Zeiger ist eine Bequemlichkeit —
// [storage.Nutzerablage.Profilzeiger] fällt auf das erste Profil zurück, und
// gibt es gar keines, ist der Bestand kaputt und das sagt die erste Abfrage
// deutlicher als eine Meldung beim Start. Was hier passiert, ist deshalb ein
// Protokolleintrag und kein Abbruch.
func (a *App) profilLaden(ctx context.Context) {
	// **Ohne Bestand gibt es nichts zu lesen.** Die drei Flächen ohne Fenster
	// bauen ihre Anwendung über `NeueApp("")` und reichen den Bestand je Naht
	// herein; ihr Profil steht im **Wurzelkontext**
	// ([storage.MitGezeigtemProfil]) und nicht in diesem Feld.
	//
	// **Ohne diese Zeile stürzt `ProfilLoeschen` an jeder dieser Flächen ab**
	// — gefunden vom Flächenlauf, nicht von einem Test dieses Pakets
	// (#1351,
	// Werkstück 5).
	if a.db == nil {
		return
	}

	profilID, fehler := storage.NeueNutzerablage(a.db).Profilzeiger(ctx)
	if fehler != nil {
		slog.Default().ErrorContext(ctx, "das gezeigte Profil ließ sich nicht lesen",
			"grund", fehler.Error())

		return
	}

	a.profil.Store(profilID)
}

// profilUmstellen setzt Kennung und Leserverbindung **in einem Zug**.
//
// # Warum es diesen Weg gibt
//
// **Drei Stellen stellen das gezeigte Profil um, und nur eine war gesichert.**
// Angestrichen vom Nachlauf zum Zweitblick über die Phasen 4.0 bis 4.13
// (2026-09-07), Befund 1:
//
//   - [App.ProfilWechseln] nahm die Sperre und richtete die Verbindung neu,
//   - [App.NutzerWechseln] tat dasselbe **ohne** Sperre,
//   - [App.ProfilLoeschen] zog nur die Kennung nach und ließ die Verbindung
//     stehen — nach dem Löschen des aktiven Profils gehörte sie noch zur
//     bisherigen Rolle.
//
// **Die erste Korrektur schloss den Fall und nicht die Klasse** (§10). Hier
// steht sie einmal, und alle drei gehen hindurch.
//
// # Was die Sperre hält und was nicht
//
// **Sie hält jede Bindung an, die [App.imProfil] geht**, solange Kennung und
// Verbindung auseinanderfallen könnten. Sie hält **nicht** den Ruf danach: wer
// schon durch `imProfil` ist, arbeitet mit der Kennung, die er dort bekam, und
// diesen Ruf zu Ende zu führen ist richtig — er gehört zu dem Stand, aus dem
// er stammt (§7, Regel 2).
//
// **Der Kontext wird ohne Sperre gebaut**, weil sie schon genommen ist und
// `sync.RWMutex` nicht wiedereintrittsfähig ist.
func (a *App) profilUmstellen(ctx context.Context, profilID int64) {
	a.profilUmstellenGesperrt(ctx, profilID)

	// **Und die Leiste folgt** (Nachlauf über Phase 6.3, N6): „Zuletzt
	// benutzt" gehört dem Profil, und bis dahin zeigte die Leiste nach einem
	// Wechsel die Sets des vorigen Menschen, bis jemand ein Set öffnete.
	// **Nach der Sperre**, denn die Leiste liest den Bestand.
	a.menueNeu(a.imProfil(ctx))
}

// profilUmstellenGesperrt stellt unter der Sperre um — alles, was dem
// vorigen Profil gehörte, geht mit.
func (a *App) profilUmstellenGesperrt(ctx context.Context, profilID int64) {
	a.umschaltung.Lock()
	defer a.umschaltung.Unlock()

	a.profil.Store(profilID)

	// **Trotz der Sitzung** (ID-02, #1701): die Umstellung ist Aufbau und
	// keine Anfrage, und beim lokalen Anmelden läuft sie, **bevor** die
	// Sitzung aufgeht. Mit der Sperre bekäme die Leseart das Profil des
	// vorigen Menschen.
	a.leserverbindungRichten(a.imProfilTrotzSitzung(ctx))

	a.profilgebundenesRaeumen()
}

// profilNachziehen liest den Zeiger und stellt darauf um.
//
// **Für die Wege, die den Zeiger nicht selbst kennen** — das Löschen eines
// Profils etwa: der Fachkern setzt den Ausweich, und das Fenster erfährt ihn
// erst durch Nachlesen.
func (a *App) profilNachziehen(ctx context.Context) {
	if a.db == nil {
		return
	}

	profilID, fehler := storage.NeueNutzerablage(a.db).Profilzeiger(ctx)
	if fehler != nil {
		slog.Default().ErrorContext(ctx, "das gezeigte Profil ließ sich nicht lesen",
			"grund", fehler.Error())

		return
	}

	a.profilUmstellen(ctx, profilID)
}

// spielbilderRaeumen wirft die gezeichneten Spielbilder weg.
//
// **Getrennt vom Ruf, weil es eines geben kann und nicht muss**: eine
// Anwendung ohne Zeichenwerk — jede Fläche ohne Fenster — hat keines.
func (a *App) spielbilderRaeumen() {
	if a.spielbilder == nil {
		return
	}

	a.spielbilder.Raeume()
}

// verbindungNachziehen richtet die Verbindung neu, wenn sich die Rechte seit
// dem letzten Richten bewegt haben — ID-03.AK-02
// (#1702).
//
// # Was hier gecacht war
//
// **Die Verbindung.** Ob das Fenster über die schreibende oder die nur
// lesende arbeitet, entschied sich beim Wechsel und danach nie wieder. Wer
// eine Freigabe zurückzog, während jemand im Profil stand — aus einem zweiten
// Prozess, über die Kommandozeile oder den Dienst —, sah ihn mit der Rolle
// von vorher weiterschreiben. Das Banner sagte längst „nur lesen": es fragt
// bei jedem Ruf neu.
//
// # Woran es sich bemerkt
//
// **An der Revision der Rechte**, die jede Änderung an Zeigern und Freigaben
// in derselben Transaktion hochzählt (Wanderung 0047). Sie ist eine Zeile und
// eine Zahl; gelesen wird sie bei **jedem** Ruf über `imProfil`, gerichtet
// wird nur, wenn sie sich bewegt hat.
//
// # Was es nicht hält
//
// **Einen Ruf, der schon läuft.** Er gehört zu dem Stand, aus dem er stammt
// (§7, Regel 2; 0078: „laufende Arbeit behält ihren Ursprung"). Der nächste
// sieht den Entzug.
func (a *App) verbindungNachziehen(ctx context.Context) {
	if !a.eigeneVerbindung() {
		return
	}

	revision, fehler := storage.NeueNutzerablage(a.db).Revision(ctx)
	if fehler != nil && ctx.Err() != nil {
		// **Ein abgebrochener Ruf ist kein klemmender Bestand** (BK-01,
		// #2109): er sagt es selbst, und als Fehler stünde er im Protokoll,
		// wann immer sich ein Fenster schließt.
		return
	}

	if fehler != nil {
		// **Kein Abbruch**: wer die Revision nicht lesen kann, hat einen
		// Bestand, der klemmt — der Ruf selbst sagt es deutlicher. Dieselbe
		// Abwägung wie in [App.leserverbindungRichten].
		slog.Default().ErrorContext(ctx, "die Revision der Rechte ließ sich nicht lesen",
			"grund", fehler.Error())

		return
	}

	if revision == a.gerichtet.Load() {
		return
	}

	// **Ohne zweiten Blick unter der Sperre**: haben zwei Rufe dieselbe
	// Bewegung bemerkt, richtet der zweite noch einmal — und findet die
	// Verbindung schon so, wie sie sein soll. Richten ist wiederholbar.
	if !sperreMitKontext(ctx, a.umschaltung.TryLock, a.umschaltung.Lock, a.umschaltung.Unlock) {
		return
	}

	defer a.umschaltung.Unlock()

	a.leserverbindungRichten(a.imProfilOhneSperre(ctx))
}

// mitBekannten legt die bekannten Werte eines Entwurfs in den Kontext — keine,
// keine Bedingung (U02b, `entwurfsannahme.go`). **Hier**, neben [App.imProfil]:
// beide setzen, was der Speicher aus dem Kontext liest, und diese Datei darf
// ihn kennen (`internal/schichten`).
func mitBekannten(ctx context.Context, bekannt []string) context.Context {
	if len(bekannt) == 0 {
		return ctx
	}

	return storage.MitBekanntenWerten(ctx, bekannt)
}
