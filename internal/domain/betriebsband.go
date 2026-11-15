package domain

// **Das Betriebsband: ein Hinweis, bevor etwas schiefgeht** — OPS02
// (#2348),
// Befund A-07 des Audits 0.9.0: nichts meldete sich von selbst bei knappem
// Platz, einer alten Sicherung, einem gestorbenen Dienst oder einem
// Webprozess, der neu geladen wurde. Die Diagnose war ein Text auf Abruf.
//
// # Was hier entschieden wird
//
// **Wann eine Lage ein Hinweis ist, und wohin er führt** — an einer Stelle,
// damit das Fenster nicht rechnet und jede Schwelle einen Namen hat. Wie die
// Lage gemessen wird, steht in `internal/app/betriebsband.go`.
//
// # Kein Flattern
//
// **Der Platz hat zwei Schwellen**: knapp wird er unter 2 GiB **oder** unter
// 5 % der Platte, gut erst wieder über 2,5 GiB **und** 6 %. Pendelt er um die
// erste, bliebe ohne die zweite der Hinweis nicht stehen — er käme und ginge
// mit jedem Abruf.

import (
	"strconv"
	"time"
)

// Die Schwellen des Platzes.
const (
	PlatzKnappBytes   = uint64(2) << 30
	PlatzKnappProzent = 5

	PlatzGutBytes   = uint64(5) << 29 // 2,5 GiB
	PlatzGutProzent = 6
)

// PlatzKnapp sagt, ob `frei` von `gesamt` knapp ist — `vorher` ist das Urteil
// des letzten Abrufs. **Ohne Größe ist nichts knapp**: eine Platte, die sich
// nicht messen ließ, bekommt keinen Hinweis, den niemand begründen kann.
func PlatzKnapp(frei, gesamt uint64, vorher bool) bool {
	if gesamt == 0 {
		return false
	}

	prozent := frei * 100 / gesamt

	if vorher {
		return frei < PlatzGutBytes || prozent < PlatzGutProzent
	}

	return frei < PlatzKnappBytes || prozent < PlatzKnappProzent
}

// Skriptdienstlage ist, was launchd über den eingerichteten Dienst für Skripte sagt.
type Skriptdienstlage struct {
	// Eingerichtet heißt: seine plist liegt da.
	Eingerichtet bool

	// Laeuft heißt: launchd nennt einen Prozess.
	Laeuft bool

	// LetzterAusgang ist der Rückgabewert seines letzten Laufs.
	LetzterAusgang int

	// Laeufe zählt, wie oft launchd ihn gestartet hat (`runs`).
	Laeufe int

	// Unbekannt heißt: launchd ließ sich nicht fragen — dann gibt es
	// **keinen** neuen Hinweis, denn ein gescheiterter Nachweis ist kein
	// Befund; ein stehender bleibt.
	Unbekannt bool
}

// Beendet sagt, ob der Dienst aus einem Fehler endete. **Mit 0 endet er
// absichtlich** — sein Schalter ist aus (OPS-10), und launchd lässt ihn aus.
func (d Skriptdienstlage) Beendet() bool {
	return d.Eingerichtet && !d.Unbekannt && !d.Laeuft && d.LetzterAusgang != 0
}

// DienstBeendet sagt, ob der Hinweis auf einen beendeten Dienst steht —
// `vorher` ist die Lage des letzten Abrufs, `warBeendet` sein Urteil.
//
// **Ein Dienst, der immer wieder abstürzt, flattert sonst**: launchd startet
// ihn nach einem Fehler neu, und ein Abruf in den Millisekunden seines Laufs
// nähme den Hinweis zurück. **Zurück geht er erst**, wenn der Dienst zwischen
// zwei Abrufen durchlief (`Laeufe` unverändert) **und dabei mindestens die
// [Erholfrist] lang gesehen wurde** (`gelaufen`), sauber endete oder nicht
// mehr eingerichtet ist.
//
// **Die Frist kam mit dem Zweitblick über Phase 4** (Befund 11): zwei Abrufe
// kurz hintereinander — die Startseite verlassen und wieder geöffnet — lagen
// beide im kurzen Lauf eines Dienstes, der danach wieder abstürzte, und er
// galt als erholt.
func DienstBeendet(jetzt, vorher Skriptdienstlage, warBeendet bool, gelaufen time.Duration) bool {
	switch {
	case !jetzt.Eingerichtet:
		return false
	case jetzt.Unbekannt:
		return warBeendet
	case jetzt.Beendet():
		return true
	}

	return warBeendet && jetzt.Laeuft && jetzt.LetzterAusgang != 0 &&
		(jetzt.Laeufe != vorher.Laeufe || gelaufen < Erholfrist)
}

