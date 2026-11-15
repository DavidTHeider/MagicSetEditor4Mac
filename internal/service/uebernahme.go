// **Ein Deck aus einem fremden Profil übernehmen** — Werkstück 12 von
// Phase 4.12
// (#1351),
// nach Vermerk [0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md).
//
// # Die Frage, die dieser Anwendungsfall beantwortet
//
// **„Das Deck da drüben hätte ich gern — und danach soll dastehen, woher es
// kommt."** So steht es in der Kommentierung: „die gehört einem zwar dann,
// kommt aber woanders her".
//
// # Warum das erst seit Werkstück 11 geht
//
// **Weil vorher niemand sagen konnte, ob man in das fremde Profil darf.** Die
// Übernahme stellt dieselbe Frage wie der Wechsel — seit Vermerk 0073 reicht
// dort `leser` —, und aus demselben Grund: wer hineingehen darf, darf auch
// mitnehmen; wer nur sieht, **dass** es das Profil gibt, sieht nicht hinein.
// `TestEinLeserDruebenDarfHolen` hält es.
//
// # Was **nicht** mitkommt
//
// **Ein Platz auf eine eigene Karte** und **das Format**. Beide sind Zeilen
// im **Quellprofil**, und dieselbe Zeile gibt es hier nicht; sie mitzunehmen
// hieße, einen Verweis über die Profilgrenze zu legen — genau das, was
// Werkstück 2 verhindert.
//
// **Genannt wird beides.** Ein Deck, von dem vierzig Karten kommen und zehn
// nicht, wird mit vierzig übernommen und zehn Befunden — dieselbe Haltung wie
// bei einem halb lesbaren MSE-Paket (§8).

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckentnahme liest ein Deck aus einem **anderen** Profil.
//
// # Die einzige Naht, die die Klemme aus Werkstück 2 absichtlich überschreitet
//
// **Und deshalb steht sie hier einzeln, mit einem Namen, den man suchen
// kann.** Sie liest und schreibt nicht, sie nennt das Profil ausdrücklich in
// der Signatur, und ihr einziger Aufrufer ist [Deckuebernahme.DeckUebernehmen] —
// der die Erlaubnis **davor** fragt.
type Deckentnahme interface {
	DeckAusProfil(ctx context.Context, profilID, deckID int64) (domain.Deck, error)

	// DeckeAusProfil nennt die Decks eines Profils — zur Wahl bei der
	// Übernahme (UX-08.AK-04, #1680). Dieselbe Grenze, derselbe Aufrufer.
	DeckeAusProfil(ctx context.Context, profilID int64) ([]domain.Deck, error)
}

// Deckuebernahmespeicher ist der Ausschnitt des Deckspeichers, den sie braucht.
//
// **Ein Ausschnitt und nicht [Deckspeicher]**: die Übernahme legt an. Ihr das
// Löschen mitzugeben hieße, einem Weg, der Daten **bringt**, das Wegnehmen in
// die Hand zu geben.
//
// **Und ein Ruf, nicht drei** —
// #1820:
// hier standen `Anlegen`, `PlatzSetzen` und `DeckherkunftSetzen` einzeln, und
// scheiterte ein späterer, blieb ein halbes Deck stehen, über dem die Fläche
// einen Fehler meldete. Eine Übernahme ist ein Anwendungsfall; der Speicher
// schreibt sie in **einer** Transaktion, oder gar nicht.
type Deckuebernahmespeicher interface {
	Uebernehmen(ctx context.Context, deck domain.Deck, plaetze []domain.Deckplatz,
		herkunft domain.Herkunft, quellDeck int64) (domain.Deck, error)
}

// Deckuebernahme ist der Anwendungsfall „ein Deck von drüben holen".
type Deckuebernahme struct {
	entnahme  Deckentnahme
	speicher  Deckuebernahmespeicher
	nutzer    *Nutzerverwaltung
	freigaben *Freigabenverwaltung
}

// NeueDeckuebernahme baut sie.
func NeueDeckuebernahme(
	entnahme Deckentnahme, speicher Deckuebernahmespeicher,
	nutzer *Nutzerverwaltung, freigaben *Freigabenverwaltung,
) *Deckuebernahme {
	return &Deckuebernahme{
		entnahme: entnahme, speicher: speicher,
		nutzer: nutzer, freigaben: freigaben,
	}
}

