package app

import (
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Der Composition Root für den Freigabezugang** —
// #1100.
//
// Eigene Datei aus demselben Grund wie `scryfallverdrahtung.go`: `app` darf
// `netz` kennen, weil es verdrahtet — aber nur an **benannten** Stellen.
// Stünde es in `app.go`, dürfte jede Bindung einen Klienten bauen, und der
// Transport wäre wieder über die Fassade verteilt (§2).

// Fassungsumgebung lenkt die Freigabeadresse um — **nur für Tests**.
//
// Dieselbe Naht wie [Scryfallumgebung] und aus demselben Grund: ohne sie ließe
// sich der Weg vom Fenster bis zur Gegenstelle nur mit echtem Netz prüfen, und
// was das wert ist, sagt §3 — nachgemessen an
// #888.
const Fassungsumgebung = "MSE_FREIGABE_ADRESSE"

// freigabeadresse ist die Grundadresse, unter der nach Freigaben gefragt wird.
var freigabeadresse = umgelenkteFreigabeadresse()

// umgelenkteFreigabeadresse liest [Fassungsumgebung] — oder gibt GitHub.
func umgelenkteFreigabeadresse() string {
	if ort := os.Getenv(Fassungsumgebung); ort != "" {
		return ort
	}

	return aktualisierung.Adresse
}

// MitFreigabeadresse lenkt die Grundadresse um — **nur für Tests**.
//
// Sie gibt zurück, was vorher dastand, damit ein Aufrufer sie zurückstellen
// kann.
func MitFreigabeadresse(neue string) (vorher string) {
	vorher, freigabeadresse = freigabeadresse, neue

	return vorher
}

// MitFreigabebund setzt den Schlüsselbund der Signaturprüfung — **nur für
// Tests**.
//
// **Weitergereicht und nicht neu erfunden**: `internal/schichten` lässt `cmd/`
// nur an `app`, `domain`, `rest` und `storage` — auch in Testdateien. Ohne
// diese Zeile bräuchte jeder Test einer Fläche den **privaten** Schlüssel des
// Herausgebers, und dann prüfte er, ob `gpg` da ist.
//
// **Was sie nicht tut: die Prüfung abschalten.** Ein Test setzt einen echten
// Schlüssel ein und signiert damit wirklich (§3).
func MitFreigabebund(neuer openpgp.EntityList) (vorher openpgp.EntityList) {
	return aktualisierung.MitFreigabebund(neuer)
}

// Downloadverzeichnis gibt das Standard-Downloadverzeichnis dieses Rechners.
//
// **Öffentlich, weil auch die Kommandozeile lädt** (§16: dieselbe Fähigkeit,
// überall dieselbe Antwort). Der Ort ist derselbe — sonst legte `mse` seine
// Abbilder woandershin als das Fenster.
func Downloadverzeichnis() (string, error) { return downloadortEcht() }

// FassungsnahtMit baut den Zugang zu den Freigaben.
//
// **Er wird eingerichtet, gefragt wird er nicht.** Ob überhaupt eine Anfrage
// hinausgeht, entscheidet die Einstellung — und ihre Vorgabe ist „aus"
// (§16, Regel 6).
//
// **Der Proxyweg kommt hinein und wird nicht hier gelesen** (N5, #1571).
// Diese Naht kennt keinen Bestand — vier Aufrufer haben einen, einer nicht
// —, und eine Naht, die sich ihre Einstellungen selbst holte, bräuchte ihn.
// `nil` heißt: die Umgebung gilt, wie überall.
func FassungsnahtMit(proxy netz.Proxyquelle) Fassungsanleger {
	// **Dasselbe Drosselwerk wie Scryfall** — eines je Prozess, und die
	// Drossel hängt am Schlüssel. Ein zweites Werk wäre ein zweiter Zustand
	// über derselben Frage, und ein Wächter zählt nach.
	// **Die Marke kommt aus der Umgebung** (#1100, Werkstück 2). Nicht aus
	// den Einstellungen: die liegen im Bestand, und der ist eine
	// SQLite-Datei im Klartext — eine Marke dort stünde in jeder Sicherung.
	return aktualisierung.NeuerKlient(freigabeadresse,
		netz.NeuerKlientMit(netz.UeblicheFristen(), proxy), drosselwerk,
		aktualisierung.MarkeAusUmgebung())
}

// Fassungsnaht ist [FassungsnahtMit] ohne eingestellten Proxy.
//
// **Für die Flächen ohne Bestand.** Sie steht daneben, damit an den vier
// Stellen, die einen haben, sichtbar wird, dass sie ihn **mitgeben** — ein
// stiller Vorgabewert sähe an beiden Stellen gleich aus.
func Fassungsnaht() Fassungsanleger { return FassungsnahtMit(nil) }
