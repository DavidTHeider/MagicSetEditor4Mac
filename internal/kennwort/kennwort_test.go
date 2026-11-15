package kennwort_test

// **Kennwörter** — Werkstück 6 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
)

// **Ein Kennwort lässt sich hüten und wiedererkennen.**
func TestEinKennwortLaesstSichHuetenUndWiedererkennen(t *testing.T) {
	t.Parallel()

	gehuetet, fehler := kennwort.Huete("ein hinreichend langes Kennwort")
	if fehler != nil {
		t.Fatalf("hüten: %v", fehler)
	}

	if kennwort.Stimmt(gehuetet, "ein hinreichend langes Kennwort") != nil {
		t.Error("das richtige Kennwort wurde nicht erkannt")
	}

	if kennwort.Stimmt(gehuetet, "ein anderes") == nil {
		t.Error("das falsche Kennwort kam durch")
	}
}

// **Zweimal dasselbe Kennwort gibt zwei verschiedene Zeichenketten.**
//
// # Warum das die halbe Zusage ist
//
// **Weil der Salzwert je Kennwort neu gezogen wird.** Ohne ihn hätten zwei
// Menschen mit demselben Kennwort denselben Hash — und wer einen davon knackt,
// hat beide. Eine vorberechnete Tabelle träfe dazu alle auf einmal.
func TestZweimalDasselbeKennwortGibtZweiZeichenketten(t *testing.T) {
	t.Parallel()

	erstes := huete(t)
	zweites := huete(t)

	if erstes == zweites {
		t.Error("zwei Hüten gaben dasselbe — der Salzwert wird nicht gezogen")
	}

	// **Und beide erkennen dasselbe Kennwort wieder.**
	if kennwort.Stimmt(zweites, "ein hinreichend langes Kennwort") != nil {
		t.Error("der zweite Hash erkennt sein Kennwort nicht")
	}
}

// **Die gehütete Form nennt ihr Verfahren und ihre Zahlen.**
//
// **Nach dem PHC-Format**, damit ein späterer Wechsel der Zahlen die alten
// Kennwörter nicht wertlos macht: was dasteht, sagt selbst, wie es entstand.
func TestDieGehueteteFormNenntIhrVerfahren(t *testing.T) {
	t.Parallel()

	gehuetet := huete(t)

	// **m=19456, t=2, p=1** — das Mindestmaß aus dem OWASP Password Storage
	// Cheat Sheet, nachgelesen am 2026-09-05: „Use Argon2id with a minimum
	// configuration of 19 MiB of memory, an iteration count of 2, and 1 degree
	// of parallelism."
	if !strings.HasPrefix(gehuetet, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Errorf("die Form ist %q", gehuetet)
	}
}

// **Ein zu kurzes Kennwort wird abgewiesen.**
func TestEinZuKurzesKennwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := kennwort.Huete("kurz"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Und ein zu langes ebenso.**
//
// **Nicht aus Prinzip, sondern weil Hüten Arbeit kostet**: Argon2id rechnet
// über die ganze Eingabe, und ein Kennwort von einem Megabyte wäre ein Weg,
// diesen Rechner zu beschäftigen.
func TestEinZuLangesKennwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := kennwort.Huete(strings.Repeat("x", 1025)); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Eine gehütete Form, die keine ist, wird abgewiesen.**
//
// **Und zwar als Fehler und nicht als „stimmt nicht".** Wer eine kaputte Zeile
// im Bestand hat, soll das erfahren — „falsches Kennwort" schickte ihn auf die
// Suche an der falschen Stelle (§8).
func TestEineKaputteFormWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"leer":               "",
		"keine Teile":        "argon2id",
		"falsches Verfahren": "$bcrypt$v=19$m=19456,t=2,p=1$c2FsemE$aGFzaA",
		"falsche Fassung":    "$argon2id$v=18$m=19456,t=2,p=1$c2FsemE$aGFzaA",
		"krumme Zahlen":      "$argon2id$v=19$m=viel,t=2,p=1$c2FsemE$aGFzaA",
		"krummer Salzwert":   "$argon2id$v=19$m=19456,t=2,p=1$!!!$aGFzaA",
		"krummer Hash":       "$argon2id$v=19$m=19456,t=2,p=1$c2FsemE$!!!",
	}

	for name, form := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := kennwort.Stimmt(form, "irgendwas"); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
			}
		})
	}
}

