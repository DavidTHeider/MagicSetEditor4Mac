package service_test

// **Eine Kartenliste einlesen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// Was gegen den **echten** Dienst nachgelesen wurde, steht im Kopf von
// `listeneinfuhr.go` samt Abrufdatum (§16, Regel 1). Hier läuft nichts ins
// Netz: gegen ein Doppel, wie §3 es verlangt.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// listenquelle antwortet je Frage verschieden.
//
// **Nicht `suchquelle`**: die gibt auf jede Frage dieselbe Antwort, und eine
// Liste stellt mehrere. Ein Doppel, das nicht unterscheidet, ließe die Zusage
// „gleiche Fragen werden einmal gestellt" ungeprüft.
type listenquelle struct {
	antworten map[string]scryfall.Treffer
	fehler    map[string]error
	alter     time.Duration

	// befunde sind, was die Quelle nebenbei meldet (Befund 10).
	//
	// **Sie kommen mit einer Karte zugleich** — das ist der Fall, den §16
	// „Pflichtfelder streng, Zusatzfelder tolerant" meint.
	befunde []domain.Befund

	// gefragt zählt je Frage die Rufe — für die Zusage über das Gedächtnis.
	gefragt map[string]int
}

// **`Druckesuche` und nicht `Suche`** (#1292): die Einfuhr fragt nach
// **Ausgaben**, nicht nach Karten. Ein Doppel, das die falsche Methode
// erfüllte, hielte eine Zusage über eine Naht, die es gar nicht gibt.
func (l *listenquelle) Druckesuche(_ context.Context, frage string) (
	scryfall.Treffer, time.Duration, []domain.Befund, error,
) {
	if l.gefragt == nil {
		l.gefragt = map[string]int{}
	}

	l.gefragt[frage]++

	if fehler, da := l.fehler[frage]; da {
		return scryfall.Treffer{}, 0, nil, fehler
	}

	return l.antworten[frage], l.alter, l.befunde, nil
}

// druck baut eine Scryfall-Karte, so weit die Einfuhr sie braucht.
func druck(kennung, name, set, nummer string) *scryfall.Karte {
	return &scryfall.Karte{
		Kennung: kennung, Name: name, Set: set, Nummer: nummer,
	}
}

// **Eine eindeutige Zeile bekommt genau einen Druck.**
func TestEineEindeutigeZeile(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Lightning Bolt" e:m11 cn:149`: {
			Karten: []*scryfall.Karte{druck("abc", "Lightning Bolt", "m11", "149")},
		},
	}}

	einfuhr := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{})

	deutung, fehler := einfuhr.Deuten(t.Context(), "4 Lightning Bolt (M11) 149")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if len(deutung.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet 1", len(deutung.Zeilen))
	}

	zeile := deutung.Zeilen[0]

	if !zeile.Eindeutig() {
		t.Errorf("die Zeile gilt als nicht eindeutig: %+v", zeile)
	}

	if zeile.Zeile.Menge != 4 {
		t.Errorf("Menge %d, erwartet 4", zeile.Zeile.Menge)
	}

	if zeile.Treffer[0].Kennung != "abc" {
		t.Errorf("Kennung %q, erwartet abc", zeile.Treffer[0].Kennung)
	}
}

// **Ein Name ohne Set ist offen — und wird nicht geraten.**
//
// Der Fall, für den es den zweiten Schritt gibt: „Counterspell" hat bei
// Scryfall 72 Drucke (nachgelesen am 2026-08-30). Einen davon still
// einzutragen füllte eine Sammlung mit Papier, das niemand besitzt.
func TestEinNameOhneSetBleibtOffen(t *testing.T) {
	t.Parallel()

	viele := make([]*scryfall.Karte, 0, 30)
	for zaehler := range 30 {
		viele = append(viele, druck(
			"k"+string(rune('a'+zaehler)), "Counterspell", "s"+string(rune('a'+zaehler)), "1"))
	}

	quelle := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Counterspell"`: {Karten: viele},
	}}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "2 Counterspell")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	zeile := deutung.Zeilen[0]

	if zeile.Eindeutig() {
		t.Error("dreißig Drucke gelten als eindeutig")
	}

	// **Die Gesamtzahl steht da, auch wenn die Liste gekappt ist** (§10).
	if zeile.Drucke != 30 {
		t.Errorf("Drucke %d, erwartet 30", zeile.Drucke)
	}

	if len(zeile.Treffer) != service.Deutungsgrenze {
		t.Errorf("%d Treffer, erwartet %d", len(zeile.Treffer), service.Deutungsgrenze)
	}
}

// **Ein Name, den es nicht gibt, ist ein Befund und kein Fehler.**
func TestEinUnbekannterName(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{antworten: map[string]scryfall.Treffer{}}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Gibtesnichtxyz")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if deutung.Zeilen[0].Befund == "" {
		t.Error("kein Befund zu einem Namen, den es nicht gibt")
	}
}

