package flaechen_test

// **Was zwischen zwei Instanzen schiefgehen kann, kontrolliert erzeugt** —
// QA-10.AK-02 und AK-03
// (#1741).
//
// # Was jeder Fall zeigt
//
// **Beide Enden sind beobachtbar**: bei Bert der Ausgang (`handelspost`, mit
// der Zahl offener Nachrichten und der Versuche), bei Anna der Vorgang
// (`handel stand`). Ein Fall ist erst bestanden, wenn beide Seiten sagen, was
// er behauptet — auch im Fehlschlag: dann steht die Nachricht **noch** im
// Ausgang und drüben **nichts**.
//
// # Was hier nicht geprüft wird
//
// **Ein Relais über DERP, ein NAT, eine Firewall, ein zweites Gerät** — der
// Vermittler ist ein Prozess auf dieser Maschine ([vermittler]). Das bleibt
// die Abnahme von Hand aus QA-10.AK-04.

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// zweiImLabor gibt Anna und Bert, die einander kennen und handeln dürfen.
func zweiImLabor(t *testing.T) (anna, bert einRechner) {
	t.Helper()

	anna, bert = stelleAuf(t), stelleAuf(t)

	kenneUndErlaube(t, anna, bert)
	kenneUndErlaube(t, bert, anna)

	return anna, bert
}

// einHandelImAusgang eröffnet bei Bert einen Handel mit Anna und legt ihn
// in den Ausgang — alles bei ihm, ohne Leitung.
func einHandelImAusgang(t *testing.T, anna, bert einRechner) {
	t.Helper()

	positionen := `[{"von":"` + anna.kennung +
		`","printing":"bolt-m10","menge":1,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"}]`

	mitEingabe(t, bert.umgebung, positionen,
		"handel", "eroeffnen", derDrahtvorgang, "tausch", anna.kennung)
	mussKommandozeile(t, bert.umgebung, "handel", "anbieten", derDrahtvorgang)
	mussKommandozeile(t, bert.umgebung,
		"handelspost-schicken", derDrahtvorgang, anna.kennung)
}

// ausgangBei liest, was bei einer Instanz für ein Gegenüber aussteht.
func ausgangBei(t *testing.T, rechner einRechner, gegenueber string) poststand {
	t.Helper()

	return poststandAus(t, mussKommandozeile(t, rechner.umgebung,
		"--json", "handelspost", gegenueber))
}

// zustellen trägt Berts Ausgang hinüber und gibt Rückgabewert und Klage.
func zustellen(t *testing.T, bert, anna einRechner) (code int, klage string) {
	t.Helper()

	aus, fehler, code := kommandozeile(t, bert.umgebung, "handelszustellung", anna.kennung)

	return code, aus + fehler
}

// nichtsBei sichert zu, dass der Vorgang bei Anna **nicht** steht.
func nichtsBei(t *testing.T, anna einRechner) {
	t.Helper()

	if aus, _, code := kommandozeile(t, anna.umgebung,
		"handel", "stand", derDrahtvorgang); code == 0 {
		t.Errorf("der Vorgang steht bei Anna, obwohl nichts ankommen konnte: %q", aus)
	}
}

// ersteZeile liest die Adresse, die ein Horcher als erste Zeile nennt.
func ersteZeile(t *testing.T, aus io.Reader) string {
	t.Helper()

	zeile, fehler := bufio.NewReader(aus).ReadString('\n')
	if fehler != nil {
		t.Fatalf("die Adresse lesen: %v", fehler)
	}

	adresse := strings.TrimSpace(zeile)
	if !strings.HasPrefix(adresse, "127.0.0.1:") {
		t.Fatalf("die genannte Adresse ist %q", adresse)
	}

	return adresse
}

// horcheMitStopp ist [horcheImHintergrund] mit einem Ende, das der Test
// selbst setzt — für den Neustart.
func horcheMitStopp(t *testing.T, rechner einRechner) (adresse string, protokoll *mitschrift, stopp func()) {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain, feste Argumente
	lauf := exec.CommandContext(ctx, gebaut.mse, "horchen", "127.0.0.1:0")
	lauf.Env = rechner.umgebung

	protokoll = &mitschrift{}
	lauf.Stderr = protokoll

	aus, fehler := lauf.StdoutPipe()
	if fehler != nil {
		t.Fatalf("die Ausgabe abgreifen: %v", fehler)
	}

	if gestartet := lauf.Start(); gestartet != nil {
		t.Fatalf("horchen starten: %v", gestartet)
	}

	adresse = ersteZeile(t, aus)

	stopp = func() {
		beenden()
		_ = lauf.Wait() //nolint:errcheck // der Abbruch ist der Weg hinaus
	}

	t.Cleanup(stopp)

	return adresse, protokoll, stopp
}

