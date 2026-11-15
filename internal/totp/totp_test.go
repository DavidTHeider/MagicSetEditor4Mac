package totp_test

// **Die Vektoren aus RFC 6238** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Warum gegen die RFC und nicht gegen eine eigene Erwartung
//
// **Weil eine selbst gerechnete Erwartung nichts belegt** (§10, §15): sie
// prüfte, dass dieselbe Rechnung zweimal dasselbe gibt. Anhang B der RFC
// nennt acht Zeitpunkte mit ihren Ziffern für drei Verfahren — das ist die
// Probe, ob **diese** Umsetzung dieselbe ist wie jede andere.
//
// **Das Geheimnis der Vektoren ist ASCII**, nicht base32: „12345678901234567890"
// für SHA-1, für SHA-256 auf 32 und für SHA-512 auf 64 Zeichen fortgesetzt
// (RFC 6238, Anhang B, Fußnote zur Tabelle).

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/totp"
)

// Die Geheimnisse aus Anhang B — ASCII, nicht base32.
const (
	seed1   = "12345678901234567890"
	seed256 = "12345678901234567890" + "123456789012"
	seed512 = "12345678901234567890" + "12345678901234567890" +
		"12345678901234567890" + "1234"
)

// **Jeder Vektor aus Anhang B stimmt.**
func TestDieVektorenAusRFC6238Stimmen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		sekunden  int64
		verfahren totp.Verfahren
		geheimnis string
		erwartet  string
	}{
		{59, totp.SHA1, seed1, "94287082"},
		{59, totp.SHA256, seed256, "46119246"},
		{59, totp.SHA512, seed512, "90693936"},
		{1111111109, totp.SHA1, seed1, "07081804"},
		{1111111109, totp.SHA256, seed256, "68084774"},
		{1111111109, totp.SHA512, seed512, "25091201"},
		{1111111111, totp.SHA1, seed1, "14050471"},
		{1111111111, totp.SHA256, seed256, "67062674"},
		{1111111111, totp.SHA512, seed512, "99943326"},
		{1234567890, totp.SHA1, seed1, "89005924"},
		{1234567890, totp.SHA256, seed256, "91819424"},
		{1234567890, totp.SHA512, seed512, "93441116"},
		{2000000000, totp.SHA1, seed1, "69279037"},
		{2000000000, totp.SHA256, seed256, "90698825"},
		{2000000000, totp.SHA512, seed512, "38618901"},
		{20000000000, totp.SHA1, seed1, "65353130"},
		{20000000000, totp.SHA256, seed256, "77737706"},
		{20000000000, totp.SHA512, seed512, "47863826"},
	}

	for _, fall := range faelle {
		t.Run(fall.erwartet, func(t *testing.T) {
			t.Parallel()

			gerechnet := totp.Code(totp.Zaehler{
				Geheimnis: []byte(fall.geheimnis),
				Verfahren: fall.verfahren,
				Stellen:   8,
			}, time.Unix(fall.sekunden, 0))

			if gerechnet != fall.erwartet {
				t.Errorf("bei %d gab %s die Ziffern %q, erwartet war %q",
					fall.sekunden, fall.verfahren, gerechnet, fall.erwartet)
			}
		})
	}
}

// **Sechs Stellen sind der Normalfall** — was jede Authenticator-App zeigt.
//
// **Die Vektoren der RFC haben acht**, und das ist kein Widerspruch: die
// Stellenzahl ist ein Parameter, und Anhang B nennt ausdrücklich acht. Wer
// beides prüft, hat die Kürzung geprüft und nicht nur die Rechnung.
func TestSechsStellenSindDieLetztenSechs(t *testing.T) {
	t.Parallel()

	zaehler := totp.Zaehler{Geheimnis: []byte(seed1), Verfahren: totp.SHA1}

	acht := totp.Code(totp.Zaehler{
		Geheimnis: []byte(seed1), Verfahren: totp.SHA1, Stellen: 8,
	}, time.Unix(59, 0))

	sechs := totp.Code(zaehler, time.Unix(59, 0))

	if sechs != acht[len(acht)-6:] {
		t.Errorf("sechs Stellen sind %q, acht sind %q", sechs, acht)
	}
}

// **Ein Code stimmt, solange er im Fenster liegt — und daneben nicht.**
//
// # Das Fenster ist eine Entscheidung und keine Vorgabe
//
// **RFC 6238 §5.2 überlässt es dem Prüfer** und rät nur, es klein zu halten.
// Hier ist es **ein** Schritt in jede Richtung: das deckt eine Uhr ab, die um
// bis zu dreißig Sekunden abweicht, und verlängert das Zeitfenster eines
// abgefangenen Codes auf höchstens neunzig Sekunden.
func TestEinCodeStimmtNurInSeinemFenster(t *testing.T) {
	t.Parallel()

	zaehler := totp.Zaehler{Geheimnis: []byte(seed1), Verfahren: totp.SHA1}
	jetzt := time.Unix(1111111111, 0)

	faelle := map[string]struct {
		versatz time.Duration
		gilt    bool
	}{
		"derselbe Schritt":   {0, true},
		"ein Schritt zurück": {-30 * time.Second, true},
		"ein Schritt vor":    {30 * time.Second, true},
		"zwei zurück":        {-60 * time.Second, false},
		"zwei vor":           {60 * time.Second, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			code := totp.Code(zaehler, jetzt.Add(fall.versatz))

			fehler := totp.Stimmt(zaehler, code, jetzt)
			if fall.gilt != (fehler == nil) {
				t.Errorf("der Code von %s gab %v", name, fehler)
			}
		})
	}
}

