package domain

// **Was ein Gegenüber von einem Exemplar sieht** — Werkstück `T3` von Phase
// 5.6 (#1468).
//
// # Warum es dafür eine eigene Gestalt gibt
//
// **Weil ein [Exemplar] ein privates Stück Papier ist.** Es trägt, was jemand
// dafür bezahlt hat, wo es liegt, was er sich dazu notiert hat und aus wessen
// Profil es kam. Nichts davon geht einen Handelspartner an, und nichts davon
// braucht er: er will wissen, **welche Karte in welchem Zustand** angeboten
// wird.
//
// **Der bequeme Weg wäre, das Exemplar mit `omitempty` hinauszuschicken.** Er
// ist der falsche: was heute leer ist, ist morgen gefüllt, und ein Feld, das
// jemand später hinzufügt, geht von selbst mit. Eine eigene Gestalt geht nur
// dann mit, wenn jemand sie einträgt — und dann sieht er dabei, dass er die
// Fläche vergrößert.
//
// # Und warum die Referenz nicht die Kennung ist
//
// **Weil eine Kennung über Sitzungen hinweg wiedererkennbar ist.** Wer zweimal
// dieselbe Zahl hört, weiß, dass es dasselbe Stück ist — und über mehrere
// Sichten hinweg, wie groß der Bestand ist und wie er sich ändert. Die
// Referenz ist deshalb **sitzungsgebunden**: sie zeigt innerhalb eines
// Gesprächs auf dieselbe Zeile und sagt außerhalb nichts.

import (
	"slices"
	"time"
)

// Sichtkarte ist eine Bestandszeile, wie sie über den Draht geht.
//
// **Sechs Felder, und jedes hat seinen Grund** — das Issue zählt sie auf:
// öffentliche Druckreferenz, Menge, Zustand, Sprache, Ausführung und die
// lesbare Bezeichnung. Dazu die Referenz, über die sich eine Position eines
// Angebots auf diese Zeile bezieht.
type Sichtkarte struct {
	// Referenz ist der **sitzungsgebundene** Name dieser Zeile.
	Referenz string `json:"referenz"`

	// Printing ist die **öffentliche** Druckreferenz — die Scryfall-Kennung.
	//
	// **Sie ist kein Geheimnis**: sie benennt einen Druck, den es millionenfach
	// gibt, und ohne sie wüsste das Gegenüber nicht, welche Karte gemeint ist.
	// Die **Zeile** dieses Bestands ([Exemplar.DruckID]) geht nicht mit.
	Printing string `json:"printing"`

	Menge   int     `json:"menge"`
	Zustand Zustand `json:"zustand"`
	Sprache string  `json:"sprache"`
	Finish  Finish  `json:"finish"`

	// Bezeichnung ist der lesbare Name — damit niemand eine UUID lesen muss.
	Bezeichnung string `json:"bezeichnung"`

	// Wunsch heißt: diese Zeile wird **gesucht** und nicht angeboten — P11
	// (#2379).
	// **Nur bei einem Wunsch im Draht**: eine ältere Fassung kennt das Feld
	// nicht und las jede Zeile als Angebot; das tut sie weiter, und eine
	// Wunschzeile von ihr bleibt so missverständlich wie vorher.
	Wunsch bool `json:"wunsch,omitempty"`
}

// AlsSichtkarte baut aus einem Exemplar, was hinausgehen darf.
//
// **Feld für Feld und ausdrücklich.** Ein Umweg über eine gemeinsame Struktur
// oder ein `omitempty` nähme diese Entscheidung aus der Hand — und dann ginge
// beim nächsten neuen Feld etwas hinaus, das niemand gewählt hat.
//
// **Die Ausführung wird beantwortet und nicht durchgereicht** (T2, 0119):
// leer heißt `nonfoil`, und das Gegenüber soll kein viertes Wort hören.
func AlsSichtkarte(exemplar Exemplar, referenz string) Sichtkarte {
	return Sichtkarte{
		Referenz:    referenz,
		Printing:    exemplar.Printing,
		Menge:       exemplar.Menge,
		Zustand:     exemplar.Zustand,
		Sprache:     exemplar.Sprache,
		Finish:      exemplar.Ausfuehrung(),
		Bezeichnung: exemplar.Bezeichnung(),
		Wunsch:      exemplar.Wunsch,
	}
}

// Freigabeumfang sagt, **was** freigegeben ist.
//
// **Fünf, und sie stehen einzeln da.** Das Issue zählt sie auf: Dubletten,
// Wünsche, ausgewählte Karten, die ganze Sammlung und digitale Objekte. Ein
// Schalter „meine Sammlung" mit einer Liste von Ausnahmen wäre kürzer und
// gefährlicher — wer eine Ausnahme vergisst, gibt mehr frei, als er wollte.
type Freigabeumfang string

const (
	// UmfangDubletten ist, was über der Behaltemenge liegt.
	UmfangDubletten Freigabeumfang = "dubletten"

	// UmfangWuensche ist, was jemand sucht — **kein Besitz**.
	UmfangWuensche Freigabeumfang = "wuensche"

	// UmfangAuswahl sind einzelne, ausdrücklich gewählte Drucke.
	UmfangAuswahl Freigabeumfang = "auswahl"

	// UmfangSammlung ist alles, was besessen wird.
	UmfangSammlung Freigabeumfang = "sammlung"

	// UmfangDigitales sind Objekte ohne Papier — eigene Karten und Sets.
	UmfangDigitales Freigabeumfang = "digitales"
)

