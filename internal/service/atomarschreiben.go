package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// schreibeAtomar legt einen Inhalt an einen Pfad, ohne unterwegs etwas
// kaputtzumachen —
// #853.
//
// # Was vorher galt
//
// `os.WriteFile` öffnet mit `O_TRUNC`: eine vorhandene Datei ist ab diesem
// Augenblick leer, und scheitert das Schreiben danach — volle Platte,
// entzogenes Medium, Fehler beim Schließen —, bleibt sie leer oder halb
// zurück. **Ein halbes PNG sieht im Dateibrowser aus wie ein ganzes.**
//
// Der Kommentar an dieser Stelle sagte „erst packen, dann schreiben" und
// stimmte für den **Kodierer** und nur für ihn.
//
// # Was jetzt gilt
//
// Geschrieben wird in eine **Nachbardatei** im selben Verzeichnis, und erst
// das Umbenennen setzt sie an ihren Platz. `os.Rename` ist innerhalb eines
// Dateisystems atomar; bis dahin steht die alte Datei unverändert.
//
// **Im selben Verzeichnis und nicht in `os.TempDir`**: über eine
// Dateisystemgrenze hinweg ist `rename` kein Umbenennen mehr, sondern ein
// Kopieren — und damit wäre die Zusage weg.
//
// # Was es kostet, und das gehört gesagt
//
// **In einem nur lesbaren Verzeichnis mit einer vorhandenen Zieldatei
// scheitert es jetzt**, wo `os.WriteFile` sie noch überschrieben hätte: eine
// Nachbardatei anzulegen braucht Schreibrecht am **Verzeichnis**, die
// vorhandene Datei zu öffnen nur eines an der Datei. Das ist der Preis der
// Atomizität, und er ist der kleinere: ein Verzeichnis, in dem sich keine
// neue Datei anlegen lässt, taugt für eine Bildausfuhr ohnehin nicht.
//
// # Und wie weit es festgeschrieben wird — nach der Rolle
//
// **Hier stand: einen Absturz des Rechners hält es nicht, ohne `fsync` gibt
// diese Zusage niemand.** Seit BK-06b
// (#2131,
// Vermerk 0174) wird die Nachbardatei **vor** dem Umbenennen und das
// Verzeichnis **danach** festgeschrieben — wie weit, sagt die Rolle der
// Datei ([haltbarkeit]). Gemessen am 2026-10-07: `fsync` kostet je Datei
// samt Verzeichnis unter 0,3 ms, `F_FULLFSYNC` rund 8 ms.
//
// # Was es hält, seit der Zweitblick nachgefragt hat
//
// **Die Rechte einer vorhandenen Datei.** `os.WriteFile` schreibt in die
// **bestehende** Datei und lässt ihren Modus; eine neue Nachbardatei entsteht
// mit 0600 und trüge ihn nicht. Wer sein `.mse-set` auf 0644 gestellt hat,
// fand es nach dem Sichern auf 0600.
//
// **Erst nur genannt, dann behoben** (Befund 6 des Zweitblicks über Phase
// 4.13, 2026-09-07). Der Einwand traf: „im Betrieb nicht überraschen" ist die
// Zusage dieser Phase, und der **Normalfall** — Modus nach dem Speichern
// nachsehen — ist ohne konstruierten Fehlerweg prüfbar. Ein Verhalten, das
// deterministisch da ist, wird nicht aufgeschrieben, sondern abgestellt.
//
// **Gibt es das Ziel noch nicht, bleibt es bei 0600** — das ist der Modus,
// den `os.WriteFile` hier vorher setzte.
func schreibeAtomar(pfad string, inhalt []byte) error {
	return schreibeAtomarMit(pfad, bytesFuellen(inhalt))
}

// haltbarkeit sagt, wie weit eine Datei auf die Platte gebracht wird, bevor
// das Schreiben als geglückt gilt — nach ihrer **Rolle** (Vermerk 0174).
type haltbarkeit int

const (
	// alsAusgabe ist, was jemand gespeichert hat und wiederfinden will:
	// `fsync` auf Datei und Verzeichnis — **dieselbe Stufe wie der Bestand**
	// selbst (`synchronous=FULL`, `datenbank.go`). Sie übersteht einen
	// Absturz des Systems; einen Stromausfall mit flüchtigem Plattenpuffer
	// sagt sie nicht zu.
	alsAusgabe haltbarkeit = iota

	// fuerDenWiederanlauf ist, was ein Neustart liest, um zu entscheiden —
	// Epochen- und Rückspielmarken, der Vermerk eines Paketersatzes:
	// `F_FULLFSYNC`, bis zum Datenträger. Selten geschrieben, teuer verloren.
	fuerDenWiederanlauf

	// alsZwischenspeicher ist, was sich neu holen lässt: gar nicht.
	alsZwischenspeicher
)

