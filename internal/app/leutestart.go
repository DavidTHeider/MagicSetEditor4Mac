// **Wer hier arbeitet und wie er hereinkommt** — der Aufbau der Nähte aus
// Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `start.go` mit der sechsten Naht über die
// Zeilengrenze aus §2 ginge — dieselbe Form wie `deckstart.go` und
// `anmeldestart.go`. Der Schnitt folgt der Frage: dort steht, **was die
// Anwendung hat**, hier, **wer sie benutzt**.

package app

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// leuteStarten setzt die Nähte um Nutzer, Profile und Konten.
//
// # Zwei Verbindungen, und der Schnitt ist die Aussage
//
// **`schreibend` ist immer die schreibende Verbindung**, `arbeit` die, über
// die das gezeigte Profil bearbeitet wird — bei einem `leser` in einem fremden
// Profil ist das die nur lesende zweite (Werkstück 16 von Phase 4.12).
//
// **Was dem MENSCHEN gehört, hängt an `schreibend`**: er selbst, seine
// Profile, seine Konten, seine Faktoren, seine Einstellungen und der Zeiger.
// Sonst fände niemand aus einem fremden Profil wieder hinaus — der Weg hinaus
// schreibt den Zeiger.
//
// **Was in ein PROFIL schreibt, hängt an `arbeit`**: Zugriffstoken,
// Profilvorgaben und die Deckübernahme. **Die Ausnahme war bis zum 2026-09-06
// zu breit**, und der Nachlauf zum Zweitblick über Phase 4.12 hat es
// angestrichen: ein Leser konnte im fremden Profil ein **Token** anlegen — und
// ein Token trägt Rechte. Gemessen an der laufenden Anwendung, nicht
// überlegt.
//
// **Freigaben bleiben an `schreibend`, und das ist keine Lücke**: sie prüfen
// die Rolle selbst (`App.alsVerwalter`), und wer auf seinem **eigenen** Profil
// etwas freigeben will, soll das auch aus einem fremden heraus können.
func (a *App) leuteStarten(schreibend, arbeit *sql.DB) {
	db := schreibend
	// **Nutzer und Profile** — Werkstück 1 von Phase 4.12 (#1351). Sie hängen
	// an nichts anderem: eine Bestandsdatei hat sie, seit Wanderung 0034
	// gelaufen ist.
	// **Eine Verwaltung für beide Nähte.** Sie kann beides, und zwei Aufbauten
	// wären zwei Stände desselben Zeigers.
	nutzerverwaltung := service.NeueNutzerverwaltung(storage.NeueNutzerablage(db))
	a.MitNutzern(nutzerverwaltung)
	a.MitProfilwechsel(nutzerverwaltung)

	// **Und wer die Profile kennt** (G8,
	// #1465):
	// ein Sitz darf an eines gebunden sein, und ein Profil darf verschwinden.
	// Was dann bleibt, ist ein Befund — und dafür muss jemand nachsehen können.
	a.MitProfilkenner(storage.NeueNutzerablage(db))
	anmeldungVerdrahten(a, db)
	nutzerverwaltung.MitFaktor(a.faktor)

	// **Und was beim Löschen mit weggeht** — Befund 2 des Zweitblicks über die
	// Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// Ein gelöschter Nutzer ließ sein Bild und sein TOTP-Geheimnis zurück,
	// während die Datenschutzhilfe „samt allem, was an ihm hängt" versprach.
	// **Hier ist der Ort, an dem beide bekannt sind**: der Schlüsselbund
	// gehört der Plattform, das Bilderverzeichnis dem Zuhause.
	//
	// **Nach `anmeldungVerdrahten`**, denn erst danach steht der Bestandsname,
	// unter dem das Geheimnis im Bund liegt.
	nutzerverwaltung.MitBeiwerk(beiwerkNaehte(a, db))

	// **Und die Einstellungen, die dem Menschen gehören** — Werkstück 8
	// desselben Issues. **Über dieselbe Ablage**: sie liest aus derselben
	// Tabelle wie die Verwaltung darüber, und ein zweiter Aufbau wäre ein
	// zweiter Stand.
	a.MitNutzereinstellungen(
		service.NeueNutzereinstellungen(storage.NeueNutzerablage(db)))
	a.MitPersonenauskunft(
		service.NeuePersonenauskunft(storage.NeueNutzerablage(db)))

	// **Die Zugriffstoken technischer Profile** — Werkstück 10 desselben
	// Issues. **Eine eigene Ablage**: sie liest eine eigene Tabelle, und die
	// Nutzerablage hat mit ihr nichts zu tun.
	// **Über `arbeit`**: ein Token gehört einem **Profil** und trägt Rechte
	// darin. Wer dort nur lesen darf, legt keines an.
	a.MitToken(service.NeueTokenverwaltung(storage.NeueZugriffstokenablage(arbeit)))

	// **Und wer in welches fremde Profil darf** — Werkstück 11 desselben
	// Issues. **Dieselbe Nutzerablage**: die Rolle hängt am Eigner eines
	// Profils, und ein zweiter Aufbau wäre ein zweiter Stand.
	freigaben := service.NeueFreigabenverwaltung(
		storage.NeueFreigabenablage(db), storage.NeueNutzerablage(db))

	a.MitFreigaben(freigaben)
	nutzerverwaltung.MitFreigaben(freigaben)
	a.MitAutorisierung(autorisierungUeber(db, freigaben))

	// **Und die Deckübernahme aus einem fremden Profil** — Werkstück 12
	// desselben Issues. **Sie bekommt dieselbe Freigabenverwaltung**: sie
	// verlangt dieselbe Rolle wie der Wechsel, und eine zweite wäre ein
	// zweiter Stand.
	// **Die Ablage über `arbeit`**: die Übernahme **schreibt** in das gezeigte
	// Profil. Die Entnahme liest aus dem fremden und darf an `schreibend`
	// bleiben — sie ändert dort nichts.
	a.MitDeckuebernahme(service.NeueDeckuebernahme(
		storage.NeueDeckentnahme(db), storage.NeueDeckablage(arbeit),
		nutzerverwaltung, freigaben))

	// **Und die Vorgaben, die auf Zeilen zeigen** — Werkstück 15 desselben
	// Issues. **Ein Fehlschlag hält den Start nicht auf**: das Angebot braucht
	// das Paketverzeichnis, und ohne Zuhause gibt es keines — die Anwendung
	// arbeitet trotzdem, nur ohne diese eine Seite (§16, Satz 5).
	// **Über `arbeit`**: eine Vorgabe sagt, womit ein neuer Datensatz **in
	// diesem Profil** anfängt.
	if verwaltung, fehler := vorgabenverwaltungUeber(arbeit); fehler == nil {
		a.MitVorgaben(verwaltung)
	}
}

