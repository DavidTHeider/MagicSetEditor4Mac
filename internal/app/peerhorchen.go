package app

// **Erreichbar sein — auf Aufruf** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum nichts von selbst horcht
//
// **Weil das Issue es verlangt**: „Nur nach bewusster Auswahl Spieler/Kontakte
// in der Nähe sichtbar werden; keine unbemerkte Dauerwerbung aller Profile."
// Wer nicht horcht, ist nicht erreichbar, und das ist der Ausgangszustand
// dieser Anwendung — auch nach einem Neustart.
//
// # Warum der Horchposten einen eigenen Kontext bekommt
//
// **Weil er den Ruf überlebt, der ihn gestartet hat.** Eine Bindung bekommt
// von Wails einen Kontext je Aufruf; der ist beendet, sobald die Antwort
// draußen ist. Ein Horchposten daran wäre nach einer Sekunde tot.
//
// **Und warum trotzdem nicht `context.Background()`**: darin stünde das Profil
// nicht, und jede hereinkommende Verbindung fragte den Bestand ohne eines.
// `context.WithoutCancel` behält die Werte und lässt das Ende fallen — genau
// die Hälfte, um die es geht.

import (
	"context"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// horchfelder ist, was eine Anwendung fürs Horchen hält.
type horchfelder struct {
	schloss sync.Mutex

	// adresse ist, wo wirklich gehorcht wird — leer heißt: nirgends.
	//
	// **Die wirkliche und nicht die gewünschte.** Wer „:0" angibt, bekommt
	// vom Betriebssystem einen Port, und nur so erfährt ein Mensch, welchen
	// er dem Gegenüber nennen muss.
	adresse string

	// erreichbar ist, was ein Gegenüber eintippen kann.
	//
	// **Einmal gerechnet und nicht bei jedem Lesen.** Die
	// Schnittstellenliste dieses Rechners kostet einen Systemaufruf, und
	// `HorchStand` läuft bei jedem Zeichnen des Panels.
	erreichbar leitung.Erreichbarkeit

	aufhoeren context.CancelFunc
	fertig    chan struct{}

	// meldetAn sagt, dass gerade eine Ankündigung entsteht — und
	// sichtbarkeit zählt, wie oft seither jemand aufgehört hat.
	//
	// **Zwei Felder für zwei Fragen** (Nachlauf zum Zweitblick über
	// Phase 5.4, Befund 1): das erste hält den zweiten Ruf auf, das zweite
	// fängt den Fall, den ein Merker nicht sieht — wer **währenddessen**
	// aufhört. `mdns.Register` dauert; wer danach eine Anmeldung einträgt,
	// die niemand mehr wollte, strahlt einen Port aus, der zu ist.
	//
	// **Ein Zähler und kein Wahrheitswert**: „aufgehört und wieder
	// angefangen" sähe sonst aus wie „nichts geschehen".
	meldetAn     bool
	sichtbarkeit uint64

	// faengtAn sagt, dass gerade einer entsteht.
	//
	// **Ohne dieses Feld ist „horcht schon" eine Prüfung auf einen Stand von
	// vorhin** — Befund 3 des Zweitblicks über Phase 5.4. Das Aufmachen
	// eines Lauschers dauert; solange steht `aufhoeren` auf `nil`, und ein
	// zweiter Ruf sähe das Profil als still.
	faengtAn bool

	// horchen ist die Naht zum Lauscher — nil heißt: der echte.
	//
	// **Als Feld, damit ein Lauf ohne Port auskommt** (§3), und aus
	// demselben Grund wie `oeffnen` und `paketort` daneben.
	horchen func(
		ctx context.Context, adresse string, ich peer.Identitaet,
	) (service.Hereinkommend, error)

	// anmeldung ist die laufende Ankündigung — nil heißt: keine.
	//
	// **Sie hängt am Horchposten und nicht daneben**, weil sie seinen Port
	// nennt: ohne ihn wäre sie eine Einladung an eine Tür, hinter der
	// niemand steht.
	anmeldung Ankuendigung

	// ankuendigen und suchen sind die Nähte zu `internal/entdeckung` —
	// nil heißt: die echten. Als Felder, damit ein Lauf ohne Netz auskommt
	// (§3), und aus demselben Grund wie `horchen` darüber.
	ankuendigen func(entdeckung.Angebot) (Ankuendigung, error)
	suchen      func(
		context.Context, entdeckung.Frage,
	) (entdeckung.Ergebnis, error)

	// erfragen ist die Naht zu den Adressen dieses Rechners — nil heißt:
	// die echte.
	//
	// **Als Naht und aus demselben Grund wie die drei darüber** (§3): sonst
	// prüfte ein Lauf, was zufällig in diesem Netz steht, und wäre auf einer
	// Maschine ohne zweite Schnittstelle rot. Was die Naht liefert, ist
	// nicht die Frage dieses Pakets — die steht in `internal/leitung` und
	// wird dort geprüft.
	erfragen func(context.Context, string) (leitung.Erreichbarkeit, error)
}

// HorchStarten macht dieses Profil unter einer Adresse erreichbar.
//
// # Was die Adresse hier bedeutet
//
// **Wo gehorcht wird, nicht wo jemand sitzt.** `:41234` horcht auf allen
// Schnittstellen dieses Rechners, `127.0.0.1:41234` nur auf der Rückschleife —
// das eine ist der Betrieb, das andere ein Lauf auf einer Maschine.
func (a *App) HorchStarten(ctx context.Context, adresse string) Horchantwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.horchenStarten(ctx, adresse)
	if fehler != nil {
		return Horchantwort{Stand: ohneListen(), Fehler: Uebersetze(ctx, fehler)}
	}

	return Horchantwort{Stand: stand}
}

