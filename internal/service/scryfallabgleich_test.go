package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Nichts wird still übernommen** (§16, Regel 6)
// (#1005).
//
// > „Nichts wird still übernommen. Was von außen in ein Kartenfeld will, zeigt
// > vorher, was sich ändert, und trägt danach seine Herkunft."
//
// # Warum das eine eigene Zusage ist und keine Frage der Oberfläche
//
// Weil sie an **allen vier** Flächen gelten soll. Stünde die Rückfrage nur im
// Fenster, übernähme MCP still — und ein Sprachmodell überschriebe eine Karte,
// während niemand hinsieht. Der Vorschlag entsteht deshalb im Fachkern, und
// erst ein zweiter Aufruf setzt ihn um.
func TestEinAbgleichZeigtVorherWasSichAendert(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), nil)

	vorschlag, fehler := abgleich.Vorschlag(t.Context(), "magic",
		map[string]string{
			"name":         "Prüfstein",
			"casting cost": "1W",
			"rule text":    "",
		}, "Prüfstein")
	if fehler != nil {
		t.Fatalf("Vorschlag: %v", fehler)
	}

	// **Jede Änderung nennt beide Seiten.** Ein Vorschlag, der nur den neuen
	// Wert zeigte, ließe niemanden entscheiden.
	nach := map[string]service.Feldvorschlag{}
	for _, eine := range vorschlag.Aenderungen {
		nach[eine.Feld] = eine
	}

	if ist := nach["casting cost"]; ist.Alt != "1W" || ist.Neu != "2W" {
		t.Errorf("casting cost: alt %q, neu %q", ist.Alt, ist.Neu)
	}

	if ist := nach["rule text"]; ist.Alt != "" || ist.Neu == "" {
		t.Errorf("rule text: alt %q, neu %q", ist.Alt, ist.Neu)
	}

	// **Ein Feld, das gleich bleibt, ist keine Änderung.**
	if _, dabei := nach["name"]; dabei {
		t.Error("ein unverändertes Feld steht als Änderung im Vorschlag")
	}

	// **Und der Vorschlag sagt, woher er kommt** (§16, Regel 6: „trägt danach
	// seine Herkunft").
	if vorschlag.Herkunft != "https://beispiel.invalid/karte/1" {
		t.Errorf("Herkunft %q", vorschlag.Herkunft)
	}
}

// **Und ein Vorschlag schreibt nichts.**
//
// Die Gegenprobe zur Zusage oben: ohne sie wäre auch ein Abgleich grün, der
// den Vorschlag berechnet **und** nebenbei schreibt.
func TestEinVorschlagSchreibtNichts(t *testing.T) {
	t.Parallel()

	schreiber := &mitschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	if _, fehler := abgleich.Vorschlag(t.Context(), "magic",
		map[string]string{"name": "Alt"}, "Prüfstein"); fehler != nil {
		t.Fatalf("Vorschlag: %v", fehler)
	}

	if len(schreiber.geschrieben) != 0 {
		t.Errorf("der Vorschlag hat geschrieben: %v", schreiber.geschrieben)
	}
}

// **Übernommen wird nur, was ausdrücklich genannt ist.**
//
// # Warum nicht „alles aus dem Vorschlag"
//
// Weil der Vorschlag zeigt, **was möglich wäre** — und die Entscheidung dem
// gehört, der ihn gesehen hat. Wer den Aromatext behalten und nur den
// Regeltext übernehmen will, soll das können, ohne den Vorschlag zu
// beschneiden.
func TestUebernommenWirdNurWasGenanntIst(t *testing.T) {
	t.Parallel()

	schreiber := &mitschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	vorschlag := mussVorschlag(t, abgleich, map[string]string{"name": "Alt", "rule text": "Alt"}, "Prüfstein")

	if _, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, vorschlag,
		[]string{"rule text"}); fehler != nil {
		t.Fatalf("Übernehmen: %v", fehler)
	}

	if len(schreiber.geschrieben) != 1 {
		t.Fatalf("%d Felder geschrieben, erwartet war eines: %v",
			len(schreiber.geschrieben), schreiber.geschrieben)
	}

	if schreiber.geschrieben[0].feld != "rule text" {
		t.Errorf("geschrieben wurde %q", schreiber.geschrieben[0].feld)
	}
}

