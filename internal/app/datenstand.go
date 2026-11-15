package app

import (
	"context"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Standleser sagt, ob ein anderer Prozess geschrieben hat.
//
// **Als Schnittstelle und nicht als Aufruf in `storage`**: `app` kennt
// `storage` nur im Composition Root (§2), und der Schichtenwächter zählt das
// nach. Die Naht steht hier, weil sie hier gebraucht wird — dieselbe Richtung
// wie überall.
type Standleser interface {
	Datenstand(ctx context.Context) (int64, error)
}

// StandAntwort trägt den Datenstand oder einen Fehler.
type StandAntwort struct {
	Stand  int64   `json:"stand"`
	Fehler *Fehler `json:"fehler"`
}

// standnaht ist, was `Datenstand` zählt: den Stand des Bestands und, was
// diese Anwendung **nebenher** geschrieben hat.
//
// # Warum es den zweiten Zähler gibt
//
// **Weil `data_version` nur fremde Prozesse zählt** (#1485) — und seit es
// den Horchposten gibt, schreibt auch dieser Prozess, ohne dass das Fenster
// es veranlasst hat: ein Handel kommt über die Leitung, eine Einladung wird
// eingelöst. Bert bot Anna einen Tausch an, er stand in ihrem Bestand, und
// ihr Fenster zeigte ihn erst, als sie die Seite neu öffnete (Hüllenprobe
// `nat04handel`, S12.NAT-04,
// #1672).
//
// **Ein Zähler und kein Ereignis**: das Fenster fragt ohnehin
// ([0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md)),
// und jede Seite, die den Datenstand liest, lädt damit nach — auch die, an
// die niemand gedacht hat.
type standnaht struct {
	stand Standleser

	// nebenher zählt, was ein Weg ohne Fenster geschrieben hat.
	nebenher atomic.Int64
}

// nebenherGeschrieben sagt dem Fenster, dass ohne sein Zutun geschrieben
// wurde.
func (s *standnaht) nebenherGeschrieben() { s.nebenher.Add(1) }

// nebenherNehmer sagt nach jedem angenommenen Zug eines Gastes, dass nebenher
// geschrieben wurde — Befund 1 des Nachlaufs zum Gesamtreview über 0.6.0
// (#1660).
//
// **Der dritte Zweck des Horchpostens.** Handel und Paarung sagten es seit
// #2060; die Partie nicht, und sie schreibt am häufigsten: jeder Zug eines
// Gastes geht im Prozess des Wirts in den Bestand. Der Tisch erfuhr es über
// sein Ereignis, die Partienliste nicht.
//
// **Am Nehmer und nicht in der Schleife**: hier geht jeder Befehl durch, und
// nur ein angenommener schreibt. In der Schleife hieße es auch für Chat und
// Präsenz, die nichts schreiben.
type nebenherNehmer struct {
	service.Befehlsnehmer

	geschrieben func()
}

// BefehlVerkuendet legt den Zug vor und sagt es weiter, wenn er angenommen
// wurde.
//
// **Vor dem Verkünden und nicht danach** (#2481): verkündet wird jetzt, bevor
// der Ruf zurückkehrt. Wer erst danach Bescheid gäbe, gäbe ihn nach der
// Meldung an das Fenster — und das Fenster läse einen Datenstand, der noch
// nicht weiß, was es gerade gemeldet bekam.
func (n nebenherNehmer) BefehlVerkuendet(
	ctx context.Context, befehl domain.Spielbefehl, verkuende service.Verkuendung,
) (service.Befehlsausgang, error) {
	return n.Befehlsnehmer.BefehlVerkuendet(ctx, befehl,
		func(ausgang service.Befehlsausgang) {
			n.geschrieben()
			verkuende(ausgang)
		})
}

// MitStand setzt die Naht zum Standleser.
//
// **Als Feld und aus demselben Grund wie `oeffnen`**: der Fehlerweg wäre sonst
// nur über eine kaputte Platte zu erreichen, und dann prüfte der Test die
// Platte statt die Übersetzung.
func (a *App) MitStand(leser Standleser) { a.stand = leser }

// Datenstand sagt, ob ein **anderer** Prozess geschrieben hat.
//
// # Warum das Fenster fragt und Go nicht meldet
//
// Der MCP-Server ist ein eigener Prozess auf derselben Datei
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)), und das
// Fenster erfuhr nichts von seinen Änderungen
// (#156).
//
// **Gefragt wird von dort, wo die Anzeige lebt.** Ein Faden in Go, der von sich
// aus meldete, führe auch dann, wenn niemand hinsieht — und er müsste wissen,
// wann das Fenster zu ist. Das Frontend weiß beides ohnehin. Die Begründung
// steht ausführlich in
// [0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md).
//
// # Was die Zahl bedeutet
//
// Nur **dass** sich etwas geändert hat, nicht was. Wer wissen will, welche
// Karte, liest nach — und das ist der Preis, den der Vermerk benennt.
func (a *App) Datenstand(ctx context.Context) StandAntwort {
	ctx = a.imProfil(ctx)

	stand, err := a.stand.Datenstand(ctx)
	if err != nil {
		return StandAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Die Summe und nicht ein Paar**: das Fenster vergleicht eine Zahl,
	// und beide Teile steigen nur. Fängt `data_version` nach einem Tausch
	// der Verbindung neu an, lädt das Fenster einmal zu viel — derselbe Preis
	// wie im Standleser.
	return StandAntwort{Stand: stand + a.nebenher.Load()}
}
