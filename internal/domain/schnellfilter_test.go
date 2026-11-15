// **Der Schnellfilter der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416),
// portiert aus MSE2 `src/data/filter.cpp:1–41` und `src/data/card.cpp:55–61`
// (Fassung f9d9356 vom 2021-07-06). Die erwarteten Werte sind aus dem
// Zerteiler dort abgelesen, Zeichen für Zeichen.
package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestSchnellfilterZerlegtDieAnfrageWieMSE(t *testing.T) {
	t.Parallel()

	teil := func(typ, wert string, soll bool) domain.Filterteil {
		return domain.Filterteil{Typ: typ, Wert: wert, Soll: soll}
	}

	faelle := map[string]struct {
		anfrage string
		teile   []domain.Filterteil
	}{
		"leer":                  {"", nil},
		"nur Leerraum":          {"   ", nil},
		"ein Wort":              {"goblin", []domain.Filterteil{teil("", "goblin", true)}},
		"zwei Wörter, Leerraum": {"  two   words ", []domain.Filterteil{teil("", "two", true), teil("", "words", true)}},
		"mit Feld":              {"name:goblin rules:haste", []domain.Filterteil{teil("name", "goblin", true), teil("rules", "haste", true)}},
		"Anführungszeichen halten zusammen": {
			`"grizzly bears" green`, []domain.Filterteil{teil("", "grizzly bears", true), teil("", "green", true)},
		},
		"Verneinung":          {"-token", []domain.Filterteil{teil("", "token", false)}},
		"Verneinung mit Feld": {"-type:token", []domain.Filterteil{teil("type", "token", false)}},
		// **Nach dem Feld ist `-` ein Zeichen wie jedes**: der Zerteiler
		// verneint nur, wenn noch kein Feld und kein Wert steht.
		"Minus nach dem Feld":      {"type:-creature", []domain.Filterteil{teil("type", "-creature", true)}},
		"der zweite Doppelpunkt":   {"a:b:c", []domain.Filterteil{teil("a", "b:c", true)}},
		"Doppelpunkt in Anführung": {`"rules:haste"`, []domain.Filterteil{teil("", "rules:haste", true)}},
		"Feld ohne Wert fällt weg": {"name:", nil},
		"offene Anführung":         {`"a b`, []domain.Filterteil{teil("", "a b", true)}},
		"Umlaute und Groß":         {"Zauberer", []domain.Filterteil{teil("", "Zauberer", true)}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			teile := domain.ZerlegeSchnellfilter(fall.anfrage)

			if len(teile) != len(fall.teile) {
				t.Fatalf("%d Teile, erwartet %d: %+v", len(teile), len(fall.teile), teile)
			}

			for i := range teile {
				if teile[i] != fall.teile[i] {
					t.Errorf("Teil %d ist %+v, erwartet %+v", i, teile[i], fall.teile[i])
				}
			}
		})
	}
}

func TestSchnellfilterTrifftKartenWieMSE(t *testing.T) {
	t.Parallel()

	baer := domain.CardRow{ID: 1, Values: []domain.Value{
		{Field: "name", Text: "Grizzly Bears"},
		{Field: "rule text", Text: "A 2/2 for two."},
		{Field: "type", Text: "Creature — Bear"},
		{Field: domain.Notizenfeld, Text: "Sammlerstück"},
	}}
	goblin := domain.CardRow{ID: 2, Values: []domain.Value{
		{Field: "name", Text: "Goblin Guide"},
		{Field: "rule text", Text: "Haste"},
		{Field: "type", Text: "Creature — Goblin Scout"},
	}}

	faelle := map[string]struct {
		anfrage string
		trifft  []int64
	}{
		"leere Anfrage lässt alle stehen":          {"", []int64{1, 2}},
		"Teilwort, Groß- und Kleinschreibung egal": {"BEAR", []int64{1}},
		"irgendein Feld":                           {"haste", []int64{2}},
		"mit Feld: der Feldname enthält den Typ":   {"name:goblin", []int64{2}},
		"das Feld enthält den Typ als Teil":        {"ame:bear", []int64{1}},
		"Typ trifft kein Feld":                     {"color:bear", nil},
		"zwei Teile müssen beide treffen":          {"creature bears", []int64{1}},
		"Verneinung":                               {"creature -goblin", []int64{1}},
		"Notizen zählen mit":                       {"sammler", []int64{1}},
		"Notizen über ihren Schlüssel":             {"notes:sammler", []int64{1}},
		"Anführungszeichen":                        {`"2/2 for"`, []int64{1}},
		"Wert im Typ trifft nicht":                 {"bear:name", nil},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			teile := domain.ZerlegeSchnellfilter(fall.anfrage)
			getroffen := []int64{}

			for _, karte := range []domain.CardRow{baer, goblin} {
				if domain.SchnellfilterTrifft(teile, karte) {
					getroffen = append(getroffen, karte.ID)
				}
			}

			if len(getroffen) != len(fall.trifft) {
				t.Fatalf("getroffen %v, erwartet %v", getroffen, fall.trifft)
			}

			for i := range getroffen {
				if getroffen[i] != fall.trifft[i] {
					t.Errorf("getroffen %v, erwartet %v", getroffen, fall.trifft)
				}
			}
		})
	}
}
