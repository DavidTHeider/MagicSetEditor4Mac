package service

// **Der Lebenszyklus einer Peersitzung** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was diese Verwaltung hält und was sie nicht tut
//
// **Sie hält die Zusagen des Issues**: Fristen, Abbruch, begrenzte
// Warteschlangen, Rückstau, Berechtigungsentzug, und beim Profilwechsel das
// Schließen samt Verwerfen verspäteter Antworten.
//
// **Sie öffnet keine Verbindung.** Das tut der [Waehler] — libp2p (N2) oder
// Pion (N3). Was hier steht, gilt für beide, und deshalb steht es nicht
// dort.
//
// # Warum eine Epoche und kein Schließmerker
//
// **Weil ein Merker zweimal dasselbe sagt.** „Geschlossen" nach einem
// Profilwechsel und „geschlossen" nach einem Ruf sehen gleich aus; wer eine
// verspätete Antwort einträgt, muss wissen, ob sie zum **heutigen** Stand
// gehört. Die Epoche zählt Wechsel — dieselbe Gestalt wie in §7, Regel 2.

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Peersitzungen führt die offenen Sitzungen dieses Profils.
type Peersitzungen struct {
	schloss sync.Mutex
	rechte  Rechteauskunft
	waehler Waehler
	ausweis Ausweisstelle
	jetzt   func() time.Time
	tiefe   int
	epoche  uint64
	offen   map[string]*Sitzungsgriff

	// entsteht sind die Sitzungen, die gerade aufgebaut werden.
	//
	// **Ohne diese Liste ist „läuft schon" eine Prüfung auf einen Stand von
	// vorhin** — Befund 3 des Zweitblicks über Phase 5.4. Zwischen der
	// Prüfung und dem Eintrag liegen Wählen **und** Handschlag; zwei
	// gleichzeitige Rufe gingen beide hinaus, und das Gegenüber erlebte
	// zwei Verbindungsaufbauten, während der zweite hier mit `ErrConflict`
	// endete. **Der Preis eines zu späten Neins zahlt der andere** (§7,
	// Regel 4).
	//
	// **Getrennt von [Peersitzungen.offen]**, weil ein Platzhalter darin
	// jeden Leser zwänge, auf `nil` zu prüfen — und wer das einmal
	// vergisst, stürzt ab.
	entsteht map[string]bool

	// zufall ist die Quelle für die Zufallszahlen des Handschlags —
	// `nil` heißt: die des Betriebssystems.
	zufall io.Reader

	// gesehen merkt sich die Zufallszahlen, die schon einmal da waren.
	//
	// **Eine je Verwaltung und nicht je Sitzung**: ein Wiedereinspielen
	// probiert es bei der nächsten Verbindung noch einmal, und eine Liste,
	// die mit der Sitzung stürbe, sähe das nie.
	gesehen *peer.Nonceliste
}

// Noncefrist ist, wie lange sich die Verwaltung eine Zufallszahl merkt.
//
// **Eine Stunde, und der Grund steht bei [peer.Nonceliste]**: über
// Verbindungen hinweg trägt die Kanalbindung, innerhalb einer Verbindung diese
// Liste. Ohne Frist wüchse sie, und ein Peer, dem noch niemand traut, füllte
// den Speicher.
const Noncefrist = time.Hour

// NeuePeersitzungen baut die Verwaltung.
//
// **Die Tiefe ist ein Argument und keine Konstante**: was eine Partie braucht,
// ist nicht, was eine Übertragung braucht — und eine feste Zahl wäre für die
// eine zu klein und für die andere zu groß.
func NeuePeersitzungen(
	rechte Rechteauskunft, waehler Waehler, ausweis Ausweisstelle,
	jetzt func() time.Time, tiefe int,
) *Peersitzungen {
	return &Peersitzungen{
		rechte:  rechte,
		waehler: waehler,
		ausweis: ausweis,
		jetzt:   jetzt,
		tiefe:   tiefe,
		offen:   map[string]*Sitzungsgriff{},
		gesehen: peer.NeueNonceliste(jetzt, Noncefrist),
	}
}

