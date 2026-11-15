//go:build abnahme && darwin

package fenster

import (
	"log/slog"
	"os"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// probepanik wirft, wenn `MSE_ABNAHME_PANIK` den Rückruf nennt — **nur im
// Abnahmebau**, damit eine Hüllenprobe eine Panik in einen Rückruf
// einspeisen kann, ohne einen echten Fehler in den Baum zu legen (F02,
// #2341). Ohne Abnahmebau tut sie nichts (`probepanik_ohne.go`).
func probepanik(was string) {
	if nennt(os.Getenv("MSE_ABNAHME_PANIK"), was) {
		panic("Probepanik in " + was)
	}
}

// probeabsturz lässt die Anwendung nach der Frist in `MSE_ABNAHME_ABSTURZ`
// abstürzen — **mit einer Panik in einer eigenen Goroutine, die niemand
// fängt**, also so, wie ein echter Fehler außerhalb jedes Riegels es täte.
// Die Hüllenprobe prüft damit die Startmarke und die Absturzspur (F04,
// #2343). Ohne gültige Frist tut sie nichts.
func probeabsturz() {
	frist, err := time.ParseDuration(os.Getenv("MSE_ABNAHME_ABSTURZ"))
	if err != nil {
		return
	}

	go func() {
		time.Sleep(frist)
		panic("Probeabsturz nach " + frist.String())
	}()
}

// probeSyncpanik wirft nach der Frist in `MSE_ABNAHME_SYNCPANIK` **im
// Rückruf eines `InvokeSync`** — der Weg, auf dem bis Wails v3.0.0-beta.27
// ein Wartender für immer hing (#2398). **Kommt der Ruf zurück, steht es im
// Protokoll**; die Hüllenprobe `syncpanik09` liest die Zeile und sieht, dass
// die Anwendung weiterläuft. Ohne gültige Frist tut sie nichts.
func probeSyncpanik() {
	frist, err := time.ParseDuration(os.Getenv("MSE_ABNAHME_SYNCPANIK"))
	if err != nil {
		return
	}

	go func() {
		time.Sleep(frist)
		application.InvokeSync(func() { panic("Probepanik im Hauptfaden") })
		slog.Info("der synchrone Ruf kam nach der Panik zurück")
	}()
}
