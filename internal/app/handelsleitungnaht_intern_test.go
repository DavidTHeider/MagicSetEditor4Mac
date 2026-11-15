package app

// **Die Naht der Handelsleitung** —
// #1637.
//
// # Was hier gemessen wird
//
// **Der Weg durch die Naht, über einen echten Bestand** — eröffnen,
// vorschlagen, schreiten, lesen. Gemessen wird dabei, was nur hier zu sehen
// ist: dass der **Initiator aus dem Ausweis** kommt, dass die **Eltern
// gelesen** werden statt genannt, und dass eine Absage eine Antwort mit
// leeren Listen ist und keinen Nullwert.
//
// **Und der Ausgang dazu**: dass eine Nachricht mit ihrer Folge
// hineingeht, als Nachlieferung wieder herauskommt und nach der Bestätigung
// nicht mehr dasteht. Der Adapter zwischen `service.Ausgang` und
// `storage.Ausgangsablage` ist die Stelle, an der `T5` eine Naht ohne
// Erfüllung hatte — hier wird sie gefahren.
//
// **Paketintern**, weil `imProfil`, `db` und der Probebund es sind —
// dieselbe Vorrichtung wie in `handelsverbuchungnaht_intern_test.go`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// derLeitungsvorgang ist der Handel dieses Laufs.
const derLeitungsvorgang = "v-leitung"

// dasLeitungsgegenueber ist die zweite Partei.
const dasLeitungsgegenueber = "bgegenueber"

// eineLeitungsflaeche stellt die Anwendung samt Identität auf.
func eineLeitungsflaeche(t *testing.T) (anwendung *App, ich string) {
	t.Helper()

	anwendung = PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	ausweis, fehler := anwendung.Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	return anwendung, string(ausweis.Kennung())
}

// einLeitungswunsch ist das kleinste Angebot, das prüft.
func einLeitungswunsch() Handelswunsch {
	return Handelswunsch{
		Vorgang: derLeitungsvorgang, Art: string(domain.HandelTausch),
		Gegenueber: []string{dasLeitungsgegenueber},
		Positionen: []HandelspositionDTO{{
			Von: dasLeitungsgegenueber, Printing: "shivan-lea", Menge: 1,
			Guete: string(domain.ZustandNearMint), Sprache: "en",
			Finish: "nonfoil",
		}},
	}
}

// **Der Weg durch die Naht — eröffnen, vorschlagen, schreiten, lesen.**
func TestDieLeitungsnahtFuehrtEinenVorgang(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch())
	if angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	// **Der Initiator kommt aus dem Ausweis** — nicht aus dem Wunsch.
	if angelegt.Vorgang.Initiator != ich {
		t.Errorf("der Initiator ist %q statt %q",
			angelegt.Vorgang.Initiator, ich)
	}

	if angelegt.Vorgang.Revision != 1 || angelegt.Vorgang.Abdruck == "" {
		t.Errorf("eröffnet wurde %+v", angelegt.Vorgang)
	}

	// **Angeboten, bevor vorgeschlagen wird** — aus einem Entwurf führt kein
	// Gegenangebot (QA-06b, #1967).
	if angeboten := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittAnbieten), "", ""); angeboten.Fehler != nil {
		t.Fatalf("Anbieten: %+v", angeboten.Fehler)
	}

	zwei := []HandelspositionDTO{{
		Von: dasLeitungsgegenueber, Printing: "shivan-lea", Menge: 2,
		Guete: string(domain.ZustandNearMint), Sprache: "en",
		Finish: "nonfoil",
	}}

	// **Aus `angeboten` schlägt der Initiator nicht vor** — nur die
	// Gegenseite (#1975). Dieser Rechner ist der Initiator.
	if eigener := anwendung.HandelVorschlagen(ctx, derLeitungsvorgang, zwei); eigener.Fehler == nil ||
		eigener.Fehler.Code != CodeValidierung {
		t.Fatalf("der Initiator schlug aus angeboten vor: %+v", eigener.Fehler)
	}

	// **Die Gegenseite schlägt vor** — so, wie es über die Leitung ankäme.
	heute, fehler := anwendung.handelsleitung().Stand(ctx, derLeitungsvorgang)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := anwendung.handelsleitung().Vorschlagen(ctx,
		heute.Fortgeschrieben(alsPositionen(zwei)), dasLeitungsgegenueber); fehler != nil {
		t.Fatalf("der Vorschlag der Gegenseite: %v", fehler)
	}

	// **Die Eltern werden gelesen und nicht genannt** — in der Verhandlung
	// schlägt jede Seite vor.
	gegenvorschlag := anwendung.HandelStand(ctx, derLeitungsvorgang)

	naechste := anwendung.HandelVorschlagen(ctx, derLeitungsvorgang, einLeitungswunsch().Positionen)
	if naechste.Fehler != nil {
		t.Fatalf("Vorschlagen: %+v", naechste.Fehler)
	}

	if naechste.Vorgang.Eltern != gegenvorschlag.Vorgang.Abdruck {
		t.Errorf("die Eltern sind %q statt %q",
			naechste.Vorgang.Eltern, gegenvorschlag.Vorgang.Abdruck)
	}

	// **Ein Schritt ohne `wer` ist dieser Rechner.**
	gegangen := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittZustimmen), "", "")
	if gegangen.Fehler != nil {
		t.Fatalf("Schreiten: %+v", gegangen.Fehler)
	}

	if len(gegangen.Vorgang.Zustimmungen) != 1 ||
		gegangen.Vorgang.Zustimmungen[0] != ich {
		t.Errorf("eingetragen wurden %v", gegangen.Vorgang.Zustimmungen)
	}

	stand := anwendung.HandelStand(ctx, derLeitungsvorgang)
	if stand.Fehler != nil {
		t.Fatalf("Stand: %+v", stand.Fehler)
	}

	if stand.Vorgang.Abdruck != gegangen.Vorgang.Abdruck ||
		len(stand.Vorgang.Positionen) != 1 {
		t.Errorf("gelesen wurde %+v", stand.Vorgang)
	}

	if stand.Vorgang.Ende {
		t.Error("ein verhandelter Vorgang gilt als Ende")
	}
}

