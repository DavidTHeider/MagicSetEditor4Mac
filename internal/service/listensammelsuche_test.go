package service_test

// **Die Sammelanfrage der Listeneinfuhr** — Befund 6 des Zweitblicks über
// Phase 4.5
// (#1348).
//
// Was gegen den **echten** Dienst nachgelesen wurde, steht im Kopf von
// `internal/scryfall/sammlung.go` samt Abrufdatum (§16, Regel 1). Hier läuft
// nichts ins Netz.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sammelquelle zählt die Anfragen und antwortet aus einem Bestand.
//
// **So streng wie Scryfall**: mehr als [scryfall.Sammelgrenze] Kennungen
// weist sie ab. Eine Vorrichtung, die großzügiger wäre als der Dienst, prüfte
// den falschen Weg (§3) — und genau diese Klasse hat Phase 4.5 viermal
// getroffen.
type sammelquelle struct {
	karten  map[scryfall.Kennung]*scryfall.Karte
	fehler  error
	befunde []domain.Befund

	// alter ist, wie alt der Stand ist (§16, Regel 5).
	alter time.Duration

	// anfragen zählt die Rufe, groessen ihre Länge.
	anfragen int
	groessen []int
}

func (s *sammelquelle) Sammlung(_ context.Context, kennungen []scryfall.Kennung) (
	scryfall.Sammelantwort, time.Duration, []domain.Befund, error,
) {
	s.anfragen++
	s.groessen = append(s.groessen, len(kennungen))

	if s.fehler != nil {
		return scryfall.Sammelantwort{}, 0, nil, s.fehler
	}

	if len(kennungen) > scryfall.Sammelgrenze {
		return scryfall.Sammelantwort{}, 0, nil, fmt.Errorf(
			"%w: %d Kennungen auf einmal", domain.ErrValidation, len(kennungen))
	}

	antwort := scryfall.Sammelantwort{
		Karten: []*scryfall.Karte{}, NichtGefunden: []scryfall.Kennung{},
	}

	for _, kennung := range kennungen {
		karte, gefunden := s.karten[kennung]
		if !gefunden {
			antwort.NichtGefunden = append(antwort.NichtGefunden, kennung)

			continue
		}

		antwort.Karten = append(antwort.Karten, karte)
	}

	return antwort, s.alter, s.befunde, nil
}

// **Eine große Liste geht in Schüben hinaus und nicht Zeile für Zeile.**
//
// # Die Abnahme aus dem Issue
//
// Bei 5.000 verschiedenen Namen und 600 ms Abstand sind serielle Anfragen rund
// **50 Minuten**. Mit 75 Kennungen je Anfrage sind es 67 Rufe.
//
// **Gezählt wird hier klein**: 200 Zeilen sind drei Schübe (75, 75, 50). Die
// Zahl ist dieselbe Rechnung, und ein Lauf über 5.000 Zeilen kostete nur
// Rechenzeit ohne mehr zu zeigen.
func TestEineGrosseListeGehtInSchueben(t *testing.T) {
	t.Parallel()

	const zeilen = 200

	quelle := &sammelquelle{karten: map[scryfall.Kennung]*scryfall.Karte{}}

	var text string

	for i := range zeilen {
		nummer := strconv.Itoa(i + 1)
		text += "1 Karte " + nummer + " (m11) " + nummer + "\n"

		kennung := scryfall.Kennung{Setcode: "m11", Nummer: nummer}
		quelle.karten[kennung] = druck("k"+nummer, "Karte "+nummer, "m11", nummer)
	}

	einzeln := &listenquelle{}

	deutung, fehler := service.NeueListeneinfuhr(einzeln, &sammlungsdoppel{}).
		MitSammelsuche(quelle).Deuten(t.Context(), text)
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	// **Drei Anfragen statt zweihundert.**
	if quelle.anfragen != 3 {
		t.Errorf("%d Sammelanfragen, erwartet waren drei: %v",
			quelle.anfragen, quelle.groessen)
	}

	// **Und keine einzige serielle.** Ohne diese Zusicherung wäre der Lauf
	// auch dann grün, wenn beide Wege liefen — und dann wäre nichts gewonnen.
	if len(einzeln.gefragt) != 0 {
		t.Errorf("%d Einzelanfragen gingen zusätzlich hinaus", len(einzeln.gefragt))
	}

	if len(deutung.Zeilen) != zeilen {
		t.Fatalf("%d Zeilen gedeutet, erwartet %d", len(deutung.Zeilen), zeilen)
	}

	// **Jede Zeile hat ihren Druck** — und zwar ihren eigenen.
	for stelle, zeile := range deutung.Zeilen {
		nummer := strconv.Itoa(stelle + 1)

		if len(zeile.Treffer) != 1 || zeile.Treffer[0].Name != "Karte "+nummer {
			t.Fatalf("Zeile %s bekam %+v", nummer, zeile.Treffer)
		}
	}
}

