// **Was eine Fassung öffnen kann** — Werkstück O6 von Phase 5.2
// (#1430).
//
// # Der Satz, um den es geht
//
// Das Issue sagt ihn selbst: **„Schema-Min/Max allein genügt nicht bei
// Checksummen/geänderten Migrationen."** Zwei Fassungen können dieselbe
// Schemanummer nennen und **andere** Wanderungen verbucht haben — eine
// nachträglich geänderte Migration hat dieselbe Nummer und einen anderen
// Inhalt. Verträglichkeit ist deshalb kein Vergleich zweier Zahlen, sondern
// zweier Listen: **Name und Prüfsumme je Wanderung**, gegen das, was der
// Bestand wirklich verbucht hat.
//
// # Drei Ausgänge, nicht zwei
//
// **`ungeklaert` ist der dritte**, und er ist die Zusage aus dem Issue: „Alte
// Releases ohne Manifest: unbekannte Kompatibilität offen benennen, keine
// erfundenen Werte." Eine Fassung ohne Manifest ist nicht unverträglich — sie
// ist unbekannt, und das ist etwas anderes.
//
// # Die Hindernisse sind Marken, keine Sätze
//
// Wie bei [Betriebszustand]: vier Flächen bauen aus derselben Marke ihren
// eigenen Satz, und der deutsche steht im Katalog des Fensters (§9). Ein Satz
// aus Go wäre an vier Flächen vier Sätze.

package domain

// Wanderungsangabe ist eine Migration, wie eine Fassung sie kennt.
type Wanderungsangabe struct {
	Name  string `json:"name"`
	Summe string `json:"sha256"`
}

// Fassungsmanifest ist, was eine Freigabe über sich selbst sagt.
//
// **Es reist in der Prüfsummendatei mit**, die schon heute mit OpenPGP
// signiert wird (`internal/aktualisierung/signatur.go`). Damit gilt die
// vorhandene Signaturprüfung unverändert auch für diese Angaben — das ist die
// „authentisierte Metadatenbindung" aus dem Issue, ohne einen zweiten
// Vertrauensanker.
//
// **Mit Apple-Notarisierung hat das nichts zu tun**
// ([0003](../../docs/entscheidungen/0003-keine-notarisierung.md)): jene sagt,
// ob Gatekeeper das Bündel startet, diese, ob die Fassung diesen Bestand
// öffnen kann. Zwei Fragen, zwei Antworten, kein gemeinsamer Weg.
type Fassungsmanifest struct {
	Fassung   string `json:"fassung"`
	Plattform string `json:"plattform"`

	// Artefakt ist die SHA-256 des ausgelieferten Bündels.
	Artefakt string `json:"artefakt"`

	// Einstellungen und Protokoll sind die Fassungen der beiden Formate, die
	// über einen Wechsel hinweg bestehen bleiben (O3 und das Protokoll).
	Einstellungen int `json:"einstellungen"`
	Protokoll     int `json:"protokoll"`

	// Wanderungen sind die Migrationen, die diese Fassung kennt.
	Wanderungen []Wanderungsangabe `json:"wanderungen"`

	// Grundlinie ist der letzte Schritt, den sie nur noch als Ganzes kennt
	// (Vermerk 0189) — leer, wenn sie jeden Schritt einzeln wandert. **Ein
	// Bestand, dem einer bis dorthin fehlt, öffnet sie nicht**, auch wenn sie
	// alle seine Schritte kennt.
	Grundlinie string `json:"grundlinie,omitempty"`
}

// Protokollfassung ist die Fassung des Protokollformats.
//
// **Sie steht hier, damit das Manifest sie gegen etwas hält** — Befund 6 des
// Zweitblicks über Phase 5.2. Das Freigabemanifest führt sie seit O6; verglichen
// wurde sie mit nichts, und eine Zahl ohne Gegenüber ist keine Aussage.
//
// **Heute ist sie eins, und niemand sonst schreibt sie.** Wer das Format der
// Protokollzeilen ändert — Felder, Namen, Bedeutung —, erhöht sie hier; ein
// Wächter hält das Freigabeskript dagegen
// (`internal/werkzeuge/manifestzahlen_test.go`).
const Protokollfassung = 1

// Vertraeglichkeitsurteil ist der Ausgang einer Prüfung.
type Vertraeglichkeitsurteil string

const (
	// UrteilVertraeglich: diese Fassung kann diesen Bestand öffnen.
	UrteilVertraeglich Vertraeglichkeitsurteil = "vertraeglich"

	// UrteilUnvertraeglich: sie kann es nicht, und die Hindernisse sagen warum.
	UrteilUnvertraeglich Vertraeglichkeitsurteil = "unvertraeglich"

	// UrteilUngeklaert: es lässt sich nicht sagen — **nicht** „geht nicht".
	UrteilUngeklaert Vertraeglichkeitsurteil = "ungeklaert"
)

// Hindernisart benennt, was im Weg steht.
type Hindernisart string

