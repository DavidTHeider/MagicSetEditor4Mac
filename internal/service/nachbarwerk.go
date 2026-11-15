package service

// **Wer in der Nähe ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum ein Fund kein Kontakt ist
//
// **Weil er nichts belegt.** Wer sich im Netz meldet, sagt eine Kennung und
// einen Namen; beides kann jeder sagen. Was daraus wird, entscheidet ein
// Mensch, der einen Fingerabdruck vergleicht — und danach die Kontaktliste.
//
// **Dieses Werk bringt beides zusammen**, und das ist seine ganze Aufgabe: es
// hängt an jeden Fund, was dieses Profil über ihn weiß.

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Nachbarfund ist, was der Transport von einem Gegenüber hört.
//
// **Roh und ungeprüft** — die Kennung ist geprüft, dass sie eine **ist**;
// dass sie dem gehört, der sie ausruft, sagt erst der Handschlag.
type Nachbarfund struct {
	Kennung peer.Kennung
	Name    string

	// Adresse ist, wohin gewählt würde — die schnellste Form.
	Adresse string

	// Merkadresse ist, was sich aufzuschreiben lohnt: der `.local`-Name, wo
	// es einen gibt, sonst dieselbe Adresse
	// (#1633).
	//
	// **Beides und nicht eines**: eine IP von heute ist nach dem nächsten
	// DHCP-Wechsel falsch, und ein Kontakt zeigte dann auf einen Rechner,
	// der inzwischen jemand anderem gehört.
	Merkadresse string
}

// Nachbarfunde ist, was eine Suche hergibt.
type Nachbarfunde struct {
	Funde []Nachbarfund

	// Verworfen zählt die Antworten, die keine Peerkennung trugen.
	Verworfen int
}

// Nachbarschaft ist die Naht zum Transport.
//
// **Eine Schnittstelle und kein Blatt** (§2): **welcher** Transport sucht,
// entscheidet der Composition Root — mDNS heute, vielleicht mehr später.
type Nachbarschaft interface {
	Suche(ctx context.Context) (Nachbarfunde, error)
}

// Nachbar ist ein Fund samt dem, was dieses Profil über ihn weiß.
type Nachbar struct {
	Kennung       peer.Kennung
	Fingerabdruck string

	// Name ist der selbstgewählte — **kein Nachweis**.
	Name string

	// Adresse ist, wohin gewählt würde.
	Adresse string

	// Merkadresse ist, was in die Kontaktadresse gehört — der Name, wo es
	// einen gibt (#1633).
	Merkadresse string

	// Bekannt heißt: dieses Profil führt ihn in seiner Kontaktliste.
	Bekannt bool

	// Stand ist seiner — leer, solange er unbekannt ist.
	Stand domain.Kontaktstand

	// GemerkteAdresse ist, was die Kontaktliste über ihn weiß — leer, wo
	// sie nichts weiß (NET-02.NAT-02,
	// #1706).
	GemerkteAdresse string
}

// AdresseGeaendert sagt, ob ein **bekannter** Nachbar woanders steht, als
// die Kontaktliste ihn führt — dann gibt es eine Adresse zu übernehmen.
//
// **Übernehmen ändert nur, wohin gewählt wird**: das Vertrauen bleibt, wie es
// ist, und die Leitung prüft weiter den Schlüssel hinter der Kennung. Ein
// Fremder an der alten Adresse bekäme deshalb nichts, und einer an der neuen
// auch nicht.
func (n Nachbar) AdresseGeaendert() bool {
	return n.Bekannt && n.Merkadresse != "" && n.GemerkteAdresse != n.Merkadresse
}

// Nachbarn ist die Liste samt dem, was dabei nicht zu deuten war.
type Nachbarn struct {
	Nachbarn  []Nachbar
	Verworfen int
}

