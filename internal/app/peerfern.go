package app

// **Über fremde Netze erreichbar sein** — Werkstück N3 von Phase 5.4
// (#1456).
//
// # Warum das ein zweiter Schalter ist und keine Einstellung am ersten
//
// **Weil es eine andere Preisgabe ist.** Wer im eigenen Netz horcht, ist für
// die erreichbar, die im selben Raum sind. Wer eine Einladung ausgibt, ist für
// jeden erreichbar, der sie bekommt — über ein Relais, das jemand anderes
// betreibt. Zwei Entscheidungen, zwei Schalter; dieselbe Linie wie zwischen
// „erreichbar sein" und „sich ankündigen" in N2.
//
// # Und warum jede Einladung stirbt, wenn das Horchen endet
//
// **Weil der Serverschlüssel nicht aufbewahrt wird.** tailcat erzeugt einen
// flüchtigen, wenn keiner gesetzt ist — und die Einladung trägt ihn. Wer
// aufhört zu horchen, macht damit **jede** ausgegebene Einladung ungültig, und
// das nächste Horchen gibt eine neue.
//
// **Das ist Widerruf und Kurzlebigkeit, ohne einen Zähler zu führen.** #1456
// verlangt „Kurzlebigkeit, Einmalnutzung, Ablauf, Versuchslimit und Widerruf"
// als unsere Schicht über der Adresse. Drei davon fallen hier zu; **Ablauf und
// Versuchslimit gibt es nicht**, und das steht hier, damit niemand sie
// annimmt: eine Einladung gilt, solange gehorcht wird.

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fernleitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fernfelder ist, was eine Anwendung für den Weg über fremde Netze hält.
//
// **Neben `peerfelder` und nicht darin**: es ist eine zweite Entscheidung mit
// einer anderen Preisgabe, und zwei Lebenszyklen in einem Verbund ließen sich
// nur gemeinsam beenden.
type fernfelder struct {
	schloss sync.Mutex

	// einladung ist, was ein Gegenüber zum Wählen braucht — leer heißt:
	// dieses Profil ist über fremde Netze nicht erreichbar.
	einladung domain.Einladung

	// horcher ist der laufende — nil heißt: keiner. Er wird gehalten, weil
	// er die Wege kennt (NET-02.AK-03).
	horcher Fernhorcher

	aufhoeren context.CancelFunc
	fertig    chan struct{}

	// faengtAn sagt, dass gerade einer entsteht — derselbe Grund wie bei
	// [horchfelder.faengtAn], und hier wiegt er schwerer: der Tunnel geht
	// durch ein Relais, und das dauert.
	faengtAn bool

	// waehlen ist die Naht zum Wählen über ein fremdes Netz — nil heißt:
	// die echte.
	//
	// **Als Feld, aus demselben Grund wie `horchen` darunter** (§3): ein
	// Lauf, der wirklich wählte, holte eine DERP-Karte aus dem Internet.
	waehlen func(
		ctx context.Context, einladung domain.Einladung,
		ich peer.Identitaet, erwartet peer.Kennung,
	) (service.Verbindung, error)

	// horchen ist die Naht zum Fernhorcher — nil heißt: der echte.
	//
	// **Als Feld, damit ein Lauf ohne Internet auskommt** (§3), aus
	// demselben Grund wie bei den Nähten des Horchpostens.
	horchen func(peer.Identitaet) (Fernhorcher, error)
}

// Fernhorcher ist, was diese Fassade von einem Horcher über fremde Netze
// braucht.
type Fernhorcher interface {
	service.Hereinkommend

	// Einladung ist, was ein Gegenüber zum Wählen braucht.
	Einladung() domain.Einladung

	// Wege sagt, wie die Verbindungen gerade gehen (NET-02.AK-03).
	Wege() []fernleitung.Fernweg

	// Angebunden sagt, wie viele Ströme offen sind (SEC-21) — Wege nennt
	// Gegenstellen, seit tailcat v0.7.0 (K01), und keine Ströme.
	Angebunden() int
}

// FernHorchStarten gibt eine Einladung aus und nimmt Verbindungen entgegen.
func (a *App) FernHorchStarten(ctx context.Context) Fernantwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.fernhorchenStarten(ctx)
	if fehler != nil {
		return Fernantwort{Stand: ohneEinladung(), Fehler: Uebersetze(ctx, fehler)}
	}

	return Fernantwort{Stand: stand}
}

// FernHorchBeenden macht dieses Profil über fremde Netze unerreichbar.
//
// **Und entwertet damit jede ausgegebene Einladung** — der Grund steht im
// Kopf dieser Datei.
func (a *App) FernHorchBeenden(ctx context.Context) Fernantwort {
	ctx = a.imProfil(ctx)

	// **Der Kontext wird gebraucht, auch wenn er hier nichts tut** — aus
	// demselben Grund wie bei [App.HorchBeenden]: ohne ihn stünde diese
	// Bindung außerhalb der Umschaltsperre (§7, Regel 2).
	_ = ctx

	a.fernhorchenBeenden()

	return Fernantwort{Stand: a.fernstand()}
}

