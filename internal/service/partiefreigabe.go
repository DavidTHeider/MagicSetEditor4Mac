package service

// **Welcher Sitz einem Gegenüber gehört** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die Frage, die eine Sitzung nicht beantwortet
//
// **Sie belegt einen Rechner, keinen Sitz.** Der Handschlag prüft die
// Peerkennung; welcher Sitz am anderen Ende spielt, steht darin nicht — und
// darf deshalb nicht von dort kommen. Hier entscheidet es der **Wirt**: er
// gibt eine Partie frei und nennt dabei die Sitze, die zu haben sind.
//
// # Warum „derselbe bekommt denselben" die eigentliche Zusage ist
//
// **Weil eine Leitung abreißt.** Wer nach einem Abriss wiederkommt, ist
// derselbe Spieler und kein neuer; bekäme er einen zweiten Sitz, säße er
// plötzlich am anderen Ende des Tisches, und die Partie hätte einen
// Teilnehmer zu viel. Genau daran hängt auch die Wiederaufnahme: sie fragt
// nach **seinem** Stand, und den gibt es nur, wenn er derselbe ist.
//
// # Und warum die Zuordnung hier steht und nicht im Kontakt
//
// **Weil sie einer Partie gehört und nicht einer Bekanntschaft.** Ein
// Peerkontakt sagt, **ob** jemand mitspielen darf ([domain.ZweckPartie]);
// **wo er sitzt**, ist je Partie verschieden — derselbe Mensch ist heute Bert
// und morgen Cora.

import (
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Sitzvergabe ist ein Sitz und das Gegenüber, das ihn hat.
type Sitzvergabe struct {
	Sitz       string
	Gegenueber peer.Kennung
}

// Partiefreigabe hält, welche Sitze zu haben sind und wer sie hat.
//
// **Mit Schloss**, weil der Horchposten für jede hereinkommende Verbindung
// eine eigene Goroutine startet: zwei Gäste können sich im selben Augenblick
// melden, und ohne Schloss bekämen beide denselben Sitz.
type Partiefreigabe struct {
	schloss sync.Mutex

	partie   string
	frei     []string
	vergeben map[peer.Kennung]string

	// eingeladen hält Sitze für ihre Gäste frei (P9b, #2377): Peer → Sitz.
	eingeladen map[peer.Kennung]string
}

// NeuePartiefreigabe gibt eine Partie mit diesen Sitzen frei.
//
// **Die Reihenfolge zählt**: wer sich zuerst meldet, bekommt den ersten. Eine
// Vergabe nach Zufall machte aus zwei gleichen Läufen zwei Ergebnisse.
func NeuePartiefreigabe(partie string, sitze []string) *Partiefreigabe {
	frei := make([]string, len(sitze))
	copy(frei, sitze)

	return &Partiefreigabe{
		partie:   partie,
		frei:     frei,
		vergeben: map[peer.Kennung]string{},
	}
}

// Partie gibt die Kennung der freigegebenen Partie.
func (f *Partiefreigabe) Partie() string { return f.partie }

// Sitz gibt den Sitz dieses Gegenübers — beim ersten Mal den nächsten freien.
func (f *Partiefreigabe) Sitz(gegenueber peer.Kennung) (string, error) {
	f.schloss.Lock()
	defer f.schloss.Unlock()

	if sitz, da := f.vergeben[gegenueber]; da {
		return sitz, nil
	}

	stelle := f.naechsterFuer(gegenueber)
	if stelle < 0 {
		// **Ein Konflikt und kein Validierungsfehler**: der Gast hat nichts
		// falsch gemacht, der Tisch ist voll.
		return "", fmt.Errorf(
			"%w: an der Partie %q ist kein Sitz mehr frei",
			domain.ErrConflict, f.partie)
	}

	sitz := f.frei[stelle]
	f.frei = append(f.frei[:stelle:stelle], f.frei[stelle+1:]...)
	f.vergeben[gegenueber] = sitz

	return sitz, nil
}

// MitEinladungen hält eingeladene Sitze für ihre Gäste frei (P9b, #2377) —
// Sitz → Peerkennung, wie die Sicht des Wirts sie nennt.
func (f *Partiefreigabe) MitEinladungen(sitzZuPeer map[string]string) *Partiefreigabe {
	f.schloss.Lock()
	defer f.schloss.Unlock()

	frei := make(map[string]bool, len(f.frei))
	for _, sitz := range f.frei {
		frei[sitz] = true
	}

	// **Nur auf einen Sitz, der zu haben ist** — eine Einladung auf einen
	// anderen hielte nichts frei und ließe den Gast ohne Platz.
	f.eingeladen = make(map[peer.Kennung]string, len(sitzZuPeer))
	for sitz, gast := range sitzZuPeer {
		if frei[sitz] {
			f.eingeladen[peer.Kennung(gast)] = sitz
		}
	}

	return f
}

// naechsterFuer gibt die Stelle des Sitzes für dieses Gegenüber in der Liste
// der freien — seinen, wenn er eingeladen ist, sonst den ersten, auf den
// niemand eingeladen ist; -1, wenn es keinen gibt.
func (f *Partiefreigabe) naechsterFuer(gegenueber peer.Kennung) int {
	reserviert := make(map[string]bool, len(f.eingeladen))
	for _, sitz := range f.eingeladen {
		reserviert[sitz] = true
	}

	eigener, eingeladen := f.eingeladen[gegenueber]

	for stelle, sitz := range f.frei {
		switch {
		case eingeladen && sitz == eigener:
			return stelle
		case !reserviert[sitz] && !eingeladen:
			return stelle
		}
	}

	return -1
}

// Vergeben gibt, wer worauf sitzt.
//
// **Eine Kopie und keine Sicht auf die Abbildung** — wer sie hielte, läse
// später einen Stand, der sich unter ihm ändert.
func (f *Partiefreigabe) Vergeben() []Sitzvergabe {
	f.schloss.Lock()
	defer f.schloss.Unlock()

	vergeben := make([]Sitzvergabe, 0, len(f.vergeben))

	for gegenueber, sitz := range f.vergeben {
		vergeben = append(vergeben, Sitzvergabe{
			Sitz: sitz, Gegenueber: gegenueber,
		})
	}

	return vergeben
}

// Einladungen nennt je eingeladenem Sitz seinen Gast — aus der Sicht des
// Wirts, denn nur dort steht er (P9b, #2377).
func Einladungen(sicht domain.Partiesicht) map[string]string {
	einladungen := map[string]string{}

	for _, sitz := range sicht.Sitze {
		if sitz.Eingeladen != "" {
			einladungen[sitz.Kennung] = sitz.Eingeladen
		}
	}

	return einladungen
}
