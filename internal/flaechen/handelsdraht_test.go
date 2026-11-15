package flaechen_test

// **Ein Handel über einen echten Draht, zwischen zwei Prozessen** —
// #1634,
// entschieden in
// [0121](../../docs/entscheidungen/0121-eine-leitung-ohne-fenster.md).
//
// # Was hier gemessen wird
//
// **Zwei `mse`-Prozesse mit getrennten Beständen, ein TCP-Anschluss mit
// TLS 1.3, und eine Handelsnachricht, die wirklich ankommt.** Bis zu diesem
// Werkstück endete jede Prüfung eines Transports bei einem Doppel oder bei
// der Schleife im selben Prozess — `internal/leitung.Schleife`.
//
// **Das ist die Ebene, die `docs/HAERTUNG.md` gefehlt hat**: zwischen der
// Pipeline im Testprozess und dem Lauf gegen einen fremden Dienst stand
// nichts.
//
// # Was er **nicht** misst
//
// **Zwei Rechner.** Beide Prozesse laufen auf dieser Maschine, und damit
// bleibt ungeprüft, was ein echtes Netz dazwischenlegt — Namensauflösung,
// MTU, eine Fritz!Box. Das ist `T12`
// (#1477),
// und es bleibt eine Handprüfung
// ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// derDrahtvorgang ist der Handel, der über die Leitung geht.
const derDrahtvorgang = "v-draht"

// einRechner ist eine Umgebung mit eigenem Bestand, eigenem Bund und eigener
// Identität.
type einRechner struct {
	umgebung []string
	kennung  string
	abdruck  string
}

// stelleAuf legt einen Rechner an und gibt seine Identität.
func stelleAuf(t *testing.T) einRechner {
	t.Helper()

	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))

	angelegt := mussKommandozeile(t, umgebung, "peer-anlegen")

	return einRechner{
		umgebung: umgebung,
		kennung:  zeilenwert(t, angelegt, "kennung"),
		abdruck:  zeilenwert(t, angelegt, "fingerabdruck"),
	}
}

// kenneUndErlaube trägt das Gegenüber ein, bestätigt es und gibt den Handel
// frei.
//
// **Alle drei Schritte, und keiner davon ist überflüssig**: ein Eintrag darf
// nichts, eine Bestätigung gibt keinen Zweck frei, und ohne Zweck kommt keine
// Sitzung zustande (N1, T3).
func kenneUndErlaube(t *testing.T, hier, dort einRechner) {
	t.Helper()

	mussKommandozeile(t, hier.umgebung,
		"kontakt-aufnehmen", dort.kennung, "Der andere")
	mussKommandozeile(t, hier.umgebung,
		"kontakt-bestaetigen", dort.kennung, dort.abdruck)
	mussKommandozeile(t, hier.umgebung,
		"zweck-erlauben", dort.kennung, "handel")
}

// horcheImHintergrund startet `mse horchen` und gibt die wirkliche Adresse.
//
// **Der Befehl bleibt, bis der Kontext endet** — das ist der Punkt dieses
// Werkstücks. Die erste Zeile seiner Ausgabe ist die Adresse; alles Weitere
// ist Protokoll und geht auf die Fehlerausgabe (0121).
func horcheImHintergrund(
	t *testing.T, rechner einRechner,
) (string, *mitschrift) {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain, feste Argumente
	lauf := exec.CommandContext(ctx, gebaut.mse, "horchen", "127.0.0.1:0")
	lauf.Env = rechner.umgebung

	protokoll := &mitschrift{}
	lauf.Stderr = protokoll

	aus, fehler := lauf.StdoutPipe()
	if fehler != nil {
		t.Fatalf("die Ausgabe abgreifen: %v", fehler)
	}

	if gestartet := lauf.Start(); gestartet != nil {
		t.Fatalf("horchen starten: %v", gestartet)
	}

	t.Cleanup(func() {
		beenden()
		_ = lauf.Wait() //nolint:errcheck // der Abbruch ist der Weg hinaus
	})

	leser := bufio.NewReader(aus)

	zeile, fehler := leser.ReadString('\n')
	if fehler != nil {
		t.Fatalf("die Adresse lesen: %v", fehler)
	}

	adresse := strings.TrimSpace(zeile)
	if adresse == "" || !strings.HasPrefix(adresse, "127.0.0.1:") {
		t.Fatalf("die genannte Adresse ist %q", adresse)
	}

	return adresse, protokoll
}

func TestEinHandelGehtUeberEinenEchtenDraht(t *testing.T) {
	anna := stelleAuf(t)
	bert := stelleAuf(t)

	kenneUndErlaube(t, anna, bert)
	kenneUndErlaube(t, bert, anna)

	// **Anna horcht**, und Bert trägt ihre wirkliche Adresse ein.
	adresse, protokoll := horcheImHintergrund(t, anna)

	mussKommandozeile(t, bert.umgebung,
		"kontakt-adresse", anna.kennung, adresse)

	// **Bert eröffnet und bietet an** — alles bei ihm, ohne Leitung.
	positionen := `[{"von":"` + anna.kennung +
		`","printing":"bolt-m10","menge":1,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"}]`

	mitEingabe(t, bert.umgebung, positionen,
		"handel", "eroeffnen", derDrahtvorgang, "tausch", anna.kennung)
	mussKommandozeile(t, bert.umgebung,
		"handel", "anbieten", derDrahtvorgang)

	// **Und legt den Stand in den Ausgang.**
	mussKommandozeile(t, bert.umgebung,
		"handelspost-schicken", derDrahtvorgang, anna.kennung)

	// **Jetzt geht er über die Leitung.**
	getragen := mussKommandozeile(t, bert.umgebung,
		"handelszustellung", anna.kennung)
	if !strings.Contains(getragen, "\t1") {
		t.Fatalf("getragen wurde: %q", getragen)
	}

	// **Und bei Anna steht der Vorgang** — sie hat ihn nie eröffnet.
	//
	// **Über eine Warteschleife**, denn die Zustellung ist drüben eine
	// eigene Goroutine: der Befehl kommt zurück, sobald die Bytes draußen
	// sind, und nicht, wenn sie verbucht sind.
	warteAufDenVorgang(t, anna, protokoll)
}

// warteAufDenVorgang liest bei Anna, bis der Handel dasteht.
//
// **Das Protokoll des Horchers gehört in die Klage.** Was drüben schiefging,
// steht auf seiner Fehlerausgabe und nirgends sonst; ohne diese Zeilen sagt
// ein roter Lauf nur, dass nichts ankam (§4).
func warteAufDenVorgang(
	t *testing.T, anna einRechner, protokoll *mitschrift,
) {
	t.Helper()

	for versuch := range 50 {
		aus, klage, code := kommandozeile(t, anna.umgebung,
			"handel", "stand", derDrahtvorgang)
		if code == 0 && strings.Contains(aus, "angeboten") {
			return
		}

		if versuch == 49 {
			t.Fatalf("der Vorgang kam nicht an: %q / %q\nder Horcher sagt:\n%s",
				aus, klage, protokoll.text())
		}

		time.Sleep(100 * time.Millisecond)
	}
}