// MitZufall setzt die Quelle für die Zufallszahlen des Handschlags.
func (p *Peersitzungen) MitZufall(zufall io.Reader) *Peersitzungen {
	p.zufall = zufall

	return p
}

// Sitzungsgriff ist eine offene Sitzung.
type Sitzungsgriff struct {
	schloss sync.Mutex

	// leeren hält einen `Leere` nach dem anderen — siehe dort.
	leeren sync.Mutex

	eltern     *Peersitzungen
	verbindung Verbindung
	stand      Sitzungsstand
	sitzung    peer.Sitzung
	epoche     uint64
	warten     []wartend
	zu         bool

	// unterwegs sagt, dass die erste wartende Nachricht gerade hinausgeht —
	// sie darf dann nicht weichen.
	unterwegs bool
}

// wartend ist eine Nachricht in der Schlange samt ihrer Zustellung — ohne
// die wüsste ein verwerfendes `Sende` nicht, was es verdrängen darf.
type wartend struct {
	nachricht []byte
	art       Zustellung
}

// Oeffne baut eine Sitzung zu einem Gegenüber auf.
//
// **Erst das Recht, dann die Leitung.** Wer zuerst wählt, hat dem Gegenüber
// schon gesagt, dass es dieses Programm gibt — und bei einem
// kostenpflichtigen Transport auch bezahlt.
func (p *Peersitzungen) Oeffne(
	ctx context.Context, kennung peer.Kennung, zweck domain.Zweck,
) (*Sitzungsgriff, error) {
	darf, fehler := p.rechte.Darf(ctx, kennung, zweck)
	if fehler != nil {
		return nil, fehler
	}

	if !darf {
		return nil, fmt.Errorf(
			"%w: %s darf %q nicht", domain.ErrValidation,
			peer.Fingerabdruck(kennung), zweck)
	}

	protokoll, gilt := peer.ProtokollZu(zweck)
	if !gilt {
		return nil, domain.PruefeZweck(zweck)
	}

	schluessel := string(kennung) + "/" + string(zweck)

	p.schloss.Lock()

	_, laeuft := p.offen[schluessel]

	if laeuft || p.entsteht[schluessel] {
		p.schloss.Unlock()

		return nil, fmt.Errorf(
			"%w: zu %s läuft schon eine Sitzung für %q", domain.ErrConflict,
			peer.Fingerabdruck(kennung), zweck)
	}

	// **Der Platz wird belegt, bevor gewählt wird** — der Grund steht bei
	// [Peersitzungen.entsteht].
	if p.entsteht == nil {
		p.entsteht = map[string]bool{}
	}

	p.entsteht[schluessel] = true

	epoche := p.epoche

	p.schloss.Unlock()

	// **Und er wird auf jedem Weg wieder frei.** Ein Fehlschlag, der die
	// Belegung stehen ließe, machte aus einem Befund eine Sperre, die erst
	// ein Profilwechsel löste.
	defer func() {
		p.schloss.Lock()
		delete(p.entsteht, schluessel)
		p.schloss.Unlock()
	}()

	// **Der Ausweis vor der Leitung.** Wer ohne ihn wählte, hätte dem
	// Gegenüber gesagt, dass es dieses Programm gibt — und stünde dann
	// da, ohne unterschreiben zu können.
	ich, fehler := p.ausweis.Ausweis(ctx)
	if fehler != nil {
		return nil, fehler
	}

	verbindung, fehler := p.waehler.Waehle(ctx, kennung, protokoll)
	if fehler != nil {
		return nil, fehler
	}

	sitzung, fehler := p.handschlagFuehren(ctx, ich, verbindung, kennung, zweck)
	if fehler != nil {
		_ = verbindung.Schliesse() //nolint:errcheck // der Weg endet hier ohnehin

		return nil, fehler
	}

	griff := p.griffAus(verbindung, sitzung, schluessel, epoche)

	if fehler := p.eintragen(griff); fehler != nil {
		return nil, fehler
	}

	return griff, nil
}

