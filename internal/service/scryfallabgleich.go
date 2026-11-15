package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremdkarten ist, was der Abgleich von einem fremden Dienst braucht.
//
// **Die Schnittstelle steht hier, wo sie gebraucht wird** (§2). `service`
// bindet damit nicht den Klienten, sondern die Fähigkeit — und ein Test
// braucht kein Netz.
type Fremdkarten interface {
	// KarteNachName gibt die Karte, **wie alt die Antwort ist** und die
	// Befunde.
	//
	// **Das Alter gehört in die Schnittstelle** (§16, Regel 5): „Der letzte
	// gute Stand bleibt, und er sagt, wie alt er ist." Ein Vorschlag aus einer
	// drei Tage alten Antwort sieht sonst aus wie einer von eben — und wer ihn
	// übernimmt, weiß nicht, dass er einem Zwischenspeicher folgt.
	KarteNachName(ctx context.Context, name string) (
		karte *scryfall.Karte, alter time.Duration, befunde []domain.Befund, fehler error)
}

// Feldvorschlag ist eine einzelne Änderung, die möglich wäre.
//
// **Beide Seiten**, denn ein Vorschlag, der nur den neuen Wert zeigte, ließe
// niemanden entscheiden.
type Feldvorschlag struct {
	Feld string
	Alt  string
	Neu  string
}

// Vorschlag ist, was eine Übernahme ändern würde.
type Vorschlag struct {
	// Aenderungen sind die Felder, die sich unterscheiden — **nur die**.
	// Ein Feld, das gleich bleibt, ist keine Änderung.
	Aenderungen []Feldvorschlag

	// Befunde sind die Auffälligkeiten der Antwort — etwa ein unbekannter
	// Aufzählungswert. Sie reisen mit, damit der sie sieht, der übernimmt,
	// und nicht nur der, der die Antwort gelesen hat.
	Befunde []domain.Befund

	// Herkunft sagt, woher der Vorschlag stammt — für die Angabe an der Karte.
	Herkunft string

	// Druck ist die Scryfall-Kennung des Drucks, aus dem der Vorschlag stammt.
	//
	// **Damit sichtbar ist, welcher es war.** Ohne Verweis sucht
	// `/cards/named` selbst einen aus; mit Verweis wird der gewählte geholt
	// und der übergebene Name **übergangen**. Beides still zu tun wäre genau
	// das, was §16 Regel 6 verbietet — hier steht, was geschehen ist.
	Druck string

	// Alter sagt, wie alt die Antwort ist, aus der dieser Vorschlag stammt.
	//
	// **Null heißt: gerade geholt.** Alles darüber kommt aus dem
	// Zwischenspeicher, und der Aufrufer soll es sagen können — §16, Regel 5.
	Alter time.Duration

	// Stand ist der **Fingerabdruck der Kartenwerte**, gegen die verglichen
	// wurde (#1241).
	//
	// # Dieselbe Klasse wie beim Ersetzen
	//
	// > Ein Bestätigungsschritt muss an **Absicht und Quellzustand** gebunden
	// > sein.
	//
	// `ScryfallUebernehmen` liest die Karte neu und rechnet den Vorschlag noch
	// einmal. Zwischen Zeigen und Übernehmen kann sie sich geändert haben —
	// dann übernimmt der Knopf **etwas anderes als das Gezeigte**.
	//
	// # Was er deckt
	//
	// Feld und **alten** Wert jeder vorgeschlagenen Änderung. Ein Feld, das
	// gleich bleibt, steht nicht im Vorschlag und bindet nichts.
	Stand string
}

