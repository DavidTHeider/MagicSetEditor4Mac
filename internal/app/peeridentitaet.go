package app

// **Die Peeridentität an der Naht** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Zwei Rufe und nicht einer
//
// **Lesen und Anlegen sind zwei Entscheidungen.** Wie es steht, darf jede
// Fläche fragen; eine Identität anzulegen ist eine Handlung, und eine zu
// **rotieren** ist eine, die jedem bekannten Kontakt eine neue Prüfung
// abverlangt ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PeeridentitaetAntwort sagt, womit sich dieses Profil ausweist.
type PeeridentitaetAntwort struct {
	// Stand ist einer der vier: `ohne`, `bereit`, `verloren`, `fremd`.
	//
	// **Drei davon sind kein Fehler** — jeder verlangt eine andere Handlung,
	// und keiner darf still behoben werden ([domain.Peerstand]).
	Stand string `json:"stand"`

	// Kennung ist die öffentliche Peerkennung — leer, wenn es keine gibt.
	Kennung string `json:"kennung"`

	// Fingerabdruck ist die Form, die zwei Menschen vergleichen.
	//
	// **Er steht hier und wird nicht von der Fläche gerechnet**: eine Fläche,
	// die ihn selbst rechnen müsste, kann ihn vergessen.
	Fingerabdruck string `json:"fingerabdruck"`

	// Seit ist der Zeitpunkt, an dem diese Kennung entstanden ist — leer,
	// wenn es keine gibt.
	Seit string `json:"seit"`

	// Vorher ist die Kennung vor der letzten Rotation — leer, wenn es keine
	// gab.
	Vorher string `json:"vorher"`

	// VorherFingerabdruck ist ihr Abdruck — er ist das, was ein Kontakt
	// bestätigt hat.
	VorherFingerabdruck string `json:"vorherFingerabdruck"`

	// Befunde erklären, was geschehen ist — **als Liste und nicht als Satz**:
	// „verloren" ohne den Grund wäre eine Meldung, mit der niemand etwas tun
	// kann (§8).
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// Peeridentitaet sagt, wie es um die eigene Identität steht — **und ändert
// nichts**.
func (a *App) Peeridentitaet(ctx context.Context) PeeridentitaetAntwort {
	ctx = a.imProfil(ctx)

	lage, fehler := a.peerwerk().Lage(ctx)
	if fehler != nil {
		return peerfehler(ctx, fehler)
	}

	return alsPeerantwort(lage)
}

// PeeridentitaetSicherstellen legt eine an, **wenn es noch keine gibt**.
//
// **Und nur dann.** Bei einem Verlust oder einem Widerspruch antwortet sie
// mit einem Konflikt und sagt, was zu tun ist — eine neue Identität dort
// anzulegen wäre das stille Übernehmen, das das Issue ausschließt.
func (a *App) PeeridentitaetSicherstellen(ctx context.Context) PeeridentitaetAntwort {
	ctx = a.imProfil(ctx)

	lage, fehler := a.peerwerk().Sicherstellen(ctx)
	if fehler != nil {
		return peerfehler(ctx, fehler)
	}

	return alsPeerantwort(lage)
}

// PeeridentitaetRotieren legt **ausdrücklich** eine neue an.
//
// **Auch aus einem Verlust und einem Widerspruch heraus** — das ist der Weg
// hinaus. Was es kostet, steht in der Antwort: wer den alten Fingerabdruck
// bestätigt hat, muss erneut prüfen.
func (a *App) PeeridentitaetRotieren(ctx context.Context) PeeridentitaetAntwort {
	ctx = a.imProfil(ctx)

	lage, fehler := a.peerwerk().Rotieren(ctx)
	if fehler != nil {
		return peerfehler(ctx, fehler)
	}

	return alsPeerantwort(lage)
}

// alsPeerantwort macht aus der Lage ihr DTO.
//
// **Die Zeitpunkte als Text in RFC 3339** — dieselbe Gestalt wie überall
// sonst an dieser Naht; eine Zeit ohne Zone wäre beim ersten Sprung über
// Mitternacht ein anderer Tag.
func alsPeerantwort(lage service.Peerlage) PeeridentitaetAntwort {
	antwort := PeeridentitaetAntwort{
		Stand:         string(lage.Stand),
		Kennung:       string(lage.Kennung),
		Fingerabdruck: lage.Fingerabdruck,
		Vorher:        string(lage.Vorher),
		Befunde:       lage.Befunde,
	}

	if !lage.Seit.IsZero() {
		antwort.Seit = lage.Seit.UTC().Format(time.RFC3339)
	}

	antwort.VorherFingerabdruck = lage.VorherAbdruck

	// **Eine leere Liste und kein `nil`** — sonst verschwände sie aus der
	// Antwort, und die Seite bliebe ohne Grund leer (§3, „zwischen den
	// Prüfebenen").
	if antwort.Befunde == nil {
		antwort.Befunde = []string{}
	}

	return antwort
}

// peerfehler ist die Antwort, wenn ein Weg scheitert.
//
// **Mit leerer Liste und nicht mit `null`** — der Wächter
// `leerelistenlaufzeit_test.go` hält das für **jede** Bindung, und der Grund
// steht bei [alsPeerantwort]: eine Liste, die aus der Antwort verschwindet,
// lässt das Fenster ohne Grund leer stehen.
func peerfehler(ctx context.Context, fehler error) PeeridentitaetAntwort {
	return PeeridentitaetAntwort{
		Befunde: []string{}, Fehler: Uebersetze(ctx, fehler),
	}
}
