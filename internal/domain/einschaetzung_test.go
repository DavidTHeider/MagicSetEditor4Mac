package domain_test

// **Was eine KI-Einschätzung sein darf** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Die Gegenproben
//
// Beide Zusagen sind einzeln entschärft worden, und beide Male wurde der Lauf
// rot (§10):
//
//   - [domain.PruefeBelege] übergangen — die erfundene Regelnummer kommt
//     durch, hier **und** eine Ebene höher in `internal/app`.
//   - [domain.NenntEinenBeleg] übergangen — eine Antwort ohne jede Fußnote
//     gilt als belegt.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiRegelbelege gibt zwei Belege, wie ein Kontextpaket sie vergibt.
func zweiRegelbelege() []domain.Beleg {
	return []domain.Beleg{
		{
			Kennung: "B1", Quelle: domain.QuelleRegelwerk,
			Titel:       "509.1 Angreifer erklären",
			Ausschnitt:  "509.1. Zuerst prüft das Spiel, ob die angegebenen Blocks legal sind.",
			Regelnummer: "509.1",
		},
		{
			Kennung: "B2", Quelle: domain.QuelleRuling,
			Titel:      "Ruling vom 2024-02-09",
			Ausschnitt: "Die Fähigkeit triggert nur einmal je Runde.",
			Urheber:    domain.UrheberWotc,
		},
	}
}

// **Eine Einschätzung, die nur vergebene Belege nennt, ist belegt.**
func TestEineEinschaetzungMitEchtenBelegenIstBelegt(t *testing.T) {
	t.Parallel()

	einschaetzung := domain.PruefeEinschaetzung(
		domain.EinschaetzungRegel,
		"Der Block ist legal [B1]; die Fähigkeit triggert einmal [B2].",
		zweiRegelbelege())

	if !einschaetzung.Belegt() {
		t.Errorf("nicht belegt: %+v", einschaetzung.Befunde)
	}

	if einschaetzung.Art != domain.EinschaetzungRegel {
		t.Errorf("Art %q", einschaetzung.Art)
	}
}

// **Eine erfundene Regelnummer macht sie unbelegt — und die Einschätzung
// bleibt trotzdem lesbar.**
//
// # Warum sie nicht verworfen wird
//
// **Weil ein Satz mit einer falschen Nummer nicht falsch sein muss** — und
// weil ein Programm, das die Nummer stillschweigend striche, den Rest als
// geprüft aussehen ließe. Gezeigt wird beides
// ([0107](../../docs/entscheidungen/0107-eine-einschaetzung-ist-keine-regelauskunft.md)).
func TestEineErfundeneRegelnummerMachtSieUnbelegt(t *testing.T) {
	t.Parallel()

	einschaetzung := domain.PruefeEinschaetzung(
		domain.EinschaetzungRegel,
		"Das steht in 702.19b und in [B1].", zweiRegelbelege())

	if einschaetzung.Belegt() {
		t.Fatal("die erfundene Nummer kam durch")
	}

	if einschaetzung.Satz == "" {
		t.Error("der Satz wurde verworfen statt gezeigt")
	}

	if len(einschaetzung.Befunde) != 1 ||
		einschaetzung.Befunde[0].Art != domain.BefundErfundeneRegelnummer {
		t.Errorf("Befunde %+v", einschaetzung.Befunde)
	}
}

// **Eine Einschätzung ohne jeden Beleg ist keine Quellenauskunft.**
//
// **Und kein Fehler**: „dazu steht in den Belegen nichts" ist eine richtige
// Antwort. Sie ist nur keine belegte — und genau das muss dastehen.
func TestEineEinschaetzungOhneBelegIstNichtBelegt(t *testing.T) {
	t.Parallel()

	einschaetzung := domain.PruefeEinschaetzung(
		domain.EinschaetzungRegel,
		"Dazu steht in den mitgegebenen Stellen nichts.", zweiRegelbelege())

	if einschaetzung.Belegt() {
		t.Fatal("eine Einschätzung ohne Beleg galt als belegt")
	}

	if len(einschaetzung.Befunde) != 1 ||
		einschaetzung.Befunde[0].Art != domain.BefundOhneBeleg {
		t.Errorf("Befunde %+v", einschaetzung.Befunde)
	}
}

// **Und sie ist nie eine amtliche Regelauskunft.**
//
// **Der Satz steht am Typ und nicht in einer Oberfläche**, weil ihn jede
// Fläche zeigen muss — und weil eine Zusage, die nur an einer Stelle steht,
// beim nächsten Umbau an drei Stellen fehlt (§16).
func TestEineEinschaetzungIstNieAmtlich(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Einschaetzungsarten() {
		einschaetzung := domain.PruefeEinschaetzung(art, "Etwas [B1].",
			zweiRegelbelege())

		if !strings.Contains(einschaetzung.Vorbehalt, "keine amtliche") {
			t.Errorf("%s: der Vorbehalt fehlt: %q", art, einschaetzung.Vorbehalt)
		}
	}
}

// **Eine unbekannte Art gibt es nicht.**
func TestEineUnbekannteEinschaetzungsartWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if fehler := domain.Einschaetzungsart("regelchen").Pruefe(); fehler == nil {
		t.Fatal("die erfundene Art kam durch")
	}

	for _, art := range domain.Einschaetzungsarten() {
		if fehler := art.Pruefe(); fehler != nil {
			t.Errorf("%s: %v", art, fehler)
		}
	}
}