// BerechneterStand gibt den Fingerabdruck der Werte, gegen die verglichen wurde.
//
// **Über die Änderungen und nicht über die ganze Karte.** Wer an einem Feld
// arbeitet, das der Vorschlag nicht anfasst, soll ihn nicht ungültig machen —
// er zeigt, was sich ändert, und das bleibt richtig.
func (v Vorschlag) BerechneterStand() string {
	if len(v.Aenderungen) == 0 {
		return ""
	}

	summe := sha256.New()

	for i := range v.Aenderungen {
		//nolint:errcheck // ein [hash.Hash] gibt nie einen Fehler
		_, _ = fmt.Fprintf(summe, "%s\x00%s\x00",
			v.Aenderungen[i].Feld, v.Aenderungen[i].Alt)
	}

	return hex.EncodeToString(summe.Sum(nil))
}

// Scryfallabgleich vergleicht eine Karte mit dem, was Scryfall hat.
//
// # Warum der Vorschlag und die Übernahme zwei Schritte sind
//
// §16, Regel 6: „Nichts wird still übernommen. Was von außen in ein Kartenfeld
// will, zeigt vorher, was sich ändert, und trägt danach seine Herkunft."
//
// **Und warum das hier steht und nicht im Fenster**: die Regel soll an allen
// vier Flächen gelten. Stünde die Rückfrage nur in der Oberfläche, übernähme
// MCP still — ein Sprachmodell überschriebe eine Karte, während niemand
// hinsieht.
type Scryfallabgleich struct {
	fremd Fremdkarten

	// drucke holt einen **bestimmten** Druck — den, auf den ein Verweis zeigt.
	drucke Fremddrucke

	// verweise sagt, ob eine Karte auf einen Druck bezogen ist.
	verweise Verweisleser

	schreiber Uebernahmeschreiber
}

// Verweisleser ist, was der Abgleich über Verweise wissen muss.
//
// **Nur lesen** (§2: die Schnittstelle sagt, was gebraucht wird). Der Abgleich
// setzt keinen Verweis; er folgt einem.
type Verweisleser interface {
	Fuer(ctx context.Context, karteID int64) (*domain.Kartenverweis, error)
}

// Uebernahmeschreiber ist, was der Abgleich zum Schreiben braucht.
//
// # Warum genau diese Gestalt
//
// Es ist die Naht, die **Fenster, MCP und Kommandozeile schon benutzen**
// (`app.Kartenschreibnaht`). Ein übernommener Regeltext landet damit in
// derselben Zeile wie ein getippter — und alles, was aus dem Bestand liest,
// sieht ihn ohne eine weitere Zeile Code: Vorschau, Druckbogen, Ausfuhr in
// `.mse-set`, MCP.
//
// **Eine Methode und nicht drei** (§2: die Schnittstelle sagt, was gebraucht
// wird). Der Abgleich entfernt keine Felder und setzt keine an eine Stelle.
//
// **Und der vorige Wert kommt zurück**, damit eine Übernahme genauso
// zurücknehmbar ist wie jede andere Änderung — das gehört zu „nichts wird
// still übernommen" dazu.
// **Und die Herkunft geht mit.** §16, Regel 6 endet mit „trägt danach seine
// Herkunft" — der Vorschlag trug sie schon, die geschriebene Zeile bis zum
// 2026-08-23 nicht. Wer den Verlauf liest, soll eine Übernahme von einer
// Tipparbeit unterscheiden können, ohne den Wert zu erraten.
//
// **Und sie schreibt mehrere Felder in einem Zug** — seit dem 2026-08-24.
// Hier stand `FeldSetzenVonAussen`, und `Uebernehmen` rief es in einer
// Schleife: scheiterte das dritte von fünf, standen zwei neu da und drei alt.
// Der Fehler kam durch, die Liste der schon geschriebenen Felder nicht
// (#1026).
type Uebernahmeschreiber interface {
	FelderVonAussen(
		ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
	) (vorher string, hatte bool, fehler error)
}

