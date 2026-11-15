package domain

import (
	"errors"
	"fmt"
	"io/fs"
	"syscall"
)

// **Was der Bestand selbst sagt** — ENG-05.AK-03
// (#1730)
// und #1812.
//
// # Warum [ErrConflict] allein nicht reicht
//
// **Weil er dreierlei hieß, und die Oberfläche nur eines sagte.** Der Katalog
// setzte für jeden Konflikt „Der Eintrag wurde zwischenzeitlich geändert.
// Bitte neu laden." Das stimmt für einen überholten Stand — und für keinen der
// anderen: ein zweites Deck mit demselben Namen, derselbe Druck ein zweites
// Mal in der Sammlung (nativer Befund F08), der letzte Nutzer, der bleibt.
// Gemessen am 2026-09-23 an der Kommandozeile: Nutzer, Profil, Turnier und
// Deck mit doppeltem Namen kamen alle als „ein Eintrag mit diesen Angaben
// liegt schon da" an, und das Fenster erklärte es als fremde Änderung.
//
// **Beide bleiben Konflikte**, damit REST weiter mit 409 antwortet (§16,
// verträglich). Unterschieden wird darunter.
var (
	// ErrDoppelt: einen Eintrag mit diesen Angaben gibt es schon — eine
	// verletzte Eindeutigkeit. Der Rat ist, eine Angabe zu ändern, nicht
	// neu zu laden.
	ErrDoppelt = fmt.Errorf("%w: das gibt es schon", ErrConflict)

	// ErrUeberholt: der Eintrag hat sich geändert, seit er gelesen wurde —
	// die Fassung stimmt nicht mehr. **Nur hier** ist „lade neu" der Rat.
	ErrUeberholt = fmt.Errorf("%w: der Stand ist überholt", ErrConflict)

	// ErrVoll: auf dem Datenträger ist kein Platz mehr (#1812).
	//
	// **Keine der drei Grundarten.** Die Eingabe ist in Ordnung, und der
	// Bestand steht ihr nicht entgegen; was fehlt, ist Platz — und der
	// nächste Schritt liegt außerhalb dieser Anwendung.
	ErrVoll = errors.New("auf dem Datenträger ist kein Platz mehr")

	// ErrGesperrt: auf diesen Ort darf die Anwendung nicht zugreifen —
	// FIX-F02 (#1746), beim Lesen wie beim Schreiben. **Wie der volle
	// Datenträger** keine der drei Grundarten: der nächste Schritt ist ein
	// anderer Ort oder eine Freigabe.
	ErrGesperrt = errors.New("auf diesen Ort darf nicht zugegriffen werden")

	// ErrBeschaedigt: die Datei ist beschädigt oder gar kein Bestand —
	// F05 (#2344). **Die einzige Lage des Starts, in der Ersetzen der Rat
	// ist**; die übrigen sagen, was am Ort, am Datenträger oder an der Zeit
	// liegt, und die Datei ist dann in Ordnung.
	ErrBeschaedigt = errors.New("die Datei ist beschädigt")

	// ErrEA: der Datenträger meldete beim Lesen oder Schreiben einen Fehler
	// — F05 (#2344). **Nicht [ErrBeschaedigt]**: ob die Datei heil ist, weiß
	// hier niemand, und wer sie gleich ersetzte, verlöre sie vielleicht ohne
	// Not.
	ErrEA = errors.New("der Datenträger meldete einen Fehler")

	// ErrKaumPlatz: geschrieben werden konnte nicht, und zugleich ist kaum
	// Platz frei — #2407. **Er umschließt den eigentlichen Fehler** — einen
	// des Datenträgers, einen gesperrten oder nur lesbaren Ort — **und ist
	// nicht [ErrVoll]**: dass ein voller Datenträger sie erzeugt, beweist nicht
	// die Umkehrung. Voll heißt es nur mit Beleg in der Kette (`ENOSPC`,
	// `SQLITE_FULL`).
	ErrKaumPlatz = errors.New("geschrieben werden konnte nicht, und es ist kaum Platz frei")

	// ErrSicherungVerworfen: die eben geschriebene Sicherung bestand ihre
	// Prüfung nicht und wurde verworfen — F05 (#2344). **Nicht
	// [ErrSicherung]**: dort taugt eine gewählte Datei nicht, und der Rat ist
	// „wähle eine andere"; hier ist die neue nichts geworden, und der Rat ist,
	// es noch einmal oder an einem anderen Ort zu versuchen.
	ErrSicherungVerworfen = errors.New("die Sicherung bestand ihre Prüfung nicht")

	// ErrSicherungUnvermerkt: die Sicherung liegt geschrieben und geprüft
	// da, nur ihr Zeitpunkt ließ sich nicht festhalten — OPS01 (#2347).
	// **Nicht [ErrSicherungVerworfen]**: dort ist nichts gesichert, hier
	// schon, und wer es nicht erführe, löschte vielleicht eine gute Datei.
	ErrSicherungUnvermerkt = errors.New("die Sicherung liegt da, ihr Zeitpunkt ist nicht vermerkt")
)

// Platzmangel ordnet einen Fehler ein, der von einem vollen Datenträger kommt.
//
// # Zwei Quellen, eine Art
//
// **SQLite meldet ihn als `SQLITE_FULL`** — auch der Treiber tut das bei
// `ENOSPC` (`vfs/os_unix.go` in ncruces v0.35.4) —, und `internal/storage`
// ordnet ihn dort ein. **Eine Datei meldet ihn als `ENOSPC`**, und zwar an
// jeder Stelle, die schreibt: Set sichern, Bilder, Druckbogen, Ausfuhr,
// Sicherung. Statt jede dieser Stellen einzeln zu fassen, fragen die Flächen
// hier — an **einer** Stelle, und eine neue Schreibstelle fällt nicht durch.
//
// **Der Pfad bleibt im Grund**: `ENOSPC` kommt als `*fs.PathError`, und
// dessen Text trägt den Pfad (§8).
//
// **Unter Windows** heißt derselbe Fall `ERROR_DISK_FULL`; `syscall.ENOSPC`
// ist dort eine erfundene Zahl und trifft ihn nicht. Beschrieben in
// `docs/PORTIERUNG.md`, gebaut wird nur für macOS (§12).
func Platzmangel(fehler error) error {
	if fehler == nil || errors.Is(fehler, ErrVoll) || !errors.Is(fehler, syscall.ENOSPC) {
		return fehler
	}

	return MitGrund(ErrVoll, fehler, "schaffe Platz und versuche es noch einmal")
}

// Zugriffsmangel ordnet einen Ort ohne Zugriffsrecht ein — `EACCES` und
// `EPERM`, beide erkennt `fs.ErrPermission`, beim Lesen wie beim Schreiben. **Dieselbe Gestalt wie
// [Platzmangel]**: an einer Stelle statt an jeder, die schreibt, und der Pfad
// bleibt im Grund (§8). Gefunden mit der Hüllenprobe am 2026-09-26: ein Set,
// in einen gesperrten Ordner gesichert, meldete „einen unerwarteten Fehler".
func Zugriffsmangel(fehler error) error {
	if fehler == nil || errors.Is(fehler, ErrGesperrt) || !errors.Is(fehler, fs.ErrPermission) {
		return fehler
	}

	return MitGrund(ErrGesperrt, fehler, "wähle einen anderen Ort oder gib diesen frei")
}
