package service

// **Eine Partie über eine Leitung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Was hier geschieht und was nicht
//
// **Hier wird verpackt, nummeriert und gekettet.** Was eine Nachricht
// bedeutet, weiß `domain.Partienachricht`; wie sie über den Draht kommt, weiß
// `internal/leitung`. Diese Stelle ist dazwischen — und sie ist die einzige,
// die beides zugleich kennt.
//
// # Warum die Zustellart aus der Art folgt
//
// **Weil sie keine Wahl des Aufrufers ist.** Ein Mauszeiger geht verwerfbar,
// ein Ereignis zuverlässig — das hängt daran, **was** die Nachricht ist, und
// nicht daran, wer sie gerade schickt. Wer es dem Aufrufer überließe, bekäme
// irgendwann ein verwerfbares Ereignis, und die Partie hätte eine Lücke, die
// niemand erklären kann.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partienaht ist, was diese Sitzung zum Senden braucht.
//
// **Eine Schnittstelle und nicht `*Sitzungsgriff`**: so lässt sich die
// Reihenfolge prüfen, ohne eine Leitung aufzubauen — dieselbe Richtung wie
// bei den Repositorien (§2).
type Partienaht interface {
	Sende(ctx context.Context, nachricht []byte, art Zustellung) error
}

// Partiesitzung führt eine Partie über eine Leitung.
type Partiesitzung struct {
	// schloss hält Vergeben und Absenden zusammen.
	//
	// # Warum es die Reihenfolge nicht allein tut
	//
	// **Weil sie nur die Nummer schützt, nicht den Weg dorthin.** Zwischen
	// dem Vergeben einer Folge und dem Absenden liegt ein Fenster; wer die
	// Fünf schon hat und dort unterbrochen wird, legt sie **nach** der Sechs
	// auf die Leitung. Beim Empfänger ist das eine **Lücke**, und eine Lücke
	// legt die Leitung nieder — ein Satz im Gespräch risse den Tisch ab.
	//
	// **Es gab den zweiten Absender vorher nicht.** Bis zum Gespräch am Tisch
	// schickte allein die Schleife einer Gastsitzung; seither schickt auch
	// das Fenster, aus einer anderen Goroutine über dieselbe Sitzung (§7).
	//
	// **Und es hält bis über das Senden hinaus, nicht nur über das Packen.**
	// Ein Schloss, das vor `naht.Sende` fiele, ließe genau das Fenster offen,
	// das es schließen soll. Der Preis ist gering: [Sitzungsgriff.Sende] legt
	// nur in die Warteschlange, hinaus geht es mit `Leere`.
	schloss sync.Mutex

	partie string
	sitz   string
	naht   Partienaht
	reihe  *Reihenfolge

	// kette ist das zuletzt vergebene Glied.
	//
	// **Sie steht hier und nicht in der Reihenfolge**: jene vergibt Nummern
	// und weiß nichts von Inhalten. Zusammengelegt hätte eine Stelle zwei
	// Gründe, sich zu ändern.
	kette string
}

// NeuePartiesitzung baut eine Sitzung für einen Sitz.
func NeuePartiesitzung(
	partie, sitz string, naht Partienaht, reihe *Reihenfolge,
) *Partiesitzung {
	return &Partiesitzung{partie: partie, sitz: sitz, naht: naht, reihe: reihe}
}

// Sende verpackt eine Nachricht und schickt sie.
//
// **Die Prüfung läuft vor dem Senden** (§8): was die Gestalt nicht einhält,
// geht gar nicht erst hinaus — sonst stünde die Meldung beim Empfänger, und
// der kann nichts dafür.
func (s *Partiesitzung) Sende(
	ctx context.Context, art domain.Nachrichtenart, inhalt []byte,
) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	nachricht, fehler := s.packe(art, inhalt)
	if fehler != nil {
		return fehler
	}

	roh, packfehler := json.Marshal(nachricht)
	if packfehler != nil {
		return fmt.Errorf("die Nachricht ließ sich nicht verpacken: %w", packfehler)
	}

	return s.naht.Sende(ctx, roh, zustellungFuer(art))
}

