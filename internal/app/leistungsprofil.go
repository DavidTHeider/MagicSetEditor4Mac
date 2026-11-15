package app

// **Ein Leistungsprofil, gezielt und begrenzt** — OPS-05.AK-03
// (#1722),
// Marke `new`.
//
// # Was es ist
//
// **Ein CPU-Profil dieses Prozesses** über eine gewählte Zahl von Sekunden,
// im Format von `go tool pprof`. Es läuft nur, wenn jemand es anstößt, und
// liegt danach neben dem Bestand; von den Profilen bleiben so viele, wie
// `messung.profile` sagt, und das älteste geht. **Die Abtastrate ist die von
// Go**, hundert Mal je Sekunde: `runtime/pprof` setzt sie beim Start selbst
// (`StartCPUProfile`, nachgelesen in `src/runtime/pprof/pprof.go` von Go
// 1.27.1), und ein eigener Wert davor gäbe nur eine Warnung.
//
// # Warum nur am Fenster
//
// **Ein Profil misst den Prozess, in dem es läuft.** Die Kommandozeile lebt
// einen Befehl lang, und Dienst und Werkzeugserver sind eigene Prozesse —
// ihnen über eine Netzfläche ein sechzigsekündiges Profil abzuverlangen, gäbe
// jedem Token einen Weg, den Dienst zu bremsen. `docs/FAEHIGKEITEN.md` nennt
// die Ausnahme.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync/atomic"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Profilordner ist der Ordner neben dem Bestand, in dem die Profile liegen.
const Profilordner = "Leistungsprofile"

// profildauerHoechstens ist die längste Dauer eines Profils in Sekunden —
// **auch im Satz der Ablehnung**, und zwar als Wortlaut: ein Feldbefund mit
// Platzhalter sähe für den Wächter aus wie ein fremder Text.
const profildauerHoechstens = 60

// LeistungsprofilAntwort sagt, wo das Profil liegt.
type LeistungsprofilAntwort struct {
	Datei    string  `json:"datei"`
	Sekunden int     `json:"sekunden"`
	Bytes    int64   `json:"bytes"`
	Fehler   *Fehler `json:"fehler"`
}

// profilLaeuft gilt für den ganzen Prozess — ein CPU-Profil gehört ihm.
var profilLaeuft atomic.Bool

// errProfilLaeuft ist die Antwort auf ein zweites Profil zur selben Zeit.
var errProfilLaeuft = fmt.Errorf("%w: Es läuft schon ein Leistungsprofil — warte, bis es fertig ist",
	domain.ErrConflict)

// LeistungsprofilLaeuft sagt, ob gerade ein Profil aufgezeichnet wird.
func (a *App) LeistungsprofilLaeuft() bool { return profilLaeuft.Load() }

// Leistungsprofil zeichnet ein CPU-Profil auf.
func (a *App) Leistungsprofil(ctx context.Context, sekunden int) LeistungsprofilAntwort {
	ctx = a.imProfil(ctx)

	pfad, groesse, fehler := a.profilAufzeichnen(ctx, sekunden)
	if fehler != nil {
		return LeistungsprofilAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Geschrieben ist geschrieben** (§7, Regel 3): ein Aufräumen, das
	// scheitert, macht das neue Profil nicht ungeschehen — es bleibt ein
	// älteres mehr liegen.
	a.profileAufraeumen(ctx, filepath.Dir(pfad))

	return LeistungsprofilAntwort{Datei: pfad, Sekunden: sekunden, Bytes: groesse}
}

// profilAufzeichnen schreibt das Profil und gibt Pfad und Größe.
//
// **Erst in den Arbeitsspeicher, dann atomar auf die Platte**: ein
// abgebrochener Lauf hinterlässt so gar nichts, und ein halbes Profil, das
// aussähe wie ein ganzes, kann nicht entstehen. Eine Minute Profil wiegt
// wenige hundert Kilobyte.
func (a *App) profilAufzeichnen(ctx context.Context, sekunden int) (pfad string, groesse int64, fehler error) {
	if sekunden < 1 || sekunden > profildauerHoechstens {
		return "", 0, domain.NewFieldError("sekunden", domain.SchluesselValidierung,
			"Ein Leistungsprofil dauert 1 bis 60 Sekunden.")
	}

	if a.datenbankpfad == "" {
		return "", 0, domain.NewFieldError("sekunden", domain.SchluesselValidierung,
			"Ohne Bestand gibt es keinen Ort für ein Leistungsprofil.")
	}

	if !profilLaeuft.CompareAndSwap(false, true) {
		return "", 0, errProfilLaeuft
	}
	defer profilLaeuft.Store(false)

	var puffer bytes.Buffer

	// **Ein fremdes Profil sperrt auch** — etwa `go test -cpuprofile`. Es
	// gehört demselben Prozess, und Go lässt nur eines zu.
	if startfehler := pprof.StartCPUProfile(&puffer); startfehler != nil {
		return "", 0, domain.MitGrund(domain.ErrConflict, startfehler,
			"Es läuft schon ein anderes Profil in diesem Prozess — warte, bis es fertig ist.")
	}

	uhr := time.NewTimer(time.Duration(sekunden) * time.Second)
	defer uhr.Stop()

	select {
	case <-ctx.Done():
		pprof.StopCPUProfile()

		return "", 0, ctx.Err()
	case <-uhr.C:
	}

	pprof.StopCPUProfile()

	ordner := filepath.Join(filepath.Dir(a.datenbankpfad), Profilordner)
	pfad = filepath.Join(ordner, "mse-"+time.Now().Format("20060102-150405.000")+".pprof")

	fehler = os.MkdirAll(ordner, 0o750)
	if fehler == nil {
		fehler = service.SchreibeAtomar(pfad, puffer.Bytes())
	}

	if fehler != nil {
		return "", 0, fehler
	}

	return pfad, int64(puffer.Len()), nil
}

// profileAufraeumen lässt die neuesten liegen, so viele wie eingestellt.
//
// **Nach dem Namen und nicht nach der Uhr der Platte**: der Name ist die
// Uhrzeit des Profils, und `os.ReadDir` gibt sortiert — die ältesten zuerst.
// Nur, was nach einem Profil dieses Programms heißt, wird angefasst.
func (a *App) profileAufraeumen(ctx context.Context, ordner string) {
	behalten := domain.MessregelAus(nil).Profile
	if werte, fehler := a.einstellungen.Alle(ctx); fehler == nil {
		behalten = domain.MessregelAus(werte).Profile
	}

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		return
	}

	namen := make([]string, 0, len(eintraege))

	for _, eintrag := range eintraege {
		if strings.HasPrefix(eintrag.Name(), "mse-") && strings.HasSuffix(eintrag.Name(), ".pprof") {
			namen = append(namen, eintrag.Name())
		}
	}

	for _, alt := range namen[:max(len(namen)-behalten, 0)] {
		_ = os.Remove(filepath.Join(ordner, alt)) //nolint:errcheck // siehe oben: ein Rest kostet Platz, nicht das Profil
	}
}

// Leistungsprofil ist die Fensterbindung.
func (b *Bindungen) Leistungsprofil(ctx context.Context, sekunden int) LeistungsprofilAntwort {
	return b.app(ctx).Leistungsprofil(b.ruf(ctx), sekunden)
}
