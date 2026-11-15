package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// anlageprobe erfüllt `mcp.Fremdanlage`.
type anlageprobe struct {
	setID  int64
	name   string
	fehler error
}

func (a *anlageprobe) Anlegen(_ context.Context, setID int64, name string) (mcp.Anlage, error) {
	a.setID, a.name = setID, name

	if a.fehler != nil {
		return mcp.Anlage{}, a.fehler
	}

	return mcp.Anlage{
		KarteID: 7, Felder: []string{"name", "rule text"},
		Befunde: []string{}, Herkunft: "https://scryfall.com/card/lea/161",
	}, nil
}

// rufeAnlage sucht das Werkzeug und ruft es.
func rufeAnlage(t *testing.T, anlage mcp.Fremdanlage, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      anlage,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "schreiben.scryfall.karte" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("es gibt kein Werkzeug schreiben.scryfall.karte")

	return "", nil
}

// **Das Werkzeug legt an und meldet die Kennung** —
// #1032.
//
// **Die Kennung ist die Zusage** (0044): zurückgenommen wird über
// `schreiben.karte.loeschen`, und ohne sie ginge das nicht.
func TestDasKartenwerkzeugLegtAn(t *testing.T) {
	probe := &anlageprobe{}

	antwort, fehler := rufeAnlage(t, probe, `{"setID": 3, "name": "Lightning Bolt"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.setID != 3 || probe.name != "Lightning Bolt" {
		t.Errorf("die Naht bekam %d/%q", probe.setID, probe.name)
	}

	if !strings.Contains(antwort, `"KarteID": 7`) &&
		!strings.Contains(antwort, `"KarteID":7`) {
		t.Errorf("Antwort = %s — die Kennung fehlt, und ohne sie ist der Aufruf "+
			"nicht umkehrbar", antwort)
	}
}

// **Eine krumme Setkennung erreicht die Naht nicht.**
func TestDasKartenwerkzeugWeistEineKrummeKennungAb(t *testing.T) {
	probe := &anlageprobe{}

	if _, fehler := rufeAnlage(t, probe, `{"setID": 0, "name": "x"}`); fehler == nil {
		t.Fatal("kein Fehler bei der Kennung 0")
	}

	if probe.name != "" {
		t.Error("die Naht wurde gerufen, obwohl die Kennung krumm ist")
	}
}

// **Ein unleserliches Argument wird abgewiesen.**
func TestDasKartenwerkzeugWeistUnleserlicheArgumenteAb(t *testing.T) {
	if _, fehler := rufeAnlage(t, &anlageprobe{}, `{"setID": "drei"}`); fehler == nil {
		t.Fatal("kein Fehler bei einem unleserlichen Argument")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestDasKartenwerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	probe := &anlageprobe{fehler: errors.New("Scryfall antwortet nicht")}

	if _, fehler := rufeAnlage(t, probe, `{"setID": 3, "name": "x"}`); fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// **Auch im Fehlerfall nennt das Werkzeug die angelegte Karte** —
// #1041.
//
// Seine eigene Erklärung sagt: „Gibt die Kennung der angelegten Karte zurück;
// zurückgenommen wird über »schreiben.karte.loeschen«." **Genau im Fehlerfall
// kam sie nicht** — der halbe Stand war über diese Fläche weder auffindbar
// noch gezielt zurücknehmbar, und ein Sprachmodell ruft, während niemand
// hinsieht (0010).
func TestDasKartenwerkzeugNenntDieKarteAuchImFehlerfall(t *testing.T) {
	probe := &halbeanlage{}

	_, fehler := rufeAnlage(t, probe, `{"setID": 3, "name": "Prüfstein"}`)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), "7") {
		t.Errorf("die Meldung nennt die Kartenkennung nicht: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "name") {
		t.Errorf("die Meldung nennt die geschriebenen Felder nicht: %v", fehler)
	}
}

// **Und wenn kein Feld durchkam, sagt sie auch das.**
//
// „gefüllt: " ohne Namen wäre eine Aufzählung ohne Glieder — und der Aufrufer
// müsste raten, ob sie leer ist oder abgeschnitten.
func TestEineKarteOhneGefuelltesFeldSagtEsAuch(t *testing.T) {
	_, fehler := rufeAnlage(t, &leereanlage{}, `{"setID": 3, "name": "x"}`)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), "kein Feld gefüllt") {
		t.Errorf("Meldung = %v", fehler)
	}
}

// halbeanlage legt an und scheitert danach.
type halbeanlage struct{}

func (halbeanlage) Anlegen(context.Context, int64, string) (mcp.Anlage, error) {
	return mcp.Anlage{KarteID: 7, Felder: []string{"name"}, Befunde: []string{}},
		errors.New("das Feld nimmt den Wert nicht an")
}

// leereanlage legt an, ohne ein Feld zu füllen.
type leereanlage struct{}

func (leereanlage) Anlegen(context.Context, int64, string) (mcp.Anlage, error) {
	return mcp.Anlage{KarteID: 7, Felder: []string{}, Befunde: []string{}},
		errors.New("der Bestand ist nicht erreichbar")
}

// **Ohne angelegte Karte bleibt der Fehler, wie er ist.**
//
// Ein Aufruf, der **vor** der Anlage scheitert, hat keinen halben Stand — und
// eine Kennung 0 daranzuhängen wäre eine Auskunft über nichts.
func TestOhneAngelegteKarteBleibtDerFehlerSchlicht(t *testing.T) {
	probe := &anlageprobe{fehler: errors.New("Scryfall antwortet nicht")}

	_, fehler := rufeAnlage(t, probe, `{"setID": 3, "name": "x"}`)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if strings.Contains(fehler.Error(), "zurückgenommen wird sie über") {
		t.Errorf("die Meldung erfindet einen halben Stand: %v", fehler)
	}
}
