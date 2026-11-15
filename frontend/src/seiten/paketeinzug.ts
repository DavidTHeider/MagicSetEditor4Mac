// Der Einzug der Paketliste und die Klasse ihrer Gruppenaufschrift — geteilt von
// `Paketzeilen.tsx` und `Paketgruppenzeile.tsx`.

// einzug gibt den linken Abstand einer Stufe — als Stil, nicht als Klasse.
//
// **Und das ist eine Ausnahme von „wie es aussieht, steht in den Klassen"**
// (§2), mit Grund. Tailwind gibt nur aus, was es im Quelltext **liest**; ein
// `pl-[${stufe * 16}px]` steht nirgends im Stylesheet und wirkt nicht. Bliebe
// eine Liste fertiger Klassen — `['pl-0', 'pl-4', …]` —, und die hat zwei
// Mängel: sie deckelt die Tiefe (die Packs gehen bis vier Stufen, mehr ist
// erlaubt), und der Zugriff darauf braucht einen Rückfall, den kein Aufruf je
// erreicht. Ein Zweig ohne Fall ist genau das, was §3 nicht will.
const einzugJeStufe = 16

export function einzug(stufe: number) {
  return { paddingLeft: `${String(stufe * einzugJeStufe)}px` }
}

// gruppenaufschrift ist die Klasse der Aufschrift einer Gruppenzeile — seit
// UXR-17 (#1925) ein Text in der Zeile, die selbst schaltet, und kein Knopf.
export const gruppenaufschrift =
  'flex w-full items-center gap-2 text-left font-semibold'