// **Ein Feld, das nicht im Vorschlag steht, wird nicht geschrieben.**
//
// # Warum das die eigentliche Wache ist
//
// Ohne sie könnte ein Aufrufer über die Feldliste **jeden** Wert setzen — und
// der Vorschlag wäre eine Höflichkeit statt einer Grenze. Wer etwas anderes
// schreiben will, benutzt den gewöhnlichen Weg; dieser hier trägt die Herkunft
// „von Scryfall", und die soll stimmen.
func TestEinFeldAusserhalbDesVorschlagsWirdNichtGeschrieben(t *testing.T) {
	t.Parallel()

	schreiber := &mitschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	vorschlag := mussVorschlag(t, abgleich, map[string]string{"name": "Prüfstein"}, "Prüfstein")

	// **`power` steht nicht im Vorschlag**, weil die Probekarte keine Stärke
	// hat und leere Werte nicht vorgeschlagen werden.
	//
	// Der erste Anlauf nahm `illustrator` — das **steht** im Vorschlag, und
	// der Test scheiterte zu Recht. Ein Fall, der die Wache prüfen soll, muss
	// ein Feld nennen, das wirklich fehlt.
	_, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, vorschlag, []string{"power"})
	if fehler == nil {
		t.Fatal("ein Feld außerhalb des Vorschlags wurde übernommen")
	}

	if len(schreiber.geschrieben) != 0 {
		t.Errorf("es wurde trotzdem geschrieben: %v", schreiber.geschrieben)
	}
}

// **Ein unbekanntes Spiel bekommt keinen Vorschlag.**
func TestEinUnbekanntesSpielBekommtKeinenVorschlag(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), nil)

	if _, fehler := abgleich.Vorschlag(t.Context(), "gibtsnicht",
		map[string]string{}, "Prüfstein"); fehler == nil {
		t.Error("für ein unbekanntes Spiel kam ein Vorschlag")
	}
}

// **Die Befunde der Antwort reisen mit dem Vorschlag.**
//
// Ein unbekannter Aufzählungswert soll den sehen, der übernimmt — nicht nur
// den, der die Antwort gelesen hat.
func TestDieBefundeReisenMitDemVorschlag(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(&fremdmitbefund{}, nil, neuesVerweisdoppel(), nil)

	vorschlag, fehler := abgleich.Vorschlag(t.Context(), "magic",
		map[string]string{}, "Seltsam")
	if fehler != nil {
		t.Fatalf("Vorschlag: %v", fehler)
	}

	if len(vorschlag.Befunde) != 1 {
		t.Errorf("%d Befunde, erwartet war einer", len(vorschlag.Befunde))
	}
}

// fremdkarte gibt eine Fremdquelle, die immer dieselbe Karte liefert.
func fremdkarte() service.Fremdkarten {
	return &fremdfest{karte: &scryfall.Karte{
		Name:       "Prüfstein",
		Manakosten: "{2}{W}",
		Typzeile:   "Kreatur — Prüfling",
		Regeltext:  "Fliegend.",
		Seltenheit: "rare",
		Kuenstler:  "Selbst Gemalt",
		Adresse:    "https://beispiel.invalid/karte/1",
	}}
}

type fremdfest struct{ karte *scryfall.Karte }

func (f *fremdfest) KarteNachName(
	context.Context, string,
) (karte *scryfall.Karte, alter time.Duration, befunde []domain.Befund, fehler error) {
	return f.karte, 0, nil, nil
}

type fremdmitbefund struct{}

func (fremdmitbefund) KarteNachName(
	context.Context, string,
) (karte *scryfall.Karte, alter time.Duration, befunde []domain.Befund, fehler error) {
	return &scryfall.Karte{Name: "Seltsam"}, 3 * time.Hour,
		[]domain.Befund{{Meldung: "unbekannter Wert"}}, nil
}

// mitschreiber merkt sich, was geschrieben wurde.
type mitschreiber struct {
	geschrieben []struct{ feld, wert, herkunft string }
}

func (m *mitschreiber) FelderVonAussen(
	_ context.Context, _, _ int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	for _, wert := range werte {
		m.geschrieben = append(m.geschrieben,
			struct{ feld, wert, herkunft string }{wert.Field, wert.Text, herkunft})
	}

	return "", false, nil
}

// **Ein Fehler beim Fragen kommt durch, und es wird nichts geschrieben.**
func TestEinFehlerBeimFragenSchreibtNichts(t *testing.T) {
	t.Parallel()

	schreiber := &mitschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdscheitert{}, nil, neuesVerweisdoppel(), schreiber)

	if _, fehler := abgleich.Vorschlag(t.Context(), "magic",
		map[string]string{}, "Prüfstein"); fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}

	if len(schreiber.geschrieben) != 0 {
		t.Errorf("es wurde geschrieben: %v", schreiber.geschrieben)
	}
}

// **Ohne Vorschlag wird nichts übernommen.**
//
// Der Fall, den ein Aufrufer herstellt, der die Reihenfolge vertauscht.
func TestOhneVorschlagWirdNichtsUebernommen(t *testing.T) {
	t.Parallel()

	schreiber := &mitschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	if _, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, nil, []string{"name"}); fehler == nil {
		t.Error("ohne Vorschlag wurde übernommen")
	}
}

