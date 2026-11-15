package app

// **Der Speicher von Anwendung und Webprozess, getrennt** — OPS02
// (#2348),
// Befund A-07 des Audits 0.9.0.
//
// **Der Go-Heap ist nicht der Speicher der Anwendung.** Die Metriken nannten
// nur ihn; der Webprozess des Fensters hält im Dauerlauf 450 bis 630 MB
// (`docs/belege/huellenprobe/dauerbetrieb.md`), außerhalb dieses Prozesses.
// Gemessen wird jetzt der **physische Fußabdruck** — die Zahl der
// Aktivitätsanzeige — dieses Prozesses und jedes WebKit-Prozesses, für den
// er verantwortlich ist (`internal/prozessspeicher`). Heap und System bleiben
// daneben stehen, als das, was sie sind.

import (
	"os"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/prozessspeicher"
)

// UngemessenSpeicher steht in [MetrikenAntwort.Ungemessen], wenn sich der
// eigene Fußabdruck nicht messen ließ.
const UngemessenSpeicher = "speicher"

// WebprozessDTO ist ein WebKit-Prozess dieser Anwendung — `webinhalt`, `gpu`
// oder `netz`.
type WebprozessDTO struct {
	Art              string `json:"art"`
	FussabdruckBytes int64  `json:"fussabdruckBytes"`
}

// gemerkteMessung misst **höchstens einmal je Frist** und gibt dazwischen die
// letzte Messung.
//
// **Ein Klient, der die Metriken in schneller Folge fragt**, ließe sonst je
// Ruf jeden Prozess des Rechners aufzählen — rund 1,4 ms bei 983 Prozessen
// (gemessen am 2026-10-15), unter dem Race-Detektor ein Vielfaches. Gefunden,
// als ein Lauf tausend Rufe maß und die Messung den Abstand überdeckte, um
// den es ihm ging.
type gemerkteMessung struct {
	frist  time.Duration
	uhr    func() time.Time
	messen func(pid int) (prozessspeicher.Stand, error)

	sperre   sync.Mutex
	gemessen time.Time
	stand    prozessspeicher.Stand
	fehler   error
}

// speichermessung ist die des Prozesses — eine Sekunde, denn die Zahl ist
// ein Stand und keine Kurve.
var speichermessung = &gemerkteMessung{frist: time.Second, uhr: time.Now, messen: prozessspeicher.Messen}

// messe gibt die letzte Messung, wenn sie jünger als die Frist ist.
func (m *gemerkteMessung) messe(pid int) (prozessspeicher.Stand, error) {
	m.sperre.Lock()
	defer m.sperre.Unlock()

	if jetzt := m.uhr(); m.gemessen.IsZero() || jetzt.Sub(m.gemessen) >= m.frist {
		m.stand, m.fehler = m.messen(pid)
		m.gemessen = jetzt
	}

	return m.stand, m.fehler
}

// speicherAnhaengen misst den Speicher über `messen` und trägt ihn ein.
func speicherAnhaengen(antwort *MetrikenAntwort, messen func(pid int) (prozessspeicher.Stand, error)) {
	antwort.Prozess.Web = []WebprozessDTO{}

	stand, fehler := messen(os.Getpid())
	if fehler != nil {
		antwort.Ungemessen = append(antwort.Ungemessen, UngemessenSpeicher)
		antwort.Prozess.WebUngemessen = fehler.Error()

		return
	}

	antwort.Prozess.FussabdruckBytes = int64(stand.Anwendung.FussabdruckBytes) // #nosec G115 -- Bytes eines Prozesses passen in int64
	antwort.Prozess.WebUngemessen = stand.Ungemessen

	for _, web := range stand.Web {
		antwort.Prozess.Web = append(antwort.Prozess.Web, WebprozessDTO{
			Art: string(web.Art), FussabdruckBytes: int64(web.FussabdruckBytes), // #nosec G115 -- dito
		})
	}
}
