// **Die kleinen Ollama-Modelle am nativen Fenster** — AI-02
// (#1743):
//
// - unter *Einstellungen › KI* stehen die gezogenen Modelle samt dem
//   Kontextfenster, das die Anwendung für sie gelesen hat,
// - am Kartenformular holt „Vorschlagen" am **Namen** einen Vorschlag von
//   einem **Ollama-Modell**, wenn das Gerätemodell aus ist — über den Vorrang,
//   ohne dass jemand ein Modell wählt; welches es nimmt, sagt die Herkunft.
//
// **Braucht eine laufende Ollama-Laufzeit unter 127.0.0.1:11434 mit einem
// Textmodell.** Fehlt sie, endet der Lauf rot und nicht grün — **und sagt es
// vor dem Start**, mit dem Bestand, den es vorfand (R02, #2107). Bis zum
// 2026-10-06 wartete er dreißig Sekunden auf eine Zeile, die ohne Laufzeit
// nie kommt, und verlangte genau vier Modelle; der Bestand ist jetzt der, den
// Ollama selbst nennt. **Gezogen wird nichts.**
import { oeffnen, zurSeite } from './uirevue.mjs'

const ollama = 'http://127.0.0.1:11434'

// modellbestand fragt Ollama, was gezogen ist — je Modell seine Fähigkeiten
// und, wie die Anwendung es liest, `<architektur>.context_length`
// (`internal/ollama/kontextfenster.go`). `{ fehlt }` nennt, woran es liegt.
async function modellbestand() {
  const frage = (weg, rumpf) => fetch(`${ollama}${weg}`, {
    method: rumpf ? 'POST' : 'GET', body: rumpf ? JSON.stringify(rumpf) : undefined, signal: AbortSignal.timeout(5000),
  }).then((r) => r.json())
  let liste
  try {
    liste = (await frage('/api/tags')).models ?? []
  } catch (grund) {
    return { fehlt: `Ollama antwortet nicht unter ${ollama} (${String(grund.cause?.code ?? grund.name ?? grund)})` }
  }
  if (liste.length === 0) return { fehlt: 'Ollama läuft, aber es ist kein Modell gezogen' }
  const bestand = []
  for (const modell of liste) {
    const schau = await frage('/api/show', { model: modell.name })
    const architektur = schau.model_info?.['general.architecture']
    bestand.push({ name: modell.name, faehig: schau.capabilities ?? [], kontext: schau.model_info?.[`${architektur}.context_length`] })
  }
  if (!bestand.some((m) => m.faehig.includes('completion'))) return { fehlt: `kein Textmodell unter ${bestand.map((m) => m.name).join(', ')}` }

  return { bestand }
}

const namensfeld = "document.querySelector('input[aria-label=\"name\"], textarea[aria-label=\"name\"]')"
const vorschlagsknopf = `[...${namensfeld}.closest('label').querySelectorAll('button')].find((b) => b.innerText.trim() === 'Vorschlagen')`

export default async function ai02(p) {
  p.notiz('Die Vorbedingung — Ollama und sein Bestand, vor dem Start')
  const vorrat = await modellbestand()
  p.pruefe('Vorbedingung: Ollama läuft mit einem Textmodell', !vorrat.fehlt, vorrat.fehlt ?? vorrat.bestand.map((m) => `${m.name} [${m.faehig.join(', ')}] ${String(m.kontext ?? 'ohne Kontextlänge')}`).join('; '))
  if (vorrat.fehlt) return

  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'einstellungen', 'setzen', 'ki.laufzeit', 'an')
  p.mse(ort, 'einstellungen', 'setzen', 'ki.geraetemodell', 'aus')

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })

  p.notiz('Einstellungen › KI — die gezogenen Modelle mit ihrem Kontextfenster')
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', () => p.seite("return !!document.querySelector('[role=tab]')"), 20000)
  await p.klickeText('KI')
  await p.warte('der Abschnitt KI steht da', () => p.seite("return [...document.querySelectorAll('main h2, main h3')].some((h) => h.innerText.trim() === 'KI')"), 20000)
  // **Die Falten öffnen** (UXR-21): die Modellliste steht in einer, und eine
  // geschlossene Falte zeigt sie nicht.
  await p.seite("for (const falte of document.querySelectorAll('main [aria-expanded=false]')) falte.click(); return true")
  // **Gewartet wird auf einen Treffer**, nicht auf eine Liste: eine leere
  // ist in JavaScript wahr, und der erste Lauf maß damit die falsche Seite.
  const fenster = await p.warte('ein Kontextfenster steht da', async () => {
    const zeilen = await p.seite("return [...document.querySelectorAll('main *')].filter((e) => e.children.length === 0 && e.innerText?.startsWith('Kontextfenster')).map((e) => e.innerText.trim())")
    return zeilen.length > 0 ? zeilen : ''
  }, 30000)
  for (const zeile of fenster) p.zeige(zeile)
  // **Mit Zahl**: das Bildmodell steht mit „nicht angegeben" daneben, und
  // eine Prüfung auf „mindestens vier Zeilen" zählte es mit. **Erwartet ist,
  // was Ollama nennt** — je Modell mit Kontextlänge eine Zeile mit Zahl.
  const mitZahl = fenster.filter((zeile) => /\d/.test(zeile))
  const erwartet = vorrat.bestand.filter((m) => Number.isInteger(m.kontext)).length
  p.pruefe('jedes Modell mit Kontextlänge nennt sein Kontextfenster', mitZahl.length === erwartet, `${String(mitZahl.length)} von ${String(erwartet)}`)
  await p.seite("[...document.querySelectorAll('main *')].find((e) => e.children.length === 0 && e.innerText?.startsWith('Kontextfenster'))?.scrollIntoView({ block: 'center', behavior: 'instant' }); return true")
  await p.bild('ai02-einstellungen')
  await p.taste('Escape')

  p.notiz('Ein Vorschlag am Namen — das Gerätemodell ist aus, also fragt der Vorrang Ollama')
  await zurSeite(p, 'Karten')
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!${namensfeld}`), 20000)
  await p.klickeAuf(vorschlagsknopf, '»Vorschlagen« am Namen')
  const ausgang = await p.warte('am Feld steht eine Antwort', () =>
    p.seite(`const feld = ${namensfeld}.closest('label'); const von = [...feld.querySelectorAll('span')].map((s) => s.innerText.trim()).find((t) => t.startsWith('von ')); if (von) return von; const meldung = feld.querySelector('[role=alert], [role=status]'); return meldung ? 'Meldung: ' + meldung.innerText.trim() : ''`), 60000)
  p.zeige(`am Feld: ${ausgang}`)
  p.pruefe('er kam von einem Ollama-Modell, über den Vorrang', ausgang.startsWith('von ollama/') && ausgang.endsWith('Vorrang'), ausgang)
  const wert = await p.seite("return [...document.querySelectorAll('span')].find((s) => s.innerText.trim().startsWith('von ollama/'))?.parentElement.querySelector('span')?.innerText.trim() ?? ''")
  p.zeige(`vorgeschlagen: ${wert}`)
  p.pruefe('mit einem Namen', wert !== '', wert)
  await p.bild('ai02-vorschlag')

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
