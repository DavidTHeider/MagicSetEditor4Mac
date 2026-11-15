// **Die Formatprüfung** — Werkstück 5 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Drei Antworten, und sie sind verschieden
//
//	ein echter Druck            → Scryfall, `legalities` je Druck
//	eine eigene Karte ohne      → `ungeprüft`
//	eine eigene Karte mit       → Scryfall, über den Druck; Herkunft dabei
//
// **`ungeprüft` ist kein `nicht zulässig`.** Es gibt keine eingebaute Liste
// verbotener Karten und keine geratene Regel. Was nicht aus Scryfall kommt,
// bleibt ungeprüft — auch dann, wenn eine Karte „offensichtlich" verboten
// wäre.

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Druckauskunft ist, was die Formatprüfung über Drucke fragt.
//
// **Sie setzt [Fremddrucke] fort statt sie zu ersetzen.** Der Sammelweg ist
// der Normalfall; der Einzelweg bleibt, weil er den Zwischenspeicher **je
// Druck** liest und damit ohne Netz noch antwortet (§16, Regel 5).
//
// **Und sie steht hier und nicht bei [Fremddrucke]**, weil der Kartenverweis
// den Sammelweg nicht braucht: er fragt nach genau einer Karte. Eine
// Schnittstelle dort zu verbreitern, wo sie niemand benutzt, verlangte von
// jedem Doppel ein Verfahren ohne Aufrufer (§2).
type Druckauskunft interface {
	Fremddrucke

	Sammlung(ctx context.Context, kennungen []scryfall.Kennung) (
		antwort scryfall.Sammelantwort, alter time.Duration,
		befunde []domain.Befund, fehler error)
}

// Deckherkunft ist, was die Prüfung an eigenen Karten nachschlägt.
//
// **Eine eigene Schnittstelle mit einem Verfahren** (§2): die Prüfung liest
// Verweise und setzt keine.
type Deckherkunft interface {
	Fuer(ctx context.Context, karteID int64) (*domain.Kartenverweis, error)
}

// Formatpruefung ist der Anwendungsfall „passt dieses Deck in dieses Format?".
type Formatpruefung struct {
	decks    Deckspeicher
	formate  Formatnachschlag
	drucke   Druckauskunft
	verweise Deckherkunft

	// karten ist die Naht zu den Feldern einer **eigenen** Karte
	// (#1622).
	//
	// **`nil` ist erlaubt**, und dann bleibt es beim Alten — derselbe Grund
	// wie bei [Deckbau.MitKartennamen]: die Kommandozeile und der
	// MCP-Server verdrahten nicht jede Naht.
	karten Kartennamensnaht
}

// MitKartenfeldern setzt die Naht zu den Feldern eigener Karten.
//
// # Wofür
//
// **Damit ein Deck aus eigenen Karten ohne Netz richtig geprüft wird** —
// #1622.
// Die Typzeile entscheidet, ob die Höchstzahl gleicher Karten diesen Platz
// ausnimmt (CR 100.2a); bis zum 2026-09-17 kam sie **nur** von Scryfall, und
// eine eigene Karte ohne Druckverweis hat dort keine. Zwanzig selbstgebaute
// Gebirge waren damit zwanzig gleiche Karten.
//
// **Was sie nicht tut**: die **Zulässigkeit** eines Drucks in einem Format
// beantworten. Die bleibt eine Auskunft von Scryfall, und ohne sie steht
// weiterhin `ungeprüft`.
func (f *Formatpruefung) MitKartenfeldern(naht Kartennamensnaht) *Formatpruefung {
	f.karten = naht

	return f
}

// NeueFormatpruefung baut sie.
func NeueFormatpruefung(
	decks Deckspeicher, formate Formatnachschlag,
	drucke Druckauskunft, verweise Deckherkunft,
) *Formatpruefung {
	return &Formatpruefung{
		decks: decks, formate: formate, drucke: drucke, verweise: verweise,
	}
}

