package msescript_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`position(of: card, in: set, order_by:, filter:)` zählt die Karten in
// ihrer Ordnung** — #2006.
//
//	} else if (order_by || filter) {
//	  ScriptObject<Set*>*  s = dynamic_cast<ScriptObject<Set*>* >(in.get());
//	  ScriptObject<CardP>* c = dynamic_cast<ScriptObject<CardP>*>(of.get());
//	  if (s && c) {
//	    return s->getValue()->positionOfCard(c->getValue(), order_by, filter);
//	  } else {
//	    throw ScriptError(_("position: using 'order_by' or 'filter' is only "
//	                        "supported for finding cards in the set"));
//	  }
//
// `src/script/functions/basic.cpp`, Zeilen 449–474, Fassung f9d9356 vom
// 2021-07-06. So rechnet `magic.mse-game` die Kartennummer (`card_number`,
// Zeile 987 des Initskripts): die Stelle der Karte im Set, geordnet nach
// Rarität und Name. Bisher suchte `position` den Wert in der Liste und
// ignorierte beide Argumente — jede Karte bekam die Nummer 0.

// satzmitKarten baut ein Set mit Karten aus Namen. Die Kennung macht die Karte
// zum Gegenstand, so wie das `set` aus `internal/service` es tut.
func satzmitKarten(namen ...string) (set msescript.Wert, karten []msescript.Wert) {
	eintraege := make([]msescript.Eintrag, 0, len(namen))

	for stelle, name := range namen {
		karte := msescript.Wert{
			Art:     msescript.ArtAbbildung,
			Kennung: fmt.Sprintf("karte:%d", stelle+1),
			Eintraege: []msescript.Eintrag{
				{Name: "name", Wert: msescript.Text(name)},
				{Name: "art", Wert: msescript.Text(strings.ToUpper(name[:1]))},
			},
		}

		karten = append(karten, karte)
		eintraege = append(eintraege, msescript.Eintrag{Name: fmt.Sprint(stelle), Wert: karte})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Kennung: "set", Eintraege: eintraege}, karten
}

// stelleimSet fragt die Stelle der Karte `stelle` (von null an) im Set.
func stelleimSet(t *testing.T, umgebung *msescript.Umgebung, karte msescript.Wert, rest string) string {
	t.Helper()

	eigene := umgebung.Darueber()
	eigene.Setze("card", karte)

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`position(of: card, in: set`+rest+`)`, eigene)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	return wert.AlsText()
}

func umgebungMitSet(namen ...string) (*msescript.Umgebung, []msescript.Wert) {
	set, karten := satzmitKarten(namen...)
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("set", set)

	return umgebung, karten
}

func TestPositionOrdnetDieKartenDesSetsNachDemSchluessel(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Cäsar", "Anton", "Berta")

	for stelle, erwartet := range []string{"2", "0", "1"} {
		if gesehen := stelleimSet(t, umgebung, karten[stelle],
			`, order_by: {card.name}`); gesehen != erwartet {
			t.Errorf("Karte %d steht an Stelle %s, erwartet %s", stelle, gesehen, erwartet)
		}
	}
}

// **Der Schlüssel ist Text, und er wird klug verglichen** — dieselbe Ordnung
// wie bei `sort_list` (`OrderCache`, `src/util/order_cache.hpp`, Zeilen
// 44–62, Fassung f9d9356 vom 2021-07-06: `smart_less(values[a.second],
// values[b.second])`).
func TestPositionVergleichtDieSchluesselKlug(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Karte 10", "Karte 2")

	if gesehen := stelleimSet(t, umgebung, karten[1], `, order_by: {card.name}`); gesehen != "0" {
		t.Errorf("»Karte 2« steht an Stelle %s, erwartet 0 — Zahlen zählen als Zahlen", gesehen)
	}
}

// **Ein Filter nimmt Karten aus der Zählung** — und die ausgenommene hat keine
// Stelle (`OrderCache::find` gibt −1, wenn der Schlüssel fehlt).
func TestPositionZaehltNurDieKartenDesFilters(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton", "Berta", "Cäsar")
	rest := `, order_by: {card.name}, filter: {card.name != "Anton"}`

	for stelle, erwartet := range []string{"-1", "0", "1"} {
		if gesehen := stelleimSet(t, umgebung, karten[stelle], rest); gesehen != erwartet {
			t.Errorf("Karte %d steht an Stelle %s, erwartet %s", stelle, gesehen, erwartet)
		}
	}
}

// **`nil` als Filter heißt: kein Filter** (`if (filter == script_nil) filter =
// ScriptValueP();`, `basic.cpp` Zeile 528). `set_filter()` gibt das für jede
// Karte, die keinen eigenen Zweig hat.
func TestPositionBehandeltEinenFilterNilAlsKeinenFilter(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Berta", "Anton")

	if gesehen := stelleimSet(t, umgebung, karten[1],
		`, order_by: {card.name}, filter: nil`); gesehen != "0" {
		t.Errorf("Stelle %s, erwartet 0", gesehen)
	}
}

// **Der Schlüssel sieht, was das Skript selbst erklärt hat** — `card_number`
// ruft `sort_name(card.name)` aus dem Initskript des Spiels. Gerechnet wird
// in der Umgebung des Aufrufers, mit `card` als der jeweiligen Karte.
func TestPositionRechnetDenSchluesselInDerUmgebungDesAufrufers(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton", "Berta")

	if err := msescript.Vorbelegen(t.Context(),
		`rueckwaerts := { if input == "Anton" then "z" else "a" }`, umgebung); err != nil {
		t.Fatalf("erklären: %v", err)
	}

	if gesehen := stelleimSet(t, umgebung, karten[0],
		`, order_by: {rueckwaerts(card.name)}`); gesehen != "1" {
		t.Errorf("Stelle %s, erwartet 1", gesehen)
	}
}

