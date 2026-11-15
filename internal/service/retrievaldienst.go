package service

// **Die hybride Suche** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum Ränge zusammengeführt werden und keine Punkte
//
// **Weil BM25 und Cosinus nicht dieselbe Einheit haben.** Das Issue verbietet
// es ausdrücklich: „BM25, Cosinus und Rerankwerte nicht roh addieren und nicht
// als Wahrscheinlichkeit ausgeben." Ein BM25-Wert wächst mit der Länge des
// Korpus, ein Cosinus liegt zwischen −1 und 1 — ihre Summe ist eine Zahl ohne
// Bedeutung, und wer sie als Güte ausgäbe, behauptete eine Genauigkeit, die
// es nicht gibt.
//
// **Reciprocal Rank Fusion** rechnet stattdessen mit dem, was beide Listen
// gemeinsam haben: der Reihenfolge. Je Liste zählt `1 / (k + Rang)`; wer in
// beiden steht, sammelt zweimal. Das Verfahren stammt von Cormack, Clarke und
// Büttcher (SIGIR 2009) und ist mit k = 60 in derselben Arbeit vorgeschlagen.
//
// # Und warum die Filter vor der Grenze greifen
//
// **Weil eine Regelstelle sonst hinter zwanzig Kartentexten verschwindet.**
// Das Issue nennt es als Abnahme: „Filterfall darf nicht allein durch
// ungefiltertes Top-k seine Treffer verlieren." Gefragt wird deshalb weiter,
// als geantwortet wird — wie weit, sagt [Kandidatenfenster], und **was das
// nicht leistet**, steht dort daneben.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Rangfusionsk ist das `k` der Rangfusion.
//
// **Sechzig**, und die Zahl ist nicht gewählt, sondern übernommen: sie steht
// so in der Arbeit, die das Verfahren beschreibt. Sie dämpft die vorderen
// Plätze — ohne sie schlüge ein erster Platz jede Übereinstimmung in beiden
// Listen, und genau das soll die Hybridsuche nicht.
const Rangfusionsk = 60

// Rangfusionsname benennt das Verfahren im Manifest.
//
// **Er steht im Index**, weil eine andere Zusammenführung eine andere
// Trefferliste ergibt — und ein Vergleich zweier Läufe sonst zwei Verfahren
// vergleicht statt zweier Korpora.
const Rangfusionsname = "rrf-" + "60"

// FrageHoechstens begrenzt eine Suchfrage.
//
// **Zweitausend Zeichen**, dieselbe Grenze wie bei einer Bildaufgabe
// ([domain.KIBildaufgabeHoechstens]): was darüber liegt, ist kein Suchbegriff
// mehr, sondern ein Text, der eingebettet werden will — und dafür gibt es
// den Indexbau.
const FrageHoechstens = 2000

// Kandidatenfenster ist, um welchen Faktor weiter gefragt wird als geantwortet.
//
// # Was es leistet und was nicht
//
// **Es rettet den Filterfall in aller Regel** — zehn gewünschte Treffer holen
// hundert Kandidaten je Liste, und ein Filter, der jeden zehnten durchlässt,
// füllt die Antwort.
//
// **Es ist keine Zusage.** Ein Filter, dessen Treffer sämtlich jenseits des
// Fensters liegen, verliert sie trotzdem; das Fenster verschiebt die Grenze,
// es hebt sie nicht auf. Wer die Zusage braucht, filtert im Bestand — und das
// ist eine andere Abfrage als diese.
const Kandidatenfenster = 10

// KandidatenMindestens und KandidatenHoechstens klemmen das Fenster.
//
// **Nach unten**, damit eine Suche nach drei Treffern nicht an dreißig
// Kandidaten scheitert; **nach oben**, damit eine Suche nach tausend nicht
// den ganzen Korpus durch die Rangfusion zieht.
const (
	KandidatenMindestens = 50
	KandidatenHoechstens = 500
)

