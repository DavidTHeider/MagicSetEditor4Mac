// Gemeinsame Einrichtung für alle Frontend-Tests.
//
// `jest-dom` bringt die Zusicherungen über den DOM mit (`toBeInTheDocument`,
// `toHaveTextContent`). `cleanup` nach jedem Fall, damit ein Test nicht den
// gerenderten Baum des vorigen sieht — Vitest räumt das ohne `globals` nicht
// von sich aus ab.
import '@testing-library/jest-dom/vitest'
import { cleanup, configure } from '@testing-library/react'
import { afterEach, beforeEach } from 'vitest'

import { gibJsdomEinLayout } from './listenmass'
import { vergissVorlagenentwuerfe } from '../haken/useVorlagenpuffer'
import { vergissVorrangreihe } from '../haken/useVorrangreihe'

// **Die Obergrenze für `waitFor` steht bei fünf Sekunden statt einer.**
//
// Nicht, um einer Zusicherung mehr Zeit zu geben — `waitFor` kehrt zurück,
// sobald sie zutrifft, und eine höhere Grenze ändert daran nichts. Sondern
// weil ein Lauf auf einem **beschäftigten** Rechner sonst an der Uhr scheitert
// statt an der Sache: `make ci` fährt die Go-Läufe daneben, und am 2026-08-20
// war `appsichern.test.tsx` genau dort einmal rot und dreimal einzeln grün.
//
// **Was das kostet**: ein Fall, der wirklich scheitert, braucht fünf Sekunden
// statt einer. Bei 656 Fällen und seltenen Fehlschlägen ist das der bessere
// Tausch.
//
// **Und was es nicht heilt**: eine Zusicherung, die schon vor der Antwort
// zutrifft (§3). Dagegen hilft keine Uhr.
//
// **Fünfzehn Sekunden seit dem 2026-09-09.** Drei Läufe von `make ci` an
// diesem Tag endeten rot an einem `findBy…` in drei verschiedenen Dateien
// (`symbolpalette`, `sammlungsgrenze`, `decks`), jeder Fall einzeln in unter
// einer Sekunde grün — während daneben zwei Blender-Prozesse mit je 350 %
// CPU rechneten und zehn Vitest-Arbeiter sich den Rest teilten. Ein Arbeiter,
// der fünf Sekunden keinen Kern bekommt, verpasst die Antwort, die längst da
// ist. Die Zusage bleibt dieselbe: `waitFor` kehrt zurück, sobald sie hält.

configure({ asyncUtilTimeout: 15_000 })

// **jsdom bekommt ein Layout** (#1166).
//
// Ohne dieses Paar sieht `useSichtfenster` nie ein Maß und zeichnet, was es
// ohne Maß zeichnet — bis zum 2026-08-27 war das die **ganze** Liste, und
// deshalb prüfte niemand je das Fenster, das im Betrieb wirklich entsteht.
// Der Grund und die Zahlen stehen in `listenmass.ts`.
//
// **Die Vorgabe räumt nichts** — sie steht nur da, bis `beforeEach` gelaufen
// ist, und wird nie gerufen: Vitest führt `beforeEach` vor jedem Fall aus.
let raeumeLayout: () => void = () => undefined

beforeEach(() => {
  raeumeLayout = gibJsdomEinLayout()
})

afterEach(() => {
  cleanup()
  raeumeLayout()

  // **Die Entwürfe der Vorlagenwerkstatt überleben das Abbauen** — das ist
  // ihr Zweck (UX-03.AK-04, #1675). Zwischen zwei Fällen derselben Datei
  // wären sie ein Rest des vorigen; im Betrieb räumt sie der Profilwechsel.
  vergissVorlagenentwuerfe()
  vergissVorrangreihe()

  // **Und der Speicher der Sitzung** (#2428): die geschlossene Einführung
  // steht dort, und ein Fall danach sähe sie sonst nie.
  sessionStorage.clear()
})