// HorchBeenden macht dieses Profil wieder unerreichbar.
func (a *App) HorchBeenden(ctx context.Context) Horchantwort {
	ctx = a.imProfil(ctx)

	// **Der Kontext wird hier nicht gebraucht und trotzdem gebunden.** Wer
	// ihn ausließe, hielte diese Bindung aus der Umschaltsperre heraus —
	// und ein Beenden mitten im Profilwechsel griffe an einem Stand, den es
	// gerade nicht mehr gibt (§7, Regel 2).
	_ = ctx

	a.horchenBeenden()

	return Horchantwort{Stand: a.horchstand()}
}

// HorchStand sagt, ob und wo gehorcht wird.
func (a *App) HorchStand(ctx context.Context) Horchantwort {
	ctx = a.imProfil(ctx)

	// **Aus demselben Grund wie beim Beenden** — ein Stand, der während
	// eines Wechsels gelesen wird, gehört zu keinem der beiden Profile.
	_ = ctx

	return Horchantwort{Stand: a.horchstand()}
}

// horchenStarten baut den Lauscher und lässt den Posten laufen.
func (a *App) horchenStarten(
	ctx context.Context, adresse string,
) (Horchzustand, error) {
	marke := a.laufmarke(domain.LaufHorchen) // vor dem Start: Nachlauf über Phase 6.12, N2

	// **Die Rolle wird gefragt, bevor der Rechner erreichbar wird** —
	// Werkstück N7 von Phase 5.4
	// (#1568).
	//
	// **Bis dahin fragte hier niemand.** `docs/AUTORISIERUNG.md` nannte für
	// diese Zeile „mitarbeiter" — und kein Ruf prüfte es; ein Leser in einem
	// fremden Profil konnte einen Port öffnen. Das war keine Lücke, die N7
	// geschaffen hat, sondern eine, die es **benannt** hat: eine Tabelle,
	// die eine Mindestrolle nennt, die nichts durchsetzt, verspricht mehr,
	// als ihr Rumpf hält (§10).
	if fehler := a.darf(ctx, domain.WirkungExternEmpfangen); fehler != nil {
		return Horchzustand{}, fehler
	}

	// **Die Horchprüfung und nicht die zum Wählen** — `:41234` horcht auf
	// jeder Schnittstelle dieses Rechners und ist zugleich keine Adresse, zu
	// der sich wählen ließe ([domain.PruefeHorchadresse]).
	if fehler := domain.PruefeHorchadresse(adresse); fehler != nil {
		return Horchzustand{}, fehler
	}

	// **Der Ausweis vor dem Port.** Ein Profil ohne Identität kann keinen
	// Handschlag führen; ein offener Port ohne Identität sagte jedem, dass
	// es dieses Programm gibt, und wiese dann jeden ab.
	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return Horchzustand{}, fehler
	}

	// **Der Platz wird belegt, bevor er gebraucht wird** — Befund 3 des
	// Zweitblicks über Phase 5.4. Wer nur läse und das Schloss wieder löste,
	// prüfte einen Stand von vorhin: das Aufmachen eines Lauschers dauert,
	// und solange steht `aufhoeren` auf `nil`. Zwei gleichzeitige Rufe
	// öffneten zwei Lauscher, und der zweite überschriebe den einzigen
	// Abbruchgriff des ersten — der horchte weiter, über das Beenden und den
	// Profilwechsel hinaus (§7, Regel 4).
	a.peerlage.horchen.schloss.Lock()
	laeuftSchon := a.peerlage.horchen.aufhoeren != nil || a.peerlage.horchen.faengtAn
	if !laeuftSchon {
		a.peerlage.horchen.faengtAn = true
	}
	a.peerlage.horchen.schloss.Unlock()

	if laeuftSchon {
		return Horchzustand{}, fmt.Errorf(
			"%w: dieses Profil horcht schon — erst beenden, dann neu "+
				"anfangen", domain.ErrConflict)
	}

	// **Und er wird auf jedem Weg wieder frei** — sonst machte ein
	// Fehlschlag aus einem Befund eine Sperre bis zum nächsten Neustart.
	defer func() {
		a.peerlage.horchen.schloss.Lock()
		a.peerlage.horchen.faengtAn = false
		a.peerlage.horchen.schloss.Unlock()
	}()

	herein, fehler := a.horcher()(ctx, adresse, ich)
	if fehler != nil {
		return Horchzustand{}, fehler
	}

	laufender, aufhoeren := context.WithCancel(context.WithoutCancel(ctx))
	fertig := make(chan struct{})

	// **Ein Befund hier hält nichts auf.** Der Port steht, die Verbindung
	// trägt — nur die Liste zum Weitergeben bleibt leer. Wer abbräche, weil
	// eine Anzeige nicht geht, nähme jemandem einen laufenden Horchposten.
	erreichbar, fehler := a.erreichbarkeit()(ctx, herein.Adresse())
	if fehler != nil {
		meldeHorchbefund(fehler)

		erreichbar = leitung.Erreichbarkeit{}
	}

	a.peerlage.horchen.schloss.Lock()
	a.peerlage.horchen.adresse = herein.Adresse()
	a.peerlage.horchen.erreichbar = erreichbar
	a.peerlage.horchen.aufhoeren = aufhoeren
	a.peerlage.horchen.fertig = fertig
	a.peerlage.horchen.schloss.Unlock()

	go func() {
		defer close(fertig)
		defer domain.Auffangen(func(zusammenbruch error) { a.horchenZusammengebrochen(fertig, zusammenbruch) })

		// **Und was angenommen wird, wird auch bedient** (G5, #1462): eine
		// Sitzung, die nur in der Übersicht stünde, läse niemand; ob eine
		// Partie offensteht, entscheidet die Aufnahme.
		service.NeuerHorchposten(
			a.sitzungsverwaltung(), meldeHorchbefund).
			MitAufnahme(a.sitzungAufnehmen).
			Bediene(laufender, herein)
	}()

	a.nachDemStart(ctx, domain.LaufHorchen, marke, func(context.Context) { a.horchenBeenden() })

	return a.horchstand(), nil
}

