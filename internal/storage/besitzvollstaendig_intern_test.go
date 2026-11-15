package storage

import (
	"context"
	"database/sql"
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Keine Methode einer Abkömmlingsablage fehlt in den Besitzläufen** —
// #1815.
//
// # Warum es diesen Wächter neben dem in `internal/werkzeuge` gibt
//
// `TestJedeAblagemethodeFragtNachDemProfil` sieht, **dass** eine Methode nach
// dem Profil fragt, nicht **wonach**. Die Turnierablage kam dort durch, weil
// sie `ProfilAus` für die Fassung des Turniers rief — und der Teilnehmer, um
// den es ging, stieg aus einem fremden Turnier aus. Gefunden hat das erst der
// Lauf aus einem fremden Profil.
//
// # Was er hält
//
// **Jede exportierte Methode der Ablagen in [abkoemmlingsablagen] wird aus
// einem fremden Profil gerufen** — in einer der Tafeln, die die Besitzläufe
// abarbeiten (Absage oder leere Antwort) — oder steht mit Grund in
// [ohneFremdeKennung]. Eine neue Methode erzwingt damit einen Lauf.
//
// # Was er nicht hält
//
// **Die übrigen Ablagen.** Sie tragen ihr Profil selbst (`reservierungen`,
// `peerkontakte`, `handelsjournal` …) oder gehören keinem; für sie gelten der
// Wächter über die Wurzelabfragen und der über die Frage nach dem Profil.

// abkoemmlingsablagen sind die Ablagen, deren Methoden Kennungen von Dingen
// annehmen, die einem Profil gehören.
var abkoemmlingsablagen = []any{
	(*Kartenschreiber)(nil), (*SetLeser)(nil), (*SetSpeicher)(nil), (*Marken)(nil),
	(*Kartenverweise)(nil), (*Verlaufsleser)(nil), (*Stilablage)(nil),
	(*Schluesselwortablage)(nil), (*Deckablage)(nil), (*Deckzuordnungen)(nil),
	(*Herkunftsablage)(nil), (*Sammlungsspeicher)(nil), (*Turnierablage)(nil),
	(*Matchablage)(nil), (*Partieablage)(nil),
}

// ohneFremdeKennung sind die Methoden dieser Ablagen, die keine Kennung von
// außen annehmen — mit Grund.
var ohneFremdeKennung = map[string]string{
	"Kartenschreiber.Als":      "ein Erbauer: gibt einen Schreiber mit anderem Urheber und liest nichts",
	"Kartenschreiber.Von":      "ein Erbauer: gibt einen Schreiber mit anderer Herkunft und liest nichts",
	"Schluesselwortablage.Als": "ein Erbauer: gibt eine Ablage mit anderem Urheber und liest nichts",
	"SetSpeicher.MitUhr":       "ein Erbauer: setzt die Uhr und liest nichts",
	"SetSpeicher.SetAnlegen":   "legt ein neues Set im gezeigten Profil an — es nimmt keine Kennung an",
	"SetLeser.Setkoepfe":       "gibt die Sets des gezeigten Profils; es nimmt keine Kennung an",
	"SetLeser.HerkunftVerwiesen": "fragt mit Absicht über alle Profile und sagt nur ja oder nein — der Eingang " +
		"des Dienstes gehört dem Rechner, und eine Quelle, die ein fremdes Profil braucht, darf nicht fort (#1944)",
	"Deckablage.Anlegen":                   "legt ein neues Deck im gezeigten Profil an — ohne Plätze, also ohne Karte",
	"Deckablage.Decks":                     "gibt die Decks des gezeigten Profils; das Spiel ist keine Kennung",
	"Deckzuordnungen.ZugeordnetJeExemplar": "summiert über die Decks des gezeigten Profils; es nimmt keine Kennung an",
	"Sammlungsspeicher.Anlegen":            "legt ein Exemplar im gezeigten Profil an — es nimmt keine Kennung an",
	"Sammlungsspeicher.AlleAnlegen":        "legt Exemplare im gezeigten Profil an — wie Anlegen, nur mehrere in einem Zug",
	"Sammlungsspeicher.Liste":              "gibt die Sammlung des gezeigten Profils; es nimmt keine Kennung an",
	"Sammlungsspeicher.Ordnungsliste":      "gibt die Sammlung des gezeigten Profils; es nimmt keine Kennung an",
	"Sammlungsspeicher.Fenster":            "wählt aus der Sammlung des gezeigten Profils; die Wahl bekommt nur deren Zeilen",
	"Sammlungsspeicher.Fuer":               "sucht einen Druck in der Sammlung des gezeigten Profils; ein Druck gehört keinem",
	"Sammlungsspeicher.NachDruck":          "sucht einen Druck in der Sammlung des gezeigten Profils; ein Druck gehört keinem",
	"Turnierablage.Turniere":               "gibt die Turniere des gezeigten Profils; das Spiel ist keine Kennung",
	"Partieablage.Partien":                 "gibt die Partien des gezeigten Profils; es nimmt keine Kennung an",
	"Partieablage.Anlegen": "legt eine Partie im gezeigten Profil an; ihre Kennung vergibt der Dienst " +
		"zufällig, und eine schon vergebene wiese der Bestand als doppelt ab",
}

// leererufe sind Rufe, die für eine fremde Kennung **eine leere Antwort**
// geben statt einer Absage — genau wie für eine, die es nicht gibt. Sie
// geben zurück, wie viel sie gefunden haben.
//
// **Warum keine Absage**: sie nehmen mehrere Kennungen an oder suchen, und
// „nichts gefunden" ist dort die Antwort auf eine fehlende wie auf eine fremde.
// Eine Absage für die fremde verriete, dass es sie gibt.
func leererufe(db *sql.DB, fremd fremdeKennungen, ctx context.Context) map[string]func() (int, error) {
	leser := NeuerSetLeser(db)

	return map[string]func() (int, error){
		"SetLeser.Kartenwerte": func() (int, error) {
			werte, f := leser.Kartenwerte(ctx, []int64{fremd.karte})
			return len(werte), f
		},
		"SetLeser.Felder": func() (int, error) { felder, f := leser.Felder(ctx, fremd.set); return len(felder), f },
		"SetLeser.Quelle": func() (int, error) {
			quelle, f := leser.Quelle(ctx, fremd.set)
			if quelle == nil {
				return 0, f
			}

			return 1, f
		},
		"SetLeser.Suche": func() (int, error) {
			treffer, f := leser.Suche(ctx, domain.SearchQuery{Text: "Fremde", SetID: fremd.set})
			return len(treffer.Hits), f
		},
		"SetLeser.Markentreffer": func() (int, error) {
			treffer, f := leser.Markentreffer(ctx, domain.SearchQuery{Text: "fremd", SetID: fremd.set})
			return len(treffer), f
		},
		"Deckablage.DecksMitKarte": func() (int, error) {
			decks, f := NeueDeckablage(db).DecksMitKarte(ctx, fremd.karte)
			return len(decks), f
		},
		"Sammlungsspeicher.NachKennungen": func() (int, error) {
			stuecke, f := NeuerSammlungsspeicher(db).NachKennungen(ctx, []int64{fremd.exemplar})
			return len(stuecke), f
		},
	}
}

// **Eine fremde Kennung gibt eine leere Antwort — und die des Eigners
// nicht.** Die Gegenprobe steht im selben Lauf: ohne sie hieße „leer" nur,
// dass der Ruf nichts findet.
func TestEineFremdeKennungGibtEineLeereAntwort(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	vorher := besitzschnappschuss(t, db)

	for name, ruf := range leererufe(db, fremd, MitProfil(t.Context(), 1)) {
		if gefunden, fehler := ruf(); fehler != nil || gefunden != 0 {
			t.Errorf("%s aus einem fremden Profil: %d gefunden (%v), erwartet nichts", name, gefunden, fehler)
		}
	}

	for name, ruf := range leererufe(db, fremd, MitProfil(t.Context(), 2)) {
		if gefunden, fehler := ruf(); fehler != nil || gefunden == 0 {
			t.Errorf("%s aus dem eigenen Profil: nichts gefunden (%v)", name, fehler)
		}
	}

	if nachher := besitzschnappschuss(t, db); nachher != vorher {
		t.Errorf("Profil 2 hat sich geändert:\n%s", unterschied(vorher, nachher))
	}
}

// **Jede Methode der Abkömmlingsablagen steht in einem Besitzlauf — oder mit
// Grund daneben.**
func TestJedeAbkoemmlingsmethodeWirdAusEinemFremdenProfilGerufen(t *testing.T) {
	t.Parallel()

	var leer fremdeKennungen

	gerufen := map[string]bool{}

	for _, rufe := range []map[string]func() error{
		kartenrufe(nil, leer, nil), deckrufe(nil, leer, nil),
		turnierrufe(nil, leer, nil), partierufe(nil, leer, nil),
	} {
		for name := range rufe {
			gerufen[name] = true
		}
	}

	for name := range leererufe(nil, leer, nil) {
		gerufen[name] = true
	}

	vorhanden := map[string]bool{}

	for _, ablage := range abkoemmlingsablagen {
		typ := reflect.TypeOf(ablage)

		for i := range typ.NumMethod() {
			name := typ.Elem().Name() + "." + typ.Method(i).Name
			vorhanden[name] = true

			if !gerufen[name] && ohneFremdeKennung[name] == "" {
				t.Errorf("%s wird aus keinem fremden Profil gerufen — in einen Besitzlauf "+
					"aufnehmen oder mit Grund in ohneFremdeKennung eintragen", name)
			}
		}
	}

	// **Und jede Ausnahme gibt es, und sie steht nicht zugleich in einem Lauf**
	// — eine Ausnahme für eine gerufene Methode wäre eine Begründung für
	// etwas, das gar nicht fehlt.
	for _, name := range slices.Sorted(maps.Keys(ohneFremdeKennung)) {
		switch {
		case !vorhanden[name]:
			t.Errorf("die Ausnahme %q nennt eine Methode, die es nicht gibt", name)
		case gerufen[name]:
			t.Errorf("die Ausnahme %q wird doch gerufen", name)
		case len(ohneFremdeKennung[name]) < 40:
			t.Errorf("die Ausnahme %q ist nicht begründet", name)
		}
	}

	for name := range gerufen {
		if !vorhanden[name] {
			t.Errorf("der Lauf %q nennt eine Methode, die es nicht gibt", name)
		}
	}
}