// **Ein Ausfall hält die Liste nicht auf** (§8).
func TestEinAusfallHaeltDieListeNichtAuf(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{
		antworten: map[string]scryfall.Treffer{
			`!"Lightning Bolt"`: {
				Karten: []*scryfall.Karte{druck("abc", "Lightning Bolt", "m11", "149")},
			},
		},
		fehler: map[string]error{
			`!"Counterspell"`: errors.New("kein Netz"),
		},
	}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Counterspell\n4 Lightning Bolt")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if len(deutung.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet 2", len(deutung.Zeilen))
	}

	if !strings.Contains(deutung.Zeilen[0].Befund, "kein Netz") {
		t.Errorf("der Befund der ersten Zeile lautet %q", deutung.Zeilen[0].Befund)
	}

	if !deutung.Zeilen[1].Eindeutig() {
		t.Error("die zweite Zeile ist nicht durchgekommen")
	}
}

// **Dieselbe Karte fragt einmal.**
//
// Eine Deckliste nennt dieselbe Karte oft mehrfach. Ohne Gedächtnis ginge je
// Zeile eine Anfrage hinaus, und die Drossel aus §16 machte daraus Wartezeit.
func TestGleicheZeilenFragenEinmal(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Lightning Bolt"`: {
			Karten: []*scryfall.Karte{druck("abc", "Lightning Bolt", "m11", "149")},
		},
	}}

	if _, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Lightning Bolt\n2 Lightning Bolt\n1 Lightning Bolt"); fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if quelle.gefragt[`!"Lightning Bolt"`] != 1 {
		t.Errorf("%d Anfragen für dieselbe Karte, erwartet 1",
			quelle.gefragt[`!"Lightning Bolt"`])
	}
}

// **Eine leere Liste ist ein fachliches Nein.**
func TestEineLeereListe(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeueListeneinfuhr(&listenquelle{}, &sammlungsdoppel{}).
		Deuten(t.Context(), "   \n  ")

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
	}
}

// **Das Alter der ältesten Antwort steht in der Deutung** (§16, Regel 5).
//
// Wer ohne Netz deutet, deutet aus dem Zwischenspeicher — und dann gehört
// dazu, wie alt die Auskunft ist.
func TestDasAlterDerAeltestenAntwort(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{
		alter: 3 * time.Hour,
		antworten: map[string]scryfall.Treffer{
			`!"Lightning Bolt"`: {
				Karten: []*scryfall.Karte{druck("abc", "Lightning Bolt", "m11", "149")},
			},
		},
	}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Lightning Bolt")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if deutung.Alter != 3*time.Hour {
		t.Errorf("Alter %v, erwartet 3h", deutung.Alter)
	}
}

// **Wenn Scryfall mehr hat, als eine Seite fasst, steht es da.**
//
// `has_more` heißt: die Zahl daneben ist eine Untergrenze. Sie ohne diesen
// Hinweis zu zeigen hieße, eine Auswahl als vollständig auszugeben, die es
// nicht ist (§10).
func TestMehrAlsEineSeite(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Lightning Bolt"`: {
			Karten:  []*scryfall.Karte{druck("abc", "Lightning Bolt", "m11", "149")},
			Weitere: true,
		},
	}}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Lightning Bolt")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if !strings.Contains(deutung.Zeilen[0].Befund, "mehr als") {
		t.Errorf("kein Hinweis auf weitere Seiten: %q", deutung.Zeilen[0].Befund)
	}
}

// **Die Befunde der Quelle gehen mit** — Befund 10 des Zweitblicks über
// Phase 4.5.
//
// `internal/scryfall` meldet einen unbekannten Aufzählungswert bewusst als
// **Befund** und liefert die Karte trotzdem (§16: Pflichtfelder streng,
// Zusatzfelder tolerant). `zeileDeuten` verwarf ihn mit `_` — und damit wurde
// gerade die zugesagte tolerante Behandlung fremder Daten **still**.
func TestDieBefundeDerQuelleGehenMit(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{
		antworten: map[string]scryfall.Treffer{
			`!"Lightning Bolt"`: {
				Karten: []*scryfall.Karte{druck("abc", "Lightning Bolt", "m11", "149")},
			},
		},
		befunde: []domain.Befund{{
			Ort: "rarity", Meldung: "unbekannter Wert »wunderbar«",
			Art: domain.ArtUmgebung,
		}},
	}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Lightning Bolt")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if !strings.Contains(deutung.Zeilen[0].Befund, "wunderbar") {
		t.Errorf("der Befund der Quelle fehlt: %q", deutung.Zeilen[0].Befund)
	}

	// **Und die Zeile bleibt eindeutig** — ein Befund ist kein Fehlschlag.
	if !deutung.Zeilen[0].Eindeutig() {
		t.Error("ein Befund hat die Zeile unbrauchbar gemacht")
	}
}

// **Und sie hängen sich an einen vorhandenen Satz, statt ihn zu ersetzen.**
func TestDieBefundeDerQuelleErsetzenNichts(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{
		antworten: map[string]scryfall.Treffer{},
		befunde: []domain.Befund{{
			Meldung: "unbekannter Wert", Art: domain.ArtUmgebung,
		}},
	}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).
		Deuten(t.Context(), "1 Gibtesnichtxyz")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	befund := deutung.Zeilen[0].Befund

	if !strings.Contains(befund, "kein Druck heißt so") ||
		!strings.Contains(befund, "unbekannter Wert") {
		t.Errorf("der Befund lautet %q — erwartet werden beide Sätze", befund)
	}
}
