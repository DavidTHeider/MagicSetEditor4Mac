package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was ein Nebenlauf von Partie und Peer tut, wenn er zusammenbricht** —
// BK-05a (#2111,
// Vermerk 0171).
//
// Eine Panik in einer Goroutine, die ein Ruf selbst startet, fängt niemand —
// der Prozess endete, mitsamt jedem anderen Tisch, Handel und der offenen
// Bearbeitung. Hier steht, was jeder dieser Läufe stattdessen tut: **er endet
// allein, und er gibt frei, was ihm gehört**.

// zusammenbruchProtokollieren schreibt den Grund ins Protokoll — mit der
// Vorgangskennung des Kontexts und über den schwärzenden Schreiber.
func zusammenbruchProtokollieren(ctx context.Context, lauf string, zusammenbruch error) {
	slog.ErrorContext(ctx, "ein Nebenlauf ist zusammengebrochen",
		"stelle", lauf, "grund", zusammenbruch.Error())
}

// amFremdenTisch startet einen Lauf des Gastes — und räumt, wenn er
// zusammenbricht, wie bei einer abgerissenen Leitung: die Sitzung geht zu,
// der Beitritt pausiert, der Platz gehört dem Gast weiter
// ([App.tischVerloren]).
//
// **Ein Fang für alle drei Läufe** — Partie erfragen, am Tisch sitzen, die
// Rückkehr melden: sie gehören derselben Sitzung, und was nach einem
// Zusammenbruch zu tun ist, ist für alle dasselbe.
func (a *App) amFremdenTisch(
	ctx context.Context, griff service.Angenommene, lauf uint64, schleife func(),
) {
	// **Gezählt für den Schluss** (BK-07a): er wartet, bis der Tisch endet,
	// bevor der Bestand zugeht.
	a.lauf.kinder.Add(1)

	go func() {
		defer a.lauf.kinder.Done()
		defer domain.Auffangen(func(zusammenbruch error) {
			zusammenbruchProtokollieren(ctx, "der Tisch eines Wirts", zusammenbruch)
			schliessePartiesitzung(ctx, griff)
			a.tischVerloren(lauf, zusammenbruch)
		})

		schleife()
	}()
}

// horchenZusammengebrochen räumt den Stand eines Horchpostens, dessen Lauf
// zusammenbrach — **wenn es noch dieser ist**: ein Beenden und ein neuer
// Start dazwischen gehören einem anderen Lauf.
//
// **Ohne auf `fertig` zu warten**, denn dieser Ruf läuft in genau der
// Goroutine, die es schließt.
func (a *App) horchenZusammengebrochen(fertig chan struct{}, zusammenbruch error) {
	meldeHorchbefund(zusammenbruch)

	a.peerlage.horchen.schloss.Lock()

	dieser := a.peerlage.horchen.fertig == fertig
	if dieser {
		a.peerlage.horchen.aufhoeren()
		a.peerlage.horchen.aufhoeren, a.peerlage.horchen.fertig = nil, nil
		a.peerlage.horchen.adresse = ""
		a.peerlage.horchen.erreichbar = leitung.Erreichbarkeit{}
	}

	a.peerlage.horchen.schloss.Unlock()

	// **Und die Ankündigung geht mit**: sie führte zu einem Port, an dem
	// niemand mehr annimmt.
	if dieser {
		a.sichtbarkeitBeenden()
	}
}

// fernhorchenZusammengebrochen ist dasselbe für fremde Netze — die Einladung
// führte sonst zu einem Posten, der nicht mehr annimmt.
func (a *App) fernhorchenZusammengebrochen(fertig chan struct{}, zusammenbruch error) {
	meldeHorchbefund(zusammenbruch)

	a.fernlage.schloss.Lock()
	defer a.fernlage.schloss.Unlock()

	if a.fernlage.fertig != fertig {
		return
	}

	a.fernlage.aufhoeren()
	a.fernlage.aufhoeren, a.fernlage.fertig = nil, nil
	a.fernlage.einladung = ""
	a.fernlage.horcher = nil
}

// meldeHorchbefund schreibt ins Protokoll, was niemanden sonst erreicht.
//
// **Ins Protokoll und nicht ans Fenster.** Ein Fremder, der Unsinn schickt,
// ist kein Ereignis für einen Menschen; wer daraus eine Meldung machte, hätte
// eine Fläche, die jeder von außen auslösen kann.
//
// **Ein Zusammenbruch steht als Fehler da** (BK-05a, #2111): er ist kein
// Unsinn von außen, sondern ein Fehler dieses Programms.
func meldeHorchbefund(fehler error) {
	if errors.Is(fehler, domain.ErrZusammengebrochen) {
		slog.Default().Error("ein Horchposten ist zusammengebrochen", "grund", fehler.Error())

		return
	}

	slog.Default().Info("eine hereinkommende Verbindung kam nicht zustande",
		"grund", fehler)
}
