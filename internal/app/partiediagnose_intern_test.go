package app

// **Was die Diagnose sagt und was sie verschweigt** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// **Der zweite Halbsatz ist der, für den es diese Läufe gibt**: „Diagnostik
// für Route/RTT/Queue/Desync/Snapshot/Framezeit **ohne Kartengeheimnisse**."

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Tisch läuft nichts — und die Liste ist trotzdem eine.**
func TestOhneTischSagtDieDiagnoseNichts(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).Partiediagnose(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Laeuft:
		t.Error("ohne Tisch läuft nichts")
	case antwort.Rolle != "":
		t.Errorf("die Rolle ist %q", antwort.Rolle)
	case antwort.Leitungen == nil:
		t.Error("die Liste ist nil statt leer")
	}
}

// **Auch die Diagnose fragt nach ihrer Wirkung.**
func TestDieDiagnoseFragtNachDemLesen(t *testing.T) {
	t.Parallel()

	wache := &abweisendeWache{}
	anwendung := freigabeanwendung(t)

	anwendung.MitWache(wache)

	antwort := anwendung.Partiediagnose(t.Context())

	switch {
	case antwort.Fehler == nil:
		t.Fatal("die Abweisung kam nicht an")
	case antwort.Leitungen == nil:
		t.Error("die Liste ist nil statt leer")
	case wache.gefragt != domain.WirkungLesen:
		t.Errorf("gefragt wurde nach %q", wache.gefragt)
	}
}

// **Ein Wirt sieht, wer an seinem Tisch sitzt und über welchen Weg.**
//
// **Und keine Zahl über seine Karten.** Die Zusicherung unten ist die Zusage
// des Issues: der Abdruck des Decks steht in der Partie und darf hier nicht
// erscheinen — er ist ein Hash über Karten, und wer ihn weitergäbe, gäbe
// etwas weiter, gegen das sich raten lässt.
func TestDieDiagnoseNenntDieLeitungenUndKeinGeheimnis(t *testing.T) {
	t.Parallel()

	wirt, gast, abdruck := einBesetzterTisch(t)

	beim := wirt.anwendung.Partiediagnose(wirt.ctx)

	switch {
	case beim.Fehler != nil:
		t.Fatalf("Fehler: %+v", beim.Fehler)
	case !beim.Laeuft:
		t.Fatal("der Tisch läuft nicht")
	case beim.Rolle != rolleWirt:
		t.Fatalf("die Rolle ist %q", beim.Rolle)
	case len(beim.Leitungen) != 1:
		t.Fatalf("erwartet eine Leitung, bekam %+v", beim.Leitungen)
	case beim.Leitungen[0].Sitz != "bert":
		t.Errorf("der Sitz ist %q", beim.Leitungen[0].Sitz)
	case beim.Leitungen[0].Weg == "":
		t.Error("der Weg ist leer")
	}

	// **Und der Gast sieht seine eine Leitung zum Wirt.**
	beimGast := gast.anwendung.Partiediagnose(gast.ctx)

	switch {
	case beimGast.Fehler != nil:
		t.Fatalf("Fehler: %+v", beimGast.Fehler)
	case beimGast.Rolle != rolleGast:
		t.Fatalf("die Rolle ist %q", beimGast.Rolle)
	case len(beimGast.Leitungen) != 1:
		t.Fatalf("erwartet eine Leitung, bekam %+v", beimGast.Leitungen)
	case beimGast.Leitungen[0].Gegenueber == "":
		t.Error("das Gegenüber fehlt")
	}

	// **Und keine der beiden Antworten trägt den Abdruck des Decks.**
	if abdruck == "" {
		t.Fatal("der Abzug hat keinen Abdruck — der Lauf prüfte nichts")
	}

	for name, antwort := range map[string]PartiediagnoseAntwort{
		"wirt": beim, "gast": beimGast,
	} {
		roh, fehler := json.Marshal(antwort)
		if fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		if strings.Contains(string(roh), abdruck) {
			t.Errorf("%s: die Diagnose nennt den Abdruck des Decks: %s",
				name, roh)
		}
	}
}

// **Und sie nennt den Sichtabdruck nicht** — den Hash, der die **eigenen**
// privaten Ergänzungen mitrechnet.
//
// **Ihn nennt das Issue beim Namen**: „Geteilter Eventhash ermöglicht keine
// simple Wörterbuchsuche auf einzelne geheime Karten." Die **Folge** daneben
// steht da; sie zählt den Strom und sagt nichts über seinen Inhalt.
//
// **Ein eigener Lauf und nicht der über zwei Anwendungen**: dort hat ein
// frisch gesetzter Gast noch keinen — der Wirt antwortet auf einen Wunsch
// über eine Partie ohne Ereignis mit `weiter`, und dabei entsteht keiner.
func TestDieDiagnoseNenntDenSichtabdruckNicht(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	anwendung.beitrittMerken(nil, "p1")
	anwendung.Merke(7, "einabdruckueberkarten")

	antwort := anwendung.Partiediagnose(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case !antwort.Laeuft || antwort.Rolle != rolleGast:
		t.Fatalf("der Stand ist %+v", antwort)
	case antwort.Stand != 7 || !antwort.StandBekannt:
		t.Errorf("die Folge ist %d (bekannt: %v)",
			antwort.Stand, antwort.StandBekannt)
	}

	roh, fehler := json.Marshal(antwort)
	if fehler != nil {
		t.Fatalf("kodieren: %v", fehler)
	}

	if strings.Contains(string(roh), "einabdruckueberkarten") {
		t.Errorf("die Diagnose nennt den Sichtabdruck: %s", roh)
	}
}