// horchenBeenden hält den Posten an und wartet, bis er steht.
//
// **Und es wartet wirklich** — sonst sagte der Stand „nicht mehr", während
// der Lauscher den Port noch hält, und das nächste Starten scheiterte an
// einer Adresse, die schon belegt ist.
func (a *App) horchenBeenden() {
	// **Erst die Ankündigung, dann der Port.** Wer zuerst den Port
	// schlösse, ließe eine Einladung im Netz stehen, die ins Leere führt.
	a.sichtbarkeitBeenden()

	a.peerlage.horchen.schloss.Lock()
	aufhoeren, fertig := a.peerlage.horchen.aufhoeren, a.peerlage.horchen.fertig
	a.peerlage.horchen.aufhoeren, a.peerlage.horchen.fertig = nil, nil
	a.peerlage.horchen.adresse = ""
	a.peerlage.horchen.erreichbar = leitung.Erreichbarkeit{}
	a.peerlage.horchen.schloss.Unlock()

	if aufhoeren == nil {
		return
	}

	aufhoeren()
	<-fertig
}

// horchstand liest den Stand.
func (a *App) horchstand() Horchzustand {
	a.peerlage.horchen.schloss.Lock()
	defer a.peerlage.horchen.schloss.Unlock()

	return Horchzustand{
		Laeuft:             a.peerlage.horchen.aufhoeren != nil,
		Adresse:            a.peerlage.horchen.adresse,
		ErreichbarNamen:    ohneNull(a.peerlage.horchen.erreichbar.Namen),
		ErreichbarAdressen: ohneNull(a.peerlage.horchen.erreichbar.Adressen),
		Sichtbar:           a.peerlage.horchen.anmeldung != nil,
	}
}

