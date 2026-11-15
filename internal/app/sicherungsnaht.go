// **Die Fassade des Sicherungswegs** — Dialog, DTOs, Fehlerübersetzung
// (#1175).
//
// # Was hier steht und warum
//
// Der Weg selbst liegt in `service.Bestandsweg`. Hier bleibt, was §2 dieser
// Schicht zuweist: der native Dateidialog, die DTOs für das Fenster und die
// Übersetzung des Fehlers. Kein `if` über einen fachlichen Zustand.
//
// **Die Orte reicht die Fassade hinein**, weil sie als einzige Datenbankpfad
// und Bilderverzeichnis zugleich kennt — und den Lebenszyklus der Verbindung
// dazu. Beides sind Nähte des Dienstes, keine Entscheidungen der Fassade.

package app

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die drei Nähte tragen hier ihren Namen weiter, damit Composition Root und
// Tests sie ansprechen können, ohne den Dienst zu importieren.
type (
	// Sicherungspruefer liest eine Sicherungsdatei, ohne sie zu verändern.
	Sicherungspruefer = service.Sicherungspruefer

	// Sicherungsbefund ist das, was eine Sicherung über sich hergibt.
	Sicherungsbefund = service.Sicherungsbefund

	// Alleinpruefung sagt, ob niemand sonst die Bestandsdatei geöffnet hat.
	Alleinpruefung = service.Alleinpruefung
)

// Rueckspielsicherung ist der Name, unter dem der alte Bestand liegen bleibt.
const Rueckspielsicherung = service.Rueckspielsicherung

// Bestandsname ist der Name der Datenbank **im Archiv**.
const Bestandsname = service.Bestandsname

// Beiwerkordner ist der Ordner der eingesetzten Bilder **im Archiv**.
const Beiwerkordner = service.Beiwerkordner

// MitBilderort setzt den Ort der eingesetzten Bilder — für Tests.
//
// **Er ist eine Naht der Fassade und keine des Dienstes**: `internal/app`
// kennt das Zuhause der Anwendung, und ein Test soll es umlenken können, ohne
// das echte anzufassen (§3).
func (a *App) MitBilderort(ort func() (string, error)) { a.bilderort = ort }

// Bilderort gibt den Ort der eingesetzten Bilder — **für Tests**.
func (a *App) Bilderort() (string, error) { return a.bilderort() }

// Eingehend ist die Endung der Kopie, die während des Tausches entsteht.
const Eingehend = service.Eingehend

// Beiwerkbeiseite ist der Name, unter dem das alte Beiwerk landet.
const Beiwerkbeiseite = service.Beiwerkbeiseite

// Rueckspielmarke ist der Name der Marke neben dem Bestand.
const Rueckspielmarke = service.Rueckspielmarke

// MitSicherungspruefer setzt die Naht — für Tests und den Composition Root.
//
// **Die Setzer stehen hier und wirken dort.** Die Fassade hält die Werte, weil
// sie verdrahtet; entschieden wird mit ihnen im Dienst, der sie bei jedem
// Vorgang über [App.bestandsweg] bekommt.
func (a *App) MitSicherungspruefer(pruefer Sicherungspruefer) { a.sicherungspruefer = pruefer }

// MitAlleinpruefung setzt die Naht — für Tests und den Composition Root.
func (a *App) MitAlleinpruefung(pruefen Alleinpruefung) { a.alleinpruefung = pruefen }

// rueckspielVollenden holt einen abgebrochenen Bildertausch nach.
//
// **Beim Start und ohne Rückfrage** — warum, steht im Dienst.
//
// **Was liegen bleibt, wird gemerkt**, damit [App.Startbefund] es dem Fenster
// sagen kann: stünde es nur im Protokoll, öffnete die Anwendung danach
// wortlos die schon getauschte Datenbank neben den alten Bildern (#1170).
func (a *App) rueckspielVollenden() {
	weg := a.bestandsweg()

	a.rueckspielrest = weg.RueckspielVollenden()

	// **Und eine liegengebliebene Rückkehr wird nachgetragen** — Befund 10
	// des Zweitblicks über Phase 5.2. Die Epoche entsteht **nach** dem
	// Tausch; scheiterte sie oder starb der Prozess dazwischen, lief die
	// Anwendung mit dem neuen Bestand weiter, und der Nachweis für die
	// unklaren Mengen fehlte (0088 sagt sein Nichtverschwinden zu).
	//
	// **Der zweite Satz überschreibt den ersten nicht**: beide Meldungen
	// betreffen denselben Start, und die erste ist die dringendere — Bilder
	// zu einem anderen Stand als die Daten.
	if rest := weg.EpocheNachholen(a.fensterkontext()); rest != "" && a.rueckspielrest == "" {
		a.rueckspielrest = rest
	}
}

