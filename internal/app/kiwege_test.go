package app_test

// **Die Fehlerwege der KI-Naht** — Werkstück A7 von Phase 5.3
// (#1446).
//
// **Damit jeder betreten worden ist** (§10): ein Fehlerweg, den nie jemand
// gegangen ist, sieht aus, als fange er etwas ab.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Bestand gibt es keinen Katalog — und die Meldung sagt es.**
func TestOhneBestandGibtEsKeinenKIKatalog(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KIModelle(probekontext(t))

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Katalog")
	}
}

// **Ein Lesefehler der Einstellungen kommt durch.**
//
// **Und wird nicht zu einem Stand**: dass die Einstellungen unlesbar sind,
// ist ein Fehler dieser Seite — anders als eine Laufzeit, die schweigt (§8).
func TestEinLesefehlerDerEinstellungenKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&stolperndeEinstellungen{})

	if antwort := anwendung.KIModelle(probekontext(t)); antwort.Fehler == nil {
		t.Fatal("der Lesefehler kam nicht durch")
	}
}

// stolperndeEinstellungen scheitern beim Lesen.
type stolperndeEinstellungen struct{}

func (stolperndeEinstellungen) Stand(context.Context) (
	map[string]service.Wahl, error,
) {
	return nil, errors.New("die Einstellungen sind unlesbar")
}

func (stolperndeEinstellungen) Alle(context.Context) (map[string]string, error) {
	return nil, errors.New("die Einstellungen sind unlesbar")
}

func (stolperndeEinstellungen) Setzen(context.Context, string, string) error {
	return errors.New("die Einstellungen sind unlesbar")
}

// **Eine Laufzeit, die ihre Modelle nicht hergibt, ist unerreichbar.**
//
// **Nicht dasselbe wie »sie antwortet nicht«**: die Fassung kam, die Liste
// nicht — und der Befund nennt beides auseinander.
func TestEineLaufzeitOhneModellisteIstUnerreichbar(t *testing.T) {
	t.Parallel()

	laufzeit := &laufzeitdoppel{fassung: "0.12.3"}
	laufzeit.listenfehler = errors.New("die Liste ist unlesbar")

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, laufzeit).KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Einrichtung != app.KIEinrichtungUnerreichbar {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}

	befunde := befundeVon(t, antwort, "ollama")

	if len(befunde) == 0 ||
		!strings.Contains(befunde[len(befunde)-1], "unlesbar") {
		t.Errorf("die Befunde sind %v", befunde)
	}
}