// NeuerScryfallabgleich baut den Abgleich.
//
// **Druckquelle und Verweisleser stehen im Konstruktor und nicht in einem
// `Mit…`** — Werkstück 2 von Phase 3.6. Ein Nachrüstaufruf, den ein
// Composition Root vergisst, ließe den Abgleich still auf den Namen
// zurückfallen: die Karte zeigte auf `lea`, und übernommen würde der Druck,
// den Scryfall gerade vorn hat. Wer es hier vergisst, bekommt einen
// Übersetzungsfehler.
func NeuerScryfallabgleich(
	fremd Fremdkarten, drucke Fremddrucke, verweise Verweisleser,
	schreiber Uebernahmeschreiber,
) *Scryfallabgleich {
	return &Scryfallabgleich{
		fremd: fremd, drucke: drucke, verweise: verweise, schreiber: schreiber,
	}
}

// Vorschlag sagt, was sich ändern würde — und ändert nichts.
//
// **Über den Namen.** Wer den gewählten Druck meint, nimmt
// [Scryfallabgleich.VorschlagFuerKarte].
func (s *Scryfallabgleich) Vorschlag(
	ctx context.Context, spiel string, jetzige map[string]string, name string,
) (*Vorschlag, error) {
	return s.ausQuelle(spiel, jetzige, func() (
		*scryfall.Karte, time.Duration, []domain.Befund, error,
	) {
		return s.fremd.KarteNachName(ctx, name)
	})
}

// VorschlagFuerKarte folgt dem Verweis der Karte, wenn sie einen hat.
//
// # Warum die Entscheidung hier steht und nicht in der Fassade
//
// §2 verbietet fachliche Entscheidungen in `internal/app`, und „welcher Druck
// ist gemeint?" ist eine. Stünde das `if` dort, träfe es nur das Fenster —
// Kommandozeile, MCP und Dienst holten weiter irgendeinen Druck, und §16 sagt:
// dieselbe Fähigkeit, überall dieselbe Antwort.
//
// **Ein Fehler beim Lesen des Verweises bricht ab.** Ein Rückfall auf den
// Namen holte still einen anderen Druck als den gewählten — und schriebe ihn
// unter einer Herkunft fort, die niemand angezweifelt hätte.
func (s *Scryfallabgleich) VorschlagFuerKarte(
	ctx context.Context, spiel string, jetzige map[string]string,
	karteID int64, name string,
) (*Vorschlag, error) {
	verweis, fehler := s.verweise.Fuer(ctx, karteID)
	if fehler != nil {
		return nil, fehler
	}

	if verweis == nil {
		return s.Vorschlag(ctx, spiel, jetzige, name)
	}

	return s.ausQuelle(spiel, jetzige, func() (
		*scryfall.Karte, time.Duration, []domain.Befund, error,
	) {
		return s.drucke.KarteNachKennung(ctx, verweis.Printing)
	})
}

// ausQuelle baut den Vorschlag aus einer geholten Karte.
//
// **Ein Rumpf für beide Wege.** Zwei gleichlautende Zweige wären zwei, von
// denen nur einer je einen Test sähe (§3).
func (s *Scryfallabgleich) ausQuelle(
	spiel string, jetzige map[string]string,
	hole func() (*scryfall.Karte, time.Duration, []domain.Befund, error),
) (*Vorschlag, error) {
	zuordnung, _, bekannt := scryfall.ZuordnungFuer(spiel)
	if !bekannt {
		return nil, fmt.Errorf(
			"%w: für das Spiel %q gibt es keine Zuordnung auf Scryfall-Felder",
			domain.ErrValidation, spiel)
	}

	karte, alter, befunde, fehler := hole()
	if fehler != nil {
		return nil, fehler
	}

	gefunden := zuordnung.Werte(karte)

	aenderungen := make([]Feldvorschlag, 0, len(gefunden))

	for feld, neu := range gefunden {
		if alt := jetzige[feld]; alt != neu {
			aenderungen = append(aenderungen, Feldvorschlag{Feld: feld, Alt: alt, Neu: neu})
		}
	}

	// **Nach Feldnamen sortiert.** Eine Map gibt ihre Schlüssel in wechselnder
	// Reihenfolge; ein Vorschlag, der bei jedem Aufruf anders aussieht, ist an
	// einer Oberfläche unbrauchbar und in einem Test nicht prüfbar.
	sort.Slice(aenderungen, func(links, rechts int) bool {
		return aenderungen[links].Feld < aenderungen[rechts].Feld
	})

	vorschlag := &Vorschlag{
		Aenderungen: aenderungen,
		Befunde:     befunde,
		Herkunft:    herkunft(karte),
		Druck:       karte.Kennung,
		Alter:       alter,
	}

	// **Der Stand gehört zum Vorschlag** (#1241) — er bindet ihn an die
	// Kartenwerte, gegen die verglichen wurde.
	vorschlag.Stand = vorschlag.BerechneterStand()

	return vorschlag, nil
}

