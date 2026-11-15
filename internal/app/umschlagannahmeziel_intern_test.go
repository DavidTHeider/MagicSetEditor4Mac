package app

// **Karte und Vorlage aus einem Umschlag annehmen** —
// #1639.
//
// # Was hier gemessen wird
//
//   - **Dass dieselbe Karte in ein anderes Set kommt** — mit **neuer**
//     lokaler Kennung und ohne die alte zu berühren. Eine Kennung aus einem
//     fremden Bestand ist hier keine; wer sie übernähme, überschriebe eine
//     Karte, von der der Absender nichts weiß.
//   - **Dass eine Vorlage in ein gewähltes Paket geht** — und dass ein Paket,
//     das es schon gibt, **nicht** überschrieben wird.
//
// # Und warum das Ziel im Ruf steht
//
// **Weil der Umschlag es nicht nennt.** Eine Karte gehört in ein Set, eine
// Vorlage in ein Paket; welches, entscheidet ein Mensch. Ein Weg, der es
// riete, legte Karten in irgendein Set.

import (
	"strings"
	"testing"
)

// **Eine Karte kommt in ein anderes Set — mit neuer Kennung.**
func TestEineKarteKommtAusDemUmschlagInEinAnderesSet(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Kartenzeilen(ctx, geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Zeilen) == 0 {
		t.Fatalf("Kartenzeilen: %+v", karten.Fehler)
	}

	herkunft := karten.Zeilen[0].ID
	ziel := einUmschlagsziel(t, "karte")

	if geschnuert := anwendung.UmschlagSchnueren(
		ctx, "karte", herkunft, ziel); geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	// **Ein zweites Set, und zwar wirklich ein anderes** — sonst sagte der
	// Lauf nur, dass eine Karte zweimal im selben Set steht.
	zweites := anwendung.SetAnlegen(ctx, "Zweites", "probe", "probe-stil")
	if zweites.Fehler != nil {
		t.Fatalf("das zweite Set: %+v", zweites.Fehler)
	}

	angenommen := anwendung.UmschlagKarte(ctx, ziel, zweites.ID)
	if angenommen.Fehler != nil {
		t.Fatalf("annehmen: %+v", angenommen.Fehler)
	}

	if angenommen.KarteID == "" || angenommen.KarteID == herkunft {
		t.Fatalf("die neue Kennung ist %q, die alte war %q",
			angenommen.KarteID, herkunft)
	}

	// **Das Herkunftsset hat genauso viele Karten wie vorher.**
	vorher := len(karten.Zeilen)

	nachher := anwendung.Kartenzeilen(ctx, geoeffnet.SetID)
	if nachher.Fehler != nil || len(nachher.Zeilen) != vorher {
		t.Errorf("das Herkunftsset führt jetzt %d statt %d Karten",
			len(nachher.Zeilen), vorher)
	}

	// **Und die Werte sind angekommen** — eine Karte ohne ihre Felder wäre
	// eine leere Zeile mit einer Kennung.
	//
	// **Verglichen mit der Herkunft und nicht nur gezählt**: eine Karte mit
	// irgendwelchen Werten wäre keine Antwort darauf, ob **diese** ankam.
	quelle := anwendung.Karte(ctx, herkunft)
	if quelle.Fehler != nil {
		t.Fatalf("die Herkunftskarte: %+v", quelle.Fehler)
	}

	angekommen := anwendung.Karte(ctx, angenommen.KarteID)
	if angekommen.Fehler != nil {
		t.Fatalf("die angenommene Karte: %+v", angekommen.Fehler)
	}

	if len(angekommen.Karte.Werte) != len(quelle.Karte.Werte) {
		t.Fatalf("angekommen sind %d Werte, hinausgegangen %d",
			len(angekommen.Karte.Werte), len(quelle.Karte.Werte))
	}

	if len(angekommen.Karte.Werte) == 0 {
		t.Fatal("die Herkunftskarte trägt keinen Wert — der Lauf greift daneben")
	}

	gefunden := map[string]string{}
	for _, wert := range angekommen.Karte.Werte {
		gefunden[wert.Feld] = wert.Wert
	}

	for _, wert := range quelle.Karte.Werte {
		if gefunden[wert.Feld] != wert.Wert {
			t.Errorf("das Feld %q trägt %q statt %q",
				wert.Feld, gefunden[wert.Feld], wert.Wert)
		}
	}
}

// **Eine Vorlage kommt in ein gewähltes Paket — und überschreibt keines.**
func TestEineVorlageKommtAusDemUmschlagInEinGewaehltesPaket(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	// **Zum Schnüren die Saat, zum Annehmen der eigene Paketort.**
	//
	// Eine Werkstatt hat **eine** Wurzel; das Paket, aus dem die Vorlage
	// kommt, liegt in der Saat, und das Paket, in das sie soll, entsteht
	// dort, wo dieses Profil seine Pakete führt. Im Betrieb ist es eine
	// Wurzel — hier sind es zwei, weil der Lauf die Herkunft nicht erst
	// installieren soll.
	anwendung.paketdateien = werkstattUeber(t, saatpaket(t))

	ziel := einUmschlagsziel(t, "vorlage")

	if geschnuert := anwendung.UmschlagSchnueren(
		ctx, "vorlage", "probe.mse-game", ziel); geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	// **Und jetzt die Wurzel, in die `PaketAnlegen` schreibt** — im Betrieb
	// ist es dieselbe, hier zeigt der Lauf sie ausdrücklich hin.
	paketort, fehler := anwendung.paketort()
	if fehler != nil {
		t.Fatalf("der Paketort: %v", fehler)
	}

	anwendung.paketdateien = werkstattUeber(t, paketort)

	angenommen := anwendung.UmschlagVorlage(ctx, ziel, "angenommen.mse-game")
	if angenommen.Fehler != nil {
		t.Fatalf("annehmen: %+v", angenommen.Fehler)
	}

	if angenommen.Paket == nil || angenommen.Paket.Name != "angenommen" {
		t.Fatalf("angelegt wurde %+v", angenommen.Paket)
	}

	// **Gelesen wird über das Verzeichnis** — der Name ist `angenommen`, auf
	// der Platte liegt `angenommen.mse-game`.
	dateien := anwendung.Paketdateien(ctx, "angenommen.mse-game")
	if dateien.Fehler != nil || len(dateien.Dateien) == 0 {
		t.Fatalf("das angenommene Paket führt keine Datei: %+v", dateien.Fehler)
	}

	// **Ein zweites Mal unter demselben Namen wird abgewiesen** — und die
	// Absage nennt den Namen, denn den hat ein Mensch gewählt.
	nochmal := anwendung.UmschlagVorlage(ctx, ziel, "angenommen.mse-game")
	if nochmal.Fehler == nil {
		t.Fatal("das zweite Annehmen kam durch")
	}

	if !strings.Contains(nochmal.Fehler.Meldung, "angenommen") {
		t.Errorf("die Absage nennt den Namen nicht: %s", nochmal.Fehler.Meldung)
	}
}
