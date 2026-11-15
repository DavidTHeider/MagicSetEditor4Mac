package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type importprobe struct {
	ergebnis *service.Ergebnis
	vorschau *service.Setvorschau
	fehler   error
	pfad     string
	spiel    string
}

func (i *importprobe) Set(_ context.Context, pfad string) (*service.Ergebnis, error) {
	i.pfad = pfad

	return i.ergebnis, i.fehler
}

// Neu legt ein leeres Set an (#891).
func (i *importprobe) Neu(_ context.Context, name, spiel, _ string) (int64, error) {
	i.pfad, i.spiel = name, spiel

	if i.fehler != nil {
		return 0, i.fehler
	}

	return 42, nil
}

func (i *importprobe) MTGJSON(_ context.Context, pfad, spiel string) (*service.Ergebnis, error) {
	i.pfad, i.spiel = pfad, spiel

	return i.ergebnis, i.fehler
}

func mitImporteur(probe *importprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitImporteur(probe)

	return anwendung
}

func TestSetOeffnenLegtDasSetAn(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 7, Karten: 312}}

	antwort := mitImporteur(probe).SetOeffnen(probekontext(t), "/tmp/probe.mse-set")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.pfad != "/tmp/probe.mse-set" {
		t.Errorf("geöffnet wurde %q", probe.pfad)
	}

	if antwort.SetID != "7" || antwort.Karten != 312 {
		t.Errorf("%+v", antwort)
	}
}

// **Die Vorbehalte reisen mit** (§8): ein Set mit 300 lesbaren und zwei
// unlesbaren Karten wird geöffnet, nicht abgewiesen — und die zwei stehen da.
func TestSetOeffnenNenntDieVorbehalte(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{
		SetID: 1, Karten: 300,
		Befunde: []domain.Befund{
			{Datei: "set", Ort: "12", Meldung: "unlesbare Karte"},
			{Datei: "PRB.json", Ort: "PRB → Karte 2 → uuid", Meldung: "unbekanntes Feld"},
		},
	}}

	antwort := mitImporteur(probe).SetOeffnen(probekontext(t), "/tmp/probe.mse-set")

	if len(antwort.Befunde) != 2 {
		t.Fatalf("%+v", antwort.Befunde)
	}

	// Mit Stelle, nicht nur mit Meldung — sonst sucht jemand in 3.000 Zeilen.
	if antwort.Befunde[0] != "set:12: unlesbare Karte" {
		t.Errorf("%q", antwort.Befunde[0])
	}

	// **Und die Stelle in der Ausdrucksweise ihres Formats.** MTG JSON hat keine
	// Zeilen, die etwas bedeuten; ein `:0:` davor wäre eine erfundene Angabe.
	if antwort.Befunde[1] != "PRB.json:PRB → Karte 2 → uuid: unbekanntes Feld" {
		t.Errorf("%q", antwort.Befunde[1])
	}
}

// **Was der Importeur über eine schon offene Datei sagt, reicht die Fassade
// durch.**
//
// # Warum hier nicht mehr die Regel geprüft wird
//
// „Eine Quelle, ein Set" stand bis zum 2026-08-23 in dieser Fassade und galt
// damit **nur** für das Fenster: die Kommandozeile ging am Importeur
// unmittelbar vorbei und legte dasselbe Set ein zweites Mal an
// (#886).
// Das war kein Versehen der Kommandozeile — eine fachliche Regel in einer
// Fassade ist eine Regel je Fassade.
//
// Belegt ist sie jetzt im Fachkern (`internal/service/schonoffen_test.go`).
// Hier bleibt die Frage der Fassade: **kommt die Auskunft an?**
func TestSetOeffnenReichtSchonOffenDurch(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 4, Karten: 12, SchonOffen: true}}

	antwort := mitImporteur(probe).SetOeffnen(probekontext(t), "/tmp/probe.mse-set")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if !antwort.SchonOffen {
		t.Error("das zweite Öffnen wurde nicht gemeldet")
	}

	if antwort.SetID != "4" || antwort.Karten != 12 {
		t.Errorf("%+v — erwartet war das bereits offene Set", antwort)
	}
}

// **Ein Fehler des Importeurs erreicht die Oberfläche.**
//
// Er entstand früher hier — beim Nachsehen, ob die Quelle schon offen ist.
// Jetzt entsteht er im Fachkern; geprüft wird, dass er nicht verlorengeht.
func TestSetOeffnenMeldetEinenFehlerDesImporteurs(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitImporteur(&importprobe{fehler: errors.New("nicht lesbar")})

	if anwendung.SetOeffnen(t.Context(), "/tmp/x.mse-set").Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Eine leere Vorbehaltsliste ist `[]`, nicht `null`.**
func TestEineLeereBefundlisteIstEineListe(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 1}}

	antwort := mitImporteur(probe).SetOeffnen(probekontext(t), "/tmp/x.mse-set")

	gebaut, err := json.Marshal(antwort.Befunde)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	if string(gebaut) != "[]" {
		t.Errorf("%s", gebaut)
	}
}

