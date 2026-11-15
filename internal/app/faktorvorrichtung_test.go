// **Die Vorrichtung des zweiten Faktors** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, damit `zweiterfaktor_test.go` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort stehen die
// Verhaltensklassen, hier, **womit** sie gefahren werden.
package app_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //#nosec G505 -- RFC 6238 §1.2: HMAC-SHA-1 ist das Verfahren
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// faktoranwendung startet eine Anwendung mit **eigenem** Bund und Sensor.
//
// # Warum nicht einfach `angelaufeneAnwendung`
//
// **Weil die den echten Schlüsselbund dieses Rechners nähme und den echten
// Sensor** — der erste bekäme Einträge, die niemand wollte, und der zweite
// öffnete einen Dialog, der auf einen Finger wartet (§3: ein Test, der nach
// draußen greift, ist grün, weil er wirklich funktioniert hat).
func faktoranwendung(t *testing.T) *app.App {
	t.Helper()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSchluesselbund(neuerBundspeicher())
	anwendung.MitSensor(&probesensor{}, func(context.Context) biometrie.Auskunft {
		return biometrie.Auskunft{Grund: "in diesem Lauf gibt es keinen Sensor"}
	})

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	return anwendung
}

// faktorcode rechnet den Code zu einem base32-Geheimnis.
//
// # Warum hier und nicht über `internal/totp`
//
// **Zwei Gründe, und der zweite ist der bessere.**
//
// **Erstens die Schichtenregel** (§2): `internal/app` darf `totp` nicht
// kennen, und ein Testimport ließe sich von einem Produktivimport nicht
// unterscheiden — der Wächter liest beide.
//
// **Zweitens ist es die stärkere Probe.** Ein Test, der dieselbe Rechnung
// ruft, die er prüft, sagt nur, dass sie mit sich selbst übereinstimmt. Diese
// hier ist eine **zweite** Umsetzung derselben zwölf Zeilen aus RFC 4226 §5.3
// — stimmen beide überein, stimmt die Naht.
func faktorcode(t *testing.T, geheimnis string) string {
	t.Helper()

	roh, fehler := base32.StdEncoding.WithPadding(base32.NoPadding).
		DecodeString(geheimnis)
	if fehler != nil {
		t.Fatalf("das Geheimnis lesen: %v", fehler)
	}

	schritte := make([]byte, 8)
	binary.BigEndian.PutUint64(schritte, uint64(time.Now().Unix())/30) //#nosec G115 -- Zeit vor 1970 gibt es hier nicht

	rechner := hmac.New(sha1.New, roh)
	rechner.Write(schritte)
	summe := rechner.Sum(nil)

	ab := summe[len(summe)-1] & 0x0f
	zahl := binary.BigEndian.Uint32(summe[ab:ab+4]) & 0x7fffffff

	return fmt.Sprintf("%06d", zahl%1_000_000)
}

// probesensor ist ein Sensor, den es in diesem Lauf nicht gibt.
type probesensor struct{}

func (*probesensor) Vorhanden(context.Context) bool { return false }

func (*probesensor) Bestaetige(context.Context, string) error { return nil }