// Nachbarwerk beantwortet „wer ist da, und kenne ich ihn?".
type Nachbarwerk struct {
	nachbarschaft Nachbarschaft
	kontakte      *Kontaktwerk
}

// NeuesNachbarwerk baut es über einem Transport und den Kontakten.
func NeuesNachbarwerk(
	nachbarschaft Nachbarschaft, kontakte *Kontaktwerk,
) *Nachbarwerk {
	return &Nachbarwerk{nachbarschaft: nachbarschaft, kontakte: kontakte}
}

// eigenekennung ist der Schlüssel, unter dem die eigene Kennung im Kontext
// steht.
type eigenekennung struct{}

// MitEigenerKennung legt die eigene Kennung in den Kontext.
//
// # Warum sie durch den Kontext reist und kein Argument ist
//
// **Weil sie nicht zur Frage gehört.** „Wer ist in der Nähe" ist dieselbe
// Frage, gleich wer sie stellt; dass die eigene Antwort darauf nicht zählt,
// ist eine Eigenschaft des Rufenden. Als Argument stünde sie an jeder Fläche
// und wäre an jeder zu vergessen.
func MitEigenerKennung(
	ctx context.Context, kennung peer.Kennung,
) context.Context {
	return context.WithValue(ctx, eigenekennung{}, kennung)
}

// Liste sagt, wer da ist und was dieses Profil über ihn weiß.
func (w *Nachbarwerk) Liste(ctx context.Context) (Nachbarn, error) {
	funde, fehler := w.nachbarschaft.Suche(ctx)
	if fehler != nil {
		return Nachbarn{}, fehler
	}

	// **Die Zusicherung wird gelesen und nicht weggeworfen** — steht keine
	// eigene Kennung im Kontext, zählt jeder Fund; das ist der Fall eines
	// Profils, das sich gar nicht angekündigt hat.
	selbst, eigene := ctx.Value(eigenekennung{}).(peer.Kennung)
	if !eigene {
		selbst = ""
	}

	nachbarn := Nachbarn{
		Nachbarn:  make([]Nachbar, 0, len(funde.Funde)),
		Verworfen: funde.Verworfen,
	}

	for _, fund := range funde.Funde {
		// **Der eigene Widerhall zählt nicht.** mDNS gibt die eigene
		// Ankündigung zurück; wer sie zeigte, böte einen Knopf an, der
		// einen Kontakt zu sich selbst erzeugte.
		if selbst != "" && fund.Kennung == selbst {
			continue
		}

		nachbar, fehler := w.alsNachbar(ctx, fund)
		if fehler != nil {
			return Nachbarn{}, fehler
		}

		nachbarn.Nachbarn = append(nachbarn.Nachbarn, nachbar)
	}

	return nachbarn, nil
}

// alsNachbar hängt an einen Fund, was die Kontaktliste über ihn sagt.
//
// **Ein Fehler des Bestands ist kein „unbekannt"** (§8): „ich kenne ihn
// nicht" und „ich weiß nicht, ob ich ihn kenne" sind zwei Dinge, und wer sie
// gleich behandelte, lüde zum doppelten Eintragen ein.
func (w *Nachbarwerk) alsNachbar(
	ctx context.Context, fund Nachbarfund,
) (Nachbar, error) {
	nachbar := Nachbar{
		Kennung:       fund.Kennung,
		Fingerabdruck: peer.Fingerabdruck(fund.Kennung),
		Name:          fund.Name,
		Adresse:       fund.Adresse,
		Merkadresse:   fund.Merkadresse,
	}

	kontakt, fehler := w.kontakte.Lesen(ctx, fund.Kennung)

	switch {
	case errors.Is(fehler, domain.ErrNotFound):
		return nachbar, nil
	case fehler != nil:
		return Nachbar{}, fehler
	}

	nachbar.Bekannt = true
	nachbar.Stand = kontakt.Stand
	nachbar.GemerkteAdresse = kontakt.Adresse

	return nachbar, nil
}