// **Der ganze Weg, einmal echt**: eine echte Paketdatei, eine echte Datenbank,
// die Anwendung dazwischen — und danach steht das Set in der Liste.
//
// # Warum das hier steht und nicht nur in `service`
//
// `service.Importeur` hat seine eigenen Prüfungen gegen echte Pakete. Was hier
// geprüft wird, ist die **Naht**: dass `App.SetOeffnen` denselben Bestand
// benutzt, den `Sets()` liest, dass die Kennung zusammenpasst und dass ein
// zweites Öffnen derselben Datei kein zweites Set anlegt.
//
// Was auch das nicht zeigt: den Dateidialog. Der braucht ein Fenster
// (`dateiwahl.go`, in `.covignore`); gesehen wurde er an der laufenden
// Anwendung.
func TestDerGanzeWegEinesOeffnens(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	pfad, err := filepath.Abs(filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	antwort := anwendung.SetOeffnen(t.Context(), pfad)
	if antwort.Fehler != nil {
		t.Fatalf("öffnen: %+v", antwort.Fehler)
	}

	if antwort.Karten != 3 {
		t.Errorf("%d Karten", antwort.Karten)
	}

	// **Der unverstandene Block reist als Rohblock mit** — und ist kein Befund.
	//
	// Der erste Anlauf dieses Tests erwartete ihn unter den Befunden. Er ist
	// keiner: der Import hat ihn nicht *gedeutet*, aber sehr wohl gelesen und
	// abgelegt. Ein Befund wäre eine Stelle, die er nicht lesen konnte.
	roh := anwendung.Vorbehalte(t.Context(), antwort.SetID)
	if roh.Fehler != nil {
		t.Fatalf("Rohblöcke: %+v", roh.Fehler)
	}

	if len(roh.Vorbehalte) == 0 {
		t.Error("kein Rohblock — die Datei führt einen unverstandenen Block")
	}

	// Und das Set steht in der Liste, die die Oberfläche liest.
	liste := anwendung.Sets(t.Context())
	if liste.Fehler != nil {
		t.Fatalf("Sets: %+v", liste.Fehler)
	}

	if len(liste.Sets) != 1 || liste.Sets[0].ID != antwort.SetID {
		t.Fatalf("%+v", liste.Sets)
	}

	if liste.Sets[0].Name != "Mein Probeset" || liste.Sets[0].Karten != 3 {
		t.Errorf("%+v", liste.Sets[0])
	}

	// Dasselbe noch einmal: kein zweites Set.
	zweites := anwendung.SetOeffnen(t.Context(), pfad)
	if !zweites.SchonOffen || zweites.SetID != antwort.SetID {
		t.Errorf("%+v", zweites)
	}

	if nachher := anwendung.Sets(t.Context()); len(nachher.Sets) != 1 {
		t.Errorf("%d Sets nach dem zweiten Öffnen", len(nachher.Sets))
	}
}

// **Ein Archiv, das sich nicht schließen lässt, nimmt das Set nicht mit** —
// Befund 1 des Nachlaufs über Phase 4.4
// (#1291).
//
// `Importeur.Set` hängt den Schließfehler des Archivs mit `errors.Join` an ein
// **geglücktes** Einfahren an. Das Set steht dann im Bestand; wer nur den
// Fehler zeigte, ließe es zurück, ohne dass jemand seine Kennung erführe.
func TestEinArchivschliessfehlerNenntDasEingefahreneSet(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitImporteur(halberImporteur{})

	antwort := anwendung.SetOeffnen(t.Context(), "/tmp/halb.mse-set")

	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.SetID != "7" {
		t.Errorf("die Kennung ging verloren: %q", antwort.SetID)
	}
}

// halberImporteur fährt ein Set ein und scheitert beim Schließen.
type halberImporteur struct{}

func (halberImporteur) Set(context.Context, string) (*service.Ergebnis, error) {
	return &service.Ergebnis{SetID: 7, Karten: 3},
		errors.New("das Archiv schließen: die Datei ist weg")
}

func (halberImporteur) MTGJSON(context.Context, string, string) (*service.Ergebnis, error) {
	return nil, nil //nolint:nilnil // dieser Weg wird hier nicht gefahren
}

func (halberImporteur) Neu(context.Context, string, string, string) (int64, error) {
	return 0, nil
}

func (halberImporteur) Vorschau(context.Context, string) (*service.Setvorschau, error) {
	return nil, nil //nolint:nilnil // dieser Weg wird hier nicht gefahren
}