// **Eine Zeile ohne Setcode bleibt bei der Einzelsuche.**
//
// Sie ist mehrdeutig: „Counterspell" hat 72 Drucke, und die Sammelanfrage gäbe
// davon **einen** — Scryfalls neuesten, nicht den, den jemand besitzt. Wer sie
// mitschickte, spart eine Anfrage und nimmt der Person die Wahl (§16, Regel 6).
func TestEineZeileOhneSetBleibtBeiDerEinzelsuche(t *testing.T) {
	t.Parallel()

	sammel := &sammelquelle{karten: map[scryfall.Kennung]*scryfall.Karte{
		{Setcode: "m11", Nummer: "149"}: druck("a", "Lightning Bolt", "m11", "149"),
	}}

	einzeln := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Counterspell"`: {Karten: []*scryfall.Karte{
			druck("b", "Counterspell", "lea", "54"),
			druck("c", "Counterspell", "dsc", "82"),
		}},
	}}

	deutung, fehler := service.NeueListeneinfuhr(einzeln, &sammlungsdoppel{}).
		MitSammelsuche(sammel).
		Deuten(t.Context(), "4 Lightning Bolt (M11) 149\n2 Counterspell\n")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if sammel.anfragen != 1 || sammel.groessen[0] != 1 {
		t.Errorf("Sammelanfragen: %d %v", sammel.anfragen, sammel.groessen)
	}

	if einzeln.gefragt[`!"Counterspell"`] != 1 {
		t.Errorf("die offene Zeile ging nicht einzeln hinaus: %v", einzeln.gefragt)
	}

	// **Und die offene Zeile behält ihre Auswahl.**
	if len(deutung.Zeilen[1].Treffer) != 2 {
		t.Errorf("die offene Zeile bekam %+v", deutung.Zeilen[1].Treffer)
	}
}

// **Was die Sammelanfrage nicht findet, bekommt seinen Befund.**
//
// Scryfall meldet es in `not_found`, und das ist kein Fehler (§8): die übrigen
// Zeilen sollen trotzdem durchkommen.
func TestWasDieSammelanfrageNichtFindetBekommtSeinenBefund(t *testing.T) {
	t.Parallel()

	sammel := &sammelquelle{karten: map[scryfall.Kennung]*scryfall.Karte{
		{Setcode: "m11", Nummer: "149"}: druck("a", "Lightning Bolt", "m11", "149"),
	}}

	deutung, fehler := service.NeueListeneinfuhr(&listenquelle{}, &sammlungsdoppel{}).
		MitSammelsuche(sammel).
		Deuten(t.Context(), "4 Lightning Bolt (M11) 149\n1 Gibtesnicht (M11) 999\n")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if len(deutung.Zeilen[0].Treffer) != 1 {
		t.Errorf("die gefundene Zeile bekam %+v", deutung.Zeilen[0])
	}

	if deutung.Zeilen[1].Befund == "" || len(deutung.Zeilen[1].Treffer) != 0 {
		t.Errorf("die fehlende Zeile bekam %+v", deutung.Zeilen[1])
	}
}

// **Ein Fehler der Sammelanfrage bricht die Deutung ab.**
//
// Anders als ein `not_found`: wer kein Netz hat, bekommt für **keine** Zeile
// eine Antwort, und eine Deutung mit lauter „kein Druck heißt so" sähe aus wie
// eine Liste voller Tippfehler.
func TestEinFehlerDerSammelanfrageBrichtAb(t *testing.T) {
	t.Parallel()

	sammel := &sammelquelle{fehler: errors.New("kein Netz")}

	_, fehler := service.NeueListeneinfuhr(&listenquelle{}, &sammlungsdoppel{}).
		MitSammelsuche(sammel).
		Deuten(t.Context(), "4 Lightning Bolt (M11) 149\n")
	if fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}
}

