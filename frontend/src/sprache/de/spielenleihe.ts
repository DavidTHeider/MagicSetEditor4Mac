// **Die Wörter der beiden Dateidialoge der Leihe** — UXR-F15
// (#2030).
//
// **Ziel und Umschlag werden gewählt und nicht getippt**: ein Pfad von Hand
// ist auf macOS kein Bedienelement (R7 in `docs/UIREVIEW.md`). Die Aufschrift
// wechselt mit der Wahl — wer schon gewählt hat, will einen *anderen* Ort,
// nicht „einen“ —, und beide Fassungen stehen verschieden da, damit ein
// Vorleseprogramm die zwei Knöpfe der Seite unterscheidet.
//
// **Eine eigene Datei**, weil `seitenspielen.ts` an der Zeilengrenze aus §2
// steht; der Katalog kommt trotzdem als **einer** an (AGENTS.md §9).
export const leihewoerter = {
  leiheZielWaehlen: 'Ort wählen …',
  leiheZielAnderes: 'Anderen Ort wählen …',
  leiheZielKein: 'Noch kein Ort gewählt.',
  leiheUmschlagWaehlen: 'Umschlag wählen …',
  leiheUmschlagAnderer: 'Anderen Umschlag wählen …',
  leiheUmschlagKein: 'Noch kein Umschlag gewählt.',
}