// **Ein Kennwort, das nicht stimmt, ist kein Formfehler.**
//
// **Der Unterschied zählt**: „stimmt nicht" heißt noch einmal tippen, „kaputte
// Form" heißt, im Bestand ist etwas faul.
func TestEinFalschesKennwortIstKeinFormfehler(t *testing.T) {
	t.Parallel()

	gehuetet := huete(t)

	fehler := kennwort.Stimmt(gehuetet, "etwas anderes")
	if !errors.Is(fehler, kennwort.ErrStimmtNicht) {
		t.Errorf("Fehler %v, erwartet war ErrStimmtNicht", fehler)
	}

	if errors.Is(fehler, domain.ErrValidation) {
		t.Error("ein falsches Kennwort kam als Formfehler an")
	}
}

// **Eine gehütete Form mit unmöglichen Zahlen wird abgewiesen.**
//
// # Warum das mehr ist als eine Formprüfung
//
// **Weil `argon2.IDKey` die genannte Menge Speicher wirklich nimmt.** Eine
// Zeile mit `m=99999999` verlangte hundert Gigabyte, und die Anwendung stünde
// — nicht durch einen Angriff von außen, sondern durch eine Bestandsdatei, die
// jemand von Hand geändert hat oder die von einem anderen Werkzeug stammt.
//
// **Und die Hashlänge geht in eine `uint32`.** Ein Wert über vier Milliarden
// liefe darin um; angestrichen von gosec (G115) am 2026-09-05, und die Antwort
// darauf ist eine Schranke und keine Ausnahme im Wächter.
func TestUnmoeglicheZahlenWerdenAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"zu viel Speicher": "$argon2id$v=19$m=99999999,t=2,p=1$" +
			"c2FsemNoZW5zYWx6Y2hlbg$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhh",
		"zu viele Durchgänge": "$argon2id$v=19$m=19456,t=99999,p=1$" +
			"c2FsemNoZW5zYWx6Y2hlbg$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhh",
		"kein Speicher": "$argon2id$v=19$m=0,t=2,p=1$" +
			"c2FsemNoZW5zYWx6Y2hlbg$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhh",
		"kein Durchgang": "$argon2id$v=19$m=19456,t=0,p=1$" +
			"c2FsemNoZW5zYWx6Y2hlbg$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhh",
		"kein Faden": "$argon2id$v=19$m=19456,t=2,p=0$" +
			"c2FsemNoZW5zYWx6Y2hlbg$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhh",
		"zu kurzer Hash": "$argon2id$v=19$m=19456,t=2,p=1$" +
			"c2FsemNoZW5zYWx6Y2hlbg$a2ss",
		"zu kurzer Salzwert": "$argon2id$v=19$m=19456,t=2,p=1$" +
			"a2ss$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhh",
	}

	for name, gehuetet := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := kennwort.Stimmt(gehuetet, "einlangesgeheimnis")
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("angenommen wurde: %v", fehler)
			}

			// **Und nicht als „stimmt nicht"**: das schickte den Nutzer auf
			// die Suche nach einem Tippfehler, den er nicht gemacht hat.
			if errors.Is(fehler, kennwort.ErrStimmtNicht) {
				t.Errorf("gemeldet wurde: %v", fehler)
			}
		})
	}
}

// probekennwort ist das Kennwort, mit dem diese Läufe arbeiten.
const probekennwort = "ein hinreichend langes Kennwort"

// huete hütet [probekennwort] und bricht ab, wenn es nicht glückt.
//
// **Damit der Rückgabewert nicht weggeworfen wird**: ein `_` daneben sähe aus
// wie eine Entscheidung und wäre eine Nachlässigkeit — und ein Test, der auf
// einer leeren Zeichenkette weiterläuft, prüft etwas anderes als sein Name
// sagt.
func huete(t *testing.T) string {
	t.Helper()

	gehuetet, fehler := kennwort.Huete(probekennwort)
	if fehler != nil {
		t.Fatalf("hüten: %v", fehler)
	}

	return gehuetet
}
