// **Eingaben überstehen einen reinen Größenwechsel** — P01
// (#2165).
//
// `politur` startet je Größe neu und misst damit die Gestalt, nicht das, was
// eine Eingabe beim Ziehen am Rand erlebt. Hier wird **ein** Formular
// gefüllt — Name, zwei zusätzliche Sitze, ein Deck — und das Fenster
// zweimal schmal und wieder breit gestellt, wie ein Ziehen am Rand es am
// Ende tut. Danach steht alles noch da, und „Wegnehmen" steht oben rechts.
import { sprachen, vorbereiten, zumStartformular } from './politur.mjs'

const stand = (t) => `return [...document.querySelectorAll('main li')].filter((li) => li.checkVisibility() && li.querySelector('[role=group]')).map((li) => {
  const name = [...li.querySelectorAll('input')].find((i) => (i.labels?.[0]?.textContent ?? '').startsWith(${JSON.stringify(t.sitzName)}))
  const weg = [...li.querySelectorAll('button')].find((b) => b.innerText.trim() === ${JSON.stringify(t.sitzWeg)})
  const r = li.getBoundingClientRect(); const w = weg?.getBoundingClientRect()
  return { name: name?.value ?? null, deck: li.querySelector('select')?.value ?? null, wegOben: w ? Math.round(w.top - r.top) : null, breite: Math.round(r.width) }
})`

export default async function p01groesse(p) {
  const ort = p.ort()
  const t = sprachen.de
  vorbereiten(p, ort)
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort })

  p.notiz('1 — das Formular gefüllt, bei 1280 × 800')
  await zumStartformular(p, t)
  // **Ein Deck im zweiten Sitz** — über den Setzer des Feldes, wie eine Wahl.
  await p.seite("const s = [...document.querySelectorAll('main li select')].filter((e) => e.checkVisibility())[1]; const wert = [...s.options].find((o) => o.value !== '')?.value; s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })); return true")
  const vorher = await p.seite(stand(t))
  p.zeige(`vorher: ${JSON.stringify(vorher)}`)
  p.pruefe('vier Sitze, ein langer Name, ein Deck', vorher.length === 4 && vorher[0].name === 'Adelheid von Wolkenstein-Trost' && vorher[1].deck !== '')

  for (const [runde, [breite, hoehe]] of [[900, 700], [1280, 800], [900, 700], [1280, 800]].entries()) {
    p.notiz(`${String(runde + 2)} — ${String(breite)} × ${String(hoehe)}`)
    await p.groesse(breite, hoehe)
    await p.warte('die Breite ist angekommen', async () => Math.abs((await p.seite('return window.innerWidth')) - breite) < 40, 10000)
    const jetzt = await p.seite(stand(t))
    p.zeige(`${String(breite)}: ${JSON.stringify(jetzt)}`)
    p.pruefe(`${String(breite)} × ${String(hoehe)}: jede Eingabe steht noch da`, JSON.stringify(jetzt.map((s) => [s.name, s.deck])) === JSON.stringify(vorher.map((s) => [s.name, s.deck])))
    p.pruefe(`${String(breite)} × ${String(hoehe)}: „Wegnehmen" steht oben im Sitz`, jetzt.every((s) => s.wegOben !== null && s.wegOben <= 40), JSON.stringify(jetzt.map((s) => s.wegOben)))
    if (runde === 0) await p.bild('p01-schmal-nach-wechsel')
  }

  p.pruefe('⌘Q beendet', await p.beende())
}