// beiwerkNaehte gibt die beiden Wege, auf denen ein gelöschter Nutzer sein
// Bild und sein Geheimnis verliert.
//
// **Eine eigene Funktion, weil beide aus verschiedenen Welten kommen**: das
// Bild liegt im Zuhause, das Geheimnis im Schlüsselbund des Rechners.
//
// # Ein unbekannter Ort ist kein stiller Erfolg
//
// **Hier stand `ort = ""`** — und `Bildablage.wegnehmen` gibt bei leerem Ort
// `nil` zurück. Ein Nutzer wurde damit gelöscht, sein Foto blieb liegen, und
// **niemand erfuhr es**: die Naht meldete Erfolg. Angestrichen vom Nachlauf
// zum Zweitblick über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 2.
//
// **Seither sagt sie es.** Das Löschen der Zeile geschieht trotzdem — der
// Mensch will gelöscht werden, und ein Zuhause, das gerade nicht auffindbar
// ist, hält ihn nicht auf (§7, Regel 3). Was sich ändert, ist die Antwort:
// sie nennt die Datei, die nachzusehen ist.
func beiwerkNaehte(a *App, db *sql.DB) (service.Nutzerbildablage, service.Nutzergeheimnisse) {
	geheimnisse := faktorgeheimnisse{
		bund: a.bundOderAnmeldebund(), kennung: bestandskennung(db),
	}

	ort, fehler := a.bilderort()
	if fehler != nil {
		return ortlos{fehler: fehler}, geheimnisse
	}

	return service.NeueBildablage(ort), geheimnisse
}

// ortlos ist die Bildablage für ein Zuhause, das sich nicht bestimmen ließ.
//
// **Sie nimmt nichts weg und sagt warum** — der Unterschied zur leeren Ablage,
// die vorher hier stand.
type ortlos struct{ fehler error }

func (o ortlos) NutzerbildWegnehmen(int64) error { return o.nennt("Bild") }

func (o ortlos) TischbildWegnehmen(int64) error { return o.nennt("Tischbild") }

func (o ortlos) FensterbildWegnehmen(int64) error { return o.nennt("Fensterbild") }

func (o ortlos) ProfilbadgeWegnehmen(int64) error { return o.nennt("Abzeichen") }

func (o ortlos) nennt(was string) error {
	return fmt.Errorf("das %s bleibt liegen, der Bilderordner ließ sich "+
		"nicht bestimmen: %w", was, o.fehler)
}

// bestandskennung gibt die Kennung und merkt sie sich.
//
// **Einmal gelesen und dann gehalten**: sie ändert sich nie (Wanderung 0039),
// und ein Ruf je Geheimnis wäre eine Abfrage je Tastendruck.
//
// **Aber erst beim ersten Gebrauch**, nicht beim Verdrahten — der Grund steht
// bei [faktorgeheimnisse.kennung].
func bestandskennung(db *sql.DB) func() (string, error) {
	var (
		einmal   sync.Once
		gemerkt  string
		gescheit error
	)

	return func() (string, error) {
		einmal.Do(func() {
			gemerkt, gescheit = storage.Bestandskennung(context.Background(), db)
		})

		return gemerkt, gescheit
	}
}

// AutorisierungAus baut die Autorisierung über einem Bestand allein — für
// die Nähte, die keine Freigabenverwaltung in der Hand haben, und für die
// Wurzeln ohne Fenster, die eine Dienstwahl mit Riegel bauen.
func AutorisierungAus(db *sql.DB) *service.Autorisierung {
	return autorisierungUeber(db, service.NeueFreigabenverwaltung(
		storage.NeueFreigabenablage(db), storage.NeueNutzerablage(db)))
}

// autorisierungUeber baut den Dienst, der den Akteur baut — **dieselbe
// Verdrahtung für jede Wurzel** (§16): Zeiger und Revision aus der
// Nutzerablage, die Rolle aus der Freigabenverwaltung, die schon da ist, und
// die Nachweise aus ihrer eigenen Tafel.
func autorisierungUeber(db *sql.DB, freigaben *service.Freigabenverwaltung) *service.Autorisierung {
	return service.NeueAutorisierung(storage.NeueNutzerablage(db), freigaben).
		MitNachweisen(service.NeueNachweise(storage.NeueNachweisablage(db)))
}
