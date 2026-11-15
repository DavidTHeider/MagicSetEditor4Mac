package app

import (
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// knapperPlatz ist, ab wann zu einem Fehler des Datenträgers der knappe
// Platz genannt wird — weniger als ein Mebibyte frei.
const knapperPlatz = 1 << 20

// vollStattEA nennt zu einem Fehler des Datenträgers beim Start den knappen
// Platz, wenn auf dem Volumen des Bestands kaum etwas frei ist — F05
// (#2344),
// seit #2407 als beides und nicht mehr als „voll".
//
// **SQLite meldet einen vollen Datenträger nicht immer als `FULL`.** Gemessen
// am nativen Fenster (`startlagen09`, 2026-10-14): auf einem randvollen
// Abbild scheitert schon `PRAGMA journal_mode = WAL` mit
// `SQLITE_IOERR_SHMSIZE` — die Nebendatei `-shm` lässt sich nicht
// vergrößern —, und `ENOSPC` steht nicht in der Kette. Gefragt wird deshalb
// das Volumen selbst, und nur, wenn schon ein Fehler des Datenträgers da ist.
//
// **Ist der Platz nicht zu erfahren, bleibt es beim Fehler des
// Datenträgers**: dessen Rat — prüfen und noch einmal versuchen — schadet
// auch bei einem vollen nicht.
func vollStattEA(fehler error, ort string, frei func(string) (int64, error)) error {
	return vollStattEABei(fehler, ort, frei, knapperPlatz)
}

// vollStattEABei ist [vollStattEA] mit einem eigenen Bedarf — für eine
// Sicherung, die so viel braucht, wie der Bestand belegt
// (`sicherungsplatz.go`).
func vollStattEABei(fehler error, ort string, frei func(string) (int64, error), bedarf int64) error {
	if !nichtGeschrieben(fehler) || errors.Is(fehler, domain.ErrVoll) {
		return fehler
	}

	// **Voll nur mit Beleg in der Kette** (#2407): ein `ENOSPC` sagt es.
	if voll := domain.Platzmangel(fehler); errors.Is(voll, domain.ErrVoll) {
		return voll
	}

	if bytes, err := frei(ort); err != nil || bytes >= bedarf {
		return fehler
	}

	// **Ohne Beleg beides**: der Fehler des Datenträgers bleibt die
	// Hauptsache, und der knappe Platz steht dabei.
	return domain.MitGrund(domain.ErrKaumPlatz, fehler,
		"schaffe Platz und versuche es noch einmal; hilft das nicht, prüfe den Datenträger und das Schreibrecht")
}

// nichtGeschrieben sagt, ob ein Fehler eine Lage ist, die ein voller
// Datenträger erzeugen kann, ohne es zu sagen: ein Fehler des Datenträgers
// (`SQLITE_IOERR_SHMSIZE` bei 12 KB frei) — und ein Ort, der sich nicht
// öffnen oder schreiben ließ (`SQLITE_CANTOPEN` bei 0 KB frei, gemessen am
// 2026-10-15; dort hieß es „gesperrt" und dann „nur lesbar", mit dem Rat,
// Schreibrecht zu geben).
func nichtGeschrieben(fehler error) bool {
	return errors.Is(fehler, domain.ErrEA) || errors.Is(fehler, domain.ErrGesperrt) || errors.Is(fehler, domain.ErrNurLesend)
}