// DieFreigabeumfaenge ist die Aufzählung in fester Reihenfolge.
//
// **Von eng nach weit**, damit eine Fläche sie in der Ordnung anbietet, in der
// jemand sie liest — und nicht mit „die ganze Sammlung" an erster Stelle.
var DieFreigabeumfaenge = []Freigabeumfang{
	UmfangAuswahl, UmfangDubletten, UmfangWuensche, UmfangSammlung,
	UmfangDigitales,
}

// IstFreigabeumfang sagt, ob eine Angabe zur Aufzählung gehört.
func IstFreigabeumfang(umfang Freigabeumfang) bool {
	for _, bekannt := range DieFreigabeumfaenge {
		if bekannt == umfang {
			return true
		}
	}

	return false
}

// Sammlungsfreigabe ist die Erlaubnis, einem Gegenüber etwas zu zeigen.
//
// # Der Satz, um den herum das gebaut ist
//
// **Ohne Freigabe sieht niemand etwas.** Das Issue sagt es zweimal: „Standard
// alles geschlossen" und „Beispielcheckboxen im Entwurf sind keine schon
// erteilte Zustimmung". Eine fehlende Freigabe ist deshalb kein Zwischenfall,
// sondern der Normalfall — und der Weg, auf dem jemand ohne sie etwas sähe,
// darf es nicht geben.
type Sammlungsfreigabe struct {
	// Empfaenger ist die **Peerkennung** des Gegenübers (N1, #1454).
	//
	// **Die Kennung und nicht der Name**: sie ist der Nachweis, ein Name ist
	// eine Behauptung. Zwei Gegenüber, die sich gleich nennen, haben
	// verschiedene Kennungen.
	Empfaenger string

	// Umfang ist, was gezeigt wird.
	Umfang Freigabeumfang

	// Zweck ist, wofür (N1): eine Freigabe zum Teilen ist keine zum Handeln.
	Zweck Zweck

	// Auswahl sind die Druckreferenzen bei [UmfangAuswahl] — sonst leer.
	Auswahl []string

	// Behalten ist die Behaltemenge bei [UmfangDubletten].
	//
	// **Sie steht an der Freigabe und nicht am Bestand** (0057): „Dublette"
	// ist kein Zustand, sondern ein Verhältnis — *mehr als ich behalten
	// will*. Wer einem Gegenüber seine Dubletten zeigt, sagt dabei, was er
	// behält; einem anderen sagt er vielleicht etwas anderes.
	Behalten int

	// Ablauf ist, bis wann sie gilt — der **Nullwert heißt: unbefristet**.
	//
	// **Und nicht „schon vorbei"**: eine Freigabe ohne Ablauf ist die
	// gewöhnliche, und wer den Nullwert als Vergangenheit läse, machte jede
	// von ihnen wirkungslos.
	Ablauf time.Time

	// Erteilt ist, wann sie gegeben wurde — für die Spur, nicht für die Regel.
	Erteilt time.Time
}

// Gilt sagt, ob diese Freigabe jetzt noch etwas erlaubt.
func (f Sammlungsfreigabe) Gilt(jetzt time.Time) bool {
	return f.Ablauf.IsZero() || jetzt.Before(f.Ablauf)
}

// Zeigt sagt, ob und mit welcher Menge diese Freigabe ein Exemplar zeigt.
//
// # Warum beides aus einer Frage kommt
//
// **Weil die Menge vom Umfang abhängt.** Wer vier Stück hat und eines behält,
// bietet unter „Dubletten" **drei** an und nicht vier; unter „Sammlung" sind
// es vier. Zwei getrennte Fragen ließen den Aufrufer die Regel ein zweites Mal
// rechnen — und die zweite Fassung ist irgendwann eine andere.
//
// # Und warum der unbekannte Umfang nichts zeigt
//
// **Weil das Nein der sichere Ausgang ist.** Ein Umfang, den dieses Programm
// nicht kennt, steht im Bestand, weil ihn eine neuere Fassung geschrieben hat
// oder weil jemand ihn von Hand eingetragen hat. Beide Male ist „ich weiß
// nicht, was das heißt" ein Grund, nichts zu zeigen — und nicht, alles zu
// zeigen.
func (f Sammlungsfreigabe) Zeigt(exemplar Exemplar) (menge int, zeigt bool) {
	switch f.Umfang {
	case UmfangWuensche:
		if exemplar.Wunsch {
			return exemplar.Menge, true
		}

	case UmfangSammlung:
		if exemplar.Besitzt() {
			return exemplar.Menge, true
		}

	case UmfangDubletten:
		if uebrig := exemplar.Menge - f.Behalten; exemplar.Besitzt() && uebrig > 0 {
			return uebrig, true
		}

	case UmfangAuswahl:
		if exemplar.Besitzt() && slices.Contains(f.Auswahl, exemplar.Printing) {
			return exemplar.Menge, true
		}

	// **Ein digitales Objekt ist kein Exemplar.** Der Umfang gibt eigene
	// Karten und Sets frei; über den Bestand sagt er nichts.
	case UmfangDigitales:
	}

	return 0, false
}
