package domain

// **Ein Schritt an einer Karte** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Warum der Typ hier steht und nicht in `storage`
//
// **Weil ihn drei Schichten nennen müssen.** Die Ablage schreibt ihn, die
// Aufsicht prüft ihn, die Naht baut ihn — und `storage` darf `service` nicht
// kennen (§2). Ein Typ, den beide brauchen, steht in `domain`; genauso ist es
// bei [Value] und [KIFeldherkunft], aus denen er besteht.
//
// # Und warum er nichts prüft
//
// **Weil er nichts weiß, was zu prüfen wäre.** Ob ein Feld zu dieser Karte
// gehört, weiß die Aufsicht; ob der Vorschlag noch gilt, das Auftragswerk. Er
// ist ein Bündel und keine Regel.

// Kartenschritt ist, was **einer** Karte in einer atomaren Gruppe geschieht.
type Kartenschritt struct {
	KarteID int64

	// SetID ist das Set, zu dem die Karte gehört.
	//
	// **Er steht dabei, weil die Aufsicht ihn braucht**: `gehoertZu` ist die
	// Prüfung, die einen Schreibweg von außen von einem beliebigen Griff in
	// den Bestand unterscheidet.
	SetID int64

	Werte []Value

	// Herkunft ist die Marke am Feld — `ki`, `einfuhr`, leer für „von hier".
	Herkunft string

	// Herkuenfte sind die Zeilen daneben: wer, worauf, wann.
	//
	// **Ohne sie ist eine Übernahme keine** — das weist die Ablage ab, und
	// zwar für jeden Schritt einzeln.
	Herkuenfte []KIFeldherkunft

	// Angaben sind die Bildparameter je Feld (A12, #1451).
	Angaben map[string][]KIBildparameter

	// Erwartet ist der Stand, auf dem die Karte stehen muss — leer heißt:
	// ohne Bedingung.
	//
	// **Er wird in der Schreibtransaktion geprüft** und nicht davor (Befund 2
	// des Zweitblicks über Phase 5.3). Eine Prüfung vor dem Schreiben ist
	// eine Auskunft über die Vergangenheit: zwei Übernahmen desselben Feldes
	// kamen beide durch sie, und die zweite übermalte die erste still.
	Erwartet string
}