// **Ein Code, der keiner ist, wird abgewiesen — und nicht gerechnet.**
func TestEinKrummerCodeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	zaehler := totp.Zaehler{Geheimnis: []byte(seed1), Verfahren: totp.SHA1}

	for name, code := range map[string]string{
		"leer":            "",
		"zu kurz":         "12345",
		"zu lang":         "1234567",
		"mit Buchstaben":  "12345a",
		"mit Leerzeichen": "123 456",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := totp.Stimmt(zaehler, code, time.Unix(1111111111, 0)); fehler == nil {
				t.Error("angenommen")
			}
		})
	}
}

// **Zwei Geheimnisse sind nie dasselbe.**
func TestJedesGeheimnisIstEinAnderes(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for range 32 {
		frisch := totp.NeuesGeheimnis()

		if len(frisch) != totp.Geheimnislaenge {
			t.Fatalf("das Geheimnis ist %d Byte lang", len(frisch))
		}

		if gesehen[string(frisch)] {
			t.Fatal("zweimal dasselbe Geheimnis")
		}

		gesehen[string(frisch)] = true
	}
}

// **Die Adresse für den Scanner ist die aus dem Key-Uri-Format.**
//
// **Nachgelesen**, nicht erfunden: `otpauth://totp/<Aussteller>:<Konto>?…`
// mit `secret`, `issuer`, `algorithm`, `digits` und `period` — die Felder,
// die Google Authenticator, 1Password und Aegis lesen.
func TestDieAdresseFuerDenScannerTraegtIhreFelder(t *testing.T) {
	t.Parallel()

	// **Ohne Verfahren und ohne Stellenzahl**: die Adresse muss beide
	// Vorgaben selbst einsetzen, denn genau so kommt ein frisch gezogenes
	// Geheimnis aus [totp.NeuesGeheimnis] hierher.
	adresse := totp.Adresse(totp.Zaehler{Geheimnis: []byte(seed1)}, "Ich")

	for _, teil := range []string{
		"otpauth://totp/MagicSetEditor4Mac:Ich?",
		"algorithm=SHA1",
		"digits=6",
		"issuer=MagicSetEditor4Mac",
		"period=30",
		"secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
	} {
		if !strings.Contains(adresse, teil) {
			t.Errorf("%q fehlt in %q", teil, adresse)
		}
	}
}

// **Ein Konto mit Sonderzeichen zerbricht die Adresse nicht.**
func TestEinKontoMitSonderzeichenWirdKodiert(t *testing.T) {
	t.Parallel()

	adresse := totp.Adresse(totp.Zaehler{
		Geheimnis: []byte(seed1), Verfahren: totp.SHA1,
	}, "wer@example.org/mit Leerzeichen")

	if strings.Contains(adresse, " ") {
		t.Errorf("ein Leerzeichen steht roh in %q", adresse)
	}

	if strings.Count(adresse, "?") != 1 {
		t.Errorf("die Abfrage beginnt mehr als einmal in %q", adresse)
	}
}

// **Ein Geheimnis lässt sich schreiben und wieder lesen.**
//
// **Als base32 ohne Auffüllen** — die Gestalt, in der jede Authenticator-App
// es entgegennimmt und in der es sich im Schlüsselbund ablegen lässt.
func TestEinGeheimnisReistAlsBase32(t *testing.T) {
	t.Parallel()

	frisch := totp.NeuesGeheimnis()

	zurueck, fehler := totp.GeheimnisAus(totp.AlsText(frisch))
	if fehler != nil {
		t.Fatalf("zurücklesen: %v", fehler)
	}

	if !bytes.Equal(zurueck, frisch) {
		t.Error("das Geheimnis kam anders zurück")
	}

	if _, krumm := totp.GeheimnisAus("kein base32 !!"); !errors.Is(krumm, domain.ErrValidation) {
		t.Errorf("angenommen wurde: %v", krumm)
	}
}

// **Der passende Schritt ist der des Codes, nicht der der Uhr** — SEC-09
// (#2150): wer ihn verbraucht, verbraucht genau diesen, und ein Code aus dem
// Fenster davor zählt als der frühere.
func TestDerPassendeSchrittIstDerDesCodes(t *testing.T) {
	zaehler := totp.Zaehler{Geheimnis: []byte("12345678901234567890")}
	jetzt := time.Unix(59, 0)

	for _, fall := range []struct {
		name    string
		von     time.Time
		schritt int64
	}{
		{"der jetzige", jetzt, 1},
		{"der davor", jetzt.Add(-totp.Schritt), 0},
		{"der danach", jetzt.Add(totp.Schritt), 2},
	} {
		schritt, err := totp.PassenderSchritt(zaehler, totp.Code(zaehler, fall.von), jetzt)
		if err != nil || schritt != fall.schritt {
			t.Errorf("%s: Schritt %d (%v), erwartet %d", fall.name, schritt, err, fall.schritt)
		}
	}

	if _, err := totp.PassenderSchritt(zaehler, "12", jetzt); err == nil {
		t.Error("eine falsche Form ergab einen Schritt")
	}
}