// Uebernehmen schreibt die genannten Felder — und nur die.
//
// # Warum die Feldliste gegen den Vorschlag geprüft wird
//
// Ohne diese Wache könnte ein Aufrufer über die Liste **jeden** Wert setzen,
// und der Vorschlag wäre eine Höflichkeit statt einer Grenze. Wer etwas
// anderes schreiben will, nimmt den gewöhnlichen Weg; dieser trägt die
// Herkunft „von Scryfall", und die soll stimmen.
//
// # Was zurückkommt
//
// **Die übernommenen Änderungen samt ihrem vorigen Wert.** Ein Aufrufer, der
// zurück will, hat sie damit in der Hand — und muss den Vorschlag nicht
// aufgehoben haben. Für MCP ist das Pflicht
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)): ein
// schreibendes Werkzeug sagt, wie es zurückzunehmen ist.
//
// **Und im Fehlerfall: nichts.** Geschrieben wird in einem Zug; scheitert er,
// steht kein Feld neu da, und die leere Liste ist die Wahrheit
// (#1026).
// Vorher schrieb dieser Weg Feld für Feld und gab bei einem Fehler mitten
// darin `nil` zurück — der Aufrufer wusste dann weder, was geschrieben war,
// noch, wie er es zurücknimmt.
func (s *Scryfallabgleich) Uebernehmen(
	ctx context.Context, setID, karteID int64, vorschlag *Vorschlag, felder []string,
) ([]Feldvorschlag, error) {
	if vorschlag == nil {
		return nil, fmt.Errorf("%w: ohne Vorschlag wird nichts übernommen", domain.ErrValidation)
	}

	for _, feld := range felder {
		stelle := slices.IndexFunc(vorschlag.Aenderungen, func(eine Feldvorschlag) bool {
			return eine.Feld == feld
		})

		if stelle < 0 {
			return nil, fmt.Errorf(
				"%w: das Feld %q steht nicht im Vorschlag und wird nicht übernommen",
				domain.ErrValidation, feld)
		}
	}

	uebernommen := make([]Feldvorschlag, 0, len(felder))
	werte := make([]domain.Value, 0, len(felder))

	for _, feld := range felder {
		stelle := slices.IndexFunc(vorschlag.Aenderungen, func(eine Feldvorschlag) bool {
			return eine.Feld == feld
		})

		uebernommen = append(uebernommen, vorschlag.Aenderungen[stelle])
		werte = append(werte,
			domain.Value{Field: feld, Text: vorschlag.Aenderungen[stelle].Neu})
	}

	// **Ein Zug für alle Felder** (#1026). Eine Schleife über einzelne
	// Schreibaufrufe konnte halb durch sein, und der Fehler sagte nicht,
	// welche Hälfte — genau die Angabe, die eine Rücknahme bräuchte.
	if _, _, fehler := s.schreiber.FelderVonAussen(
		ctx, setID, karteID, werte, vorschlag.Herkunft); fehler != nil {
		return nil, fehler
	}

	return uebernommen, nil
}

// herkunft sagt, woher ein Vorschlag stammt.
//
// **Die Adresse, wenn es eine gibt, sonst die Kennung.** Beides zeigt auf
// dieselbe Karte bei Scryfall; die Adresse lässt sich anklicken.
func herkunft(karte *scryfall.Karte) string {
	if karte.Adresse != "" {
		return karte.Adresse
	}

	return "Scryfall " + karte.Kennung
}
