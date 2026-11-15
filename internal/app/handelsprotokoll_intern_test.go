package app

// **Das Protokoll an der Naht** — TRADE-03.AK-03
// (#1717).

import "testing"

func TestDasProtokollKommtAnDerNahtAn(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	if geschritten := anwendung.HandelSchritt(ctx, derLeitungsvorgang, "anbieten", "", ""); geschritten.Fehler != nil {
		t.Fatalf("anbieten: %+v", geschritten.Fehler)
	}

	antwort := anwendung.HandelProtokoll(ctx, derLeitungsvorgang)
	if antwort.Fehler != nil || len(antwort.Zeilen) != 2 ||
		antwort.Zeilen[0].Von != "" || antwort.Zeilen[0].Nach != "entwurf" ||
		antwort.Zeilen[1].Von != "entwurf" || antwort.Zeilen[1].Nach != "angeboten" ||
		antwort.Zeilen[1].Zeit == "" {
		t.Fatalf("Protokoll: %+v", antwort)
	}
}

func TestDasProtokollOhneProfilIstEineAbsage(t *testing.T) {
	antwort := PeerflaecheUeber(bestandFuerDenRiegel(t)).HandelProtokoll(t.Context(), "v")

	if antwort.Fehler == nil || antwort.Zeilen == nil {
		t.Fatalf("Antwort: %+v", antwort)
	}
}
