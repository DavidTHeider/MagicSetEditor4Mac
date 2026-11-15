package austausch

// **Die Quarantäne** — was geprüft wird, **bevor** etwas ausgepackt wird.
//
// # Warum davor und nicht danach
//
// **Weil danach die Bytes schon auf der Platte liegen.** Eine Archivbombe, die
// beim Auspacken auffällt, hat das Verzeichnis dann bereits gefüllt; ein
// Symlink, der beim Auspacken auffällt, zeigt bereits irgendwohin.
//
// # Was eine Unterschrift hier nicht erspart
//
// **Nichts.** #1457 sagt es wörtlich: „Signiert bedeutet bekannte Herkunft,
// nicht sichere ausführbare Vorlage." Ein bekannter Absender kann ein
// übergroßes Bild schicken, und ein Konto, das jemand übernommen hat, ist
// weiterhin ein bekannter Absender.

import (
	"fmt"
	"path"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Grenzen eines Umschlags.
//
// **Enger als die einer Sicherung**, und das ist der Unterschied zwischen
// „meine eigenen Daten" und „was mir jemand geschickt hat":
// `sicherung.HoechstensGesamt` erlaubt 64 GB, weil ein Bestand so groß sein
// darf. Ein Umschlag ist eine Sendung.
const (
	// HoechstensObjekte ist die Zahl der Objekte in einem Umschlag.
	HoechstensObjekte = 4096

	// HoechstensObjekt ist die Größe eines einzelnen Objekts.
	HoechstensObjekt int64 = 256 << 20

	// HoechstensGesamt ist die Summe aller Objekte, ausgepackt.
	//
	// **Ausgepackt und nicht gepackt** — das ist der ganze Punkt bei einer
	// Archivbombe: gepackt sind es ein paar Kilobyte.
	HoechstensGesamt int64 = 2 << 30

	// HoechstensTiefe ist die Zahl der Schrägstriche in einem Namen.
	HoechstensTiefe = 8

	// HoechstensName ist die Länge eines Namens im Umschlag.
	HoechstensName = 255

	// HoechstensManifest ist die Größe von `austausch.json` — SEC-03
	// (#2138).
	// Gerechnet am Schreiber: höchstens [HoechstensObjekte] Einträge zu je
	// gut 400 Bytes — Name bis [HoechstensName], Art, Größe, Summe — und ein
	// Kopf; vier MiB lassen mehr als das Doppelte.
	HoechstensManifest int64 = 4 << 20

	// HoechstensUnterschrift ist die Größe von `austausch.sig` — eine
	// Unterschrift von Ed25519 hat 64 Bytes.
	HoechstensUnterschrift = 1 << 10

	// HoechstensUmschlagmanifest ist die Größe des äußeren Manifests: es
	// führt dieselben Objekte und die zwei Teile darüber.
	HoechstensUmschlagmanifest int64 = 4 << 20
)

// pruefeName hält einen Namen gegen das, was einer sein darf.
//
// # Was hier abgewiesen wird und warum
//
//   - **Leere Namen** — es gibt keinen Eintrag ohne Namen.
//   - **Absolute Wege** (`/etc/passwd`) — ein Umschlag beschreibt, was *in*
//     ihm liegt, nicht wo es auf einem fremden Rechner hingehört.
//   - **Rückwärtsschritte** (`../`) — der klassische Weg aus dem Zielordner
//     heraus. Geprüft wird die **gesäuberte** Form: `a/../../b` sieht
//     unverdächtig aus und ist es nicht.
//   - **Windows-Trenner** (`\`) — auf macOS ein gültiges Zeichen im
//     Dateinamen, auf dem Zielsystem ein Trenner. Wer einen Umschlag auf dem
//     einen packt und auf dem anderen auspackt, bekäme zwei verschiedene
//     Wege aus derselben Zeichenkette.
//   - **Laufwerksbuchstaben** (`C:`) — aus demselben Grund.
//   - **Zu tiefe und zu lange Namen** — beides sind Grenzen, an denen
//     Dateisysteme unterschiedlich versagen.
func pruefeName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf(
			"%w: ein Objekt ohne Namen", domain.ErrValidation)

	case strings.HasPrefix(name, "/"):
		return fmt.Errorf("%w: %q ist ein absoluter Weg",
			domain.ErrValidation, name)

	case strings.Contains(name, `\`):
		return fmt.Errorf("%w: %q trägt einen Rückwärtsschrägstrich",
			domain.ErrValidation, name)

	case len(name) > HoechstensName:
		return fmt.Errorf("%w: der Name ist %d Zeichen lang, erlaubt sind %d",
			domain.ErrValidation, len(name), HoechstensName)

	case strings.Count(name, "/") > HoechstensTiefe:
		return fmt.Errorf("%w: %q liegt tiefer als %d Ebenen",
			domain.ErrValidation, name, HoechstensTiefe)
	}

	// **Erst säubern, dann urteilen.** `a/../../b` ist unverdächtig zu lesen
	// und führt zwei Ebenen nach oben; `path.Clean` macht daraus `../b`.
	sauber := path.Clean(name)
	if sauber == ".." || strings.HasPrefix(sauber, "../") ||
		strings.HasPrefix(sauber, "/") {
		return fmt.Errorf("%w: %q führt aus dem Umschlag heraus",
			domain.ErrValidation, name)
	}

	// **Und was nicht hinausführt, steht trotzdem so da, wie es gemeint
	// ist.** `a/../b.json` nennt `b.json`, und `.` nannte das Verzeichnis
	// selbst: ausgepackt wurde das Objekt an dessen Stelle. Gefunden vom
	// Saatlauf von `FuzzUmschlag` (QA-06b, #1967).
	if sauber == "." || sauber != name {
		return fmt.Errorf("%w: %q ist kein Name im Umschlag — gemeint wäre %q",
			domain.ErrValidation, name, sauber)
	}

	// **Und ein Laufwerksbuchstabe ist kein Name.** `C:datei` ist auf
	// Windows ein Weg relativ zum aktuellen Verzeichnis von `C:`.
	if vorn, _, hat := strings.Cut(name, ":"); hat && len(vorn) <= 2 {
		return fmt.Errorf("%w: %q sieht aus wie ein Laufwerk",
			domain.ErrValidation, name)
	}

	return nil
}

// pruefeManifest hält ein gelesenes Manifest gegen die Grenzen.
//
// **Bevor irgendetwas ausgepackt wird**, und die Reihenfolge ist die Zusage:
// wer erst auspackt und dann prüft, hat die Bytes schon.
func pruefeManifest(manifest Manifest) error {
	if manifest.Fassung != Fassung {
		return fmt.Errorf("%w (%d gegen %d)",
			domain.ErrUmschlagFassung, manifest.Fassung, Fassung)
	}

	if len(manifest.Objekte) == 0 {
		return fmt.Errorf(
			"%w: der Umschlag nennt keine Objekte", domain.ErrValidation)
	}

	if len(manifest.Objekte) > HoechstensObjekte {
		return fmt.Errorf("%w: %d Objekte, erlaubt sind %d",
			domain.ErrValidation, len(manifest.Objekte), HoechstensObjekte)
	}

	gesehen := make(map[string]bool, len(manifest.Objekte))

	var gesamt int64

	for _, objekt := range manifest.Objekte {
		if fehler := pruefeObjekt(objekt, gesehen); fehler != nil {
			return fehler
		}

		gesamt += objekt.Groesse
		if gesamt > HoechstensGesamt {
			return fmt.Errorf(
				"%w: der Umschlag enthält ausgepackt mehr als %d Byte",
				domain.ErrValidation, HoechstensGesamt)
		}
	}

	// **Und was gebraucht wird, liegt im Umschlag.** Ein Verweis nach
	// draußen wäre eine Frage nach dem Bestand des Empfängers, und #1457
	// verbietet sie: „keine beliebige Besitz-/Existenzabfrage".
	for _, objekt := range manifest.Objekte {
		for _, gebraucht := range objekt.Braucht {
			if !gesehen[gebraucht] {
				return fmt.Errorf(
					"%w: %q braucht %q, und das liegt nicht im Umschlag",
					domain.ErrValidation, objekt.Name, gebraucht)
			}
		}
	}

	return nil
}

// pruefeObjekt hält ein einzelnes Objekt und merkt seinen Namen.
func pruefeObjekt(objekt Objekt, gesehen map[string]bool) error {
	if fehler := pruefeName(objekt.Name); fehler != nil {
		return fehler
	}

	if gesehen[objekt.Name] {
		return fmt.Errorf("%w: %q steht zweimal im Umschlag",
			domain.ErrValidation, objekt.Name)
	}

	gesehen[objekt.Name] = true

	if !objekt.Art.gilt() {
		return fmt.Errorf("%w: %q ist keine Objektart",
			domain.ErrValidation, objekt.Art)
	}

	if objekt.Groesse < 0 || objekt.Groesse > HoechstensObjekt {
		return fmt.Errorf("%w: %q ist %d Byte groß, erlaubt sind %d",
			domain.ErrValidation, objekt.Name, objekt.Groesse,
			HoechstensObjekt)
	}

	if len(objekt.Summe) != 64 {
		return fmt.Errorf("%w: %q hat keine SHA-256",
			domain.ErrValidation, objekt.Name)
	}

	return nil
}