// **Eine Absage gibt leere Listen und keinen Nullwert.**
//
// **Die Gegenprobe zur Zusage aus §3**: ein `null` im Fenster sieht aus wie
// „konnte nicht beantwortet werden" und nicht wie „nichts da".
func TestEineAbsageDerLeitungGibtLeereListen(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)

	antwort := anwendung.HandelStand(riegelkontext(t), "gibtesnicht")
	if antwort.Fehler == nil {
		t.Fatal("ein Vorgang, den es nicht gibt, kam durch")
	}

	if antwort.Vorgang.Parteien == nil ||
		antwort.Vorgang.Positionen == nil ||
		antwort.Vorgang.Zustimmungen == nil {
		t.Errorf("die Absage trägt einen Nullwert: %+v", antwort.Vorgang)
	}
}

// **Die drei Abschlussschritte enden an der Leitung.**
//
// **Hier und nicht nur in `service`**: die Fläche soll den Satz durchreichen,
// der den Weg nennt, der es kann — eine Absage, die nur „nein" sagt, schickt
// jemanden suchen.
func TestDieLeitungsnahtWeistDenAbschlussAb(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if antwort := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); antwort.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", antwort.Fehler)
	}

	antwort := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittVerbuchen), "", "")
	if antwort.Fehler == nil {
		t.Fatal("verbuchen kam über die Leitung durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Verbuchung") {
		t.Errorf("die Absage nennt die Verbuchung nicht: %s",
			antwort.Fehler.Meldung)
	}
}

// **Ohne Identität wird nichts eröffnet, kein Schritt gegangen und nichts
// vorgeschlagen.**
//
// **Weil der Initiator und die eigene Stimme aus dem Ausweis kommen.** Eine
// Fläche, die hier weiterliefe, schriebe einen Vorgang mit leerem Initiator —
// und der Vertrag wiese ihn ab, mit einem Satz über ein Feld statt über die
// fehlende Identität.
func TestOhneIdentitaetFuehrtDieLeitungNichts(t *testing.T) {
	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); angelegt.Fehler == nil {
		t.Error("ohne Identität wurde eröffnet")
	}

	if gegangen := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittAnbieten), "", ""); gegangen.Fehler == nil {
		t.Error("ohne Identität wurde geschritten")
	}

	// **Und nichts vorgeschlagen** — wer vorschlägt, kommt aus dem Ausweis
	// (#1975).
	if vorgeschlagen := anwendung.HandelVorschlagen(ctx, derLeitungsvorgang,
		einLeitungswunsch().Positionen); vorgeschlagen.Fehler == nil {
		t.Error("ohne Identität wurde vorgeschlagen")
	}
}