// Alle gibt den Stand jeder offenen Sitzung.
func (p *Peersitzungen) Alle() []Sitzungsstand {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	staende := make([]Sitzungsstand, 0, len(p.offen))
	for _, griff := range p.offen {
		staende = append(staende, griff.jetzigerStand())
	}

	return staende
}

// Pausiere schließt jede Sitzung und macht jede spätere Antwort ungültig.
//
// **Für den Profil- und Nutzerwechsel**, und das Issue verlangt beides:
// „Sessions pausieren und Schlüsselkontext schließen, verspätete Antworten
// verwerfen."
func (p *Peersitzungen) Pausiere() error {
	p.schloss.Lock()

	p.epoche++
	griffe := make([]*Sitzungsgriff, 0, len(p.offen))

	for _, griff := range p.offen {
		griffe = append(griffe, griff)
	}

	p.offen = map[string]*Sitzungsgriff{}

	p.schloss.Unlock()

	var befund error

	for _, griff := range griffe {
		// **Jede wird geschlossen, auch wenn eine klagt.** Wer beim ersten
		// Fehler aufhörte, ließe die übrigen Leitungen offen — und die
		// gehören zu einem Profil, das niemand mehr benutzt.
		if fehler := griff.schliesseStill(); fehler != nil {
			befund = fehler
		}
	}

	return befund
}

// vergiss nimmt eine Sitzung aus der Übersicht.
func (p *Peersitzungen) vergiss(schluessel string, griff *Sitzungsgriff) {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	// **Nur die eigene.** Nach einem Wechsel steht dort eine andere, und
	// wer sie löschte, nähme jemandem die laufende Sitzung weg — dieselbe
	// Überlegung wie beim Bildabbruch (#1452).
	if p.offen[schluessel] == griff {
		delete(p.offen, schluessel)
	}
}

// griffAus baut den Griff zu einer fertig ausgehandelten Sitzung.
func (p *Peersitzungen) griffAus(
	verbindung Verbindung, sitzung peer.Sitzung, schluessel string,
	epoche uint64,
) *Sitzungsgriff {
	return &Sitzungsgriff{
		eltern:     p,
		verbindung: verbindung,
		sitzung:    sitzung,
		epoche:     epoche,
		stand: Sitzungsstand{
			Kennung:       schluessel,
			Gegenueber:    sitzung.Gegenueber,
			Fingerabdruck: sitzung.Fingerabdruck,
			Zweck:         sitzung.Zweck,
			Kanal:         sitzung.Kanal,
			Seit:          p.jetzt(),
		},
	}
}

// eintragen hängt einen Griff in die Übersicht — wenn der Stand noch gilt.
//
// **Zwischen dem Aushandeln und dem Eintragen kann ein Wechsel liegen.** Ein
// Handschlag geht über das Netz und dauert; wer in dieser Lücke das Profil
// wechselt, bekäme sonst eine Sitzung, die niemand mehr führt — und die
// Leitung bliebe offen.
func (p *Peersitzungen) eintragen(griff *Sitzungsgriff) error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	if p.epoche != griff.epoche {
		_ = griff.verbindung.Schliesse() //nolint:errcheck // der Weg endet hier ohnehin

		return fmt.Errorf(
			"%w: das Profil hat gewechselt, während die Sitzung entstand",
			domain.ErrConflict)
	}

	if vorhanden, laeuft := p.offen[griff.stand.Kennung]; laeuft &&
		vorhanden != griff {
		_ = griff.verbindung.Schliesse() //nolint:errcheck // dito

		return fmt.Errorf(
			"%w: zu %s läuft schon eine Sitzung für %q", domain.ErrConflict,
			griff.stand.Fingerabdruck, griff.stand.Zweck)
	}

	p.offen[griff.stand.Kennung] = griff

	return nil
}
