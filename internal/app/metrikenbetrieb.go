package app

// **Was die Metriken seit OPS-05 dazu sagen** — OPS-05.AK-01/04
// (#1722).
//
// # Die Einheit steht im Namen
//
// **`…Ms`, `…Bytes`, `…Mikros`** — wer die Zahl liest, soll nicht raten, ob
// Sekunden oder Millisekunden gemeint sind. Zählungen haben keine Einheit
// und keinen Zusatz.
//
// # Vollständigkeit gehört in die Antwort
//
// **Was dieser Prozess nicht misst, steht dabei** (AK-04): die Bindungen des
// Fensters haben keine gemeinsame Stelle, an der ein Ruf endet — gemessen
// werden Dienst und Werkzeugserver —, und die Kommandozeile lebt einen
// Befehl lang. Ohne diese Zeile sähe eine leere Verteilung am Fenster aus
// wie ein Fenster, das nichts getan hat.

import (
	"context"
	"database/sql"
	"runtime"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Was die Laufzeiten nicht kennen.
const (
	UngemessenFenster       = "fenster"
	UngemessenKommandozeile = "kommandozeile"
)

// prozessbeginn ist der Anfang des Erhebungszeitraums.
var prozessbeginn = time.Now()

// DatenbankDTO sind die Zahlen der Verbindung zum Bestand.
type DatenbankDTO struct {
	Offen       int     `json:"offen"`
	InBenutzung int     `json:"inBenutzung"`
	Frei        int     `json:"frei"`
	Gewartet    int64   `json:"gewartet"`
	WartezeitMs float64 `json:"wartezeitMs"`
}

// SpeicherDTO sind die Zahlen der Zwischenspeicher fremder Dienste — die
// Gestalt aus `internal/netz`, übersetzt in `scryfallnaehte.go`, denn dort
// darf diese Schicht `netz` kennen.
type SpeicherDTO struct {
	Treffer        int64 `json:"treffer"`
	Geholt         int64 `json:"geholt"`
	Veraltet       int64 `json:"veraltet"`
	Fehlgeschlagen int64 `json:"fehlgeschlagen"`
}

// ProzessDTO ist der Ressourcenverbrauch dieses Prozesses.
type ProzessDTO struct {
	HeapBytes   int64 `json:"heapBytes"`
	SystemBytes int64 `json:"systemBytes"`
	Goroutinen  int   `json:"goroutinen"`
	CPUMs       int64 `json:"cpuMs"`

	// FussabdruckBytes ist der physische Fußabdruck dieses Prozesses, Web
	// der seiner WebKit-Prozesse — **leer am Dienst und am Werkzeugserver**,
	// die keine Webansicht halten —, und WebUngemessen sagt, warum sie nicht
	// gemessen sind (OPS02, `metrikenspeicher.go`).
	FussabdruckBytes int64           `json:"fussabdruckBytes"`
	Web              []WebprozessDTO `json:"web"`
	WebUngemessen    string          `json:"webUngemessen"`
}

// betriebAnhaengen ergänzt eine Antwort um Laufzeiten, Spur, Speicher,
// Datenbank und Prozess — und um Zeitraum und Aufwand.
func betriebAnhaengen(antwort *MetrikenAntwort, db *sql.DB) {
	beginn := time.Now()

	antwort.Laufzeiten = service.Laufzeitstand()
	antwort.EimerMs = service.Laufzeiteimer
	antwort.Spur = service.Spurstand()
	antwort.Speicher = speicherzahlen()
	antwort.Ungemessen = []string{UngemessenFenster, UngemessenKommandozeile}

	if db != nil {
		stand := db.Stats()
		antwort.Datenbank = &DatenbankDTO{
			Offen: stand.OpenConnections, InBenutzung: stand.InUse, Frei: stand.Idle,
			Gewartet: stand.WaitCount, WartezeitMs: float64(stand.WaitDuration) / float64(time.Millisecond),
		}
	}

	var speicher runtime.MemStats
	runtime.ReadMemStats(&speicher)

	antwort.Prozess = ProzessDTO{
		HeapBytes:   int64(speicher.HeapAlloc), // #nosec G115 -- Bytes eines Prozesses passen in int64
		SystemBytes: int64(speicher.Sys),       // #nosec G115 -- dito
		Goroutinen:  runtime.NumGoroutine(), CPUMs: prozesszeitMs(),
	}

	speicherAnhaengen(antwort, speichermessung.messe)

	jetzt := time.Now()
	antwort.Seit = prozessbeginn.Format(time.RFC3339Nano)
	antwort.Bis = jetzt.Format(time.RFC3339Nano)
	antwort.AufwandMikros = jetzt.Sub(beginn).Microseconds()
}

// MetrikenUeber gibt die Metriken eines Prozesses, der seinen Bestand über
// eine eigene Verbindung hält — Dienst und Werkzeugserver.
func MetrikenUeber(ctx context.Context, db *sql.DB) MetrikenAntwort { return metrikenAus(ctx, db) }

// Rufmessung ist ein gemessener Ruf über Dienst oder Werkzeugserver — die
// Gestalt aus `service`, damit die Wurzeln `service` nicht kennen müssen.
type Rufmessung = service.Laufzeitpunkt

// ErfasseRuf nimmt einen Ruf über Dienst oder Werkzeugserver auf.
func ErfasseRuf(ctx context.Context, punkt Rufmessung) { service.ErfasseRuf(ctx, punkt) }