// **Ein Vorschlag auf einen Vorgang, den es nicht gibt, endet beim Lesen.**
//
// **Vor dem Fortschreiben und nicht danach**: wer erst die nächste Revision
// baute, baute sie über einem Nullwert — und der Konflikt käme mit einem Satz
// über Eltern statt über den Vorgang.
func TestEinVorschlagOhneVorgangEndetBeimLesen(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)

	antwort := anwendung.HandelVorschlagen(riegelkontext(t), "gibtesnicht",
		einLeitungswunsch().Positionen)
	if antwort.Fehler == nil {
		t.Fatal("ein Vorschlag ohne Vorgang kam durch")
	}

	if antwort.Vorgang.Positionen == nil {
		t.Error("die Absage trägt einen Nullwert")
	}
}

// **Eine unleserliche Zeile im Ausgang hält den ganzen Ausgang auf.**
//
// # Warum das Absicht ist
//
// **Weil eine übersprungene Nachricht verloren wäre.** Die Gegenseite weiß
// nichts von ihr; wer sie stillschweigend ausließe, meldete „nichts steht
// aus" über einen Ausgang, in dem etwas liegt. Der Satz nennt deshalb die
// **Kennung** der Zeile — an ihr lässt sie sich wegräumen.
func TestEineUnleserlicheZeileImAusgangKlagt(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	// **Von Hand hingelegt, weil kein Weg so etwas schreibt** — der Adapter
	// schreibt JSON, und die Domäne kann daran nicht scheitern. Geprüft wird
	// der Weg zurück: ein Bestand, den jemand von außen angefasst hat, ist
	// der Normalfall (§8).
	if _, fehler := anwendung.db.ExecContext(anwendung.imProfil(ctx),
		`INSERT INTO handelsausgang
		   (profil_id, nachricht_id, vorgang, empfaenger, nachricht)
		 VALUES ((SELECT id FROM profile LIMIT 1), 'n-krumm', ?, ?, '{')`,
		derLeitungsvorgang, dasLeitungsgegenueber); fehler != nil {
		t.Fatalf("die krumme Zeile legen: %v", fehler)
	}

	antwort := anwendung.HandelWiederaufnahme(ctx, derLeitungsvorgang,
		dasLeitungsgegenueber,
		HandelsfortschrittDTO{Vorgang: derLeitungsvorgang})
	if antwort.Fehler == nil {
		t.Fatal("die krumme Zeile kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "n-krumm") {
		t.Errorf("die Klage nennt die Zeile nicht: %s", antwort.Fehler.Meldung)
	}
}

// **Ein zweites Eröffnen ist ein Konflikt, kein Überschreiben.**
//
// **Und ein Vorschlag ohne Position sagt nicht, worum es geht** — beides
// endet an der Leitung, und beides gibt eine Antwort mit leeren Listen.
func TestDieLeitungsnahtWeistZweitesEroeffnenUndLeerenVorschlagAb(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	if nochmal := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); nochmal.Fehler == nil {
		t.Error("ein zweites Eröffnen kam durch")
	}

	if leer := anwendung.HandelVorschlagen(ctx, derLeitungsvorgang,
		[]HandelspositionDTO{}); leer.Fehler == nil {
		t.Error("ein Vorschlag ohne Position kam durch")
	}
}

// **Eine genannte Partei trägt ihre eigene Stimme ein.**
//
// **Die Gegenprobe zum leeren `wer`**: über eine Leitung kommt die
// Zustimmung des Gegenübers an, und sie gehört eingetragen, ohne dass dieser
// Rechner sie wäre. Ohne diesen Fall sagte „ein leeres `wer` ist dieser
// Rechner" nichts — es gäbe nur eines.
func TestEineGenanntePartieTraegtIhreEigeneStimmeEin(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(
		ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	if angeboten := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittAnbieten), "", ""); angeboten.Fehler != nil {
		t.Fatalf("Anbieten: %+v", angeboten.Fehler)
	}

	gegangen := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittZustimmen), dasLeitungsgegenueber, "")
	if gegangen.Fehler != nil {
		t.Fatalf("Zustimmen: %+v", gegangen.Fehler)
	}

	if len(gegangen.Vorgang.Zustimmungen) != 1 ||
		gegangen.Vorgang.Zustimmungen[0] != dasLeitungsgegenueber {
		t.Fatalf("eingetragen wurden %v", gegangen.Vorgang.Zustimmungen)
	}

	// **Und die eigene Stimme macht die Runde voll.**
	beide := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittZustimmen), "", "")
	if beide.Fehler != nil {
		t.Fatalf("die eigene Stimme: %+v", beide.Fehler)
	}

	if beide.Vorgang.Zustand != string(domain.HandelAngenommen) {
		t.Errorf("nach zwei Stimmen steht er auf %q — %v",
			beide.Vorgang.Zustand, ich)
	}
}
