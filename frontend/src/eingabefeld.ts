// **Ob der Fokus in einem Feld steht, das selbst tippt.**
//
// **Aus `rahmen/Rahmen.tsx` herausgelöst** — Werkstück 1 von Phase 3.7
// (#1168):
// die Kartenliste stellt dieselbe Frage für ⌘C/⌘X/⌘V. Zwei Antworten darauf
// liefen auseinander, und die zweite fiele still aus.

// tippendeInputs sind die Eingabearten, die **nicht** tippen.
//
// **Eine Verneinung und keine Aufzählung der tippenden.** Eine neue Art —
// HTML bekommt sie gelegentlich — ist im Zweifel eine, in die man schreibt;
// wer die tippenden aufzählte, nähme ihr das ⌘C, ohne es zu merken.
const tippenNicht = new Set([
  'button',
  'checkbox',
  'color',
  'file',
  'image',
  'radio',
  'range',
  'reset',
  'submit',
])

// inEingabe sagt, ob der Fokus in einem Feld steht, das selbst tippt.
//
// **Über den Knotennamen und das Attribut**, nicht über eine Liste von
// Kennungen: ein Feld, das jemand später hinzufügt, wäre in einer Liste nicht
// drin — und ⌘Z nähme dort eine fremde Änderung zurück, während der Cursor in
// einer ungesicherten steht.
//
// **`closest` statt `isContentEditable`.** Die Eigenschaft wäre die genauere
// Frage, aber jsdom setzt sie nicht — der Zweig wäre einer, den kein Test je
// beträte, und damit eine Behauptung darüber, was im Betrieb geschieht (§3).
// `closest` beantwortet dieselbe Frage über das Attribut und trägt die
// Vererbung mit: ein `<b>` in einem bearbeitbaren Bereich findet ihn.
// `contenteditable="false"` passt auf keinen der beiden Wähler und gilt
// deshalb zu Recht nicht als Feld.
//
// # Die Eingabeart zählt mit, und das ist eine Änderung
//
// Hier stand `ziel.tagName === 'INPUT'` ohne weitere Frage. Ein **Ankreuzfeld**
// ist ein `INPUT` und tippt nicht — es steht in jeder Zeile der Kartenliste,
// und wer eine Karte anhakt, hat den Fokus darauf. ⌘C meinte dort die Karten
// und traf nichts.
//
// **Was sich dadurch sonst ändert**: ⌘Z auf einem Ankreuzfeld nimmt jetzt die
// letzte fachliche Änderung zurück, statt nichts zu tun. Das ist die
// gemeinte Bedeutung — ein Ankreuzfeld kennt keine eigene Rücknahme, die
// näher läge (§15, Regel 4: gesagt, nicht stillschweigend).
export function inEingabe(ziel: EventTarget | null): boolean {
  if (!(ziel instanceof HTMLElement)) {
    return false
  }

  if (ziel instanceof HTMLInputElement) {
    return !tippenNicht.has(ziel.type)
  }

  return (
    ziel.tagName === 'TEXTAREA' ||
    ziel.closest('[contenteditable=""], [contenteditable="true"]') !== null
  )
}
