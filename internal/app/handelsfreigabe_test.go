package app

// **Die Freigabe eines verbindlichen Handelsschritts** — Werkstück `T10` von
// Phase 5.6
// (#1475).
//
// **Ein interner Lauf**, weil er `laufendeVerbindung` und `imProfil` braucht
// — dieselbe Gestalt wie `handelsverbuchungnaht_intern_test.go` daneben.
//
// # Was hier geprüft wird
//
//   - Ein Schritt, den es nicht gibt, bekommt keine Freigabe.
//   - **Ohne Bestand auch keine** — die Bindung sagt es, statt abzustürzen.
//   - Und ein Vorgang, den es nicht gibt, ebenfalls nicht: ein Mensch soll
//     seine Zustimmung nicht an einem Handel verbrauchen, den es nicht gibt.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein unbekannter Schritt bekommt keine Freigabe.**
func TestEinUnbekannterSchrittBekommtKeineFreigabe(t *testing.T) {
	anwendung := PeerflaecheUeber(bestandFuerDenRiegel(t))

	antwort := anwendung.HandelFreigabeErteilen(
		anwendung.imProfil(riegelkontext(t)), "v-1", "tanzen", "")
	if antwort.Fehler == nil {
		t.Fatal("der Schritt bekam eine Freigabe")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "tanzen") {
		t.Errorf("die Absage nennt den Schritt nicht: %s", antwort.Fehler.Meldung)
	}
}

// **Ohne Bestand gibt es keine Freigabe — und keinen Absturz.**
func TestOhneBestandGibtEsKeineFreigabe(t *testing.T) {
	anwendung := NeueApp("")

	if antwort := anwendung.HandelFreigabeErteilen(
		t.Context(), "v-1", "verbuchen", ""); antwort.Fehler == nil {
		t.Error("ohne Bestand kam eine Freigabe heraus")
	}
}

// **Einen Vorgang, den es nicht gibt, gibt es nicht freizugeben.**
//
// **Sonst verbrauchte ein Mensch seine Zustimmung an einem Handel, den es
// nicht gibt** — und beim nächsten Klick fehlte sie.
func TestEinUnbekannterVorgangBekommtKeineFreigabe(t *testing.T) {
	anwendung := PeerflaecheUeber(bestandFuerDenRiegel(t))

	if antwort := anwendung.HandelFreigabeErteilen(
		anwendung.imProfil(riegelkontext(t)), "v-gibtesnicht", "verbuchen", "",
	); antwort.Fehler == nil {
		t.Error("der unbekannte Vorgang bekam eine Freigabe")
	}
}

// **Über die Kommandozeile gibt es keine Freigabe.**
//
// **Das ist die Zusage aus 0078**, und sie steht nicht in einer Aufzählung,
// sondern im Modus: `Nachweise.Ausstellen` weist jeden ab, der nicht
// `fenster` ist. Dieser Lauf betritt genau den Zweig — und er ist derselbe,
// über den auch MCP und der Dienst stolpern.
//
// **Der Vorgang ist dabei echt**: sonst endete der Ruf schon davor, und
// gemessen wäre das Lesen und nicht der Riegel.
func TestUeberDieKommandozeileGibtEsKeineFreigabe(t *testing.T) {
	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	legeHandelHin(t, anwendung.db, ich.Kennung())

	antwort := anwendung.HandelFreigabeErteilen(
		service.MitDienstmodus(ctx, domain.ModusKommandozeile),
		derNahtvorgang, "verbuchen", "")
	if antwort.Fehler == nil {
		t.Fatal("die Kommandozeile bekam eine Freigabe")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Fenster") {
		t.Errorf("die Absage nennt das Fenster nicht: %s", antwort.Fehler.Meldung)
	}
}