// **jsdom kennt `ResizeObserver` nicht** — es ist dort schlicht nicht
// implementiert, und ein `new ResizeObserver(…)` wirft.
//
// Die Kartenliste braucht ihn seit
// #891: sie
// zeichnet nur, was zu sehen ist, und muss deshalb wissen, wenn ihr Fenster
// sich ändert. Ein `typeof`-Wächter im Quelltext wäre ein Zweig, den es nur
// für die Prüfung gäbe — die Lücke steht hier, wo sie hingehört.
//
// **Und er misst nichts von selbst.** jsdom rechnet kein Layout; die Höhen
// kommen deshalb aus `listenmass.ts`, das sie oben für jeden Lauf setzt.
// Was das Fenster **rechnet**, prüft `sichtfenster.test.ts` mit eigenen Maßen.
// **Eine Zuweisung und kein `??=`.** TypeScript kennt `ResizeObserver` aus der
// DOM-Bibliothek und hält ihn für immer vorhanden; `??=` wäre für den Linter
// eine Bedingung, die nie greift. Vorhanden ist er hier trotzdem nicht.
globalThis.ResizeObserver = class {
  observe() {
    return undefined
  }

  unobserve() {
    return undefined
  }

  disconnect() {
    return undefined
  }
}

// **jsdom kennt `matchMedia` nicht** — UX-07
// (#1679).
//
// Dieselbe Lage wie beim `ResizeObserver` darüber, und derselbe Grund, es
// hier zu beantworten: `MantineProvider` fragt danach, um dem Fenster zu
// folgen — ohne ihn wirft **schon der Provider**, und damit jeder Lauf, der
// die Anwendung zeichnet.
//
// **Die Antwort ist immer »nein«**, und das ist eine Aussage: unter jsdom gibt
// es keine Systemeinstellung, der man folgen könnte. Wer eine Abfrage prüfen
// will — `prefers-reduced-motion` etwa —, gibt sie selbst mit, so wie
// `bewegungErwuenscht` es verlangt.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (frage: string): MediaQueryList => ({
    matches: false,
    media: frage,
    onchange: null,
    addListener: () => undefined,
    removeListener: () => undefined,
    addEventListener: () => undefined,
    removeEventListener: () => undefined,
    dispatchEvent: () => false,
  }),
})

// **jsdom kennt `scrollIntoView` nicht** (#1168).
//
// Dieselbe Lage wie beim `ResizeObserver` darüber und dieselbe Antwort: jsdom
// rechnet kein Layout und rollt deshalb auch nicht. Der Menüeintrag „Über"
// führt auf einen Abschnitt weit unten in der Hilfe und rollt ihn heran; ein
// `typeof`-Wächter im Quelltext wäre wieder ein Zweig, den es nur für die
// Prüfung gäbe.
//
// **Was das kostet**: dass wirklich gerollt wird, sieht kein Test. Es steht in
// der Sichtprüfung (§3), und die Naht davor — der Abschnitt hat den Anker, auf
// den der Effekt zeigt — prüft `menueleiste.test.tsx`.
// **Zwei Kinder mit demselben Schlüssel sind ein Fehler und keine Warnung.**
// Gesehen am 2026-09-24 in WebKit: auf *Handeln* heißen Bereich und Seite
// gleich, der Schlüssel der Brotkrume war der Name, und beim nächsten Wechsel
// blieb eine Station stehen. React sagt das nur auf der Konsole; hier
// scheitert der Lauf, der es auslöst.
const meldeFehler = console.error.bind(console)

console.error = (...teile: unknown[]) => {
  const satz = teile[0]

  if (
    typeof satz === 'string' &&
    satz.includes('Encountered two children with the same key')
  ) {
    throw new Error(satz.replace('%s', String(teile[1])))
  }

  meldeFehler(...teile)
}

Element.prototype.scrollIntoView = function scrollIntoView() {
  return undefined
}
