package app

import (
	"context"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Wer hier arbeitet und wie er hereinkommt** — die Nähte zu Nutzern,
// Profilen und Konten, Phase 4.12
// (#1351).
//
// **Eine eigene Datei und ein eingebetteter Verbund** — dieselbe Form wie
// `fremddienste.go` und aus demselben Grund: `anwendung.go` ging mit der
// sechsten Naht über die Zeilengrenze aus §2, und der Schnitt folgt der
// Frage. Dort steht, **woraus** die Anwendung besteht; hier, **wer** sie
// benutzt.
//
// **Eingebettet und nicht als Feld**: Go reicht die Felder eines
// eingebetteten Verbunds durch, und deshalb heißt es weiterhin `a.nutzer` und
// nicht `a.leutenaehte.nutzer`. Ein benanntes Feld wäre derselbe Schnitt mit
// einer Umbenennung an sechzig Stellen — Aufwand ohne Ertrag.

// leutenaehte bündelt, wer hier arbeitet und wie er hereinkommt.
type leutenaehte struct {
	// nutzer führt die Nutzer und ihre Profile — `nutzer.go`.
	nutzer Nutzerfuehrer

	// profilwechsel stellt das gezeigte Profil um — `profilwechsel.go`.
	profilwechsel Profilwechsler

	// anmeldung führt über einen **Anbieter** herein, konten führt die
	// Verbindungen dorthin — `anmeldung.go`, Werkstück 4.
	anmeldung *Anmeldenaht
	konten    Kontenfuehrer

	// kontoarten ist die Naht zu den vier **eigenen** Kontoarten — Adresse,
	// Rufnummer, anonym, ohne Konto (Werkstück 6). Die Anbieter stehen
	// darüber: dort steht, wer jemand **anderswo** ist, hier, wer er **hier**
	// ist.
	kontoarten Kontofuehrer

	// faktor richtet zweite Faktoren ein und prüft sie — Werkstück 7
	// desselben Issues. **Ein Zeiger auf den Dienst und keine Schnittstelle**:
	// die Fassade braucht mehr als eine Frage von ihm, und `service` liegt
	// ohnehin unter dieser Schicht (§2).
	faktor *service.Zweiterfaktor

	// sitzung ist die Sitzung **des Fensters** — ID-02 (#1701). Gesetzt beim
	// Start und nur dort; `nil` heißt: keine Sitzung, also jede Fläche ohne
	// Fenster. **Atomar**, weil der Start sie setzt und jede Bindung sie
	// liest. sitzungsablage ist ihre Naht zum Bestand, gesetzt beim
	// Verdrahten.
	sitzung        atomic.Pointer[service.Sitzung]
	sitzungsablage service.Sitzungsablage

	// biometrie sagt, ob dieser Rechner Touch ID kann — für die **Auskunft**
	// an die Seite, nicht für die Prüfung. Die geht durch den Dienst.
	//
	// **Ein Feld und kein fester Ruf**, damit ein Lauf ohne Sensor eine
	// Antwort setzen kann; `nil` heißt: es wird nichts behauptet.
	biometrie func(context.Context) biometrie.Auskunft

	// sensor prüft sie — `nil` heißt: der dieses Rechners.
	sensor service.Sensor

	// nutzereinstellungen führt, was dem **Menschen** gehört — Werkstück 8
	// desselben Issues. **Neben `einstellungen`, nicht darin**: jene gelten
	// für das Profil, in dem gearbeitet wird.
	nutzereinstellungen Nutzereinstellungsfuehrer

	// personenauskunft gibt heraus, was über einen Menschen gespeichert ist
	// — Punkt 10 desselben Werkstücks.
	personenauskunft Personenauskunfterin

	// freigaben sagt, wer in welches fremde Profil darf — Werkstück 11
	// (#1351).
	freigaben Freigabenfuehrer

	// autorisierung baut den Akteur eines Aufrufs — Werkstück O1 von Phase
	// 5.2 (#1425). **Der konkrete Dienst und keine Schnittstelle**: er hat
	// keinen zweiten Aufbau, und ein Doppel hier prüfte die Übersetzung an
	// einem Akteur, den es so nicht gibt.
	autorisierung *service.Autorisierung

	// wache ersetzt die Autorisierung für Läufe ohne echte Rollen (§3) —
	// Werkstück N7 von Phase 5.4
	// (#1568).
	wache Wache

	// **Die zwei Dienstposten** — Werkstück N6 von Phase 5.4
	// (#1572).
	// Sie laufen nur auf Aufruf und enden mit dem Fenster und beim
	// Profilwechsel.
	dienste dienstfelder

	// dienststarter ersetzt das Starten eines Programms — für Läufe ohne
	// gebautes `mse-rest` (§3).
	dienststarter Dienststarter

	// deckuebernahme holt ein Deck aus einem fremden Profil — Werkstück 12
	// desselben Issues. **Neben `decks` und nicht darin**: sie fasst zwei
	// Profile an und ist die einzige Naht, die das tut.
	deckuebernahme Deckuebernehmer

	// vorgaben führt die Vorgaben eines Profils, die auf **Zeilen** zeigen —
	// Werkstück 15 desselben Issues. **Neben `einstellungen`, nicht darin**:
	// jene halten aufzählbare Werte, diese zeigen auf den Bestand.
	vorgaben Vorgabenfuehrer

	// token führt die Zugriffstoken technischer Profile — Werkstück 10
	// desselben Issues. **Verwaltet wird an jeder Fläche, geprüft nur am
	// REST-Zugang**: er ist der einzige, der auf einem Port lauscht.
	token Tokenfuehrer

	// gerichtet ist die Revision der Rechte, für die die Verbindung zuletzt
	// gerichtet wurde — ID-03.AK-02 (#1702). Warum und wo verglichen:
	// [App.verbindungNachziehen].
	gerichtet atomic.Int64
}
