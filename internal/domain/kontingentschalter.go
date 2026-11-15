// **Die Kontingente lassen sich abschalten** — QA-08.AK-03
// (#1739),
// Marke `new`: MSE2 hat weder einen Dienst noch einen Werkzeugserver.
//
// # Was abgeschaltet wird, und was nicht
//
// **Nur die Guthaben** — je Principal und global ([Kontingent]). Sie sind die
// einzige Grenze, die eine **Rate** ist: wie viel legitime Arbeit in der
// Minute durchgeht. In einer kontrollierten Umgebung — ein Lastlauf, ein
// Stapel über Nacht — sind sie eine Bremse ohne Gegenüber.
//
// **Was bleibt, gleich wie der Schalter steht**: die Plätze je
// [Ressourcenklasse], die Grenze für Rumpf und Antwort, die Sperre gegen das
// Raten, der Zugang, TLS und die Fristen. Sie schützen Kerne, Speicher und
// Zugang, und die hat eine kontrollierte Umgebung genauso
// ([0161](../../docs/entscheidungen/0161-die-kontingente-lassen-sich-abschalten.md)).
//
// # Warum nicht `dienst.kontingente`
//
// **Weil „nur lokal" jeden Dienstschalter abschaltet** ([mitNurLokal] in
// `geltung.go`), und `dienst.kontingente` hätte die Gestalt eines Schalters:
// zwei Teile unter `dienst.`. Wer die Richtlinie wählte, hätte damit die
// Kontingente abgeschaltet — eine Schranke hätte eine andere geöffnet.

package domain

// Der Schalter und seine zwei Werte.
const (
	GrenzenKontingente = "grenzen.kontingente"
	KontingenteAn      = "an"
	KontingenteAus     = "aus"
)

// kontingentschalter ist die Zeile für die Reihe — neben dem Zeugnis der
// Dienste, denn sie betrifft dieselben zwei Programme.
func kontingentschalter() einstellung {
	return einstellung{
		Schluessel: GrenzenKontingente, Wirkung: WirkungNeustart, Bereich: BereichNetz,
		Werte: []string{KontingenteAn, KontingenteAus}, Vorgabe: KontingenteAn,
	}
}

// OhneKontingente sagt, ob die Einstellungen die Guthaben abschalten —
// **nur ein ausdrückliches `aus`**: ein unlesbarer Wert lässt sie gelten.
func OhneKontingente(werte map[string]string) bool {
	return werte[GrenzenKontingente] == KontingenteAus
}

// NachEinstellungen gibt die Grenzen, wie die Einstellungen sie wollen —
// alles andere bleibt, wie es war.
func (g Haushaltsgrenzen) NachEinstellungen(werte map[string]string) Haushaltsgrenzen {
	g.OhneKontingente = OhneKontingente(werte)

	return g
}

// Kontingentstand nennt den Schalter so, wie die Einstellung ihn schreibt.
func (g Haushaltsgrenzen) Kontingentstand() string {
	if g.OhneKontingente {
		return KontingenteAus
	}

	return KontingenteAn
}
