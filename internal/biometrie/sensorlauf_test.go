//go:build sensorlauf && darwin

package biometrie_test

// **Die echte Bindung am Sensor** — SEC-19
// (#2153).
//
// Die übrigen Tests prüfen `biometrie` gegen ein Doppel der drei Rufe ins
// Framework; ungemessen blieb, was `LocalAuthentication` wirklich antwortet.
// Hier läuft die **echte** Bindung — `tmc/apple` über `purego` — gegen den
// Sensor dieses Rechners.
//
// # Wie er läuft
//
//	go test -tags sensorlauf -count=1 -v -run <Name> ./internal/biometrie/
//
// **Nicht in `make ci`**: er öffnet den Touch-ID-Dialog dieses Rechners.
// `TestEineAblehnungIstKeinErfolg` wartet darauf, dass jemand — oder ein
// Werkzeug, das den Dialog bedient — „Abbrechen" drückt; ohne das endet er
// nach seiner Frist rot.
//
// # Was er nicht prüft
//
// **Eine Zustimmung.** Dazu müsste ein Finger auf den Sensor, und das tut
// kein Lauf. Die Zusage „Zustimmung → Erfolg" bleibt damit ungemessen.

import (
	"context"
	"errors"
	"runtime/debug"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestDieGeladenenFassungen hält, was geladen ist: genau die gepinnten
// Fassungen der beiden unfertigen Schichten (AK: „Genaue tmc/apple v0.6.18
// und purego v0.11.0-alpha.6 geladen"). Seit dem 2026-10-11 ist purego eine
// Endfassung, v0.11.1 (#2305) — tmc/apple verlangt weiter alpha.6, und
// geladen wird nach MVS die höhere.
func TestDieGeladenenFassungen(t *testing.T) {
	info, gelesen := debug.ReadBuildInfo()
	if !gelesen {
		t.Fatal("keine Bauauskunft")
	}

	erwartet := map[string]string{
		"github.com/tmc/apple":         "v0.6.18",
		"github.com/ebitengine/purego": "v0.11.1",
	}

	for _, modul := range info.Deps {
		if soll, gibt := erwartet[modul.Path]; gibt {
			if modul.Version != soll {
				t.Errorf("%s ist %s geladen, gepinnt ist %s", modul.Path, modul.Version, soll)
			}

			delete(erwartet, modul.Path)
		}
	}

	for pfad := range erwartet {
		t.Errorf("%s ist nicht geladen", pfad)
	}
}

// TestDerSensorIstDa ist der positive Beleg vor jeder Verneinung (§3):
// ohne Sensor hieße „kein Erfolg" nur, dass niemand fragen kann.
func TestDerSensorIstDa(t *testing.T) {
	auskunft := biometrie.Stand(t.Context())
	if !auskunft.Vorhanden || auskunft.Art != "Touch ID" {
		t.Fatalf("kein Touch ID: %+v", auskunft)
	}
}

// TestEineFristEndetOhneSpaetenErfolg lässt den Dialog unbeantwortet: die
// Frist des Kontexts beendet die Abfrage mit einem Fehler, und danach kommt
// nichts mehr — der Ruf hat geantwortet, und `Invalidate` schließt den
// Dialog.
func TestEineFristEndetOhneSpaetenErfolg(t *testing.T) {
	frist := 4 * time.Second

	ctx, abbrechen := context.WithTimeout(t.Context(), frist)
	defer abbrechen()

	anfang := time.Now()
	fehler := biometrie.Bestaetige(ctx, "SEC-19: eine Frist läuft ab — bitte nichts tun")
	dauer := time.Since(anfang)

	t.Logf("nach %s: %v", dauer.Round(time.Millisecond), fehler)

	if fehler == nil {
		t.Fatal("eine unbeantwortete Abfrage galt als bestätigt")
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist keine Ablehnung der Prüfung: %v", fehler)
	}

	if dauer < frist-500*time.Millisecond || dauer > frist+5*time.Second {
		t.Errorf("die Abfrage endete nach %s, die Frist war %s", dauer, frist)
	}
}

// TestEineAblehnungIstKeinErfolg wartet auf „Abbrechen" im Dialog — die
// Ablehnung durch den, der davorsitzt. **Sie muss vor der Frist kommen**:
// endet der Ruf erst an der Frist, hat niemand abgelehnt, und der Test sagt
// es.
func TestEineAblehnungIstKeinErfolg(t *testing.T) {
	frist := 90 * time.Second

	ctx, abbrechen := context.WithTimeout(t.Context(), frist)
	defer abbrechen()

	t.Log("Der Dialog steht gleich da — jetzt »Abbrechen« drücken.")

	anfang := time.Now()
	fehler := biometrie.Bestaetige(ctx, "SEC-19: bitte ablehnen")
	dauer := time.Since(anfang)

	t.Logf("nach %s: %v", dauer.Round(time.Millisecond), fehler)

	if fehler == nil {
		t.Fatal("eine abgelehnte Abfrage galt als bestätigt")
	}

	if ctx.Err() != nil {
		t.Fatalf("die Abfrage endete an der Frist, nicht an einer Ablehnung: %v", fehler)
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist keine Ablehnung der Prüfung: %v", fehler)
	}
}

// TestEinAbbruchEndetOhneSpaetenErfolg bricht die Abfrage mitten im Dialog ab
// — so, wie die Anwendung es tut, wenn jemand das Fenster schließt. Danach
// kommt nichts mehr: der Ruf antwortet mit dem Abbruch, `Invalidate`
// schließt den Dialog, und wer jetzt noch den Finger auflegte, bestätigte
// nichts mehr.
func TestEinAbbruchEndetOhneSpaetenErfolg(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	uhr := time.AfterFunc(2*time.Second, abbrechen)

	defer uhr.Stop()

	anfang := time.Now()
	fehler := biometrie.Bestaetige(ctx, "SEC-19: ein Abbruch kommt — bitte nichts tun")
	dauer := time.Since(anfang)

	t.Logf("nach %s: %v", dauer.Round(time.Millisecond), fehler)

	if fehler == nil {
		t.Fatal("eine abgebrochene Abfrage galt als bestätigt")
	}

	if !errors.Is(fehler, domain.ErrValidation) || !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("erwartet ein Abbruch als Ablehnung der Prüfung: %v (Kontext: %v)", fehler, ctx.Err())
	}

	if dauer > 7*time.Second {
		t.Errorf("der Abbruch nach 2 s wirkte erst nach %s", dauer)
	}
}