// **Ohne gesetzten Weg geht alles einzeln.**
//
// Die Gegenprobe zu [service.Listeneinfuhr.MitSammelsuche]: eine Fläche, die
// ihn nicht setzt, bekommt dasselbe Ergebnis — nur langsamer.
func TestOhneSammelsucheGehtAllesEinzeln(t *testing.T) {
	t.Parallel()

	einzeln := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Lightning Bolt" e:m11 cn:149`: {Karten: []*scryfall.Karte{
			druck("a", "Lightning Bolt", "m11", "149"),
		}},
	}}

	deutung, fehler := service.NeueListeneinfuhr(einzeln, &sammlungsdoppel{}).
		Deuten(t.Context(), "4 Lightning Bolt (M11) 149\n")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if einzeln.gefragt[`!"Lightning Bolt" e:m11 cn:149`] != 1 {
		t.Errorf("die Zeile ging nicht einzeln hinaus: %v", einzeln.gefragt)
	}

	if len(deutung.Zeilen[0].Treffer) != 1 {
		t.Errorf("die Zeile bekam %+v", deutung.Zeilen[0])
	}
}

// **Eine Zeile mit Set aber ohne Nummer fragt über den Namen.**
//
// Scryfalls Kennungsschemata sind `name,set` **oder** `collector_number,set`.
// Wer beides schickte, stellte eine Frage, die dort nicht steht.
func TestEineZeileOhneNummerFragtUeberDenNamen(t *testing.T) {
	t.Parallel()

	sammel := &sammelquelle{karten: map[scryfall.Kennung]*scryfall.Karte{
		{Name: "Lightning Bolt", Setcode: "m11"}: druck("a", "Lightning Bolt", "m11", "149"),
	}}

	deutung, fehler := service.NeueListeneinfuhr(&listenquelle{}, &sammlungsdoppel{}).
		MitSammelsuche(sammel).Deuten(t.Context(), "4 Lightning Bolt (M11)\n")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if len(deutung.Zeilen[0].Treffer) != 1 {
		t.Errorf("die Zeile bekam %+v", deutung.Zeilen[0])
	}
}

// **Dieselbe adressierte Zeile wird nur einmal gefragt.**
//
// Eine Deckliste nennt dieselbe Karte mehrfach — vier Berge sind vier Zeilen,
// wenn jemand sie einzeln schreibt. Ohne diese Zusage stünde derselbe Druck
// mehrfach im Schub und verbrauchte Plätze, die eine andere Karte braucht.
func TestDieselbeAdressierteZeileWirdEinmalGefragt(t *testing.T) {
	t.Parallel()

	sammel := &sammelquelle{karten: map[scryfall.Kennung]*scryfall.Karte{
		{Setcode: "m11", Nummer: "149"}: druck("a", "Lightning Bolt", "m11", "149"),
	}}

	deutung, fehler := service.NeueListeneinfuhr(&listenquelle{}, &sammlungsdoppel{}).
		MitSammelsuche(sammel).Deuten(t.Context(),
		"2 Lightning Bolt (M11) 149\n2 Lightning Bolt (M11) 149\n")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if sammel.anfragen != 1 || sammel.groessen[0] != 1 {
		t.Errorf("Anfragen: %d %v — erwartet war eine mit einer Kennung",
			sammel.anfragen, sammel.groessen)
	}

	// **Und beide Zeilen bekommen ihre Deutung.** Ohne diese Zusicherung wäre
	// der Lauf auch dann grün, wenn die zweite Zeile leer bliebe.
	for stelle, zeile := range deutung.Zeilen {
		if len(zeile.Treffer) != 1 {
			t.Errorf("Zeile %d bekam %+v", stelle+1, zeile.Treffer)
		}
	}
}

// **Das Alter der Sammelantwort kommt an.**
//
// §16, Regel 5: der letzte gute Stand bleibt, „und er sagt, wie alt er ist".
// Eine Deutung, die ihr Alter verschwiege, sähe frisch aus.
func TestDasAlterDerSammelantwortKommtAn(t *testing.T) {
	t.Parallel()

	sammel := &sammelquelle{
		karten: map[scryfall.Kennung]*scryfall.Karte{
			{Setcode: "m11", Nummer: "149"}: druck("a", "Lightning Bolt", "m11", "149"),
		},
		alter: 3 * time.Hour,
	}

	deutung, fehler := service.NeueListeneinfuhr(&listenquelle{}, &sammlungsdoppel{}).
		MitSammelsuche(sammel).Deuten(t.Context(), "2 Lightning Bolt (M11) 149\n")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if deutung.Alter != 3*time.Hour {
		t.Errorf("Alter %v, erwartet waren drei Stunden", deutung.Alter)
	}
}
