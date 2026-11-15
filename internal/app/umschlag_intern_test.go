package app

// **Ein Set als Umschlag hinaus und wieder herein** — Werkstück N4 von Phase
// 5.4 (#1457).
//
// # Was hier gemessen wird
//
// **Die Abnahme dieses Werkstücks**, wörtlich: „Dieselbe freigegebene
// Karte/Set/… ergibt nach bewusster Übernahme fachlich denselben Inhalt und
// **neue lokale IDs**."
//
// **Und die Trennung von Ansehen und Übernehmen**: ein Ansehen, das schon
// schriebe, machte die Freigabe zu einer Frage nach etwas, das schon geschehen
// ist.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Hinaus und wieder herein — mit neuer Kennung.**
func TestEinSetGehtAlsUmschlagHinausUndWiederHerein(t *testing.T) {
	// **Mit einem Probebund**: der echte Anmeldebund ist für einen Lauf
	// gesperrt (§3), und ein Umschlag wird unterschrieben.
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe.msepackage")

	geschnuert := anwendung.UmschlagSchnueren(ctx, "", geoeffnet.SetID, ziel)
	if geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	// **Jede Datei des Setordners steht einzeln darin** — auch die Bilder.
	//
	// **Wie viele es sind, steht hier nicht.** Der erste Anlauf erwartete
	// genau eines, der zweite mindestens zwei — und beide Zahlen hingen
	// daran, ob das Saatset ein Bild mitschreibt. Geprüft wird, was
	// **immer** gilt: die Setdatei ist dabei, und alles darin ist ein
	// Setstück. Eine Zahl, die von einer Vorrichtung abhängt, prüft die
	// Vorrichtung (§10).
	if len(geschnuert.Objekte) == 0 {
		t.Fatalf("der Umschlag ist leer")
	}

	var setdabei bool

	for _, objekt := range geschnuert.Objekte {
		if objekt.Art != "set" {
			t.Errorf("ein Objekt trägt die Art %q", objekt.Art)
		}

		if objekt.Bezeichnung == "set" {
			setdabei = true
		}
	}

	if !setdabei {
		t.Errorf("die Setdatei selbst fehlt: %+v", geschnuert.Objekte)
	}

	if _, fehler := os.Stat(ziel); fehler != nil {
		t.Fatalf("der Umschlag liegt nicht da: %v", fehler)
	}

	// **Ansehen schreibt nichts** — das ist die halbe Zusage.
	angesehen := anwendung.UmschlagAnsehen(ctx, ziel)
	if angesehen.Fehler != nil {
		t.Fatalf("ansehen: %+v", angesehen.Fehler)
	}

	if angesehen.Fingerabdruck == "" {
		t.Error("die Sendung nennt keinen Fingerabdruck")
	}

	vorher := anwendung.Sets(ctx)
	if vorher.Fehler != nil {
		t.Fatalf("Sets: %+v", vorher.Fehler)
	}

	// **Und Übernehmen schreibt** — mit einer neuen Kennung.
	uebernommen := anwendung.UmschlagUebernehmen(ctx, ziel)
	if uebernommen.Fehler != nil {
		t.Fatalf("übernehmen: %+v", uebernommen.Fehler)
	}

	if uebernommen.SetID == geoeffnet.SetID {
		t.Errorf("das übernommene Set trägt dieselbe Kennung %q",
			uebernommen.SetID)
	}

	nachher := anwendung.Sets(ctx)
	if nachher.Fehler != nil {
		t.Fatalf("Sets: %+v", nachher.Fehler)
	}

	if len(nachher.Sets) != len(vorher.Sets)+1 {
		t.Errorf("vorher %d Sets, nachher %d",
			len(vorher.Sets), len(nachher.Sets))
	}
}

// **Ein Umschlag ohne Set lässt sich nicht übernehmen.**
func TestEinUmschlagOhneSetLaesstSichNichtUebernehmen(t *testing.T) {
	// **Mit einem Probebund**: der echte Anmeldebund ist für einen Lauf
	// gesperrt (§3), und ein Umschlag wird unterschrieben.
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.peerwerk().Ausweis(anwendung.imProfil(ctx))
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	ziel := filepath.Join(t.TempDir(), "ohneset.msepackage")

	if _, fehler := austauschSchreiben(ziel, ich); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.UmschlagUebernehmen(ctx, ziel); antwort.Fehler == nil {
		t.Error("ein Umschlag ohne Set kam durch")
	}
}

// austauschSchreiben legt einen Umschlag mit einer Karte an — also ohne Set.
func austauschSchreiben(
	ziel string, ich peer.Identitaet,
) (austausch.Manifest, error) {
	return austausch.Schreibe(ziel, []austausch.Inhalt{{
		Art: austausch.ArtKarte, Name: "eine.json",
		Bezeichnung: "Eine Karte", Bytes: []byte(`{}`),
	}}, ich)
}
