package app

// **Wer am Tisch saß** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// # Was hier gemessen wird
//
// **Dass ein bekanntes Gegenüber wiedererkannt wird** — über den
// Fingerabdruck, ohne zweite Paarung.
//
// **Dass „kenne ich nicht" kein Fehler ist.** Wer als Gast an einem Tisch
// saß, muss nicht in der Kontaktliste stehen.
//
// **Und dass ein leerer Sitz eine Auskunft bekommt und keinen leeren
// Kontakt.**

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

func TestEinBekanntesGegenueberWirdAmSitzWiedererkannt(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := anwendung.imProfil(t.Context())

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(ctx, string(fremd.Kennung()),
		"Turfer"); antwort.Fehler != nil {
		t.Fatalf("Kontakt: %+v", antwort.Fehler)
	}

	// **Der Sitz kommt aus der Freigabe** — ohne laufende Freigabe sitzt
	// dort niemand, und genau das ist die erste Auskunft.
	leer := anwendung.Handelspartner(ctx, "a")
	if leer.Fehler == nil {
		t.Error("ein Sitz ohne Freigabe gab kein Nein")
	}

	if leer.Bekannt {
		t.Error("ein Sitz ohne Freigabe gab ein Gegenüber")
	}

	// **Und der Suchweg selbst erkennt den Fingerabdruck.**
	gefunden := anwendung.kontaktZumAbdruck(ctx, "a", fremd.Fingerabdruck())
	if !gefunden.Bekannt {
		t.Fatalf("das bekannte Gegenüber wurde nicht erkannt: %+v", gefunden)
	}

	if gefunden.Kontakt.Kennung != string(fremd.Kennung()) {
		t.Errorf("es kam %q", gefunden.Kontakt.Kennung)
	}

	// **Ein fremder Abdruck ist kein Fehler** — er ist die Auskunft „kenne
	// ich nicht".
	unbekannt := anwendung.kontaktZumAbdruck(ctx, "a", "0000-0000-0000-0000")
	if unbekannt.Fehler != nil {
		t.Errorf("ein unbekannter Abdruck gab einen Fehler: %+v", unbekannt.Fehler)
	}

	if unbekannt.Bekannt {
		t.Error("ein unbekannter Abdruck gab ein Gegenüber")
	}
}

// **Die Fehlerwege des Sitzpartners.**
//
// **Ohne Profil geht gar nichts**, und ein Sitz, den es nicht gibt, bekommt
// eine Auskunft statt eines leeren Kontakts.
func TestJederWegDesSitzpartnersReichtSeinenFehlerDurch(t *testing.T) {
	anwendung := mitIdentitaet(t)

	// **Ohne Profil klagt schon der Freigabestand.**
	ohne := anwendung.Handelspartner(t.Context(), "a")
	if ohne.Fehler == nil {
		t.Error("der Sitz kam ohne laufende Freigabe durch")
	}

	if ohne.Kontakt.Rechte == nil {
		t.Error("die Rechte sind nil statt leer")
	}

	// **Und der Kontaktweg klagt, wenn der Bestand fort ist.**
	geschlossen := mitIdentitaet(t)
	if fehler := geschlossen.db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	klagend := geschlossen.kontaktZumAbdruck(
		geschlossen.imProfil(t.Context()), "a", "0000")
	if klagend.Fehler == nil {
		t.Error("der geschlossene Bestand kam durch")
	}

	if klagend.Kontakt.Rechte == nil {
		t.Error("die Rechte sind nil statt leer")
	}
}

// **Wer auf einem Sitz sitzt, steht in der Vergabe.**
//
// **Einzeln gefahren**, weil eine laufende Freigabe mit besetztem Sitz zwei
// Rechner braucht — die Frage selbst braucht eine Liste.
func TestDerAbdruckAmSitzWirdGefunden(t *testing.T) {
	t.Parallel()

	vergeben := []SitzvergabeDTO{
		{Sitz: "a", Fingerabdruck: "1111"},
		{Sitz: "b", Fingerabdruck: "2222"},
	}

	if abdruck := abdruckAmSitz(vergeben, "b"); abdruck != "2222" {
		t.Errorf("auf b sitzt %q", abdruck)
	}

	// **Ein Sitz, den niemand hat, gibt nichts** — und das ist die Auskunft,
	// aus der „da sitzt niemand von außen" wird.
	if abdruck := abdruckAmSitz(vergeben, "c"); abdruck != "" {
		t.Errorf("auf c sitzt %q", abdruck)
	}

	if abdruck := abdruckAmSitz(nil, "a"); abdruck != "" {
		t.Errorf("ohne Vergabe sitzt %q", abdruck)
	}
}

// **Und ein klagender Freigabestand kommt als Klage an.**
func TestEinKlagenderFreigabestandKommtAlsKlageAn(t *testing.T) {
	anwendung := mitIdentitaet(t)

	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	antwort := anwendung.Handelspartner(t.Context(), "a")
	if antwort.Fehler == nil {
		t.Error("der geschlossene Bestand kam durch")
	}

	if antwort.Kontakt.Rechte == nil {
		t.Error("die Rechte sind nil statt leer")
	}
}
