package service

// **Was eine Sicherung über sich hergibt** — herausgelöst aus
// `bestandsrueckspielen.go`, als OPS-06
// (#1723)
// die Pakete dazubrachte und die Datei an die Zeilengrenze aus §2 ging. Der
// Schnitt folgt der Frage: dort steht, **wie getauscht wird**, hier, **was man
// vorher sieht**.

import "context"

// Sicherungsschau ist, was eine Sicherung über sich hergibt — **ohne dass
// sich am Bestand etwas ändert**.
//
// Wer nur wissen will, was in einer Datei liegt, bekommt die Auskunft; und wer
// übernehmen will, hat vorher gesehen, was er bekommt.
type Sicherungsschau struct {
	AngelegtAm  string
	Schemastand string
	Sets        int
	Karten      int
	Bilder      string

	// Verschluesselt sagt, ob die Sicherung ein Kennwort trägt (OPS-06).
	Verschluesselt bool

	// PaketeEnthalten sind die eigenen Pakete, die mitkommen;
	// PaketeFehlen die bezogenen, die hier nicht liegen und neu zu beziehen
	// sind — die Vollständigkeit, bevor jemand bestätigt (OPS-06.AK-03).
	PaketeEnthalten []string
	PaketeFehlen    []string
}

// Die drei Werte von [Sicherungsschau.Bilder]
// (#1162).
//
// **Warum das Fenster es erfahren muss**: was mit den eingesetzten Bildern
// geschieht, hängt am **Format** der gewählten Datei — und ohne diese Angabe
// sagte die Rückfrage für jede Sicherung dasselbe. Bei einer nackten `.db`
// sagte sie damit das Gegenteil dessen, was geschieht, und genau dafür
// verlangt #1162 einen Hinweis (Befund 2 des Nachlaufs).
const (
	// BilderErsetzt: die Sicherung trägt Bilder, sie treten an die Stelle der
	// vorhandenen.
	BilderErsetzt = "ersetzt"

	// BilderGeleert: die Sicherung ist ein Archiv **ohne** Bilder. Das ist
	// eine Aussage — sie beschreibt einen Stand, in dem keine da waren.
	BilderGeleert = "geleert"

	// BilderUnberuehrt: eine nackte `.db` aus 0.1.0 oder 0.2.0. Sie sagt über
	// Bilder **nichts**, und deshalb bleiben die vorhandenen stehen.
	BilderUnberuehrt = "unberuehrt"
)

// Beschauen liest eine Sicherungsdatei, ohne sie zu übernehmen.
//
// **Der Dialog steht nicht hier.** Welche Datei gemeint ist, entscheidet das
// Fenster (§6); dieser Weg bekommt einen Pfad und gibt Auskunft — und ist
// damit auch von einer Fläche ohne Fenster erreichbar (§16).
func (b *Bestandsweg) Beschauen(ctx context.Context, pfad string) (Sicherungsschau, error) {
	// **Angesehen wird der Bestand, nicht das Archiv.** Ein ZIP ist keine
	// Datenbank; wer die Vorschau darauf ansetzte, bekäme „unbrauchbar" für
	// jede gültige Sicherung.
	ausgepackt, auspackfehler := b.auspacken(ctx, pfad, b.orte.Datenbank, Bestandsname)
	if auspackfehler != nil {
		return Sicherungsschau{}, auspackfehler
	}

	defer ausgepackt.Wegraeumen()

	befund, fehler := b.pruefer(ctx, ausgepackt.Bestand)
	if fehler != nil {
		return Sicherungsschau{}, fehler
	}

	schau := Sicherungsschau{
		AngelegtAm:  befund.AngelegtAm,
		Schemastand: befund.Schemastand,
		Sets:        befund.Sets,
		Karten:      befund.Karten,
		Bilder:      ausgepackt.Bilderlage(),

		Verschluesselt: ausgepackt.Verschluesselt,
	}
	schau.PaketeEnthalten, schau.PaketeFehlen = b.paketlage(ausgepackt.Pakete)

	return schau, nil
}