// Suchmodus sagt, womit gesucht wurde.
//
// **Er steht in jedem Ergebnis**, weil eine lexikalische Antwort etwas
// anderes ist als eine hybride — und niemand das an den Treffern sieht.
type Suchmodus string

// Die drei Modi.
const (
	// ModusHybrid ist beides — Volltext und Vektoren.
	ModusHybrid Suchmodus = "hybrid"

	// ModusLexikalisch ist der Rückfall: Volltext allein.
	ModusLexikalisch Suchmodus = "lexikalisch"

	// ModusOhneIndex heißt, dass gar kein Index aktiv ist.
	ModusOhneIndex Suchmodus = "ohne-index"
)

// Retrievalablage ist, was der Dienst vom Bestand braucht.
//
// **Sie steht hier und wird in `storage` erfüllt** (§2). Der Dienst kennt
// zwei Kandidatenlisten und ein Manifest; wie sie zustande kommen, ist nicht
// seine Frage.
type Retrievalablage interface {
	Lexikalisch(ctx context.Context, frage string, grenze int) (
		[]domain.Retrievaltreffer, error)
	Semantisch(ctx context.Context, frage []float32, grenze int) (
		[]domain.Retrievaltreffer, error)
	Manifest(ctx context.Context) (domain.Indexmanifest, error)
}

// Einbetter ist eine Einbettungsquelle, die ihren Vektorraum nennt.
//
// **Beides an einem Objekt, und das ist der Punkt.** Wer den Raum getrennt
// vom Einbetter reichte, könnte den Raum des einen mit den Vektoren des
// anderen paaren — und das ist genau der Fehler, den
// [domain.Vektorraum.PasstZu] verhindern soll.
type Einbetter interface {
	Einbettungsquelle

	Raum() domain.Vektorraum
}

// Suchwunsch ist eine Frage an den Index.
type Suchwunsch struct {
	Frage  string
	Grenze int
	Filter Retrievalfilter
}

// Suchergebnis ist die Antwort samt der Auskunft, wie sie zustande kam.
type Suchergebnis struct {
	Treffer []domain.Retrievaltreffer

	Modus Suchmodus

	// Grund sagt, warum es nicht der volle Modus war — leer bei [ModusHybrid].
	Grund string

	// Manifest ist der Stand, gegen den gesucht wurde.
	Manifest domain.Indexmanifest
}

// Retrievaldienst beantwortet Suchfragen gegen den lokalen Index.
type Retrievaldienst struct {
	ablage    Retrievalablage
	einbetter Einbetter

	// ausfallgrund ersetzt den allgemeinen Satz, wenn der Aufrufer den
	// genaueren kennt.
	//
	// **„Das Modell dieses Index ist nicht mehr installiert" ist etwas
	// anderes als „es ist keines eingerichtet"**, und der Unterschied sagt
	// jemandem, was zu tun ist. Der Dienst kann ihn nicht kennen: er sieht
	// nur, dass er keinen Einbetter hat.
	ausfallgrund string
}

// MitAusfallgrund sagt, warum kein Einbetter da ist.
func (d *Retrievaldienst) MitAusfallgrund(grund string) *Retrievaldienst {
	d.ausfallgrund = grund

	return d
}

// NeuerRetrievaldienst baut ihn.
//
// **Der Einbetter darf fehlen.** Das Issue verlangt es: „exakte Suche
// funktioniert ohne installierten Encoder" — ohne ihn bleibt die Suche
// lexikalisch und sagt es.
func NeuerRetrievaldienst(ablage Retrievalablage, einbetter Einbetter) *Retrievaldienst {
	return &Retrievaldienst{ablage: ablage, einbetter: einbetter}
}