const (
	// HindernisKeinManifest: die Freigabe sagt nichts über sich — ein
	// Release von vor diesem Werkstück.
	HindernisKeinManifest Hindernisart = "keinManifest"

	// HindernisPlattform: sie ist für einen anderen Rechner gebaut.
	HindernisPlattform Hindernisart = "plattform"

	// HindernisWanderungFehlt: der Bestand hat eine Migration verbucht, die
	// diese Fassung nicht kennt — der Rückschritt auf eine ältere.
	HindernisWanderungFehlt Hindernisart = "wanderungFehlt"

	// HindernisVorDerGrundlinie: dem Bestand fehlt ein Schritt bis zu ihrer
	// Grundlinie, und einzeln wandert sie ihn nicht mehr (Vermerk 0189).
	HindernisVorDerGrundlinie Hindernisart = "vorDerGrundlinie"

	// HindernisSummeAnders: dieselbe Migration, anderer Inhalt. **Das ist
	// der Fall, den eine Schemanummer nicht sieht.**
	HindernisSummeAnders Hindernisart = "summeAnders"

	// **Die vier der Vorprüfung** — sie sagen nichts über die Fassung,
	// sondern über den Augenblick: kann dieser Rechner den Tausch jetzt
	// überstehen?

	// HindernisPlatz: auf dem Volumen des Bestands ist zu wenig frei.
	HindernisPlatz Hindernisart = "platz"

	// HindernisEinstellungsformat: das Manifest nennt die Fassung des
	// Einstellungsaustauschs nicht — der Katalog ist unvollständig (Befund 6).
	HindernisEinstellungsformat Hindernisart = "einstellungsformatFehlt"

	// HindernisProtokollformat: dasselbe für das Protokollformat.
	HindernisProtokollformat Hindernisart = "protokollformatFehlt"

	// HindernisVorgaengeLaufen: es arbeitet noch etwas — ein Tausch mitten
	// in einer Einfuhr ließe halbe Arbeit zurück.
	HindernisVorgaengeLaufen Hindernisart = "vorgaengeLaufen"

	// HindernisKeineSicherung: es ist keine gemacht worden.
	HindernisKeineSicherung Hindernisart = "keineSicherung"

	// HindernisSicherungUnlesbar: es ist eine da, und sie lässt sich nicht
	// lesen. **Das ist der schlimmere Fall** — eine Sicherung, die niemand
	// zurückspielen kann, ist schlechter als keine, weil man sich auf sie
	// verlässt.
	HindernisSicherungUnlesbar Hindernisart = "sicherungUnlesbar"

	// HindernisHalter: ein anderes Programm hält den Bestand offen — es
	// liefe nach dem Tausch mit dem Code von vorher weiter. `Stelle` nennt
	// es (M02, #2346). **Der eingerichtete Dienst zählt nicht**: ihn startet
	// die neue Fassung selbst neu.
	HindernisHalter Hindernisart = "halter"
)

// Aktualisierungsbefund ist das Ergebnis der ganzen Vorprüfung.
type Aktualisierungsbefund struct {
	Urteil      Vertraeglichkeitsurteil `json:"urteil"`
	Hindernisse []Hindernis             `json:"hindernisse"`

	// Bereit sagt, ob der Tausch vorbereitet werden darf.
	//
	// **Nur bei [UrteilVertraeglich] und ohne jedes Hindernis.** Bei
	// [UrteilUngeklaert] ist er falsch, und das ist kein Nein zur Fassung:
	// es ist ein Nein zum **automatischen** Weg. Der Mensch kann von Hand
	// weitermachen, und die Oberfläche sagt ihm wie — das ist die zweite
	// Abnahme aus dem Issue.
	Bereit bool `json:"bereit"`
}

// Bereitschaft rechnet aus Urteil und Hindernissen, ob vorbereitet werden darf.
func Bereitschaft(urteil Vertraeglichkeitsurteil, hindernisse []Hindernis) Aktualisierungsbefund {
	if hindernisse == nil {
		hindernisse = []Hindernis{}
	}

	return Aktualisierungsbefund{
		Urteil:      urteil,
		Hindernisse: hindernisse,
		Bereit:      urteil == UrteilVertraeglich && len(hindernisse) == 0,
	}
}

// Hindernis ist ein Grund, warum eine Fassung diesen Bestand nicht öffnet.
type Hindernis struct {
	Art Hindernisart `json:"art"`

	// Stelle benennt es genauer: der Name einer Wanderung, die Plattform der
	// Freigabe. Leer, wo es nichts zu benennen gibt.
	Stelle string `json:"stelle"`
}