// bestandsweg baut den Dienst aus den Orten, die nur hier bekannt sind.
func (a *App) bestandsweg() *service.Bestandsweg {
	weg := service.NeuerBestandsweg(
		service.Sicherungsorte{Datenbank: a.datenbankpfad, Bilder: a.bilderort},
		a.sicherungspruefer, a.alleinpruefung, a.bestandssicherer,
		// **Die Tauschwege und nicht `Beenden`/`Starten`** (A02, #2216):
		// der Schluss des Programms ließe das Tor hinter dem Tausch zu. Das
		// Tor selbst hält der Rahmen ([App.imTausch]) um den ganzen Weg.
		a.zumTauschSchliessen, a.wiederOeffnen, Fassung,
	)

	weg.MitBestandschiffre(a.bestandschiffre)

	// **Und das Buch der Sicherungen** (OPS01, #2347) — solange ein Bestand
	// offen ist; die Wiederherstellung beim Start hat keinen.
	if a.db != nil {
		weg.MitSicherungsbuch(vermerkbuch{holen: a.laufendeVerbindung})
	}

	// **Und die eigenen Pakete** (OPS-06.AK-01) — solange es einen Katalog
	// gibt; die Wiederherstellung beim Start hat keinen offenen Bestand.
	if a.pakete != nil {
		weg.MitPaketen(paketliste{a.pakete}, func() (string, error) { return paketablage(false) })
	}

	// **Und die drei Nähte des Rücksprungvergleichs** (O7, #1431). Sie stehen
	// hier und nicht im Composition Root, weil sie zu **diesem** Weg gehören:
	// er entsteht bei jedem Ruf neu, und die Verbindung darunter wechselt.
	weg.MitRuecksprungschau(
		bestandsschaunaht,
		eigenstandnaht{holen: a.laufendeVerbindung}.stand,
		a.rueckkehrnaht,
	)

	return weg
}

// paketliste reicht den Katalog des Verwalters an die Sicherung.
type paketliste struct{ verwalter Paketfuehrer }

func (p paketliste) Pakete(ctx context.Context) ([]domain.Paket, error) {
	return p.verwalter.Liste(ctx)
}

// Sicherungsauswahl ist, was das Fenster über eine gewählte Sicherung erfährt.
//
// **Vor der Übernahme**, damit jemand entscheiden kann: der Pfad, der
// Zeitpunkt der Anlage, der Schemastand und was darin liegt.
type Sicherungsauswahl struct {
	Pfad        string  `json:"pfad"`
	AngelegtAm  string  `json:"angelegtAm"`
	Schemastand string  `json:"schemastand"`
	Sets        int     `json:"sets"`
	Karten      int     `json:"karten"`
	Bilder      string  `json:"bilder"`
	Abgebrochen bool    `json:"abgebrochen"`
	Fehler      *Fehler `json:"fehler"`

	// Verschluesselt sagt, ob die Sicherung ein Kennwort trägt — auch dann,
	// wenn es fehlte und deshalb [Sicherungsauswahl.Fehler] gesetzt ist: das
	// Fenster fragt dann danach (OPS-06.AK-02).
	Verschluesselt bool `json:"verschluesselt"`

	// PaketeEnthalten sind die eigenen Pakete darin, PaketeFehlen die
	// bezogenen, die hier nicht liegen (OPS-06.AK-03).
	PaketeEnthalten []string `json:"paketeEnthalten"`
	PaketeFehlen    []string `json:"paketeFehlen"`
}

// SicherungWaehlen fragt nach einer Datei und liest, was darin steht.
//
// **Sie übernimmt nichts.** Wer nur wissen will, was in einer Datei liegt,
// bekommt die Auskunft, ohne dass sich am Bestand etwas ändert — und wer
// übernehmen will, hat vorher gesehen, was er bekommt.
func (a *App) SicherungWaehlen(ctx context.Context) Sicherungsauswahl {
	ctx = a.imProfil(ctx)

	texte := DialogtexteFuer(a.sprache)

	pfad, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  texte.BestandOeffnen,
		Filter: texte.BestandFilter,

		// **Beide Formate**, seit die Sicherung ein Archiv ist (#1162). Wer
		// eine `.db` aus 0.1.0 oder 0.2.0 zurückspielen will, soll sie im
		// Dialog auch sehen.
		Muster: "*" + service.Sicherungsendung + ";*.db",
	})
	if fehler != nil {
		return ohneSicherung(Sicherungsauswahl{Fehler: Uebersetze(ctx, fehler)})
	}

	if pfad == "" {
		return ohneSicherung(Sicherungsauswahl{Abgebrochen: true})
	}

	return a.sicherungBeschauen(ctx, pfad, "")
}

