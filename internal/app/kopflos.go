package app

// **Der Sensor einer Fläche, vor der niemand sitzt** — SEC-07
// (#2148),
// gebaut von [NutzerflaecheKopflos].

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kopflos ist kein Sensor — und sagt das, ohne den Rechner zu fragen.
type kopflos struct{}

// Vorhanden ist nie wahr: `ueberDenSensor` meldet dann den fehlenden Faktor,
// und `BiometrieEinrichten` richtet nichts ein.
func (kopflos) Vorhanden(context.Context) bool { return false }

// Bestaetige wird über [kopflos.Vorhanden] nie erreicht; wer es doch ruft,
// bekommt dieselbe Antwort wie ohne Sensor.
func (kopflos) Bestaetige(context.Context, string) error {
	return domain.FehlenderFaktor(domain.FaktorBiometrie)
}

// Auskunft sagt, warum es hier keine Biometrie gibt — **ohne** den Rechner
// zu fragen, denn die Antwort hinge an der Ausstattung und nicht an der Fläche.
func (kopflos) Auskunft(context.Context) biometrie.Auskunft {
	return biometrie.Auskunft{
		Grund: "Touch ID und Face ID gibt es nur am Fenster und an der Kommandozeile dieses Rechners.",
	}
}

var _ service.Sensor = kopflos{}

// NutzerflaecheKopflos baut sie für eine Fläche, vor der niemand sitzt —
// REST und MCP (SEC-07, #2148).
//
// **Der Sensor dieses Rechners gehört dem Menschen davor.** Wer über einen
// Port oder ein Sprachmodell kommt, sitzt nicht davor: ein leerer Nachweis
// öffnete dort einen Touch-ID-Dialog, den jemand bestätigt, der nicht weiß,
// wofür. Hier gibt es deshalb **keinen Sensor** — ein leerer Nachweis endet in
// [domain.FehlenderFaktor], und TOTP und die Wiederherstellungscodes bleiben
// der Weg. Fenster und Kommandozeile behalten den Sensor
// ([NutzerflaecheUeber]): an beiden sitzt der Mensch, der ihn bedient.
//
// **Eingesetzt vor der Verdrahtung**, denn der zweite Faktor fängt den
// Sensor beim Bauen ein (siehe [NutzerflaecheOhneSensor]).
func NutzerflaecheKopflos(db *sql.DB) *Nutzerflaeche {
	return nutzerflaeche(db, kopflos{}, kopflos{}.Auskunft)
}