func TestEinHandelGehtUeberEinenVermittler(t *testing.T) {
	anna, bert := zweiImLabor(t)

	adresse, protokoll := horcheImHintergrund(t, anna)
	weg := neuerVermittler(t, adresse)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())
	einHandelImAusgang(t, anna, bert)

	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("über den Vermittler: %d %s", code, klage)
	}

	warteAufDenVorgang(t, anna, protokoll)

	if weg.durchgereicht.Load() == 0 {
		t.Error("die Zustellung lief nicht über den Vermittler")
	}
}

func TestBeiAusfallBleibtDerHandelImAusgang(t *testing.T) {
	anna, bert := zweiImLabor(t)

	adresse, protokoll := horcheImHintergrund(t, anna)
	weg := neuerVermittler(t, adresse)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())
	einHandelImAusgang(t, anna, bert)

	weg.sperren(true)

	if code, _ := zustellen(t, bert, anna); code == 0 {
		t.Fatal("die Zustellung meldete Erfolg, obwohl der Weg fort war")
	}

	// **Beide Enden sagen dasselbe**: hier steht es noch aus, drüben nichts.
	if stand := ausgangBei(t, bert, anna.kennung); stand.Offen != 1 {
		t.Errorf("nach dem Ausfall steht im Ausgang %+v", stand)
	}

	nichtsBei(t, anna)

	// **Und wenn der Weg zurück ist, geht dieselbe Nachricht.**
	weg.sperren(false)

	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("nach dem Ausfall: %d %s", code, klage)
	}

	warteAufDenVorgang(t, anna, protokoll)
}

func TestEinVerlorenerWegWirdWiederholt(t *testing.T) {
	anna, bert := zweiImLabor(t)

	adresse, protokoll := horcheImHintergrund(t, anna)
	weg := neuerVermittler(t, adresse)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())
	einHandelImAusgang(t, anna, bert)

	// **Die erste Verbindung geht verloren** — ohne ein Byte hinüber.
	weg.kappeDieNaechsten(1)

	if code, _ := zustellen(t, bert, anna); code == 0 {
		t.Fatal("die Zustellung meldete Erfolg über eine gekappte Verbindung")
	}

	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("die Wiederholung: %d %s", code, klage)
	}

	warteAufDenVorgang(t, anna, protokoll)
}

func TestLatenzHaeltDieZustellungNichtAuf(t *testing.T) {
	anna, bert := zweiImLabor(t)

	adresse, protokoll := horcheImHintergrund(t, anna)
	weg := neuerVermittler(t, adresse)
	weg.verzoegern(150 * time.Millisecond)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())
	einHandelImAusgang(t, anna, bert)

	anfang := time.Now()

	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("mit Latenz: %d %s", code, klage)
	}

	// **Und sie lag wirklich an** — sonst prüfte dieser Fall den schnellen Weg.
	if dauer := time.Since(anfang); dauer < 300*time.Millisecond {
		t.Errorf("die Zustellung dauerte %v — die Verzögerung griff nicht", dauer)
	}

	warteAufDenVorgang(t, anna, protokoll)
}

func TestNachDemNeustartDesHorchersGehtEsWeiter(t *testing.T) {
	anna, bert := zweiImLabor(t)

	adresse, _, stopp := horcheMitStopp(t, anna)
	weg := neuerVermittler(t, adresse)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())
	einHandelImAusgang(t, anna, bert)

	// **Anna ist fort** — der Anschluss horcht nicht mehr.
	stopp()

	if code, _ := zustellen(t, bert, anna); code == 0 {
		t.Fatal("die Zustellung meldete Erfolg, obwohl niemand horchte")
	}

	// **Und wieder da, auf einem anderen Anschluss.**
	neueAdresse, protokoll := horcheImHintergrund(t, anna)
	weg.umlenken(neueAdresse)

	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("nach dem Neustart: %d %s", code, klage)
	}

	warteAufDenVorgang(t, anna, protokoll)
}

func TestNachEinemSchluesselwechselGehtNichtsZumFalschen(t *testing.T) {
	anna, bert := zweiImLabor(t)

	einHandelImAusgang(t, anna, bert)

	// **Anna löst ihre Identität ab** und horcht unter der neuen.
	mussKommandozeile(t, anna.umgebung, "peer-rotieren")

	adresse, _ := horcheImHintergrund(t, anna)
	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, adresse)

	// **Bert erwartet die alte Kennung** — und der Handschlag ist auf sie
	// festgenagelt. Die Nachricht bleibt, wo sie ist.
	code, klage := zustellen(t, bert, anna)
	if code == 0 {
		t.Fatal("die Zustellung ging an eine Gegenstelle mit einer anderen Kennung")
	}

	if strings.TrimSpace(klage) == "" {
		t.Error("die Abweisung sagt nicht, woran es liegt")
	}

	if stand := ausgangBei(t, bert, anna.kennung); stand.Offen != 1 {
		t.Errorf("nach dem Schlüsselwechsel steht im Ausgang %+v", stand)
	}

	nichtsBei(t, anna)
}