// **Jeder Weg reicht seinen Fehler durch** (§8).
//
// **Jeder einzeln, weil jeder seinen eigenen `if fehler != nil` hat** — ein
// vergessener machte aus einem Fehlschlag ein stilles Ja, und das ist bei
// einem zweiten Faktor der schlimmste Ausgang.
func TestJederFaktorwegAmFensterReichtSeinenFehlerDurch(t *testing.T) {
	anwendung := faktoranwendung(t)

	// **Ohne Faktor gibt es nichts zu entfernen und nichts zu bestätigen** —
	// beide Wege melden das, statt still zu glücken.
	wege := map[string]func() *app.Fehler{
		"bestätigen ohne Einrichtung": func() *app.Fehler {
			return anwendung.FaktorBestaetigen(t.Context(), "1", "123456").Fehler
		},
		"entfernen, was es nicht gibt": func() *app.Fehler {
			return anwendung.FaktorEntfernen(t.Context(), "1", "totp").Fehler
		},
		"Biometrie ohne Sensor": func() *app.Fehler {
			return anwendung.BiometrieEinrichten(t.Context(), "1").Fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			if fehler := weg(); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Die Rechnerauskunft kommt aus der Naht und nicht aus dem Rechner.**
//
// **Sonst prüfte der Lauf die Ausstattung des Prüfenden** (§10) — und wäre
// auf einem Mac mit Touch ID grün und auf einem ohne rot.
func TestDieRechnerauskunftKommtAusDerNaht(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSchluesselbund(neuerBundspeicher())
	anwendung.MitSensor(&probesensor{}, func(context.Context) biometrie.Auskunft {
		return biometrie.Auskunft{Vorhanden: true, Art: "Face ID"}
	})

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	antwort := anwendung.Faktoren(t.Context(), "1")
	if !antwort.Biometrie || antwort.Biometrieart != "Face ID" {
		t.Errorf("angekommen ist %+v", antwort)
	}
}

// **Und der Flächenweg reicht alle fünf durch.**
//
// **Über einen echten Bestand**, denn das ist die Zusage dieser Naht: die
// drei Flächen ohne Fenster bekommen dieselbe Anwendung (§16).
func TestDerFaktorwegUeberDieFlaeche(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	if antwort := flaeche.Faktoren(probekontext(t), "1"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Die schreibenden Wege melden ihren Fehler** — einen Bund gibt es in
	// diesem Lauf nicht, und einen Sensor auch nicht.
	if flaeche.FaktorBestaetigen(probekontext(t), "1", "123456").Fehler == nil {
		t.Error("das Bestätigen ohne Einrichtung glückte")
	}

	if flaeche.FaktorEntfernen(probekontext(t), "1", "totp").Fehler == nil {
		t.Error("das Entfernen ohne Faktor glückte")
	}

	// **Und die Biometrie gegen einen Sensor, den es hier nicht gibt.**
	//
	// # Warum der Lauf sich seinen Sensor selbst hinstellt
	//
	// **Weil er sonst den echten anfasst — und der zeigt ein Fenster.**
	// `BiometrieEinrichten` ruft `Bestaetige`, und das wartet auf einen
	// Finger; `sensornaht.go` sagt das in seinem Kopf. Bis zum 2026-09-06
	// stand hier derselbe Ruf **ohne** Doppel, und die Zusicherung lautete
	// „die Biometrie ohne Sensor glückte" — die Bedingung dafür stand
	// nirgends im Lauf, sondern in `scripts/ci.sh`, das mit leerem Zuhause
	// fährt.
	//
	// **Was das gekostet hat**: auf diesem Rechner **mit** Sensor wurde der
	// Lauf rot, und mitten in einem Pipelinelauf ging ein Touch-ID-Fenster
	// auf und wartete auf den Auftraggeber. Eine Vorrichtung, die anders ist
	// als der Betrieb, ist §3s erste Warnung — hier war sie nicht großzügiger,
	// sondern griff nach draußen.
	//
	// **Jetzt stellt der Lauf seine Voraussetzung selbst her** (§10: die
	// Zusage gehört dorthin, wo sie eingelöst wird): `NutzerflaecheOhneSensor`
	// setzt einen Sensor, den es nicht gibt, und damit ist „ohne Sensor" wahr
	// **durch Bauart** statt durch Zufall — `Bestaetige` wird nie erreicht.
	//
	// **Ein eigener Bauweg und kein Ruf danach**: der zweite Faktor fängt den
	// Sensor beim Bauen ein. Der erste Versuch setzte ihn hinterher, war grün
	// und bewirkte nichts — die Gegenprobe mit einem Sensor, den es **gibt**,
	// blieb ebenfalls grün und hat es angestrichen.
	ohneSensor := app.NutzerflaecheOhneSensor(
		offeneProbedatenbank(t), &probesensor{}, keineBiometrie)

	if ohneSensor.BiometrieEinrichten(probekontext(t), "1").Fehler == nil {
		t.Error("die Biometrie ohne Sensor glückte")
	}

	// **Und das Einrichten legt ein Geheimnis ab** — im **echten** Bund
	// dieses Rechners, und deshalb steht es hier nicht: der Weg wird in
	// `internal/flaechen` gefahren, wo ein eigener Bund danebensteht (§3).
}

// **Und das Einrichten über die Fläche legt ein Geheimnis ab.**
//
// # Warum dieser Lauf einen eigenen Schlüsselbund anlegt
//
// **Weil `NutzerflaecheUeber` ihre eigene Anwendung baut** — ein Feld auf der
// `App` erreicht sie nicht, und ohne diesen Weg schriebe der Lauf in den
// echten Bund dieses Rechners (§3: ein Test, der nach draußen greift, ist
// grün, **weil** er wirklich funktioniert hat).
func TestDasEinrichtenUeberDieFlaecheLegtEinGeheimnisAb(t *testing.T) {
	t.Setenv(schluesselbund.Bundort, probebundOrt(t))

	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	eingerichtet := flaeche.FaktorEinrichten(probekontext(t), "1")
	if eingerichtet.Fehler != nil {
		t.Fatalf("%+v", eingerichtet.Fehler)
	}

	if eingerichtet.Geheimnis == "" || len(eingerichtet.Codes) == 0 {
		t.Errorf("angekommen ist %+v", eingerichtet)
	}
}

// **Ein Bund, der nicht antwortet, reißt ab.**
func TestEinStummerBundReisstAb(t *testing.T) {
	speicher := neuerBundspeicher()
	speicher.fehler = errors.New("der Bund antwortet nicht")

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSchluesselbund(speicher)
	anwendung.MitSensor(&jasensor{}, func(context.Context) biometrie.Auskunft {
		return biometrie.Auskunft{Vorhanden: true, Art: "Touch ID"}
	})

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	if antwort := anwendung.FaktorEinrichten(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("das Einrichten glückte ohne Bund")
	}

	// **Und mit einem Sensor, der ja sagt, entsteht der biometrische Faktor.**
	if antwort := anwendung.BiometrieEinrichten(t.Context(), "1"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort := anwendung.Faktoren(t.Context(), "1"); len(antwort.Faktoren) != 1 {
		t.Errorf("eingetragen ist %+v", antwort.Faktoren)
	}
}

// **Eine Anwendung ohne Auskunft behauptet nichts.**
//
// **Nicht „keine Biometrie", sondern keine Aussage** — ein `nil` heißt hier,
// dass niemand gefragt wurde, und eine erfundene Antwort wäre schlimmer als
// keine (§10).
func TestEineAnwendungOhneAuskunftBehauptetNichts(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitZweitemFaktor(service.NeuerZweiterfaktor(
		&stilleFaktoren{}, &stilleFaktoren{}, &probesensor{}))

	antwort := anwendung.Faktoren(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Biometrie || antwort.Biometriegrund != "" {
		t.Errorf("behauptet wurde %+v", antwort)
	}
}

// **Und eine Ablage, die nicht antwortet, reißt ab.**
func TestEineStummeFaktorablageReisstAb(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitZweitemFaktor(service.NeuerZweiterfaktor(
		&stilleFaktoren{fehler: errors.New("kaputt")}, &stilleFaktoren{},
		&probesensor{}))

	antwort := anwendung.Faktoren(t.Context(), "1")
	if antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}

	if antwort.Faktoren == nil {
		t.Error("die Liste ist nil")
	}
}

// probebundOrt legt einen eigenen Schlüsselbund an und gibt seinen Ort.
func probebundOrt(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.keychain")

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	if aus, err := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort).CombinedOutput(); err != nil {
		t.Fatalf("den Probebund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		// **Ein eigener Kontext**, weil der des Laufs beim Aufräumen schon
		// abgebrochen ist.
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		//nolint:gosec,errcheck // dito
		_ = exec.CommandContext(ctx, "/usr/bin/security", "delete-keychain", ort).Run()
	})

	//nolint:gosec,errcheck // dito
	_ = exec.CommandContext(t.Context(), "/usr/bin/security",
		"unlock-keychain", "-p", "leer", ort).Run()

	return ort
}

// jasensor ist ein Sensor, der alles bestätigt.
type jasensor struct{}

func (*jasensor) Vorhanden(context.Context) bool { return true }

func (*jasensor) Bestaetige(context.Context, string) error { return nil }

// stilleFaktoren ist die kleinste Ablage, die diese Läufe tragen.
type stilleFaktoren struct{ fehler error }

func (s *stilleFaktoren) Faktoren(context.Context, int64) ([]domain.Faktor, error) {
	return nil, s.fehler
}

func (s *stilleFaktoren) FaktorAnlegen(_ context.Context, neu domain.Faktor) (
	domain.Faktor, error,
) {
	return neu, s.fehler
}

func (s *stilleFaktoren) FaktorLoeschen(
	context.Context, int64, domain.Faktorart,
) error {
	return s.fehler
}

func (s *stilleFaktoren) CodesSetzen(context.Context, int64, []string) error {
	return s.fehler
}

func (s *stilleFaktoren) FreieCodes(context.Context, int64) (
	[]domain.Wiederherstellungscode, error,
) {
	return nil, s.fehler
}

func (s *stilleFaktoren) CodeVerbrauchen(context.Context, int64) error { return s.fehler }

func (s *stilleFaktoren) Setze(context.Context, int64, string) error { return s.fehler }

func (s *stilleFaktoren) Hole(context.Context, int64) (string, error) {
	return "", s.fehler
}

func (s *stilleFaktoren) Nimm(context.Context, int64) error { return s.fehler }

// keineBiometrie ist die Auskunft eines Rechners ohne Sensor.
//
// **Neben dem Sensor und nicht in ihm**: `MitSensor` verlangt beide, weil
// „prüft" und „fragt nach" dieselbe Frage von zwei Seiten stellen — wer sie
// getrennt setzte, könnte einen Rechner bauen, der prüft, ohne es zu können.
func keineBiometrie(context.Context) biometrie.Auskunft {
	return biometrie.Auskunft{Grund: "in diesem Lauf gibt es keinen Sensor"}
}