// SicherungBeschauen liest eine schon gewählte Sicherung mit ihrem Kennwort —
// der zweite Schritt, wenn [App.SicherungWaehlen] eine verschlüsselte fand.
func (a *App) SicherungBeschauen(ctx context.Context, pfad, kennwort string) Sicherungsauswahl {
	ctx = a.imProfil(ctx)

	return a.sicherungBeschauen(ctx, pfad, kennwort)
}

func (a *App) sicherungBeschauen(ctx context.Context, pfad, kennwort string) Sicherungsauswahl {
	verschluesselt := service.SicherungVerschluesselt(pfad)

	schau, schaufehler := a.bestandsweg().Beschauen(
		service.MitSicherungswunsch(ctx, service.Sicherungswunsch{Kennwort: kennwort}), pfad)
	if schaufehler != nil {
		return ohneSicherung(Sicherungsauswahl{
			Pfad: pfad, Verschluesselt: verschluesselt, Fehler: Uebersetze(ctx, schaufehler),
		})
	}

	return Sicherungsauswahl{
		Pfad:           pfad,
		AngelegtAm:     schau.AngelegtAm,
		Schemastand:    schau.Schemastand,
		Sets:           schau.Sets,
		Karten:         schau.Karten,
		Bilder:         schau.Bilder,
		Verschluesselt: verschluesselt,

		PaketeEnthalten: schau.PaketeEnthalten,
		PaketeFehlen:    schau.PaketeFehlen,
	}
}

// ohneSicherung füllt eine Auswahl, zu der es keine gelesene Sicherung gibt —
// abgebrochen oder gescheitert —, so, wie das Fenster sie typisiert: die
// Listen leer statt `null`, und die Bilder **unberührt**, denn ohne Datei
// geschieht mit ihnen nichts. Bis C01 (#2226) stand dort ein leerer Text,
// den keine der drei Lagen kennt.
func ohneSicherung(auswahl Sicherungsauswahl) Sicherungsauswahl {
	auswahl.Bilder = service.BilderUnberuehrt
	auswahl.PaketeEnthalten = []string{}
	auswahl.PaketeFehlen = []string{}

	return auswahl
}

// Rueckspielantwort sagt, ob der Bestand ersetzt wurde.
//
// **Und wohin der alte kam.** Ein Zurückspielen ohne diese Angabe ist ein
// Vorgang, dessen Rücknahme niemand kennt.
//
// **Und nach einem Fehler, ob der bisherige Stand wieder ganz dasteht** —
// die ausdrückliche Rückkehrmeldung aus A05 (Nachlauf über Phase A).
// `false` heißt: auch der Rückweg scheiterte, und Teile liegen beiseite
// ([service.ErrRueckwegGescheitert]).
type Rueckspielantwort struct {
	Ersetzt      bool    `json:"ersetzt"`
	Beiseite     string  `json:"beiseite"`
	Unveraendert bool    `json:"unveraendert"`
	Fehler       *Fehler `json:"fehler"`

	// Bestandsepoche ist die Epoche nach dem Tausch (J05, #2269): eine
	// Kennung, die der Doppelklick aus einer älteren meldete, verwirft das
	// Fenster.
	Bestandsepoche uint64 `json:"bestandsepoche"`
}

// SicherungZurueckspielen ersetzt den Bestand durch die Sicherung.
//
// **Die Reihenfolge, die Rücknahme und die Marke stehen im Dienst**
// ([service.Bestandsweg.Zurueckspielen]) — hier bleibt die Übersetzung.
func (a *App) SicherungZurueckspielen(ctx context.Context, pfad, kennwort string) Rueckspielantwort {
	ctx = a.imProfil(ctx)
	ctx = service.MitSicherungswunsch(ctx, service.Sicherungswunsch{Kennwort: kennwort})

	var beiseite string

	fehler := a.imTausch(func() error {
		var tauschfehler error
		beiseite, tauschfehler = a.bestandsweg().Zurueckspielen(ctx, pfad)

		return tauschfehler
	})
	if fehler != nil {
		return Rueckspielantwort{
			Unveraendert:   !errors.Is(fehler, service.ErrRueckwegGescheitert),
			Fehler:         Uebersetze(ctx, fehler),
			Bestandsepoche: a.lauf.epoche.Load(),
		}
	}

	return Rueckspielantwort{Ersetzt: true, Beiseite: beiseite, Bestandsepoche: a.lauf.epoche.Load()}
}