// Pruefe beurteilt ein Deck gegen sein gewähltes Format.
//
// **Die Befunde sind kein Fehler.** Sie sagen, was unterwegs schiefging —
// eine Gegenstelle, die schwieg, ein Wort, das niemand kennt. Die Prüfung
// läuft weiter und meldet für die betroffene Karte `ungeprüft`; dieselbe
// Haltung wie bei `msepaket` (§8) und §16, Regel 5.
func (f *Formatpruefung) Pruefe(ctx context.Context, deckID int64) (
	domain.Deckurteil, []domain.Befund, error,
) {
	deck, err := f.decks.Deck(ctx, deckID)
	if err != nil {
		return domain.Deckurteil{}, nil, err
	}

	if deck.FormatID == 0 {
		return domain.Deckurteil{}, nil, fmt.Errorf(
			"%w: dieses Deck hat kein Format gewählt — gegen was soll geprüft werden?",
			domain.ErrValidation)
	}

	format, lesefehler := f.formate.Format(ctx, deck.FormatID)
	if lesefehler != nil {
		return domain.Deckurteil{}, nil, lesefehler
	}

	urteile, posten, befunde := f.urteile(ctx, deck, format)

	return domain.Deckurteil{
		Formatname: format.Name,
		Hauptdeck:  deck.Zaehle(domain.BereichHauptdeck),
		Sideboard:  deck.Zaehle(domain.BereichSideboard),
		Urteile:    urteile,
		Groessen:   domain.PruefeGroessen(posten, format),
	}, befunde, nil
}

// urteile beurteilt jeden Platz.
//
// **Drei Schritte statt einem**: erst steht fest, welcher Druck für welchen
// Platz zu fragen ist, dann werden **alle** in wenigen Anfragen geholt, dann
// wird geurteilt. Der zweite Schritt ist der Grund — er war einmal je Platz
// eine Anfrage (Befund 6 des Zweitblicks über Phase 4.6).
//
// **Sie gibt die Posten mit heraus**, weil dieselbe Antwort von Scryfall
// beides trägt: das Wort zur Zulässigkeit und die Typzeile, an der hängt, ob
// die Höchstzahl gleicher Karten diesen Platz ausnimmt
// (#1370).
// Ein zweiter Weg dafür wäre eine zweite Anfrage an eine Gegenstelle, die eine
// Grenze nennt (§16, Regel 4).
func (f *Formatpruefung) urteile(
	ctx context.Context, deck domain.Deck, format domain.Format,
) ([]domain.Kartenurteil, []domain.Deckposten, []domain.Befund) {
	var (
		urteile = make([]domain.Kartenurteil, 0, len(deck.Plaetze))
		posten  = make([]domain.Deckposten, 0, len(deck.Plaetze))
	)

	bezuege, befunde := f.bezuege(ctx, deck, format)

	antworten, weitere := f.hole(ctx, bezuege)
	befunde = append(befunde, weitere...)

	for stelle, platz := range deck.Plaetze {
		urteil, zuege := f.beurteile(platz, bezuege[stelle], antworten, format)

		urteile = append(urteile, urteil)
		posten = append(posten, domain.Deckposten{
			Platz: platz,
			Standardland: domain.IstStandardland(
				f.typzeileZu(ctx, platz, zuege.Typzeile)),

			// **Und die Orakelkennung** (#1375): zwei englische Namen können
			// dieselbe Karte sein, und dann zählen sie zusammen. Ohne diese
			// Zeile gingen vier von einem und vier vom anderen Namen durch.
			Orakelkennung: zuege.Orakelkennung,

			// **Der Name entscheidet die Höchstzahl** — CR 100.2a begrenzt
			// Karten mit einem Namen, nicht Drucke (Befund 1 des Zweitblicks
			// über Phase 4.6). Er steht schon im Urteil; ohne diese Zeile
			// kam er nur nicht dorthin, wo gezählt wird.
			Name: urteil.Bezeichnung,
		})
	}

	return urteile, posten, befunde
}