// Vertraeglichkeit vergleicht eine Freigabe mit dem, was ein Bestand verbucht
// hat.
//
// **`verbucht` ist der echte Ledger** — Name auf Prüfsumme, so wie
// `storage.angewendeteMitSummen` ihn liest. Nicht die Migrationen, die im
// Bündel liegen: die sagen, was eine Fassung **kann**, nicht, was dieser
// Bestand **hat**.
//
// **Ohne Manifest gibt es kein Urteil.** Ein leeres [Fassungsmanifest] heißt
// „die Freigabe sagt nichts über sich", und die Antwort darauf ist
// [UrteilUngeklaert] samt [HindernisKeinManifest] — nicht ein erfundenes Ja
// und nicht ein vorsorgliches Nein.
func Vertraeglichkeit(
	manifest Fassungsmanifest, verbucht map[string]string, plattform string,
) (Vertraeglichkeitsurteil, []Hindernis) {
	if manifest.Fassung == "" || len(manifest.Wanderungen) == 0 {
		return UrteilUngeklaert, []Hindernis{{Art: HindernisKeinManifest}}
	}

	hindernisse := make([]Hindernis, 0, 1)

	if manifest.Plattform != plattform {
		hindernisse = append(hindernisse, Hindernis{
			Art: HindernisPlattform, Stelle: manifest.Plattform,
		})
	}

	// **Die Formatfassungen tragen durch** — Befund 6 des Zweitblicks über
	// Phase 5.2. Das Manifest führte sie, und die Entscheidung sah sie nicht
	// an; zwei zugesagte Katalogdimensionen waren damit ohne Wirkung.
	//
	// **Ein Katalog, der sie verschweigt, ist unvollständig** — und darauf
	// ist die Antwort `ungeklaert` und kein erfundener Wert: dieses Programm
	// kann dann nicht sagen, ob Einstellungen und Protokoll den Wechsel
	// überstehen.
	//
	// **Eine Zahl, die kleiner ist als die eigene, steht hier nicht.** Sie
	// wäre die richtige Regel — eine Fassung muss lesen, was dieser Bau
	// schreibt —, und beide Formate stehen heute auf **eins**: der Zweig
	// ließe sich von keinem Lauf betreten, und §3 nennt genau das einen
	// Fehlerpfad, der eine Behauptung ist. Wer eines der Formate erhöht,
	// schreibt die Regel dazu.
	for _, format := range []struct {
		art  Hindernisart
		wert int
	}{
		{HindernisEinstellungsformat, manifest.Einstellungen},
		{HindernisProtokollformat, manifest.Protokoll},
	} {
		if format.wert <= 0 {
			return UrteilUngeklaert, []Hindernis{{Art: format.art}}
		}
	}

	hindernisse = append(hindernisse, wanderungshindernisse(manifest, verbucht)...)

	if len(hindernisse) != 0 {
		return UrteilUnvertraeglich, hindernisse
	}

	return UrteilVertraeglich, hindernisse
}

// wanderungshindernisse vergleicht den Ledger mit dem, was die Fassung kennt.
//
// **In der Reihenfolge des Ledgers**, damit zwei Läufe dieselbe Liste geben
// und ein Vergleich etwas bedeutet.
func wanderungshindernisse(
	manifest Fassungsmanifest, verbucht map[string]string,
) []Hindernis {
	kennt := make(map[string]string, len(manifest.Wanderungen))
	for _, angabe := range manifest.Wanderungen {
		kennt[angabe.Name] = angabe.Summe
	}

	hindernisse := make([]Hindernis, 0, len(verbucht))

	for _, name := range sortiert(verbucht) {
		erwartet, bekannt := kennt[name]
		switch {
		case !bekannt:
			hindernisse = append(hindernisse, Hindernis{
				Art: HindernisWanderungFehlt, Stelle: name,
			})

		// **Eine leere Summe im Ledger ist kein Hindernis.** Ein Bestand aus
		// der Zeit vor der Prüfsummenspalte trägt sie nicht; sie gegen eine
		// gefüllte zu halten hieße, jeden alten Bestand für unverträglich zu
		// erklären, weil er alt ist.
		case verbucht[name] != "" && verbucht[name] != erwartet:
			hindernisse = append(hindernisse, Hindernis{
				Art: HindernisSummeAnders, Stelle: name,
			})
		}
	}

	return append(hindernisse, grundlinienhindernis(manifest, verbucht)...)
}

// grundlinienhindernis prüft, ob der Bestand jeden Schritt bis zur
// Grundlinie der Fassung führt (Vermerk 0189). **Ein Hindernis und nicht
// eines je Schritt**: einem Bestand aus der Zeit davor fehlen Dutzende, und
// der Rat ist für alle derselbe. Die Stelle ist die Grundlinie.
func grundlinienhindernis(manifest Fassungsmanifest, verbucht map[string]string) []Hindernis {
	if manifest.Grundlinie == "" {
		return nil
	}

	for _, angabe := range manifest.Wanderungen {
		if _, da := verbucht[angabe.Name]; !da && angabe.Name <= manifest.Grundlinie {
			return []Hindernis{{Art: HindernisVorDerGrundlinie, Stelle: manifest.Grundlinie}}
		}
	}

	return nil
}