// Erholfrist ist, wie lange ein Lauf des Dienstes gesehen sein muss, bevor
// ein beendeter als erholt gilt — **ein Takt des Bands** (30 s am Fenster):
// launchd startet einen abstürzenden Dienst frühestens nach zehn Sekunden
// neu, und ein Lauf, der einen Takt übersteht, ist kein Absturz in Schleife.
const Erholfrist = 30 * time.Second

// Betriebslage ist, was ein Abruf über den Betrieb weiß.
type Betriebslage struct {
	PlatzKnapp bool
	FreiBytes  uint64

	SicherungAlt  bool
	SicherungTage int

	// DienstBeendet ist das Urteil aus [DienstBeendet], DienstAusgang der
	// Rückgabewert seines letzten Laufs.
	DienstBeendet bool
	DienstAusgang int

	// Webneustarts zählt, wie oft der Webprozess seit dem Start neu
	// geladen wurde.
	Webneustarts int
}

// Betriebshinweisart ist, worüber ein Hinweis spricht.
type Betriebshinweisart string

// Die vier Hinweise.
const (
	HinweisPlatzKnapp    Betriebshinweisart = "platzKnapp"
	HinweisSicherungAlt  Betriebshinweisart = "sicherungAlt"
	HinweisDienstBeendet Betriebshinweisart = "dienstBeendet"
	HinweisWebprozessNeu Betriebshinweisart = "webprozessNeu"
)

// Die Ziele — wohin ein Hinweis führt.
const (
	ZielDatenUndSicherung = "einstellungen/daten"
	ZielDiagnose          = "einstellungen/diagnose"
)

// Betriebshinweis ist ein Hinweis mit seinem Ziel und den Werten für seinen
// Satz.
type Betriebshinweis struct {
	Art   Betriebshinweisart
	Ziel  string
	Werte map[string]string
}

// Betriebshinweise gibt je Lage **genau einen** Hinweis.
func Betriebshinweise(lage Betriebslage) []Betriebshinweis {
	var hinweise []Betriebshinweis

	if lage.PlatzKnapp {
		hinweise = append(hinweise, Betriebshinweis{
			Art: HinweisPlatzKnapp, Ziel: ZielDatenUndSicherung,
			Werte: map[string]string{"frei": strconv.FormatUint(lage.FreiBytes, 10)},
		})
	}

	if lage.SicherungAlt {
		hinweise = append(hinweise, Betriebshinweis{
			Art: HinweisSicherungAlt, Ziel: ZielDatenUndSicherung,
			Werte: map[string]string{"tage": strconv.Itoa(lage.SicherungTage)},
		})
	}

	if lage.DienstBeendet {
		hinweise = append(hinweise, Betriebshinweis{
			Art: HinweisDienstBeendet, Ziel: ZielDiagnose,
			Werte: map[string]string{"ausgang": strconv.Itoa(lage.DienstAusgang)},
		})
	}

	if lage.Webneustarts > 0 {
		hinweise = append(hinweise, Betriebshinweis{
			Art: HinweisWebprozessNeu, Ziel: ZielDiagnose,
			Werte: map[string]string{"anzahl": strconv.Itoa(lage.Webneustarts)},
		})
	}

	return hinweise
}