// Suchen gibt die zusammengeführten Treffer.
func (d *Retrievaldienst) Suchen(ctx context.Context, wunsch Suchwunsch) (
	Suchergebnis, error,
) {
	if fehler := pruefeSuchwunsch(wunsch); fehler != nil {
		return Suchergebnis{}, fehler
	}

	manifest, fehler := d.ablage.Manifest(ctx)
	if fehler != nil {
		if errors.Is(fehler, domain.ErrNotFound) {
			return Suchergebnis{
				Treffer: []domain.Retrievaltreffer{},
				Modus:   ModusOhneIndex,
				Grund: "Es ist kein Suchindex aktiv. Wer noch keinen gebaut " +
					"hat, findet nichts — auch nicht lexikalisch.",
			}, nil
		}

		return Suchergebnis{}, fehler
	}

	fenster := kandidatenfenster(wunsch.Grenze)

	lexikalisch, fehler := d.ablage.Lexikalisch(ctx, wunsch.Frage, fenster)
	if fehler != nil {
		return Suchergebnis{}, fehler
	}

	semantisch, modus, grund := d.semantischeHaelfte(ctx, wunsch, manifest, fenster)

	return Suchergebnis{
		Treffer:  schneide(gefiltert(zusammengefuehrt(lexikalisch, semantisch), wunsch.Filter), wunsch.Grenze),
		Modus:    modus,
		Grund:    grund,
		Manifest: manifest,
	}, nil
}

// semantischeHaelfte holt die zweite Liste — oder sagt, warum es keine gibt.
//
// **Ein Ausfall ist hier kein Fehler.** Das Issue: „Ausfall ergibt sichtbaren
// Moduswechsel nach deklarierter Policy, keine andere Quelle/Cloud." Die
// lexikalische Hälfte trägt weiter; was fehlt, steht im Grund.
func (d *Retrievaldienst) semantischeHaelfte(
	ctx context.Context, wunsch Suchwunsch,
	manifest domain.Indexmanifest, fenster int,
) (treffer []domain.Retrievaltreffer, modus Suchmodus, grund string) {
	if d.einbetter == nil {
		if d.ausfallgrund != "" {
			return nil, ModusLexikalisch, d.ausfallgrund
		}

		return nil, ModusLexikalisch, "Es ist kein Einbettungsmodell " +
			"eingerichtet. Gesucht wird im Volltext."
	}

	raum := d.einbetter.Raum()
	if !raum.PasstZu(manifest.Raum()) {
		return nil, ModusLexikalisch, fmt.Sprintf(
			"Das eingerichtete Modell %q rechnet in einem anderen Vektorraum "+
				"als der Index, den %q gebaut hat. Gesucht wird im Volltext, "+
				"bis der Index neu gebaut ist.", raum.Modell, manifest.Modell)
	}

	vektoren, fehler := d.einbetter.Einbetten(ctx,
		[]string{raum.Vertrag.FuerFrage(wunsch.Frage)})
	if fehler != nil {
		return nil, ModusLexikalisch, mitSatz(
			"Das Einbettungsmodell hat nicht geantwortet", fehler)
	}

	if len(vektoren) != 1 {
		return nil, ModusLexikalisch, fmt.Sprintf(
			"Auf eine Frage kamen %d Vektoren zurück. Gesucht wird im Volltext.",
			len(vektoren))
	}

	treffer, fehler = d.ablage.Semantisch(ctx, vektoren[0], fenster)
	if fehler != nil {
		return nil, ModusLexikalisch, mitSatz(
			"Der Vektorteil des Index ist nicht lesbar", fehler)
	}

	return treffer, ModusHybrid, ""
}

// pruefeSuchwunsch weist ab, was keine Frage ist.
func pruefeSuchwunsch(wunsch Suchwunsch) error {
	if strings.TrimSpace(wunsch.Frage) == "" {
		return fmt.Errorf("%w: eine Suche braucht eine Frage", domain.ErrValidation)
	}

	if len(wunsch.Frage) > FrageHoechstens {
		return fmt.Errorf(
			"%w: die Frage hat %d Zeichen, höchstens %d — was länger ist, ist "+
				"kein Suchbegriff mehr", domain.ErrValidation,
			len(wunsch.Frage), FrageHoechstens)
	}

	if wunsch.Grenze <= 0 {
		return fmt.Errorf(
			"%w: eine Suche nach %d Treffern beantwortet sich von selbst",
			domain.ErrValidation, wunsch.Grenze)
	}

	return nil
}

