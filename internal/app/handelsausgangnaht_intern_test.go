package app

// **Der Ausgang der Handelsleitung** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass eine Nachricht mit ihrer Folge hineingeht, als Nachlieferung wieder
// herauskommt und nach der Bestätigung nicht mehr dasteht.** Der Adapter
// zwischen [service.Ausgang] und `storage.Ausgangsablage` ist die Stelle, an
// der `T5` eine **Naht ohne Erfüllung** hatte — die Gestalten gingen
// auseinander, und niemand hat es gemerkt, weil sie keinen Aufrufer hatte.
//
// **Eine eigene Datei neben `handelsleitungnaht_intern_test.go`**: dort steht
// der Vertrag, hier der Weg hinaus — der Schnitt folgt der Frage (§2), und
// die Zeilengrenze macht ihn sichtbar.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Ausgang: hinein, nachliefern, bestätigen.**
func TestDieLeitungsnahtFuehrtDenAusgang(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if antwort := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); antwort.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", antwort.Fehler)
	}

	// **Aus einem Entwurf geht nichts hinaus** — erst der Schritt macht ihn
	// zu einer Nachricht.
	if leer := anwendung.HandelSchicken(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber); leer.Fehler == nil {
		t.Fatal("der Entwurf ging hinaus")
	}

	if antwort := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittAnbieten), "", ""); antwort.Fehler != nil {
		t.Fatalf("Anbieten: %+v", antwort.Fehler)
	}

	geschickt := anwendung.HandelSchicken(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber)
	if geschickt.Fehler != nil {
		t.Fatalf("Schicken: %+v", geschickt.Fehler)
	}

	offen := anwendung.HandelWiederaufnahme(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber,
		HandelsfortschrittDTO{Vorgang: derLeitungsvorgang})
	if offen.Fehler != nil {
		t.Fatalf("Wiederaufnahme: %+v", offen.Fehler)
	}

	if len(offen.Nachzuliefern) != 1 ||
		offen.Nachzuliefern[0].NachrichtID != geschickt.NachrichtID {
		t.Fatalf("nachzuliefern ist %+v", offen.Nachzuliefern)
	}

	// **Die Folge kommt aus der Nachricht und zählt ab eins.**
	if offen.Nachzuliefern[0].Folge != 1 ||
		offen.Nachzuliefern[0].Art != string(domain.NachrichtVorschlag) {
		t.Errorf("die Nachlieferung ist %+v", offen.Nachzuliefern[0])
	}

	if weg := anwendung.HandelPostbestaetigung(
		ctx, geschickt.NachrichtID); weg.Fehler != nil {
		t.Fatalf("Bestätigen: %+v", weg.Fehler)
	}

	danach := anwendung.HandelWiederaufnahme(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber,
		HandelsfortschrittDTO{Vorgang: derLeitungsvorgang})
	if danach.Fehler != nil || len(danach.Nachzuliefern) != 0 {
		t.Errorf("nach der Bestätigung: %+v", danach)
	}

	// **Eine Bestätigung ins Leere ist ein Nein** — und eine ohne Kennung
	// endet vor dem Bestand.
	if ohne := anwendung.HandelPostbestaetigung(ctx, ""); ohne.Fehler == nil {
		t.Error("eine Bestätigung ohne Kennung kam durch")
	}
}

// **Und ein Widerspruch der Fortschritte ist eine Klärung und kein Fehler.**
func TestEinWiderspruchImAusgangIstEineKlaerung(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch())
	if angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	// **Dieselbe Revision, ein anderer Abdruck** — dann haben beide Seiten
	// ordentlich gearbeitet, und trotzdem stehen zwei Verläufe da.
	antwort := anwendung.HandelWiederaufnahme(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber, HandelsfortschrittDTO{
			Vorgang: derLeitungsvorgang, Revision: 1, Abdruck: "einanderer",
		})
	if antwort.Fehler != nil {
		t.Fatalf("Wiederaufnahme: %+v", antwort.Fehler)
	}

	if !antwort.Klaerung {
		t.Error("der Widerspruch gilt nicht als Klärung")
	}

	if antwort.Nachzuliefern == nil {
		t.Error("die Klärung trägt einen Nullwert statt einer leeren Liste")
	}
}

// **Ein Ausgang, der nicht zu lesen ist, klagt durch.**
//
// # Warum die Tabelle weggenommen wird und nicht der ganze Bestand
//
// **Weil der Vorgang noch zu lesen sein muss.** Die Wiederaufnahme liest
// zuerst den eigenen Stand; ein geschlossener Bestand endete dort, und der
// Adapter käme nie dran. Was gemessen werden soll, ist **seine** Absage: er
// reicht weiter, was die Ablage sagt — eine leere Liste daneben hieße
// „nichts steht aus", und das wäre die falsche Auskunft über einen Ausgang,
// den niemand lesen konnte.
func TestEinUnlesbarerAusgangKlagtDurch(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	if _, fehler := anwendung.db.ExecContext(ctx,
		`DROP TABLE handelsausgang`); fehler != nil {
		t.Fatalf("die Tabelle wegnehmen: %v", fehler)
	}

	if antwort := anwendung.HandelWiederaufnahme(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber,
		HandelsfortschrittDTO{Vorgang: derLeitungsvorgang},
	); antwort.Fehler == nil {
		t.Error("der fehlende Ausgang kam durch")
	}
}
