// Package service hält die Anwendungsfälle, Transaktionsgrenzen und die
// fachlichen Regeln, die mehr als eine Entität betreffen (AGENTS.md §2).
//
// Die Repository-**Schnittstellen** stehen hier, **implementiert** werden sie in
// `internal/storage`. Das ist eine fachliche Zuordnung, keine Importrichtung:
// wegen Gos struktureller Typisierung erfüllt `storage` sie ohne einen Import
// von hier — und `internal/schichten` kann das deshalb nicht prüfen. Es steht
// dort als das, was der Wächter nicht sieht.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// SetSpeicher legt ein eingelesenes Set ab.
//
// Die Schnittstelle sagt, was `service` **braucht** — nicht, was `storage`
// anzubieten hat. Der Unterschied ist der Grund für die Richtung.
type SetSpeicher interface {
	SetAnlegen(ctx context.Context, set *domain.Set, roh []domain.RawBlock) (int64, error)
}

// Ergebnis ist der Ausgang eines Imports.
//
// **Befunde und Erfolg schließen einander nicht aus.** Ein Set mit 300 lesbaren
// und zwei unlesbaren Karten wird mit 300 Karten und zwei benannten Befunden
// geöffnet, nicht abgewiesen (AGENTS.md §8).
//
// **Die Befunde sind `domain.Befund` und nicht die von `msepaket`.** Seit es
// zwei Leser gibt — Pakete und MTG JSON —, ist die Stelle nicht mehr immer eine
// Zeilennummer. Der gemeinsame Typ steht in der Domäne und beschreibt die
// Stelle als Text; jeder Leser sagt sie in der Ausdrucksweise seines Formats.
type Ergebnis struct {
	SetID   int64
	Karten  int
	Befunde []domain.Befund

	// Angelegt sind die Kennungen aller Sets, die dabei entstanden sind.
	//
	// **Meist genau eine — bei einer AllSets-Datei mehrere.** Sie steht hier,
	// weil ein Import auf halber Strecke scheitern kann: jedes Set wird in
	// einer eigenen Transaktion abgelegt, und was vorher durchkam, bleibt.
	// Ohne diese Liste bekäme der Aufrufer einen Fehler und einen Bestand mit
	// Sets, von denen niemand weiß. Der Zweitblick über Phase 9 hat es
	// gefunden.
	Angelegt []int64

	// SchonOffen sagt, dass diese Quelle bereits im Bestand lag und nichts
	// angelegt wurde (#886).
	//
	// **Ein eigenes Feld und keine Ableitung aus `len(Angelegt) == 0`.** Die
	// beiden fallen heute zusammen und müssen es nicht: ein Import, der an
	// seinem ersten Set scheitert, legt auch nichts an — und wäre dann
	// fälschlich „schon offen".
	SchonOffen bool
}

// Importeur liest Pakete in den Bestand.
type Importeur struct {
	haushaltswahl

	speicher SetSpeicher

	// koepfe sagt, was schon im Bestand liegt — für die Regel „eine Quelle,
	// ein Set" ([schonOffen]).
	koepfe Kopfliste
}

// NeuerImporteur baut den Importeur über einem Speicher.
func NeuerImporteur(speicher SetSpeicher) *Importeur {
	return &Importeur{speicher: speicher}
}

// Set liest ein Setpaket in den Bestand.
func (i *Importeur) Set(ctx context.Context, pfad string) (*Ergebnis, error) {
	freigeben, platzfehler := i.platz(ctx, domain.RessourceEinfuhr)
	if platzfehler != nil {
		return nil, platzfehler
	}

	defer freigeben()

	// **Eine Quelle, ein Set** — die Regel steht in `schonoffen.go`.
	if schon, err := i.schonOffen(ctx, pfad); err != nil || schon != nil {
		return schon, err
	}

	gelesen, err := setLesen(pfad)
	if err != nil {
		return nil, err
	}

	set, roh := gelesen.set, gelesen.roh

	// **Die letzte Frage vor dem Schreiben** (FMT-03): bis hierher ist
	// nichts im Bestand, und das Set entsteht danach in einer Transaktion.
	einSet := domain.Fortschritt{Gesamt: 1, Einheit: domain.EinheitSet}
	if abbruch := Schritt(ctx, einSet); abbruch != nil {
		return nil, abbruch
	}

	kennung, fehler := i.speicher.SetAnlegen(ctx, set, roh)
	if fehler != nil {
		return nil, fmt.Errorf("Set ablegen: %w", fehler)
	}

	einSet.Erledigt = 1
	Melde(ctx, einSet)

	return &Ergebnis{
		SetID:    kennung,
		Karten:   len(set.Cards),
		Befunde:  gelesen.befunde,
		Angelegt: []int64{kennung},
	}, nil
}