// **Ein Schreibfehler kommt durch.**
func TestEinSchreibfehlerKommtDurch(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiberscheitert{&mitschreiber{}})

	vorschlag := mussVorschlag(t, abgleich, map[string]string{"rule text": "Alt"}, "Prüfstein")

	if _, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, vorschlag,
		[]string{"rule text"}); fehler == nil {
		t.Error("der Schreibfehler kam nicht durch")
	}
}

// **Ohne Adresse steht die Kennung in der Herkunft.**
//
// Beides zeigt auf dieselbe Karte; die Adresse lässt sich anklicken, die
// Kennung nicht — aber eine Herkunft ohne beides wäre keine.
func TestOhneAdresseStehtDieKennungInDerHerkunft(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(
		&fremdfest{karte: &scryfall.Karte{Name: "Ohne", Kennung: "abc-123"}},
		nil, neuesVerweisdoppel(), nil)

	vorschlag, fehler := abgleich.Vorschlag(t.Context(), "magic",
		map[string]string{}, "Ohne")
	if fehler != nil {
		t.Fatalf("Vorschlag: %v", fehler)
	}

	if vorschlag.Herkunft != "Scryfall abc-123" {
		t.Errorf("Herkunft %q", vorschlag.Herkunft)
	}
}

// mussVorschlag holt einen Vorschlag und scheitert, wenn es nicht geht.
//
// **Immer für `magic`** — das einzige Spiel mit einer Zuordnung. Wer ein
// zweites aufnimmt, gibt es hier mit.
func mussVorschlag(
	t *testing.T, abgleich *service.Scryfallabgleich,
	jetzige map[string]string, name string,
) *service.Vorschlag {
	t.Helper()

	vorschlag, fehler := abgleich.Vorschlag(t.Context(), "magic", jetzige, name)
	if fehler != nil {
		t.Fatalf("Vorschlag: %v", fehler)
	}

	return vorschlag
}

type fremdscheitert struct{}

func (fremdscheitert) KarteNachName(
	context.Context, string,
) (karte *scryfall.Karte, alter time.Duration, befunde []domain.Befund, fehler error) {
	return nil, 0, nil, domain.ErrNotFound
}

type schreiberscheitert struct{ *mitschreiber }

func (schreiberscheitert) FelderVonAussen(
	context.Context, int64, int64, []domain.Value, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, domain.ErrConflict
}

// **Das Alter der Antwort reist mit dem Vorschlag** (§16, Regel 5).
//
// > „Der letzte gute Stand bleibt, und er sagt, wie alt er ist."
//
// # Warum das im Vorschlag steht und nicht nur im Klienten
//
// Weil der es sehen soll, der **übernimmt**. Ein Vorschlag aus einer drei Tage
// alten Antwort sieht sonst aus wie einer von eben, und wer ihn annimmt, weiß
// nicht, dass er einem Zwischenspeicher folgt.
func TestDasAlterDerAntwortReistMitDemVorschlag(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(&fremdmitbefund{}, nil, neuesVerweisdoppel(), nil)

	vorschlag := mussVorschlag(t, abgleich, map[string]string{}, "Seltsam")

	if vorschlag.Alter != 3*time.Hour {
		t.Errorf("Alter %v, erwartet waren 3h", vorschlag.Alter)
	}
}

// **Die Übernahme gibt zurück, was sie geändert hat — samt vorigem Wert.**
//
// # Warum das zur Zusage gehört
//
// Für MCP ist es Pflicht (0010): ein schreibendes Werkzeug sagt, wie es
// zurückzunehmen ist. Ohne diese Rückgabe müsste ein Gegenüber den Vorschlag
// aufgehoben haben — und ein Modell, das dazwischen den Faden verliert, hätte
// keinen Rückweg.
func TestDieUebernahmeGibtZurueckWasSieGeaendertHat(t *testing.T) {
	t.Parallel()

	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), &mitschreiber{})

	vorschlag := mussVorschlag(t, abgleich,
		map[string]string{"rule text": "Alt", "flavor text": "Auch alt"}, "Prüfstein")

	uebernommen, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, vorschlag,
		[]string{"rule text"})
	if fehler != nil {
		t.Fatalf("Übernehmen: %v", fehler)
	}

	if len(uebernommen) != 1 {
		t.Fatalf("%d zurückgegeben, erwartet war eines", len(uebernommen))
	}

	if uebernommen[0].Feld != "rule text" || uebernommen[0].Alt != "Alt" {
		t.Errorf("zurückgegeben wurde %+v", uebernommen[0])
	}
}