// kandidatenfenster sagt, wie weit gefragt wird.
func kandidatenfenster(grenze int) int {
	return min(max(grenze*Kandidatenfenster, KandidatenMindestens),
		KandidatenHoechstens)
}

// zusammengefuehrt ist die Rangfusion.
//
// **Die Reihenfolge der Eingabe ist der Rang** und nicht das Feld `Rang`:
// eine Ablage, die vertauschte Ränge lieferte, verschöbe sonst still die
// Gewichte. Was zurückkommt, trägt seinen Rang **in dieser** Liste.
func zusammengefuehrt(listen ...[]domain.Retrievaltreffer) []domain.Retrievaltreffer {
	punkte := map[string]float64{}
	gefunden := map[string]domain.Retrievaltreffer{}
	reihenfolge := make([]string, 0)

	for _, liste := range listen {
		for stelle, treffer := range liste {
			kennung := treffer.Dokument.Kennung

			if _, kennt := gefunden[kennung]; !kennt {
				gefunden[kennung] = treffer
				reihenfolge = append(reihenfolge, kennung)
			}

			// **Der Ausschnitt kommt aus der Liste, die einen hat.** Die
			// semantische Hälfte kennt keine Textstelle; ohne diese Zeile
			// verlöre ein Treffer aus beiden Listen seine Erklärung.
			if vorhanden := gefunden[kennung]; vorhanden.Ausschnitt == "" &&
				treffer.Ausschnitt != "" {
				vorhanden.Ausschnitt = treffer.Ausschnitt
				gefunden[kennung] = vorhanden
			}

			punkte[kennung] += 1 / float64(Rangfusionsk+stelle+1)
		}
	}

	// **Die Ordnung ist die Punktzahl, der Gleichstand die Kennung.** Ohne
	// den zweiten Teil gäbe dieselbe Frage zweimal zwei Listen — `map` ordnet
	// in Go absichtlich nicht.
	slices.SortStableFunc(reihenfolge, func(links, rechts string) int {
		if punkte[links] != punkte[rechts] {
			if punkte[links] > punkte[rechts] {
				return -1
			}

			return 1
		}

		return strings.Compare(links, rechts)
	})

	zusammen := make([]domain.Retrievaltreffer, 0, len(reihenfolge))

	for stelle, kennung := range reihenfolge {
		treffer := gefunden[kennung]
		treffer.Rang = stelle + 1

		if treffer.Ausschnitt == "" {
			treffer.Ausschnitt = anfangsausschnitt(treffer.Dokument.Text)
		}

		zusammen = append(zusammen, treffer)
	}

	return zusammen
}

// gefiltert wirft weg, was der Filter nicht durchlässt — und zählt neu.
func gefiltert(
	treffer []domain.Retrievaltreffer, filter Retrievalfilter,
) []domain.Retrievaltreffer {
	if filter.leer() {
		return treffer
	}

	geblieben := make([]domain.Retrievaltreffer, 0, len(treffer))

	for _, einzelner := range treffer {
		if filter.laesstDurch(einzelner.Dokument) {
			einzelner.Rang = len(geblieben) + 1
			geblieben = append(geblieben, einzelner)
		}
	}

	return geblieben
}

// schneide kürzt auf die gewünschte Zahl.
func schneide(
	treffer []domain.Retrievaltreffer, grenze int,
) []domain.Retrievaltreffer {
	if len(treffer) <= grenze {
		return treffer
	}

	return treffer[:grenze]
}

// mitSatz hängt den Satz eines Fehlers an, wenn er einen hat.
//
// **Und nur dann.** [domain.SatzVon] gibt für einen fremden Fehler nichts —
// was hier hinausgeht, ist geschriebener Text und keine Fehlerkette (§8).
func mitSatz(kopf string, fehler error) string {
	if satz := domain.SatzVon(fehler); satz != "" {
		kopf += " (" + satz + ")"
	}

	return kopf + ". Gesucht wird im Volltext."
}