// einBesetzterTisch gibt einen Wirt mit einem sitzenden Gast — und den
// Abdruck des Decks, das dabei eingefroren wurde.
//
// **Der Abdruck kommt mit**, weil die Zusicherung darüber der Zweck dieses
// Laufs ist: eine Zahl, die man von Hand hinschriebe, prüfte eine Zahl und
// nicht die Partie.
func einBesetzterTisch(t *testing.T) (wirt, gast *seite, abdruck string) {
	t.Helper()

	wirt, gast = zweiAnwendungen(t)

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, _ := einTischBeimWirt(t, wirt)

	sicht := wirt.anwendung.Partie(wirt.ctx, partieID, "anna")
	if sicht.Fehler != nil {
		t.Fatalf("Partie: %+v", sicht.Fehler)
	}

	for _, sitz := range sicht.Partie.Sitze {
		if sitz.Kennung == "bert" {
			abdruck = sitz.Deckabdruck
		}
	}

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{})

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	return wirt, gast, abdruck
}

// **Die Runde wird gemessen, wenn eine zurückkam — und sonst nicht.**
//
// **Beides in einem Lauf**, denn die Aussage ist eine: „fehlende Messgrößen
// erscheinen unbekannt" heißt, dass die Null vorher **keine Auskunft** ist.
func TestDieRundeErscheintErstNachDerRueckkehr(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.beitrittMerken(nil, "p1")

	if antwort := anwendung.Partiediagnose(t.Context()); antwort.RueckkehrBekannt {
		t.Fatalf("vor der ersten Rückkehr steht schon eine Zahl: %d",
			antwort.RueckkehrMillis)
	}

	// **Der Wunsch ging vor einer Weile hinaus** — die Uhr wird gesetzt und
	// nicht abgewartet: ein Lauf, der schliefe, prüfte die Uhr dieser
	// Maschine (§3).
	anwendung.gefragt(time.Now().Add(-50 * time.Millisecond))
	anwendung.Merke(1, "abdruck")

	antwort := anwendung.Partiediagnose(t.Context())

	switch {
	case !antwort.RueckkehrBekannt:
		t.Fatal("die Rückkehr wurde nicht gemessen")
	case antwort.RueckkehrMillis < 50:
		t.Errorf("die Runde ist %d ms", antwort.RueckkehrMillis)
	}

	// **Und eine zweite Antwort auf denselben Wunsch misst nicht noch
	// einmal** — sonst zählte sie die Zeit seit der ersten mit.
	anwendung.Merke(2, "abdruck")

	zweite := anwendung.Partiediagnose(t.Context())
	if zweite.RueckkehrMillis != antwort.RueckkehrMillis {
		t.Errorf("die Runde wanderte von %d auf %d ms",
			antwort.RueckkehrMillis, zweite.RueckkehrMillis)
	}
}

// **Geht der Rückkehrwunsch nicht hinaus, endet die Sitzung** — und die
// angefangene Messung wird verworfen.
//
// # Warum die Messung dabei verworfen wird
//
// **Weil sie sonst den Tisch überlebte.** Eine offene Messung schlösse beim
// **nächsten** Tisch und zählte die Zeit dazwischen mit; im Fenster stünde
// dann eine Runde von Minuten, die niemand gedreht hat (§10).
func TestEineGescheiterteRueckkehrEndetDieSitzung(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	sitzung := eineAngenommene(t, domain.ZweckPartie)
	sitzung.sendefehler = errors.New("die Leitung ist zu")

	anwendung.melden = nil
	anwendung.beitrittMerken(nil, "p1")

	gast := service.NeuerPartiegast("p1",
		service.NeuePartiesitzung("p1", "gast", sitzung,
			service.NeueReihenfolge(1)),
		anwendung.tischmelder(t.Context(), "p1"))

	anwendung.rueckkehrMelden(t.Context(), gast, sitzung, "p1")

	if !sitzung.istZu() {
		t.Fatal("die Sitzung steht noch offen")
	}

	// **Und es steht keine Runde da** — gemessen wurde nichts.
	if antwort := anwendung.Partiediagnose(
		t.Context()); antwort.RueckkehrBekannt {
		t.Fatalf("es steht eine Runde da: %d", antwort.RueckkehrMillis)
	}
}