// ausBaum macht aus dem Datenbaum ein Set, seine Rohblöcke und seine Befunde.
func ausBaum(paket *msepaket.Paket) (*domain.Set, []domain.RawBlock, []msepaket.Befund) {
	set := &domain.Set{
		Name:   namenAus(paket),
		Game:   paket.Spiel,
		Style:  wertVon(paket.Kopf.Ersten("stylesheet")),
		Origin: paket.Pfad,
	}

	var (
		roh     []domain.RawBlock
		befunde []msepaket.Befund
	)

	// **Was einmal gedeutet wurde, wird beim zweiten Mal roh.**
	//
	// Ein Schlüssel aus `gespeichert` schloss bisher seinen ganzen Bereich vom
	// Rohweg aus — auch das Vorkommen, das der Leser gar nicht verwertet: ein
	// zweites `game:` verschwand, ein zweites `set info:` überschrieb das erste
	// (#88).
	gedeutet := map[string]bool{}

	for _, knoten := range paket.Kopf.Wurzel {
		if knoten.Zeilenart != msepaket.ArtEintrag {
			// **Kommentare und Leerzeilen der Wurzelebene reisen mit.** Sie
			// stehen in der Quelldatei und kommen beim Schreiben heraus (0011);
			// im Bestand fehlten sie, und damit sahen Suche, Statistik und der
			// MCP-Server sie nicht.
			roh = append(roh, rohblockAus(knoten))

			continue
		}

		switch {
		// **`set_info` und `set info` sind derselbe Block.** In einer von MSE2
		// gespeicherten Datei steht der kanonische Name (#396).
		case msepaket.Entkanonisch(knoten.Schluessel) == "set info" && !gedeutet["set info"]:
			gedeutet["set info"] = true
			set.Values = werteAus(knoten)

		// **Die setweiten Stiloptionen werden ein Setwert.**
		//
		// Derselbe Weg wie bei der Karte: `has styling` und `styling data`
		// liegen dort als gewöhnliche Feldwerte, und der Planbauer deutet sie
		// (`kartenstiloptionen.go`). Eine eigene Tabelle wäre eine zweite
		// Antwort auf dieselbe Frage — und eine Migration für Daten, die
		// zeilenweise ohnehin flach sind
		// (#557).
		// **Die fünf Formatangaben des Setkopfs** (#1325). Sie stehen in
		// jeder MSE-Datei und standen bis dahin unter „Nicht gedeutet".
		case kopfangabe(msepaket.Entkanonisch(knoten.Schluessel)) != "" &&
			!gedeutet[knoten.Schluessel]:
			gedeutet[knoten.Schluessel] = true
			kopfSetzen(&set.Kopf, msepaket.Entkanonisch(knoten.Schluessel), knoten)

		case knoten.Schluessel == "styling" && !gedeutet["styling"]:
			gedeutet["styling"] = true
			set.Values = append(set.Values, setstilwertAus(knoten))

		// **Die Schlüsselwörter des Sets** (#1062). Sie fielen bis dahin in
		// den Rohweg: byte-genau erhalten und im Bestand unsichtbar.
		case knoten.Schluessel == "keyword":
			set.Keywords = append(set.Keywords, schluesselwortAus(knoten))

		// **Die Packarten des Sets** (#1081). Sie fielen bis dahin ebenso in
		// den Rohweg, und `Packbereich` zog aus denen des Spiels — obwohl MSE
		// zuerst im Set sucht.
		case msepaket.Entkanonisch(knoten.Schluessel) == "pack type":
			set.Packs = append(set.Packs, packartAus(knoten))

		case knoten.Schluessel == "card":
			// **Der skalare Rest eines `card: irgendwas` reist mit.** `werteAus`
			// sieht nur Kinder; was hinter dem Doppelpunkt steht, sah niemand.
			if knoten.Wert != "" {
				roh = append(roh, rohblockAus(knoten))
			}

			karte := &domain.Card{Values: werteAus(knoten)}

			// **Eine leere Karte bleibt eine Karte.** Sie zuerst zu melden und
			// dann wegzuwerfen war der Fehler: der Befund sagte „hier stimmt
			// etwas nicht", und der Datensatz war trotzdem fort — beim Export
			// nicht wiederherstellbar. Der Befund bleibt, die Karte auch.
			if len(karte.Values) == 0 {
				befunde = append(befunde, msepaket.Befund{
					Stelle:  knoten.Stelle,
					Meldung: "eine Karte ohne einen einzigen Feldwert",
				})
			}

			set.Cards = append(set.Cards, karte)

		case !gespeichert[knoten.Schluessel] || gedeutet[knoten.Schluessel]:
			roh = append(roh, rohblockAus(knoten))

		default:
			// Ein gespeicherter Schlüssel beim **ersten** Mal: `game` und
			// `stylesheet` liest der Setkopf oben über `Ersten`.
			gedeutet[knoten.Schluessel] = true
		}

		// **Kommentare und Leerzeilen hängen am Knoten davor.**
		//
		// Der Zerteiler hängt sie in die gerade offene Ebene — ein Kommentar
		// hinter `game: probe` wird damit zum **Kind** von `game`. Wird `game`
		// gedeutet, reist er nicht mit, und der Bestand ist blind für ihn (#88).
		roh = append(roh, beiwerkAus(knoten)...)
	}

	return set, roh, befunde
}

