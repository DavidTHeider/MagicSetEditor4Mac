package app

// **Der Vorgang sagt, wo er steht und was geht** — TRADE-02.AK-02
// (#1716).
//
// # Was hier gemessen wird
//
// **Dass jeder Weg, der einen Vorgang zurückgibt, dieselben vier Angaben
// trägt** — eröffnen, schreiten, ansehen und die Liste. Die Regel steht in
// `domain`; hier steht, dass sie an der Naht ankommt.
//
// **Und für wen sie gelten** — `ich`, die Partei dieses Profils
// (#2026).
// Ohne sie wusste das Fenster, dass seine Stimme fehlt, und nicht, wessen
// Stimme das ist.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerVorgangNenntLageSchritteUndPhase(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch())
	if angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	vorgang := angelegt.Vorgang
	if vorgang.Ich != ich || vorgang.Lage != "dran" || vorgang.Naechster != "anbieten" ||
		vorgang.Phase != "angebot" ||
		!slices.Equal(vorgang.Moegliche, []string{"anbieten", "abbrechen"}) {
		t.Fatalf("eröffnet: %+v", vorgang)
	}

	// **Nach dem Anbieten wartet die Initiatorin** auf ihr Gegenüber — und
	// darf doch zustimmen, abbrechen und ablaufen lassen.
	angeboten := anwendung.HandelSchritt(ctx, derLeitungsvorgang, "anbieten", "", "")
	if angeboten.Fehler != nil || angeboten.Vorgang.Phase != "angebot" ||
		!slices.Contains(angeboten.Vorgang.Moegliche, "zustimmen") ||
		slices.Contains(angeboten.Vorgang.Moegliche, "anbieten") {
		t.Fatalf("angeboten: %+v", angeboten)
	}

	gesehen := anwendung.HandelStand(ctx, derLeitungsvorgang)
	if gesehen.Fehler != nil || gesehen.Vorgang.Ich != ich ||
		gesehen.Vorgang.Lage != angeboten.Vorgang.Lage ||
		!slices.Equal(gesehen.Vorgang.Moegliche, angeboten.Vorgang.Moegliche) {
		t.Fatalf("gesehen: %+v gegen %+v", gesehen.Vorgang, angeboten.Vorgang)
	}

	liste := anwendung.HandelVorgaenge(ctx, dasLeitungsgegenueber)
	if liste.Fehler != nil || len(liste.Vorgaenge) != 1 || liste.Vorgaenge[0].Ich != ich ||
		!slices.Equal(liste.Vorgaenge[0].Moegliche, angeboten.Vorgang.Moegliche) {
		t.Fatalf("Liste: %+v", liste)
	}
}

func TestOhneIdentitaetStehenKeineSchritte(t *testing.T) {
	// **Ohne eigene Kennung gibt es keine Partei**, also nichts, was „ich"
	// darf — leer und kein `null`.
	dto := alsHandelsvorgangDTO(domain.Handelsangebot{
		Vorgang: "v", Art: domain.HandelTausch, Initiator: "ba",
		Parteien: []string{"ba", "bb"}, Zustand: domain.HandelEntwurf,
	}, "")
	if dto.Ich != "" || dto.Moegliche == nil || len(dto.Moegliche) != 0 ||
		dto.Naechster != "" || dto.Phase != "angebot" {
		t.Fatalf("DTO: %+v", dto)
	}
}