// DeckeZurUebernahme nennt die Decks eines fremden Profils — zur Wahl.
//
// **Dieselbe Erlaubnis wie die Übernahme, und davor** (Vermerk 0071): wer
// nicht hinein darf, erfährt auch nicht, welche Decks es dort gibt. Bis
// hierher war die Kennung eines Decks von drüben zu tippen, und niemand im
// eigenen Profil konnte sie sehen (UX-08.AK-04,
// #1680).
func (u *Deckuebernahme) DeckeZurUebernahme(
	ctx context.Context, quellProfil int64,
) ([]domain.Deck, error) {
	if _, _, fehler := u.erlaubnisPruefen(ctx, quellProfil); fehler != nil {
		return nil, fehler
	}

	return u.entnahme.DeckeAusProfil(ctx, quellProfil)
}

// DeckUebernehmen holt ein Deck aus einem fremden Profil in das eigene.
//
// **Die Befunde sind Teil des Ergebnisses und kein Fehler.** Was nicht
// mitkonnte, steht darin; das Deck ist trotzdem da.
func (u *Deckuebernahme) DeckUebernehmen(
	ctx context.Context, quellProfil, quellDeck int64,
) (domain.Deck, []string, error) {
	quelle, bezeichnung, fehler := u.erlaubnisPruefen(ctx, quellProfil)
	if fehler != nil {
		return domain.Deck{}, nil, fehler
	}

	// **Gelesen wird erst danach.** Wer nicht hinein darf, soll auch nicht
	// erfahren, was drinsteht — dieselbe Reihenfolge wie beim Wechsel.
	fremdes, lesefehler := u.entnahme.DeckAusProfil(ctx, quellProfil, quellDeck)
	if lesefehler != nil {
		return domain.Deck{}, nil, lesefehler
	}

	// **Erst entscheiden, was mitkommt — dann in einem Zug schreiben.**
	befunde := u.formatbefund(fremdes)
	plaetze := make([]domain.Deckplatz, 0, len(fremdes.Plaetze))

	for _, platz := range fremdes.Plaetze {
		if befund := platzbefund(platz); befund != "" {
			befunde = append(befunde, befund)

			continue
		}

		plaetze = append(plaetze, domain.Deckplatz{
			Printing:    platz.Printing,
			Menge:       platz.Menge,
			Bereich:     platz.Bereich,
			Bezeichnung: platz.Bezeichnung,
		})
	}

	// **Die Herkunft kommt mit** — sie sagt „dieses Deck kam von dort", und
	// sie entsteht im selben Zug wie das Deck.
	angelegt, schreibfehler := u.speicher.Uebernehmen(ctx, domain.Deck{
		Name:  fremdes.Name,
		Spiel: fremdes.Spiel,
		Notiz: fremdes.Notiz,
	}, plaetze, domain.Herkunft{
		Art:         domain.HerkunftProfil,
		Quelle:      bezeichnung,
		QuellProfil: quelle.ID,
	}, quellDeck)
	if schreibfehler != nil {
		return domain.Deck{}, nil, schreibfehler
	}

	return angelegt, befunde, nil
}