// beiwerkAus sammelt die Kommentare und Leerzeilen unter einem Knoten.
//
// **Nur die unmittelbaren Kinder.** Was tiefer liegt, gehört zu einem Feldwert;
// was hier steht, gehört zwischen zwei Einträge der Wurzelebene.
func beiwerkAus(knoten *msepaket.Knoten) []domain.RawBlock {
	var beiwerk []domain.RawBlock

	for _, kind := range knoten.Kinder {
		if kind.Zeilenart == msepaket.ArtEintrag || strings.TrimSpace(kind.Roh) == "" {
			continue
		}

		beiwerk = append(beiwerk, rohblockAus(kind))
	}

	return beiwerk
}

// rohblockAus macht aus einem Knoten einen Rohblock.
func rohblockAus(knoten *msepaket.Knoten) domain.RawBlock {
	return domain.RawBlock{
		File: knoten.Stelle.Datei,
		Line: knoten.Stelle.Zeile,
		Text: rohtext(knoten),
	}
}

// werteAus macht aus den Kindern eines Knotens Feldwerte, in ihrer Reihenfolge.
func werteAus(knoten *msepaket.Knoten) []domain.Value {
	werte := make([]domain.Value, 0, len(knoten.Kinder))

	for _, kind := range knoten.Kinder {
		// **Eine verirrte Stilzeile aus einer früheren Abschrift** (F05,
		// #2247) — sie gäbe einen zweiten setweiten Stilblock neben dem echten.
		if kind.Zeilenart != msepaket.ArtEintrag || verirrterStilblock(kind) {
			continue
		}

		werte = append(werte, domain.Value{
			Field: msepaket.Entkanonisch(kind.Schluessel),
			Text:  rechtschreibmarke.ReplaceAllString(wertText(kind), ""),
		})
	}

	if len(werte) == 0 {
		return nil
	}

	return werte
}

// namenAus bestimmt den Namen des Sets.
//
// `set info → title` zuerst, sonst der Dateiname. Ein Set ohne Titel ist keine
// Ausnahme — MSE legt eines an, sobald jemand auf „Neu" klickt.
func namenAus(paket *msepaket.Paket) string {
	// **Auch als Block** (C02, #2227): ein Titel, der eine Klammer offen
	// lässt, steht geschrieben als Block — wie ein langer Kartentext.
	if titel := msepaket.WertOderBlock(paket.Kopf.Ersten("set info").Ersten("title")); titel != "" {
		return titel
	}

	name := paket.Pfad
	if stelle := strings.LastIndexByte(name, '/'); stelle >= 0 {
		name = name[stelle+1:]
	}

	return name
}

func wertVon(knoten *msepaket.Knoten) string {
	if knoten == nil {
		return ""
	}

	return knoten.Wert
}
