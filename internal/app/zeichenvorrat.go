package app

// **Die Flächen ohne Fenster behalten ihr Zeichenwerk** —
// #2005.
//
// # Warum
//
// [Zeichennaht] baute es bis dahin **je Ruf**: jede Karte und jeder Ausdruck
// über `mse-mcp`, `mse-rest` und die Kommandozentrale las das Stilpaket neu,
// zerteilte seine Skripte, suchte jede Schrift durch das ganze Pack und
// entschlüsselte jeden Rahmen. Gemessen im Leistungsvergleich (QA-04): 1,8 s
// je Karte in einer laufenden `mse-mcp`-Sitzung — das Fenster, das sein Werk
// einmal baut, braucht 43–61 ms. Wohin die Zeit ging, steht samt Messung in
// `docs/LEISTUNGSVERGLEICH.md` §8: zu drei Vierteln in die Schriftsuche.
//
// # Wann es neu gebaut wird
//
// **Wenn sich ändert, woraus es gebaut ist**: der Bestand (ein anderes
// `*sql.DB`), das Bilderverzeichnis, die Schriftwurzel, die Paketwurzeln —
// und **die Änderungszeit jeder Wurzel**. Ein Paket, das in eine Wurzel
// kommt oder aus ihr geht, ändert sie; dann entsteht das Werk neu, wie es
// vorher bei jedem Ruf entstand (FMT-02, #1691).
//
// # Wer den Rest sieht
//
// Was die Wurzeln nicht ändert, sieht das Werk selbst — **dieselben Wege wie
// im Fenster**:
//
//   - **Stil oder Spiel eines Sets**, auch aus einem anderen Prozess: der
//     Stilspeicher liest den Setkopf je Ruf.
//   - **Eine Datei in einem Paket**: der Stempel des Stilspeichers, höchstens
//     eine Sekunde alt (`stempelpause`). Er meldet den Wechsel, und das Werk
//     räumt Bildquelle und Ebenenspeicher ([NeuesZeichenwerk]).
//   - **Eine Schrift, die jemand in ein vorhandenes Paket legt**: die
//     Schriftsuche nicht — die Grenze steht bei `schriften.paketnamensindex`
//     und gilt im Fenster genauso.

import (
	"database/sql"
	"os"
	"strconv"
	"strings"
	"sync"
)

// grenzeWerkvorrat ist die Zahl der Bestände, für die ein Werk behalten wird.
//
// **Vier**: ein Prozess ohne Fenster arbeitet über einem Bestand, höchstens
// zweien — `db` und `arbeit` in `cmd/mse-mcp`. Mehr sind ein Prozess, der
// Bestände wechselt; ein Werk je Bestand für immer hielte jeden alten fest,
// samt seinen Bildern.
const grenzeWerkvorrat = 4

// werkvorrat hält je Bestand ein Zeichenwerk samt dem Stand, aus dem es
// gebaut wurde.
type werkvorrat struct {
	grenze int

	sperre sync.Mutex
	werke  map[*sql.DB]gehaltenesWerk
}

// gehaltenesWerk ist ein Werk und sein Stand ([werkstand]).
type gehaltenesWerk struct {
	stand string
	werk  *Zeichenwerk
}

// zeichenvorrat ist der Vorrat dieses Prozesses.
var zeichenvorrat = neuerWerkvorrat(grenzeWerkvorrat)

// neuerWerkvorrat baut einen leeren Vorrat.
func neuerWerkvorrat(grenze int) *werkvorrat {
	return &werkvorrat{grenze: grenze, werke: map[*sql.DB]gehaltenesWerk{}}
}

// werk gibt das behaltene Werk eines Bestands — oder baut eines, wenn es
// keines gibt oder sein Stand nicht mehr gilt.
//
// **Gebaut wird unter der Sperre.** Zwei erste Rufe zugleich bauten sonst zwei
// Werke, und das spätere verdrängte das frühere samt allem, was es schon
// gelesen hat. Der Bau selbst ist billig — teuer ist der erste Gebrauch, und
// der liegt außerhalb.
//
// **An der Grenze räumt er ganz**, nicht klug — derselbe Zug wie bei der
// Bildquelle und dem Ebenenspeicher.
func (v *werkvorrat) werk(db *sql.DB, stand string, bauen func() *Zeichenwerk) *Zeichenwerk {
	v.sperre.Lock()
	defer v.sperre.Unlock()

	gehalten, da := v.werke[db]
	if da && gehalten.stand == stand {
		return gehalten.werk
	}

	if !da && len(v.werke) >= v.grenze {
		clear(v.werke)
	}

	werk := bauen()
	v.werke[db] = gehaltenesWerk{stand: stand, werk: werk}

	return werk
}

// werkstand faltet zusammen, woraus ein Werk gebaut wird: die Orte und die
// Änderungszeit jeder Paketwurzel.
//
// **Eine Wurzel, die sich nicht lesen lässt, steht ohne Zeit da** — entsteht
// sie, ändert sich der Stand.
func werkstand(wurzeln []string, bilder, schriftwurzel string) string {
	var stand strings.Builder

	stand.WriteString(bilder + "\x00" + schriftwurzel)

	for _, wurzel := range wurzeln {
		stand.WriteString("\x00" + wurzel + "\x00")

		if angaben, err := os.Stat(wurzel); err == nil {
			stand.WriteString(strconv.FormatInt(angaben.ModTime().UnixNano(), 10))
		}
	}

	return stand.String()
}
