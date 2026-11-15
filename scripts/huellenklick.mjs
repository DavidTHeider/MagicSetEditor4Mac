// **Ein nativer Klick, und was er traf** — aus `huellenprobe.mjs`
// herausgelöst, als die Belege dafür (R04,
// #2094)
// die Datei über die Zeilengrenze aus §2 brachten.
//
// Der Klick geht über die Abnahmeleitung in die Ereignisschlange von AppKit;
// die Seite meldet über einen `click`-Horcher, ob er das gemeinte Element
// traf. **Was das nicht belegt**: die Hand vor AppKit — eine echte Maus.
import { schlafe } from './huellenleitung.mjs'

// klickeAuf sucht das Element über JavaScript und klickt über die
// Ereignisschlange. **Gerollt wird ohne Bewegung** (`instant`): ein sanftes
// Rollen maß die Stelle vorher, und der Klick ging an die alte.
//
// **Und es kehrt erst zurück, wenn der Klick angekommen ist** — am `mouseup`
// in der Seite, nach dem auch der Fokus gewechselt hat. Der Klick kam bis zum
// 2026-10-05 später an als der nächste Ruf in die Seite: `ops06` las danach
// den Wert des Feldes, das den Fokus noch hatte. **Kommt keines an** — ein
// natives Menü nimmt es —, geht es nach drei Sekunden weiter wie vorher.
//
// **Umschalter und Klickzahl gehen mit** (`art`, P4 #2372): ⌘-Klick wählt am
// Tisch, ein Doppelklick stellt die Zweithandlung. Belegt wird der erste
// Klick der Folge.
export async function klickeAuf(probe, ausdruck, beschreibung = ausdruck, { mit = [], anzahl = 1 } = {}) {
  // **Ein Ziel, das sich noch bewegt, wird abgewartet** — zweimal dieselbe
  // Mitte, höchstens eine halbe Sekunde (R04, #2094).
  const messen = `const e = (${ausdruck}); if (!e || !e.checkVisibility()) return false; e.scrollIntoView({ block: 'center', behavior: 'instant' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`
  let stelle = await probe.warte(`da ist: ${beschreibung}`, () => probe.seite(messen))
  for (let versuch = 0; versuch < 10; versuch += 1) {
    await schlafe(50)
    const jetzt = await probe.seite(messen)
    if (jetzt && jetzt.x === stelle.x && jetzt.y === stelle.y) break
    stelle = jetzt || stelle
  }
  // **Und belegt wird, was der Klick traf** — bis zum 2026-10-06 wartete
  // die Probe auf irgendein `mouseup` und ging nach drei Sekunden weiter,
  // ob der Klick ankam oder nicht.
  //
  // **Und eine neue Seite erkennt die Marke der alten nicht** — so gilt ein
  // Klick, der navigiert, als angekommen. **Ein Lesefehler gilt nicht**
  // (Zweitblick über A und B, Befund 5): bis dahin fing der Leser jeden
  // Fehler und meldete „traf", auch eine Leitung, die gar nicht antwortete.
  const marke = `klick-${String(Date.now())}-${String(Math.random()).slice(2, 8)}`
  await probe.seite(`window.__mseSeite = '${marke}'; window.__mseKlick = null; const ziel = (${ausdruck}); document.addEventListener('click', (ereignis) => { window.__mseKlick = { traf: !!ziel && ziel.contains(ereignis.target), was: ((ereignis.target.innerText ?? '') || ereignis.target.tagName).trim().replace(/\\s+/g, ' ').slice(0, 60) } }, { capture: true, once: true }); return true`)
  await probe.frage({ was: 'klicke', x: stelle.x, y: stelle.y, ...(mit.length > 0 ? { mit } : {}), ...(anzahl > 1 ? { anzahl } : {}) })
  const lesen = `return window.__mseSeite === '${marke}' ? (window.__mseKlick ?? null) : { traf: true, was: 'eine neue Seite' }`
  const bis = Date.now() + 3000
  let klick = null
  let lesefehler = null
  while (Date.now() < bis && !klick) {
    try {
      klick = await probe.seite(lesen)
      lesefehler = null
    } catch (fehler) {
      lesefehler = fehler
    }
    if (!klick) await schlafe(20)
  }
  if (!klick && lesefehler) throw new Error(`der Klickbeleg zu ${beschreibung} ließ sich nicht lesen: ${lesefehler.message}`)
  if (!klick) throw new Error(`der Klick auf ${beschreibung} kam nicht an (${String(Math.round(stelle.x))}, ${String(Math.round(stelle.y))})`)
  if (!klick.traf) throw new Error(`der Klick auf ${beschreibung} traf »${klick.was}« (${String(Math.round(stelle.x))}, ${String(Math.round(stelle.y))})`)
}

// klappeAuf öffnet eine Falte nativ und verlangt, dass sie danach offen
// ist — `aria-expanded` am Knopf (R04, #2094).
export async function klappeAuf(probe, ausdruck, beschreibung = ausdruck) {
  await probe.klickeAuf(ausdruck, beschreibung)
  await probe.warte(`${beschreibung} ist offen`, () => probe.seite(`return (${ausdruck})?.getAttribute('aria-expanded') === 'true'`), 5000)
}
