package service_test

// **Lokale Entsperrung über den zweiten Faktor** — ID-02.AK-02/03
// (#1701).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Möglich heißt: eingerichtet und auf diesem Rechner vorhanden** — gefragt
// wird dabei nicht.
func TestDieLokaleEntsperrungIstNurMitSensorMoeglich(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name      string
		faktoren  []domain.Faktor
		vorhanden bool
		moeglich  bool
	}{
		{"ohne Faktor", nil, true, false},
		{"nur TOTP", []domain.Faktor{{Art: domain.FaktorTOTP}}, true, false},
		{"Touch ID ohne Sensor", []domain.Faktor{{Art: domain.FaktorBiometrie}}, false, false},
		{"Touch ID mit Sensor", []domain.Faktor{{Art: domain.FaktorBiometrie}}, true, true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			sensor := &sensordoppel{vorhanden: fall.vorhanden}
			ablage := &faktordoppel{faktoren: fall.faktoren}
			dienst := service.NeuerZweiterfaktor(ablage, ablage, sensor)

			moeglich, fehler := dienst.SensorMoeglich(t.Context(), 1)
			if fehler != nil || moeglich != fall.moeglich {
				t.Errorf("möglich: %v, %v", moeglich, fehler)
			}

			if sensor.gefragt != "" {
				t.Errorf("die Auskunft fragte den Sensor: %q", sensor.gefragt)
			}

			pruefung := dienst.SensorPruefen(t.Context(), 1)
			if fall.moeglich != (pruefung == nil) {
				t.Errorf("Prüfung: %v", pruefung)
			}

			if !fall.moeglich && !errors.Is(pruefung, domain.ErrValidation) {
				t.Errorf("unmöglich, aber kein Validierungsfehler: %v", pruefung)
			}
		})
	}
}

func TestDieLokaleEntsperrungNenntIhrenGrund(t *testing.T) {
	t.Parallel()

	sensor := &sensordoppel{vorhanden: true}
	ablage := &faktordoppel{faktoren: []domain.Faktor{{Art: domain.FaktorBiometrie}}}
	dienst := service.NeuerZweiterfaktor(ablage, ablage, sensor)

	if fehler := dienst.SensorPruefen(t.Context(), 1); fehler != nil {
		t.Fatal(fehler)
	}

	if sensor.gefragt != "um diese Anwendung zu entsperren" {
		t.Errorf("Grund: %q", sensor.gefragt)
	}

	sensor.fehler = errors.New("abgebrochen")
	if fehler := dienst.SensorPruefen(t.Context(), 1); fehler == nil {
		t.Error("ein abgebrochener Sensor gilt")
	}

	ablage.kaputt = errors.New("die Platte klemmt")
	if _, fehler := dienst.SensorMoeglich(t.Context(), 1); fehler == nil {
		t.Error("ohne Faktorliste möglich")
	}

	if fehler := dienst.SensorPruefen(t.Context(), 1); fehler == nil {
		t.Error("ohne Faktorliste geprüft")
	}
}

// **Nur der Zettel stellt wieder her** — gefunden wird der Code hier,
// verbraucht wird er im selben Geschäft wie das neue Kennwort
// (`storage.Nutzerablage.KennwortMitCodeSetzen`).
func TestEinCodeVomZettelWirdGefunden(t *testing.T) {
	t.Parallel()

	gehuetet, fehler := kennwort.Huete("zettel-code-1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{Art: domain.FaktorTOTP}},
		codes:    []domain.Wiederherstellungscode{{ID: 7, Gehuetet: gehuetet}},
	}
	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{})

	if _, fehler := dienst.CodeFinden(t.Context(), 1, "falsch"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("falsch: %v", fehler)
	}

	if id, fehler := dienst.CodeFinden(t.Context(), 1, "zettel-code-1"); fehler != nil || id != 7 {
		t.Errorf("gefunden: %d %v", id, fehler)
	}

	// **Finden verbraucht nichts** — der Code steht danach noch da.
	if len(ablage.codes) != 1 {
		t.Errorf("verbraucht: %+v", ablage.codes)
	}

	ablage.lesefehler = errors.New("die Platte klemmt")
	if _, fehler := dienst.CodeFinden(t.Context(), 1, "zettel-code-1"); !errors.Is(fehler, ablage.lesefehler) {
		t.Errorf("Lesefehler: %v", fehler)
	}
}

// **Auch Sensor und Zettel sagen, was nicht stimmt.**
func TestSensorUndZettelSagenWasNichtStimmt(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}
	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{})

	var feld *domain.FieldError
	if fehler := dienst.SensorPruefen(t.Context(), 1); !errors.As(fehler, &feld) || feld.Key != domain.SchluesselLokaleEntsperrungFehlt {
		t.Errorf("Sensor: %v", fehler)
	}

	if _, fehler := dienst.CodeFinden(t.Context(), 1, "falsch"); !errors.As(fehler, &feld) || feld.Key != domain.SchluesselCodeFalsch {
		t.Errorf("Zettel: %v", fehler)
	}
}
