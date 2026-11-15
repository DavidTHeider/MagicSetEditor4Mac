package domain_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorlage ist ein kleiner Ausschnitt des deutschen Katalogs.
var vorlage = domain.Katalogvorlage{
	"allgemein.titel":                     "Sets",
	"fehler.symbol":                       "Die Datei ist größer als {grenze}.",
	"seiten.decks.umschlag":               "Name und Spiel **gehen mit**, die Notiz `bleibt`.",
	"allgemein.zaehlwoerter.karten.eine":  "Karte",
	"allgemein.zaehlwoerter.karten.viele": "Karten",
}

// **Eine Übersetzung wird gegen ihre Vorlage geprüft** — LOC-02.AK-02
// (#1688):
// Platzhalter, Zählformen, sichere Auszeichnung. Ein Sprachpaket ist Text;
// was wie Code aussieht, kommt nicht hinein.
func TestEineUebersetzungGegenIhreVorlage(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		pfad, text string
		schluessel domain.Schluessel
	}{
		{name: "ein Wort", pfad: "allgemein.titel", text: "Conjuntos"},
		{name: "mit Platzhalter", pfad: "fehler.symbol", text: "El archivo supera {grenze}."},
		{name: "Platzhalter zweimal", pfad: "fehler.symbol", text: "{grenze}: más de {grenze}."},
		{name: "mit Auszeichnung", pfad: "seiten.decks.umschlag", text: "**viajan** y `se queda`."},
		{name: "mehrzeilig", pfad: "allgemein.titel", text: "Con\njuntos"},
		{name: "die dritte Zählform", pfad: "allgemein.zaehlwoerter.karten.wenige", text: "карты"},
		{
			name: "unbekannter Pfad", pfad: "allgemein.gibtsnicht", text: "x",
			schluessel: domain.SchluesselValidierung,
		},
		{
			name: "wenige ohne Zählform", pfad: "allgemein.wenige", text: "x",
			schluessel: domain.SchluesselValidierung,
		},
		{name: "leer", pfad: "allgemein.titel", text: "  ", schluessel: domain.SchluesselValidierung},
		{
			name: "zu lang", pfad: "allgemein.titel", text: strings.Repeat("a", 4001),
			schluessel: domain.SchluesselValidierung,
		},
		{
			name: "Platzhalter fehlt", pfad: "fehler.symbol", text: "El archivo es grande.",
			schluessel: domain.SchluesselPlatzhalterAbweichend,
		},
		{
			name: "Platzhalter erfunden", pfad: "allgemein.titel", text: "{name}",
			schluessel: domain.SchluesselPlatzhalterAbweichend,
		},
		{
			name: "HTML", pfad: "allgemein.titel", text: "<b>Sets</b>",
			schluessel: domain.SchluesselAuszeichnungUnsicher,
		},
		{
			name: "Skriptadresse", pfad: "allgemein.titel", text: "JavaScript:alert(1)",
			schluessel: domain.SchluesselAuszeichnungUnsicher,
		},
		{
			name: "offener Fettdruck", pfad: "seiten.decks.umschlag", text: "**viajan y `se queda`.",
			schluessel: domain.SchluesselAuszeichnungUnsicher,
		},
		{
			name: "offener Code", pfad: "seiten.decks.umschlag", text: "**viajan** y `se queda.",
			schluessel: domain.SchluesselAuszeichnungUnsicher,
		},
		{
			name: "Steuerzeichen", pfad: "allgemein.titel", text: "Sets\x1b[31m",
			schluessel: domain.SchluesselAuszeichnungUnsicher,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := vorlage.PruefeUebersetzung(fall.pfad, fall.text)
			if fall.schluessel == "" {
				if fehler != nil {
					t.Fatalf("abgewiesen: %v", fehler)
				}

				return
			}

			var feld *domain.FieldError
			if !errors.As(fehler, &feld) || feld.Key != fall.schluessel {
				t.Fatalf("%v statt %s", fehler, fall.schluessel)
			}

			// **Der Befund steht am Pfad** — dort zeigt ihn der Assistent.
			if feld.Field != fall.pfad || feld.Werte["pfad"] != fall.pfad {
				t.Errorf("am Feld %q mit %v", feld.Field, feld.Werte)
			}
		})
	}
}

// **Was fehlt, was veraltet ist und was die Vorlage nicht kennt** —
// LOC-02.AK-01 und AK-03. Veraltet heißt: der deutsche Satz ist ein anderer
// als der, gegen den übersetzt wurde.
func TestDerAbgleichEinesPakets(t *testing.T) {
	t.Parallel()

	paket := domain.Sprachpaket{Texte: map[string]domain.Uebersetzungstext{
		"allgemein.titel":                      {Text: "Conjuntos", Quelle: "Sets"},
		"fehler.symbol":                        {Text: "Mayor que {grenze}.", Quelle: "Die Datei ist zu groß: {grenze}."},
		"allgemein.zaehlwoerter.karten.wenige": {Text: "карты", Quelle: "Karte"},
		"allgemein.weg":                        {Text: "fort", Quelle: "weg"},
	}}

	abgleich := vorlage.Abgleich(paket)

	if abgleich.Gesamt != 5 || abgleich.Uebersetzt != 1 {
		t.Errorf("%d von %d übersetzt", abgleich.Uebersetzt, abgleich.Gesamt)
	}

	pruefe := func(name string, ist, soll []string) {
		t.Helper()

		if !slices.Equal(ist, soll) {
			t.Errorf("%s: %v statt %v", name, ist, soll)
		}
	}

	pruefe("fehlend", abgleich.Fehlend, []string{
		"allgemein.zaehlwoerter.karten.eine", "allgemein.zaehlwoerter.karten.viele",
		"seiten.decks.umschlag",
	})
	pruefe("veraltet", abgleich.Veraltet, []string{
		"allgemein.zaehlwoerter.karten.wenige", "fehler.symbol",
	})
	pruefe("unbekannt", abgleich.Unbekannt, []string{"allgemein.weg"})
}

// **Die Quelle einer Übersetzung** ist der deutsche Satz — bei der dritten
// Zählform der von `viele`.
func TestDieQuelleEinerUebersetzung(t *testing.T) {
	t.Parallel()

	for pfad, quelle := range map[string]string{
		"allgemein.titel":                      "Sets",
		"allgemein.zaehlwoerter.karten.wenige": "Karten",
		"allgemein.gibtsnicht":                 "",
	} {
		if got := vorlage.Quelle(pfad); got != quelle {
			t.Errorf("%s: %q statt %q", pfad, got, quelle)
		}
	}
}

// **Ein Satz unter einem unbekannten Pfad** wird nur daraufhin geprüft, ob
// er Text ist.
func TestEineFremdeUebersetzung(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeFremdeUebersetzung("neu.satz", "Novo **ja**"); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}

	var feld *domain.FieldError
	if fehler := domain.PruefeFremdeUebersetzung("neu.satz", "<script>"); !errors.As(fehler, &feld) ||
		feld.Key != domain.SchluesselAuszeichnungUnsicher || feld.Field != "neu.satz" {
		t.Errorf("durchgelassen: %v", fehler)
	}
}
