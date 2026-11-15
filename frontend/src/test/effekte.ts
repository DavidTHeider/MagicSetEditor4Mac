// Reacts Effekte nachholen, bevor ein Test in einen Horcher hineinruft
// (#761).
import { act } from '@testing-library/react'

// effekteNachholen lässt React seine ausstehenden Effekte laufen.
//
// # Was `await findByRole(…)` nicht zusagt
//
// **Dass ein `useEffect` gelaufen ist.** Testing Library schaltet für die Dauer
// jeder asynchronen Abfrage die act-Umgebung ausdrücklich **ab** und gibt
// danach genau eine Makroaufgabe als Frist:
//
//	asyncWrapper: async cb => {
//	  const previousActEnvironment = getIsReactActEnvironment()
//	  setReactActEnvironment(false)
//	  try {
//	    const result = await cb()
//	    // Drain microtask queue.
//	    await new Promise(resolve => { setTimeout(() => { resolve() }, 0) })
//
// `@testing-library/react@16.3.2`, `dist/pure.js`, Zeilen 83–99 — der Kommentar
// „We just want to run `waitFor` without IS_REACT_ACT_ENVIRONMENT" steht dort
// darüber, es ist also Absicht und kein Versehen.
//
// React legt seine passiven Effekte über den Scheduler in eine **andere**
// Aufgabe. Meist ist er vor der Frist fertig. Unter Last nicht — und dann steht
// der Knoten schon im DOM, während der Horcher daneben noch nicht hängt.
//
// # Warum `waitFor` das nicht heilt
//
// **Es wiederholt die Zusicherung, nicht das Ereignis.** Ein Tastendruck, der
// ins Leere ging, kommt nicht wieder; die Zusicherung danach läuft in ihre
// Zeitgrenze, gleich wie lange man wartet.
//
// # Gezählt, nicht erschlossen
//
// **Vierzig volle Frontendläufe am 2026-08-21, vier davon rot** — und in
// **zwei** verschiedenen Tests derselben Datei, was den gemeinsamen Grund
// zeigt: einmal ein verlorener Tastendruck (`aria-current` blieb `null`),
// einmal ein `runtime`-Horcher, den die `Map` noch nicht kannte und den `?.()`
// still fallen ließ.
//
// **Und die Ursache steht als Zahl da**: zwölf volle Läufe unter Last auf zehn
// Kernen, dabei `window.addEventListener` mitgezählt. Einmal von zwölf hingen
// nach `await band()` **null** Tastenhorcher — und nach dieser Funktion einer.
// Dieselbe Messung belegt beide Hälften: den Grund und die Abhilfe.
// # Und warum sie **nicht** global hängt
//
// **Weil sie einen anderen Fehler zudeckt.** Am 2026-09-02 lag der Vorschlag
// da, sie über `configure({ asyncWrapper })` an **jede** asynchrone Abfrage
// zu hängen — die Abhilfe stand in 15 Testdateien, während 137 die gefährdete
// Gestalt hatten. Der Anlass war ein flackernder Lauf von
// `formatzeile.test.tsx`.
//
// **Die Messung hat den Vorschlag widerlegt**, und zwar indem sie den
// Flackerer endlich zu fassen bekam: acht volle Läufe ohne den Wrapper, einer
// davon rot, und die Meldung lautete
//
//	Unable to find a label with the text of: Sideboard ohne Grenze
//	<section aria-label="Formate">…Noch kein Format.…
//
// Der Test wartete auf `findByRole('region', { name: 'Formate' })` — und den
// Bereich gibt es **auch im leeren Zustand**, mit demselben Namen. Die
// Zusicherung traf zu, bevor die Antwort da war (§3, „eine Zusicherung, die
// schon vor der Antwort zutrifft"). Das ist ein Fehler **im Test**, kein
// Wettlauf um passive Effekte.
//
// **Ein globales `act` hätte ihn grün gemacht**: es treibt die ausstehende
// Zustandsänderung mit durch, der geladene Zustand steht da, und die Abfrage
// findet ihr Feld. Grün, und die Zusicherung bleibt falsch — an jeder der 137
// Stellen, für immer.
//
// Deshalb bleibt sie eine Funktion, die man **hinschreibt**: dort, wo jemand
// begründen kann, dass ein Effekt gelaufen sein muss.
export async function effekteNachholen(): Promise<void> {
  // **Die Wirkung liegt im Verlassen, nicht im Rumpf.** `act` treibt dabei
  // Reacts Arbeitsschleife leer, und dazu gehören die ausstehenden passiven
  // Effekte. Ein leerer Rumpf täte dasselbe — der Linter liest ihn allerdings
  // als Versehen, und ein `disable` daneben wäre eine Ausnahme für nichts.
  await act(async () => {
    await Promise.resolve()
  })
}