// **Gleiche Schlüssel bleiben in der Reihenfolge der Liste** — MSE nimmt
// `std::sort`, hier ist es stabil, so wie bei `sort_list`.
func TestPositionLaesstGleicheSchluesselInDerReihenfolgeDerListe(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton", "Adam", "Anna")

	for stelle, erwartet := range []string{"0", "1", "2"} {
		if gesehen := stelleimSet(t, umgebung, karten[stelle],
			`, order_by: {card.art}`); gesehen != erwartet {
			t.Errorf("Karte %d steht an Stelle %s, erwartet %s", stelle, gesehen, erwartet)
		}
	}
}

// **Ohne Set und ohne Karte gibt es nur den Fehler, den MSE wirft.**
func TestPositionMitOrdnungBrauchtKarteUndSet(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton")

	for _, fall := range []struct{ name, quelle string }{
		{"eine Zahl statt der Karte", `position(of: 1, in: set, order_by: {card.name})`},
		{"eine Liste statt des Sets", `position(of: card, in: [card], order_by: {card.name})`},
		{"ein Filter allein", `position(of: card, in: [card], filter: {true})`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			eigene := umgebung.Darueber()
			eigene.Setze("card", karten[0])

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, eigene)
			if err == nil || !strings.Contains(err.Error(), "nur für Karten im Set") {
				t.Errorf("Fehler %v", err)
			}
		})
	}
}

// **Und ohne `order_by` kann ein Filter nicht ordnen** — MSE prüft das mit
// einem `assert(order_by)` und stürzte im Release ab.
func TestPositionMitFilterOhneOrdnungMeldetDenFehler(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton")
	eigene := umgebung.Darueber()
	eigene.Setze("card", karten[0])

	_, err := msescript.AuswertenQuelle(t.Context(),
		`position(of: card, in: set, filter: {true})`, eigene)
	if err == nil || !strings.Contains(err.Error(), "order_by") {
		t.Errorf("Fehler %v", err)
	}
}

// **Ein Ausdruck, der scheitert, scheitert die Position mit.**
func TestPositionGibtDenFehlerDesSchluesselsWeiter(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton")
	eigene := umgebung.Darueber()
	eigene.Setze("card", karten[0])

	for _, rest := range []string{
		`order_by: {unbekannt()}`,
		`order_by: {card.name}, filter: {unbekannt()}`,
		`order_by: {card.name}, filter: {card.name}`,
		`order_by: "kein Block"`,
	} {
		if _, err := msescript.AuswertenQuelle(t.Context(),
			`position(of: card, in: set, `+rest+`)`, eigene); err == nil {
			t.Errorf("%s: kein Fehler", rest)
		}
	}
}

// **Ein Text auf einer der beiden Seiten bleibt die Textsuche**, auch mit
// `order_by:` — MSE prüft die Zeichenketten zuerst (`basic.cpp`, Zeile 452),
// und sucht dabei `in` in `of` (`stelletext_test.go`).
func TestPositionSuchtMitOrdnungInEinemTextWeiterImText(t *testing.T) {
	t.Parallel()

	umgebung, _ := umgebungMitSet("Anton")

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`position(of: "abc", in: "c", order_by: {input})`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "2" {
		t.Errorf("Stelle %s, erwartet 2", wert.AlsText())
	}
}

// **Im Zeichenpfad führt `set` seine Felder und keine Karten**, und die Karte
// trägt keine Kennung: `position` gibt −1, rechnet keinen Schlüssel und wirft
// nichts — `card_number` ergibt dort wie bisher 0 (#2006).
func TestPositionInEinemSetOhneKartenGibtMinusEins(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("set", msescript.Wert{
		Art: msescript.ArtAbbildung, Kennung: "set",
		Eintraege: []msescript.Eintrag{{Name: "title", Wert: msescript.Text("Probe")}},
	})
	umgebung.Setze("card", msescript.Wert{
		Art:       msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{{Name: "name", Wert: msescript.Text("Anton")}},
	})

	// Der Schlüssel scheiterte, würde er je gerechnet: `title` hat kein `name`.
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`position(of: card, in: set, order_by: {card.name.zeichen})`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "-1" {
		t.Errorf("Stelle %s, erwartet -1", wert.AlsText())
	}
}

// **Auch viele gleiche Schlüssel bleiben in der Reihenfolge der Liste.** Bei
// einem Dutzend sortiert jedes Verfahren von selbst stabil, und bei lauter
// gleichen Schlüsseln lässt jedes die Liste, wie sie ist; erst viele Karten in
// drei Gruppen trennen `sort.SliceStable` von `sort.Slice`.
func TestPositionLaesstVieleGleicheSchluesselInDerReihenfolgeDerListe(t *testing.T) {
	t.Parallel()

	const gruppe = 67

	namen := make([]string, 3*gruppe)
	for i := range namen {
		namen[i] = fmt.Sprintf("%c Karte %03d", 'A'+rune(i%3), i)
	}

	umgebung, karten := umgebungMitSet(namen...)

	for stelle, karte := range karten {
		// Der Schlüssel `card.art` ist der erste Buchstabe: A, B oder C.
		erwartet := (stelle%3)*gruppe + stelle/3

		if gesehen := stelleimSet(t, umgebung, karte,
			`, order_by: {card.art}`); gesehen != fmt.Sprint(erwartet) {
			t.Fatalf("Karte %d steht an Stelle %s, erwartet %d", stelle, gesehen, erwartet)
		}
	}
}