// erlaubnisPruefen fragt, ob der arbeitende Mensch aus diesem Profil nehmen
// darf — und gibt es zurück, weil sein **Name** in die Herkunft gehört.
func (u *Deckuebernahme) erlaubnisPruefen(ctx context.Context, quellProfil int64) (
	quelle domain.Profil, bezeichnung string, fehler error,
) {
	gezeigt, lesefehler := u.nutzer.Gezeigt(ctx)
	if lesefehler != nil {
		return domain.Profil{}, "", lesefehler
	}

	// **Aus dem eigenen Profil übernimmt niemand.** Ein Deck, das schon hier
	// ist, kommt nicht von woanders — und eine Herkunft darüber behauptete
	// genau das.
	if gezeigt.ID == quellProfil {
		return domain.Profil{}, "", fmt.Errorf(
			"%w: aus dem eigenen Profil lässt sich nichts übernehmen",
			domain.ErrValidation)
	}

	quelle, bezeichnung, suchfehler := u.quelleSuchen(ctx, quellProfil)
	if suchfehler != nil {
		return domain.Profil{}, "", suchfehler
	}

	// **Dieselbe Frage wie beim Wechsel und nicht eine strengere.** Wer aus
	// einem Profil nehmen darf, ist genau, wer hineingehen darf: eine
	// Übernahme, die mehr verlangte, wäre mit einem Wechsel zu umgehen —
	// hineingehen, ansehen, abtippen. Der Nachweis ist leer, weil dieser Weg
	// niemanden nach einem Code fragt.
	if erlaubnis := u.nutzer.Zutritt(ctx, quelle, ""); erlaubnis != nil {
		return domain.Profil{}, "", erlaubnis
	}

	return quelle, bezeichnung, nil
}

// quelleSuchen findet das Quellprofil **und** seine Bezeichnung in **einem**
// Gang über die Übersicht.
//
// # Warum in einem und nicht in zweien
//
// **Weil zwei Lesevorgänge zwei Stände sind** (§7, Regel 2). Der erste Anlauf
// suchte das Profil über die Freigabenverwaltung und den Namen danach noch
// einmal; dazwischen kann jemand löschen, und dann steht ein Rückfall im
// Code, den **kein Lauf erreicht**, solange der Bestand sich nicht mitten im
// Ruf ändert. Ein Zweig, den man nur über einen Wettlauf betritt, ist ein
// Zweig, den niemand prüft.
//
// **Ein Gang gibt beides oder keines** — und braucht keinen Rückfall.
//
// # Warum der Nutzer davorsteht
//
// **Weil „Mein Profil" der Vorgabename jedes ersten Profils ist.** Von einem
// Fremden gelesen sagt er nichts; „Zweiter · Mein Profil" sagt, von wem es
// kam. Dieselbe Gestalt wie bei [domain.Exemplar.Bezeichnung] — und sie wird
// **in Go** zusammengesetzt und nicht an drei Flächen (§16).
//
// **Abgeschrieben und nicht nachgeschlagen**: sie hält den Stand des
// Übernahmezeitpunkts fest. Wer sich später umbenennt, ändert damit nicht,
// woher etwas kam ([0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md)).
func (u *Deckuebernahme) quelleSuchen(ctx context.Context, quellProfil int64) (
	quelle domain.Profil, bezeichnung string, fehler error,
) {
	zeilen, lesefehler := u.nutzer.Uebersicht(ctx)
	if lesefehler != nil {
		return domain.Profil{}, "", lesefehler
	}

	for _, zeile := range zeilen {
		for _, profil := range zeile.Profile {
			if profil.ID == quellProfil {
				return profil, zeile.Nutzer.Name + " · " + profil.Name, nil
			}
		}
	}

	return domain.Profil{}, "", fmt.Errorf(
		"%w: das Profil %d führt der Bestand nicht", domain.ErrNotFound, quellProfil)
}

// platzbefund sagt, ob ein Platz dort bleibt — und warum.
//
// **Ein Platz auf eine eigene Karte des anderen Profils bleibt dort**: die
// Karte ist eine Zeile im Quellprofil, und dieselbe Zeile gibt es hier nicht.
func platzbefund(platz domain.Deckplatz) string {
	if platz.KarteID == 0 {
		return ""
	}

	return fmt.Sprintf(
		"Der Platz %q zeigt auf eine eigene Karte des anderen Profils und "+
			"bleibt dort.", platz.Bezeichnung)
}

// formatbefund sagt, dass das Format zurückbleibt — wenn es eines gab.
//
// **Ein Deck ohne Format ist ungeprüft, nicht falsch** (Vermerk 0061). Die
// Absage ist deshalb ein Befund und keine Ablehnung.
func (u *Deckuebernahme) formatbefund(fremdes domain.Deck) []string {
	if fremdes.FormatID == 0 {
		return nil
	}

	return []string{
		"Das Format des Decks ist eine Zeile im anderen Profil und bleibt dort; " +
			"das übernommene Deck ist zunächst ungeprüft.",
	}
}