// schreibeMarke ist [schreibeAtomar] für eine Datei, die ein Neustart liest.
func schreibeMarke(pfad string, inhalt []byte) error {
	return schreibeNach(pfad, bytesFuellen(inhalt), fuerDenWiederanlauf, false)
}

// schreibeZwischenspeicher ist [schreibeAtomar] für eine Datei, die sich
// neu holen lässt.
func schreibeZwischenspeicher(pfad string, inhalt []byte) error {
	return schreibeNach(pfad, bytesFuellen(inhalt), alsZwischenspeicher, false)
}

// SchreibeAtomarVertraulich ist [SchreibeAtomar] für eine Ausgabe, die über
// den Menschen spricht — Personendatenauskunft, Diagnose,
// Einstellungsausfuhr (SEC-20,
// #2133).
//
// **Sie endet mit 0600, auch über einer Datei mit 0644.** Eine gewöhnliche
// Ausgabe übernimmt den Modus der Datei, die sie ersetzt ([modusUebernehmen]);
// hier hieße das, eine Auskunft über einen Menschen für jeden auf dem Rechner
// lesbar zu lassen, weil die Datei davor es war.
func SchreibeAtomarVertraulich(pfad string, inhalt []byte) error {
	return schreibeNach(pfad, bytesFuellen(inhalt), alsAusgabe, true)
}

// Die Nähte für die Prüfung der Fehlerwege (BK-06b): eine Platte, die beim
// Festschreiben oder Umbenennen scheitert, gibt der Betrieb nicht her.
var (
	nachbarAnlegen = func(verzeichnis, muster string) (nachbardatei, error) {
		return os.CreateTemp(verzeichnis, muster)
	}
	umbenennen = os.Rename

	// vollFestschreiben nimmt `F_FULLFSYNC` — unter macOS das, was
	// `(*os.File).Sync` tut.
	vollFestschreiben = func(d nachbardatei) error { return d.Sync() }
)

// nachbardatei ist, was eine Nachbardatei können muss — `*os.File` kann es.
type nachbardatei interface {
	io.WriteCloser
	Name() string
	Sync() error
	Fd() uintptr
}

// festschreibenAls bringt `d` so weit auf die Platte, wie die Rolle verlangt.
func festschreibenAls(d nachbardatei, rolle haltbarkeit) error {
	switch rolle {
	case alsAusgabe:
		return leichtFestschreiben(d)
	case fuerDenWiederanlauf:
		return vollFestschreiben(d)
	default:
		return nil
	}
}

// verzeichnisFestschreiben schreibt den Eintrag eines neuen Namens fest —
// ohne ihn übersteht das Umbenennen den Absturz nicht, auch wenn die Datei
// darunter es tut.
func verzeichnisFestschreiben(verzeichnis string, rolle haltbarkeit) error {
	if rolle == alsZwischenspeicher {
		return nil
	}

	griff, fehler := os.Open(verzeichnis) // #nosec G304 -- das Verzeichnis der eigenen Ausgabe

	// **Ein Verzeichnis ohne Leserecht** lässt sich beschreiben, aber nicht
	// öffnen — hier hinein zu schreiben ging vorher, und ein Fehler wäre ein
	// Rückschritt. Der neue Name ist dort nicht festgeschrieben; die Grenze
	// steht im Vermerk 0174.
	if errors.Is(fehler, fs.ErrPermission) {
		return nil
	}

	if fehler != nil {
		return fehler
	}

	return errors.Join(festschreibenAls(griff, rolle), griff.Close())
}

// schreibeAtomarMit ist [schreibeAtomar] für einen Inhalt, der **strömt** —
// ein `.mse-installer` trägt ganze Pakete samt Bildern (FMT-02.AK-04,
// #1691),
// und ihn erst im Arbeitsspeicher zu bauen hieße, ihn zweimal zu halten.
//
// **Dieselbe Regel und keine zweite Fassung**: [schreibeAtomar] ist dieser
// Weg mit einem festen Inhalt.
func schreibeAtomarMit(pfad string, fuellen func(io.Writer) error) error {
	return schreibeNach(pfad, fuellen, alsAusgabe, false)
}

