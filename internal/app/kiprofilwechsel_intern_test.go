package app

// **Ein Profilwechsel räumt die offenen KI-Vorschläge** — Befund 1 des
// Nachlaufs über Phase 5.3.
//
// # Der Befund
//
// **Das Auftragsbuch ist seit dem Zweitblick geteilt** — und es überlebte
// einen Profilwechsel. Ein in Profil A angelegter Vorschlag stand danach in
// Profil B weiter offen; übernommen hätte er die Karte von A geschrieben,
// während die Herkunft das Profil von B trüge.
//
// **Zwei Zusagen halten das jetzt**, und dieser Lauf prüft die zweite: die
// Grenze am Auftrag weist ab, was durchkommt ([service.KIAuftragswerk]), und
// der Wechsel räumt weg, was liegen bliebe (§7, Regel 1). Die erste steht in
// `internal/service/kiprofilgrenze_test.go`.

import (
	"testing"
)

func TestEinProfilwechselRaeumtDasAuftragswerk(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&nurEingeschaltete{})
	anwendung.MitKILaufzeit(func(string) KILaufzeitquelle {
		return festeLaufzeit{modelle: einModellFuerDenBauplan()}
	})

	staende, _, standfehler := anwendung.frageAnbieter(riegelkontext(t))
	if standfehler != nil {
		t.Fatalf("frageAnbieter: %v", standfehler)
	}

	if _, _, fehler := anwendung.werkzeugAus(
		riegelkontext(t), staende, ollamaanbieter, kideckung{}); fehler != nil {
		t.Fatalf("werkzeugAus: %v", fehler)
	}

	if anwendung.kiwerk == nil {
		t.Fatal("nach dem Bauen steht kein Werk")
	}

	anwendung.profilUmstellen(riegelkontext(t), 9)

	if anwendung.kiwerk != nil {
		t.Fatal("das Werk überlebte den Profilwechsel")
	}

	// **Und danach gibt es nichts zu übernehmen** — die Meldung sagt das,
	// statt einen Auftrag zu suchen, den es in diesem Profil nie gab.
	if _, fehler := anwendung.kiwerkOhneQuelle(riegelkontext(t)); fehler == nil {
		t.Fatal("nach dem Wechsel kam ein Werk ohne Quelle")
	}
}

// **Und ohne Profil im Kontext übernimmt niemand.**
//
// **Der Weg dorthin ist kein Versehen**: eine Fläche, die ihren
// Wurzelkontext selbst ausstattet, kann ihn ohne Profil bauen, und
// [App.imProfil] überschreibt dann nichts. Die Übernahme braucht es aber —
// sie ist seit Befund 1 des Nachlaufs an ein Profil gebunden.
func TestOhneProfilImKontextUebernimmtNiemand(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&nurEingeschaltete{})
	anwendung.MitKILaufzeit(func(string) KILaufzeitquelle {
		return festeLaufzeit{modelle: einModellFuerDenBauplan()}
	})

	staende, _, standfehler := anwendung.frageAnbieter(riegelkontext(t))
	if standfehler != nil {
		t.Fatalf("frageAnbieter: %v", standfehler)
	}

	if _, _, fehler := anwendung.werkzeugAus(
		riegelkontext(t), staende, ollamaanbieter, kideckung{}); fehler != nil {
		t.Fatalf("werkzeugAus: %v", fehler)
	}

	// **Ein nackter Kontext** — `a.profil` steht auf null, und `imProfil`
	// legt dann keines dazu.
	if antwort := anwendung.KIVorschlagUebernehmen(
		t.Context(), "ki-1", []string{"name"}); antwort.Fehler == nil {
		t.Error("die einzelne Übernahme kam ohne Profil durch")
	}

	if antwort := anwendung.KIGruppeUebernehmen(t.Context(),
		[]KIGruppenwahlDTO{{Kennung: "ki-1", Felder: []string{"name"}}},
	); antwort.Fehler == nil {
		t.Error("die Gruppenübernahme kam ohne Profil durch")
	}
}
