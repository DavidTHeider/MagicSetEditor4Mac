package app

// **Die Liste der Handelsvorgänge an der Naht** — TRADE-01
// (#1715).
//
// Über einen echten Bestand, mit derselben Vorrichtung wie
// `handelsleitungnaht_intern_test.go`: ein Entwurf steht in der Liste, wird
// bearbeitet, kopiert und gelöscht — und was hinausging, nicht.

import (
	"testing"
	"time"
)

func TestDieListeFuehrtEinenEntwurf(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	if angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch()); angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	liste := anwendung.HandelUebersicht(ctx, HandelsfilterDTO{})
	if liste.Fehler != nil || len(liste.Zeilen) != 1 {
		t.Fatalf("Liste: %+v", liste)
	}

	zeile := liste.Zeilen[0]
	if zeile.Lage != "dran" || zeile.Naechster != "anbieten" || zeile.Unbestaetigt ||
		len(zeile.Gegenueber) != 1 || zeile.Gegenueber[0].Kennung != dasLeitungsgegenueber ||
		zeile.Karten == nil || zeile.Marken == nil || zeile.Geaendert == "" {
		t.Errorf("Zeile: %+v", zeile)
	}

	wunsch := einLeitungswunsch()
	wunsch.Positionen[0].Menge = 3

	if geaendert := anwendung.HandelEntwurfAendern(ctx, wunsch); geaendert.Fehler != nil ||
		geaendert.Vorgang.Positionen[0].Menge != 3 {
		t.Errorf("ändern: %+v", geaendert)
	}

	kopie := anwendung.HandelKopieren(ctx, derLeitungsvorgang)
	if kopie.Fehler != nil || kopie.Vorgang.Vorgang == derLeitungsvorgang {
		t.Fatalf("kopieren: %+v", kopie)
	}

	// **Was läuft, geht nicht ins Archiv.**
	if antwort := anwendung.HandelArchivieren(ctx, derLeitungsvorgang, true); antwort.Fehler == nil ||
		antwort.Fehler.Code != CodeKonflikt {
		t.Errorf("archivieren: %+v", antwort.Fehler)
	}

	if antwort := anwendung.HandelEntwurfLoeschen(ctx, derLeitungsvorgang); antwort.Fehler != nil {
		t.Errorf("löschen: %+v", antwort.Fehler)
	}

	// **Ein Tag und ein Zeitpunkt grenzen gleich ein**: heute bis heute.
	heute := zeile.Geaendert[:10]
	for _, filter := range []HandelsfilterDTO{
		{Seit: heute, Bis: heute},
		{Seit: heute + "T00:00:00Z", Lagen: []string{"dran"}},
	} {
		if liste = anwendung.HandelUebersicht(ctx, filter); liste.Fehler != nil || len(liste.Zeilen) != 1 ||
			liste.Zeilen[0].Vorgang.Vorgang != kopie.Vorgang.Vorgang {
			t.Errorf("nach %+v: %+v", filter, liste)
		}
	}
}

// **Jede Absage trägt eine leere Liste oder einen leeren Vorgang** — und
// den Grund am Feld, wo es eines gibt.
func TestDieListeSagtAb(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	for feld, filter := range map[string]HandelsfilterDTO{
		"seit":    {Seit: "gestern"},
		"bis":     {Bis: "morgen"},
		"ordnung": {Ordnung: "zufall"},
	} {
		antwort := anwendung.HandelUebersicht(ctx, filter)
		if antwort.Fehler == nil || antwort.Fehler.Feld != feld || antwort.Zeilen == nil {
			t.Errorf("%s: %+v", feld, antwort)
		}
	}

	if antwort := anwendung.HandelEntwurfAendern(ctx, einLeitungswunsch()); antwort.Fehler == nil ||
		antwort.Fehler.Code != CodeNichtGefunden {
		t.Errorf("ändern ohne Entwurf: %+v", antwort.Fehler)
	}

	if antwort := anwendung.HandelKopieren(ctx, "gibt-es-nicht"); antwort.Fehler == nil {
		t.Error("kopieren ohne Vorgang ging durch")
	}

	// **Ohne Identität gibt es kein „ich"** — und damit keine Liste.
	ohne := PeerflaecheUeber(bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	if antwort := ohne.HandelUebersicht(ctx, HandelsfilterDTO{}); antwort.Fehler == nil {
		t.Error("eine Liste ohne Identität")
	}

	if antwort := ohne.HandelEntwurfAendern(ctx, einLeitungswunsch()); antwort.Fehler == nil {
		t.Error("ändern ohne Identität")
	}

	if antwort := ohne.HandelKopieren(ctx, derLeitungsvorgang); antwort.Fehler == nil {
		t.Error("kopieren ohne Identität")
	}
}

// **Eine unlesbare Zeit kommt leer an** und nicht als Jahr eins — die
// Ablage macht aus ihr die Nullzeit.
func TestEineNullzeitIstLeer(t *testing.T) {
	if text := zeitAlsText(time.Time{}); text != "" {
		t.Errorf("die Nullzeit hieß %q", text)
	}
}