// MitHorcher setzt die Naht zum Lauscher — für Läufe.
//
// **Als Naht und nicht als Schalter** (§3): ein Lauf, der einen echten Port
// öffnete, hinge an der lokalen Netzwerkberechtigung dieses Rechners und an
// dem, was gerade sonst auf ihm horcht.
func (a *App) MitHorcher(
	horchen func(
		ctx context.Context, adresse string, ich peer.Identitaet,
	) (service.Hereinkommend, error),
) *App {
	a.peerlage.horchen.horchen = horchen

	return a
}

// horcher gibt die eingesetzte Naht — oder die echte.
func (a *App) horcher() func(
	context.Context, string, peer.Identitaet,
) (service.Hereinkommend, error) {
	if a.peerlage.horchen.horchen != nil {
		return a.peerlage.horchen.horchen
	}

	return echterHorcher
}

// echterHorcher öffnet einen Port und packt ihn in die Schnittstelle.
func echterHorcher(
	ctx context.Context, adresse string, ich peer.Identitaet,
) (service.Hereinkommend, error) {
	horcher, fehler := leitung.Horche(ctx, adresse, ich)
	if fehler != nil {
		return nil, fehler
	}

	return horcherkiste{innen: horcher}, nil
}

// horcherkiste hängt `internal/leitung` an [service.Hereinkommend].
//
// **Zwei Zeilen, und sie sind der Grund für den Composition Root** (§2): der
// Rückgabetyp ist ein `*leitung.Leitung`, die Schnittstelle will eine
// `service.Verbindung`, und keines der beiden Pakete darf das andere kennen.
type horcherkiste struct{ innen *leitung.Horcher }

func (h horcherkiste) Nimm(
	ctx context.Context,
) (service.Verbindung, error) {
	return h.innen.Nimm(ctx)
}

func (h horcherkiste) Schliesse() error { return h.innen.Schliesse() }

// Adresse gibt, wo dieser Lauscher wirklich steht.
func (h horcherkiste) Adresse() string { return h.innen.Adresse() }

// MitErreichbarkeit setzt die Naht zu den Adressen dieses Rechners.
func (a *App) MitErreichbarkeit(
	statt func(context.Context, string) (leitung.Erreichbarkeit, error),
) *App {
	a.peerlage.horchen.erfragen = statt

	return a
}

// erreichbarkeit gibt die eingesetzte Naht — oder die echte.
func (a *App) erreichbarkeit() func(
	context.Context, string,
) (leitung.Erreichbarkeit, error) {
	if a.peerlage.horchen.erfragen != nil {
		return a.peerlage.horchen.erfragen
	}

	return leitung.ErreichbarUnter
}

// ohneNull macht aus einer fehlenden Liste eine leere.
//
// **Weil `nil` als `null` hinausgeht.** Der Zustand ohne Horchposten trägt
// keine Erreichbarkeit, und das Fenster bekäme dann keine leere Liste,
// sondern gar keine — die Klasse aus §3, die
// `leerelistenlaufzeit_test.go` hält.
func ohneNull(liste []string) []string {
	if liste == nil {
		return []string{}
	}

	return liste
}