// packe baut die Nachricht und schreibt Folge und Kette fort.
func (s *Partiesitzung) packe(
	art domain.Nachrichtenart, inhalt []byte,
) (domain.Partienachricht, error) {
	if fehler := art.Pruefe(); fehler != nil {
		return domain.Partienachricht{}, fehler
	}

	nachricht := domain.Partienachricht{
		Art:    art,
		Partie: s.partie,
		Von:    s.sitz,
		Inhalt: inhalt,
	}

	// **Flüchtiges verbraucht keine Folgenummer.** Sonst entstünde bei jedem
	// Mauszeiger eine Lücke im Strom der Ereignisse — und die blockiert beim
	// Empfänger.
	if !art.WirdAufbewahrt() {
		nachricht.Epoche, _ = s.reihe.Stand()

		return nachricht, nachricht.Pruefe()
	}

	nachricht.Epoche, nachricht.Folge = s.reihe.Naechste()

	if art.GehtInDieKette() {
		s.kette = kettenglied(s.kette, nachricht)
		nachricht.Kette = s.kette
	}

	return nachricht, nachricht.Pruefe()
}

// Kette gibt das zuletzt vergebene Glied.
func (s *Partiesitzung) Kette() string {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	return s.kette
}

// zustellungFuer folgt aus der Art — siehe den Kopf dieser Datei.
func zustellungFuer(art domain.Nachrichtenart) Zustellung {
	if art.WirdAufbewahrt() {
		return Zuverlaessig
	}

	return Verwerfbar
}

// kettenglied bindet ein Ereignis an das vorige.
//
// **Dieselbe Gestalt wie in G1** (#1458):
// das vorige Glied geht mit ein, und deshalb ändert ein ausgetauschtes
// Ereignis jedes Glied danach.
func kettenglied(vorher string, nachricht domain.Partienachricht) string {
	// **`hash.Hash` klagt nie** — das steht so in seiner Zusage, und deshalb
	// gehen die Rückgaben hier ausdrücklich ins Leere statt in eine Prüfung,
	// die nie zutrifft.
	summe := sha256.New()
	summe.Write([]byte(vorher))           //nolint:errcheck // hash.Hash klagt nie
	summe.Write([]byte(nachricht.Partie)) //nolint:errcheck // dito
	summe.Write([]byte(nachricht.Von))    //nolint:errcheck // dito
	summe.Write(fmt.Appendf(nil, "%d:%d", //nolint:errcheck // dito
		nachricht.Epoche, nachricht.Folge))
	summe.Write(nachricht.Inhalt) //nolint:errcheck // dito

	return hex.EncodeToString(summe.Sum(nil))
}

// leerbar ist eine Naht, die auch hinausgeben kann — im Betrieb der
// [Sitzungsgriff].
type leerbar interface {
	Leere(ctx context.Context) error
}

// Leere gibt hinaus, was wartet — wenn die Naht das kann.
//
// **Für den, der in einem Zug mehr schickt, als die Schlange fasst** — den
// Wirt beim Nachreichen. Die Schlange ist ein Rückstau mit
// `app.Sitzungstiefe` Plätzen; wer erst am Ende leerte, fände sie voll
// (Messlauf über Phase 6.7, #1668). **Eine Naht ohne Leeren** hat nichts
// zurückgehalten und deshalb nichts zu tun.
func (s *Partiesitzung) Leere(ctx context.Context) error {
	if naht, kann := s.naht.(leerbar); kann {
		return naht.Leere(ctx)
	}

	return nil
}

// schliessbar ist eine Naht, die ihre Leitung niederlegen kann.
type schliessbar interface {
	Schliesse() error
}

// Schliesse legt die Leitung nieder — wenn die Naht das kann.
//
// **Für den, der nebenher schickt** (Nachlauf zum Zweitblick über Phase 7,
// R3): scheitert sein Versand, gibt es niemanden, dem er den Fehler
// zurückgeben könnte. Die Leitung niederzulegen sagt es der Schleife, die
// sie liest — sie endet, und mit ihr die Sitzung.
func (s *Partiesitzung) Schliesse() error {
	if naht, kann := s.naht.(schliessbar); kann {
		return naht.Schliesse()
	}

	return nil
}
