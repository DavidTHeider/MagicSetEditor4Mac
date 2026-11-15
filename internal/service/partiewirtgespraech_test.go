package service_test

// **Das Gespräch am Tisch, von der Seite des Wirts** — Werkstück G5 von
// Phase 5.5 (#1462).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Anmaßung, die bei einem Befehl schon einmal durchging.** Der
// Wirt setzt `Spielbefehl.Absender` aus der Sitzung, weil ein Gast sonst im
// Namen eines anderen zöge. Bei Chat und Präsenz stand dieselbe Lücke offen:
// `Von` kam ungeprüft aus der Nutzlast, und ein Gast konnte im Namen eines
// anderen reden und dessen Zeiger bewegen.
//
// **Und gegen den Tisch, der aus Zwiegesprächen bestand.** Ein Satz von Anna
// erreichte den Wirt und endete dort; Bert am selben Tisch hörte nichts.

import (
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einGespraechswirt baut einen Wirt samt Fenster und Runde.
//
// **Der Sitz heißt hier „anna" und nicht „gast"**, damit der Lauf zeigen kann,
// dass er aus der Sitzung kommt: die Nutzlast behauptet etwas anderes.
func einGespraechswirt(t *testing.T) (
	*service.Partiewirt, *nebendoppel, *platzdoppel,
) {
	t.Helper()

	naht := &nahtdoppel{}
	sitzung := service.NeuePartiesitzung(
		"p1", "wirt", naht, service.NeueReihenfolge(1))

	neben := &nebendoppel{}
	runde := &service.Partierunde{}

	bert, bertsLeitung := einPlatz("bert")
	runde.Setze(bert)

	wirt := service.NeuerPartiewirt("p1", "anna", &nehmerdoppel{}, sitzung).
		MitNebenstrom(neben).
		MitRunde(runde)

	return wirt, neben, bertsLeitung
}

// eineGastnachricht baut, was ein Gast über die Leitung schickt.
func eineGastnachricht(
	t *testing.T, art domain.Nachrichtenart, nutzlast any,
) domain.Partienachricht {
	t.Helper()

	inhalt, fehler := json.Marshal(nutzlast)
	if fehler != nil {
		t.Fatalf("die Nutzlast ließ sich nicht verpacken: %v", fehler)
	}

	nachricht := domain.Partienachricht{
		Art: art, Partie: "p1", Von: "gast", Epoche: 1, Inhalt: inhalt,
	}

	if art.WirdAufbewahrt() {
		nachricht.Folge = 1
	}

	return nachricht
}

// **Wer im Namen eines anderen redet, redet trotzdem im eigenen.**
func TestEinGastRedetNichtImNamenEinesAnderen(t *testing.T) {
	t.Parallel()

	wirt, neben, bertsLeitung := einGespraechswirt(t)

	if fehler := wirt.Angekommen(t.Context(), eineGastnachricht(t,
		domain.ArtChat, domain.Chat{Von: "bert", Text: "ich bin bert"},
	)); fehler != nil {
		t.Fatalf("angekommen: %v", fehler)
	}

	if len(neben.bekommen) != 1 {
		t.Fatalf("%d Nachrichten am Fenster, erwartet eine",
			len(neben.bekommen))
	}

	if amFenster := neben.bekommen[0].Chat; amFenster.Von != "anna" {
		t.Errorf("am Fenster steht %q — die Behauptung des Gastes hat gewonnen",
			amFenster.Von)
	}

	// **Und derselbe Satz geht mit demselben Sprecher an den Tisch.** Zwei
	// Wahrheiten über einen Satz wären schlimmer als keine.
	weiter := bertsLeitung.alle()
	if len(weiter) != 1 {
		t.Fatalf("bert bekam %d Nachrichten, erwartet eine", len(weiter))
	}

	var beiBert domain.Chat
	if fehler := json.Unmarshal(weiter[0].Inhalt, &beiBert); fehler != nil {
		t.Fatalf("kein Satz: %v", fehler)
	}

	if beiBert.Von != "anna" || beiBert.Text != "ich bin bert" {
		t.Errorf("bei bert steht %q von %q", beiBert.Text, beiBert.Von)
	}
}

// **Ein Zeiger bekommt beim Weiterreichen den Stand der Runde.**
//
// # Warum er umnummeriert wird
//
// **Weil der Stand des Gastes für seinen Sprung gilt und nicht für den
// nächsten.** Verbindet er neu, fängt er wieder bei eins an — ein Gast am
// anderen Ende wiese dann alles ab, was danach von ihm kommt. Der Wirt ist
// die Autorität am Tisch und zählt für ihn weiter.
func TestEinWeitergereichterZeigerBekommtDenStandDerRunde(t *testing.T) {
	t.Parallel()

	wirt, _, bertsLeitung := einGespraechswirt(t)

	for _, stand := range []uint64{7, 8} {
		if fehler := wirt.Angekommen(t.Context(), eineGastnachricht(t,
			domain.ArtPraesenz, domain.Praesenz{
				Von: "bert", Art: domain.PraesenzZeiger, Stand: stand,
				X: 100, Y: 200,
			},
		)); fehler != nil {
			t.Fatalf("angekommen: %v", fehler)
		}
	}

	weiter := bertsLeitung.alle()
	if len(weiter) != 2 {
		t.Fatalf("bert bekam %d Zeiger, erwartet zwei", len(weiter))
	}

	staende := make([]uint64, 0, len(weiter))

	for _, nachricht := range weiter {
		var gezeigt domain.Praesenz
		if fehler := json.Unmarshal(nachricht.Inhalt, &gezeigt); fehler != nil {
			t.Fatalf("keine Präsenz: %v", fehler)
		}

		if gezeigt.Von != "anna" {
			t.Errorf("der Zeiger gehört %q statt anna", gezeigt.Von)
		}

		staende = append(staende, gezeigt.Stand)
	}

	if staende[0] != 1 || staende[1] != 2 {
		t.Errorf("die Stände sind %v, erwartet [1 2] — die Runde zählt",
			staende)
	}
}

// **Eine überholte Präsenz erreicht weder das Fenster noch den Tisch.**
func TestEineUeberholtePraesenzErreichtNiemanden(t *testing.T) {
	t.Parallel()

	wirt, neben, bertsLeitung := einGespraechswirt(t)

	zeige := func(stand uint64) {
		t.Helper()

		if fehler := wirt.Angekommen(t.Context(), eineGastnachricht(t,
			domain.ArtPraesenz, domain.Praesenz{
				Von: "anna", Art: domain.PraesenzZeiger, Stand: stand,
			},
		)); fehler != nil {
			t.Fatalf("angekommen: %v", fehler)
		}
	}

	zeige(5)
	zeige(4)

	if len(neben.bekommen) != 1 {
		t.Errorf("%d Nachrichten am Fenster, erwartet eine — die zweite war "+
			"überholt", len(neben.bekommen))
	}

	if bei := bertsLeitung.alle(); len(bei) != 1 {
		t.Errorf("bert bekam %d Zeiger, erwartet einen", len(bei))
	}
}

// **Ohne Runde geschieht am Tisch nichts, und das ist kein Fehler.**
//
// **Die Gegenprobe** (§10): ein Wirt ohne offenen Tisch — an der
// Kommandozeile etwa — führt die Partie trotzdem. Ein Satz, den niemand
// weiterreicht, hält sie nicht an.
func TestOhneRundeHaeltEinSatzNichtsAn(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := service.NeuePartiesitzung(
		"p1", "wirt", naht, service.NeueReihenfolge(1))

	neben := &nebendoppel{}
	wirt := service.NeuerPartiewirt("p1", "anna", &nehmerdoppel{}, sitzung).
		MitNebenstrom(neben)

	if fehler := wirt.Angekommen(t.Context(), eineGastnachricht(t,
		domain.ArtChat, domain.Chat{Von: "anna", Text: "hallo"},
	)); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}

	if len(neben.bekommen) != 1 {
		t.Error("ohne Runde erreichte der Satz nicht einmal das Fenster")
	}
}

// **Was keine Gestalt hat, hält an — und erreicht niemanden.**
func TestEineFormloseNebennachrichtHaeltDenWirtAn(t *testing.T) {
	t.Parallel()

	wirt, neben, bertsLeitung := einGespraechswirt(t)

	if fehler := wirt.Angekommen(t.Context(), eineGastnachricht(t,
		domain.ArtChat, domain.Chat{Von: "anna", Text: "   "},
	)); fehler == nil {
		t.Fatal("ein leerer Satz ging durch")
	}

	if len(neben.bekommen) != 0 || len(bertsLeitung.alle()) != 0 {
		t.Error("ein leerer Satz wurde trotzdem weitergereicht")
	}
}