// schreibeNach ist der eine Weg hinter allen: Nachbardatei füllen,
// festschreiben, schließen, umbenennen, Verzeichnis festschreiben.
func schreibeNach(pfad string, fuellen func(io.Writer) error, rolle haltbarkeit, vertraulich bool) error {
	// **`CreateTemp` legt mit 0600 an.** Ist das Ziel schon da, tritt sein
	// Modus an die Stelle — siehe oben —, außer bei einer vertraulichen
	// Ausgabe.
	nachbar, err := nachbarAnlegen(filepath.Dir(pfad), "."+filepath.Base(pfad)+".*")
	if err != nil {
		return fmt.Errorf("die Nachbardatei anlegen: %w", err)
	}

	fehler := fuelleStromUndSchliesse(nachbar, fuellen, func() error {
		return festschreibenAls(nachbar, rolle)
	})
	if fehler == nil {
		if !vertraulich {
			modusUebernehmen(nachbar.Name(), pfad)
		}

		fehler = umbenennen(nachbar.Name(), pfad)
	}

	if fehler != nil {
		// **Der Aufräumfehler geht mit**, statt den ersten zu verdecken:
		// dieselbe Haltung wie in `storage.Sichern`.
		return errors.Join(fehler, os.Remove(nachbar.Name()))
	}

	// **Die Datei steht schon da** — scheitert das Verzeichnis, ist das
	// trotzdem keine Erfolgsmeldung: ihr Name übersteht dann keinen Absturz.
	if fehler := verzeichnisFestschreiben(filepath.Dir(pfad), rolle); fehler != nil {
		return fmt.Errorf("%s steht da, ist aber nicht festgeschrieben: %w", filepath.Base(pfad), fehler)
	}

	return nil
}

// modusUebernehmen gibt der Nachbardatei die Rechte des Ziels.
//
// **Gibt es das Ziel nicht, bleibt es bei 0600** — und das ist kein Fehler,
// sondern der häufigere Fall: eine Ausfuhr schreibt meist an eine Stelle, an
// der noch nichts liegt.
//
// **Und kein Fehlschlag hält den Lauf auf** — deshalb gibt sie keinen Fehler.
// Weder ein `Stat`, das aus einem anderen Grund als „nicht da" scheitert, noch
// ein `Chmod` auf eine Datei, die dieser Prozess eben selbst angelegt hat. Die
// Nachbardatei steht dann mit 0600 da, und das ist genau das, was
// `os.WriteFile` an dieser Stelle täte: **der Inhalt zählt mehr als der
// Modus**.
//
// **Ein `error` wäre hier ein Zweig, den kein Aufruf betritt** (§3) — und
// damit eine Behauptung darüber, was im Ernstfall geschieht.
func modusUebernehmen(nachbar, ziel string) {
	if angabe, fehler := os.Stat(ziel); fehler == nil {
		_ = os.Chmod(nachbar, angabe.Mode().Perm()) //nolint:errcheck,gosec // siehe oben
	}
}

// SchreibeAtomar ist [schreibeAtomar] für die Fassade.
//
// **Weil `internal/app` dieselben vier Ausgabewege hat** — Diagnose,
// Personendatenauskunft, Vorlagenausgabe und ihre Beigaben (Befund 2 des
// Zweitblicks über Phase 4.13). Sie schreiben dorthin, wohin jemand im
// Dateidialog gezeigt hat, und dafür gilt dieselbe Zusage.
//
// **Eine Hülle und keine zweite Fassung**: zwei Umsetzungen derselben Regel
// liefen auseinander, und die schwächere sähe genauso grün aus.
func SchreibeAtomar(pfad string, inhalt []byte) error {
	return schreibeAtomar(pfad, inhalt)
}

// fuelleUndSchliesse schreibt und schließt — und meldet beides.
//
// **Der Schließfehler zählt.** Ein gepufferter Schreibvorgang scheitert erst
// beim Schließen; wer ihn verwürfe, hielte eine halb geschriebene Datei für
// fertig und benennte sie um.
func fuelleUndSchliesse(datei io.WriteCloser, inhalt []byte) error {
	return fuelleStromUndSchliesse(datei, bytesFuellen(inhalt), func() error { return nil })
}

// bytesFuellen schreibt einen festen Inhalt.
func bytesFuellen(inhalt []byte) func(io.Writer) error {
	return func(ziel io.Writer) error {
		_, err := ziel.Write(inhalt)

		return err
	}
}

// fuelleStromUndSchliesse ist [fuelleUndSchliesse] für einen Inhalt, der
// strömt.
//
// **`festschreiben` läuft zwischen Füllen und Schließen** — danach gibt es
// keinen Griff mehr, über den es ginge.
func fuelleStromUndSchliesse(datei io.WriteCloser, fuellen func(io.Writer) error, festschreiben func() error) error {
	if err := fuellen(datei); err != nil {
		return errors.Join(fmt.Errorf("die Nachbardatei füllen: %w", err), datei.Close())
	}

	if err := festschreiben(); err != nil {
		return errors.Join(fmt.Errorf("die Nachbardatei festschreiben: %w", err), datei.Close())
	}

	if err := datei.Close(); err != nil {
		return fmt.Errorf("die Nachbardatei schließen: %w", err)
	}

	return nil
}