// typzeileZu gibt die Typzeile, an der die Höchstzahl hängt — CR 100.2a.
//
// # Die Reihenfolge ist keine Vorliebe
//
// **Was Scryfall sagt, gilt.** Ein echter Druck hat eine gedruckte Typzeile,
// und der Bestand kennt für ihn ohnehin keine. Der Bestand tritt **nur** ein,
// wo dort nichts steht — bei einer eigenen Karte ohne Druckverweis, und bei
// einem Format, das gar nicht erst fragt.
//
// # Und eine leere bleibt leer
//
// **Nicht jedes Spiel führt einen Supertyp.** Ein leeres Feld ist keine
// Auskunft; wer daraus „Standardland" machte, ließe jede Karte ohne Typzeile
// in beliebiger Zahl durch (§10: eine Ausnahme auf Verdacht ist keine).
func (f *Formatpruefung) typzeileZu(
	ctx context.Context, platz domain.Deckplatz, vonScryfall string,
) string {
	if vonScryfall != "" || f.karten == nil || platz.KarteID == 0 {
		return vonScryfall
	}

	// **Zwei Felder und nicht das gerechnete `type`.** In einem MSE-Paket ist
	// `type` ein Feld mit `save value: false` (`docs/FORMATE.md`) — es steht
	// im Bestand oft gar nicht. `super type` und `sub type` stehen dort, und
	// [domain.TypzeileAus] deutet seit `M2` genau diese zwei
	// (#1604).
	//
	//nolint:errcheck // Der leere Wert **ist** die Antwort — siehe oben.
	ober, _ := f.karten.FeldZurKarte(ctx, platz.KarteID, "super type")
	//nolint:errcheck // dito
	unter, _ := f.karten.FeldZurKarte(ctx, platz.KarteID, "sub type")

	// **Zwei leere Felder geben keine Zeile**, und das entscheidet
	// [domain.TypzeileText].
	return domain.TypzeileText(ober, unter)
}

// druckbezug ist, welcher Druck für einen Platz zu fragen ist — und woher er
// bekannt ist.
type druckbezug struct {
	druck    string
	herkunft string
}

// bezuege sagt für jeden Platz, welcher Druck zu fragen ist.
//
// **Ohne Scryfallnamen wird gar nicht erst gefragt** — und dann auch kein
// Verweis gelesen. Eine Anfrage, deren Antwort niemand lesen kann, ist
// verbrauchtes Kontingent (§16, Regel 4).
func (f *Formatpruefung) bezuege(
	ctx context.Context, deck domain.Deck, format domain.Format,
) ([]druckbezug, []domain.Befund) {
	bezuege := make([]druckbezug, len(deck.Plaetze))

	if format.Scryfallname == "" {
		return bezuege, nil
	}

	var befunde []domain.Befund

	for stelle, platz := range deck.Plaetze {
		druck, herkunft, gefunden := f.druckZu(ctx, platz)

		bezuege[stelle] = druckbezug{druck: druck, herkunft: herkunft}
		befunde = append(befunde, gefunden...)
	}

	return bezuege, befunde
}

// beurteile beantwortet **einen** Platz aus dem, was der Schub brachte.
func (f *Formatpruefung) beurteile(
	platz domain.Deckplatz, bezug druckbezug,
	antworten map[string]druckantwort, format domain.Format,
) (urteil domain.Kartenurteil, zuege kartenzuege) {
	urteil = domain.Kartenurteil{
		PlatzID: platz.ID, Menge: platz.Menge, Bereich: platz.Bereich,
		Zulaessigkeit: domain.ZulaessigUngeprueft,
	}

	// **Der Name bleibt damit unbekannt**, und die Höchstzahl fällt auf den
	// Druck zurück. Das ist die ehrliche Vorgabe: wer nicht fragt, weiß
	// nichts.
	if format.Scryfallname == "" {
		urteil.Grund = fmt.Sprintf(
			"Das Format %q hat keinen Scryfallnamen; dort steht nichts über seine "+
				"Zulässigkeiten.", format.Name)

		return urteil, kartenzuege{}
	}

	if bezug.druck == "" {
		urteil.Grund = "Diese Karte ist ein eigener Entwurf und verweist auf keinen " +
			"Druck; Scryfall kennt sie naturgemäß nicht."

		return urteil, kartenzuege{}
	}

	antwort := antworten[druckschluessel(bezug.druck)]
	if antwort.karte == nil {
		urteil.Grund = antwort.grund

		return urteil, kartenzuege{}
	}

	return wortZu(urteil, antwort.karte, bezug.herkunft, format)
}