// FernHorchStand sagt, ob und unter welcher Einladung gehorcht wird.
func (a *App) FernHorchStand(ctx context.Context) Fernantwort {
	ctx = a.imProfil(ctx)
	_ = ctx

	return Fernantwort{Stand: a.fernstand()}
}

// fernhorchenStarten baut den Horcher und lässt den Posten laufen.
func (a *App) fernhorchenStarten(ctx context.Context) (Fernzustand, error) {
	// **Die Marke vor dem Start** (Nachlauf über Phase 6.12, N2).
	marke := a.laufmarke(domain.LaufFernhorchen)

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
		return Fernzustand{}, fehler
	}

	// **Der Ausweis vor dem Tunnel**, wie in N2: ein Profil ohne Identität
	// kann keinen Handschlag führen, und eine Einladung ohne ihn wäre eine
	// Tür, hinter der niemand steht.
	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return Fernzustand{}, fehler
	}

	// **Der Platz wird belegt, bevor er gebraucht wird** — derselbe Befund
	// und derselbe Grund wie beim Horchposten; hier dauert es länger, denn
	// der Tunnel geht durch ein Relais.
	a.fernlage.schloss.Lock()
	laeuftSchon := a.fernlage.aufhoeren != nil || a.fernlage.faengtAn
	if !laeuftSchon {
		a.fernlage.faengtAn = true
	}
	a.fernlage.schloss.Unlock()

	if laeuftSchon {
		return Fernzustand{}, domain.ErrFernhorchenLaeuft
	}

	defer func() {
		a.fernlage.schloss.Lock()
		a.fernlage.faengtAn = false
		a.fernlage.schloss.Unlock()
	}()

	horcher, fehler := a.fernhorcher()(ich)
	if fehler != nil {
		return Fernzustand{}, fehler
	}

	laufender, aufhoeren := context.WithCancel(context.WithoutCancel(ctx))
	fertig := make(chan struct{})

	a.fernlage.schloss.Lock()
	a.fernlage.einladung = horcher.Einladung()
	a.fernlage.horcher = horcher
	a.fernlage.aufhoeren = aufhoeren
	a.fernlage.fertig = fertig
	a.fernlage.schloss.Unlock()

	go func() {
		defer close(fertig)
		defer domain.Auffangen(func(zusammenbruch error) { a.fernhorchenZusammengebrochen(fertig, zusammenbruch) })

		// **Und was angenommen wird, wird auch bedient** (G5, #1462): eine
		// Sitzung, die nur in der Übersicht stünde, läse niemand; ob eine
		// Partie offensteht, entscheidet die Aufnahme.
		service.NeuerHorchposten(
			a.sitzungsverwaltung(), meldeHorchbefund).
			MitAufnahme(a.sitzungAufnehmen).
			Bediene(laufender, horcher)
	}()

	a.nachDemStart(ctx, domain.LaufFernhorchen, marke, func(context.Context) { a.fernhorchenBeenden() })

	return a.fernstand(), nil
}

// fernhorchenBeenden hält den Posten an und wartet, bis er steht.
func (a *App) fernhorchenBeenden() {
	a.fernlage.schloss.Lock()
	aufhoeren, fertig := a.fernlage.aufhoeren, a.fernlage.fertig
	a.fernlage.aufhoeren, a.fernlage.fertig = nil, nil
	a.fernlage.einladung = ""
	a.fernlage.horcher = nil
	a.fernlage.schloss.Unlock()

	if aufhoeren == nil {
		return
	}

	aufhoeren()
	<-fertig
}

// fernstand liest den Stand.
func (a *App) fernstand() Fernzustand {
	a.fernlage.schloss.Lock()
	stand := Fernzustand{
		Laeuft:    a.fernlage.aufhoeren != nil,
		Einladung: string(a.fernlage.einladung),
	}
	horcher := a.fernlage.horcher
	a.fernlage.schloss.Unlock()

	// **Die Wege außerhalb des Schlosses**: tailcat fragt dafür seinen
	// eigenen Stand ab, und ein Beenden soll darauf nicht warten.
	stand.Wege = []FernwegDTO{}
	if horcher != nil {
		stand.Wege = fernwege(horcher.Wege())
		stand.Angebunden = horcher.Angebunden()
	}

	return stand
}

// MitFernhorcher setzt die Naht — für Läufe.
func (a *App) MitFernhorcher(
	statt func(peer.Identitaet) (Fernhorcher, error),
) *App {
	a.fernlage.horchen = statt

	return a
}