// druckZu sagt, welcher Druck für einen Platz zu fragen ist.
//
// **Zwei Wege zu demselben Ziel**: ein Platz nennt einen Druck selbst, oder er
// nennt eine eigene Karte, die auf einen verweist.
func (f *Formatpruefung) druckZu(ctx context.Context, platz domain.Deckplatz) (
	druck, herkunft string, befunde []domain.Befund,
) {
	if platz.Printing != "" {
		return platz.Printing, "Scryfall", nil
	}

	verweis, err := f.verweise.Fuer(ctx, platz.KarteID)
	if err != nil {
		return "", "", []domain.Befund{{
			Ort: fmt.Sprintf("karte:%d", platz.KarteID),
			Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Der Verweis der Karte %d ließ sich nicht lesen: %v. "+
					"Sie bleibt ungeprüft.", platz.KarteID, err),
		}}
	}

	if verweis == nil || verweis.Printing == "" {
		return "", "", nil
	}

	return verweis.Printing, "Scryfall, über den Verweis dieser Karte", nil
}

// wortZu liest die Zulässigkeit aus einer geholten Karte.
func wortZu(
	urteil domain.Kartenurteil, karte *scryfall.Karte,
	herkunft string, format domain.Format,
) (domain.Kartenurteil, kartenzuege) {
	urteil.Bezeichnung = karte.Name

	wort, gelistet := karte.Legalitaeten[format.Scryfallname]
	if !gelistet || wort == "" {
		urteil.Grund = fmt.Sprintf(
			"Scryfall führt für diesen Druck kein Wort zum Format %q.",
			format.Scryfallname)

		// **Die Typzeile geht trotzdem mit.** Ob dieser Druck ein
		// Standardland ist, hängt nicht daran, ob Scryfall ihn in diesem
		// Format führt — und ein Deck mit vierundzwanzig ungelisteten Wäldern
		// bekäme sonst einen zweiten, falschen Befund obendrein.
		return urteil, zuegeVon(karte)
	}

	urteil.Zulaessigkeit = domain.Zulaessigkeit(wort)
	urteil.Quelle = herkunft

	return urteil, zuegeVon(karte)
}

// kartenzuege ist, was die Antwort von Scryfall **neben** dem Urteil trägt.
//
// # Warum ein Verbund und keine zwei Rückgabewerte
//
// Es waren zwei — Urteil und Typzeile —, und mit der Orakelkennung wären es
// drei geworden
// (#1375).
// Drei namenlose Rückgabewerte derselben Herkunft sind eine Reihenfolge, die
// man sich merken muss; ein Verbund ist eine, die dasteht.
type kartenzuege struct {
	// Typzeile entscheidet, ob die Höchstzahl diesen Platz ausnimmt
	// (#1370).
	Typzeile string

	// Orakelkennung trägt die austauschbaren Namen mit — zwei englische
	// Namen, eine Karte (#1375, CR 100.2a und CR 201.3b).
	Orakelkennung string
}

// zuegeVon liest sie aus einer geholten Karte.
func zuegeVon(karte *scryfall.Karte) kartenzuege {
	return kartenzuege{Typzeile: karte.Typzeile, Orakelkennung: karte.Orakelkennung}
}
