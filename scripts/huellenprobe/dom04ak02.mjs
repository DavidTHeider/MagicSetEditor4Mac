// **Eine Berichtigung am nativen Fenster** — DOM-04.AK-02
// (#1699):
//
// - vier Namen, zwei gespielte Runden — die Kommandozeile legt sie an,
// - im Fenster das Ergebnis der ersten Paarung umgedreht — Runde 1
//   aufgeklappt, *Ergebnis* an der Paarung: statt einzutragen steht die
//   Frage da, nennt, wer Platz wechselt, und dass Runde 2 bleibt,
// - *Nicht eintragen* lässt die Rangliste, wie sie war,
// - *Trotzdem eintragen* schreibt — und die Kommandozeile liest die neue
//   Rangliste.

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

const frage = "[...document.querySelectorAll('[role=group]')].find((g) => g.checkVisibility() && g.innerText.startsWith('Diese Berichtigung verschiebt die Rangliste.'))"

const ersterPlatz = (p, ort, turnierID) => p.mse(ort, 'turnier', 'rangliste', turnierID).split('\n')[0]

async function klicke(p, name) {
  await p.seite(`${knopf(name)}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf(name), name)
}

export default async function dom04ak02(p) {
  const ort = p.ort()
  const turnierID = p.mse(ort, 'turnier', 'neu', 'Freitag', 'magic').trim()
  for (const name of ['Anna', 'Bert', 'Cleo', 'Dora']) p.mse(ort, 'turnier', 'anmelden', turnierID, name)

  let erste = ''
  for (let runde = 1; runde <= 2; runde++) {
    const gelegt = JSON.parse(p.mse(ort, 'turnier', 'runde', turnierID, '--json')).runde
    if (erste === '') erste = gelegt.paarungen[0]
    for (const paarung of gelegt.paarungen) p.mse(ort, 'turnier', 'ergebnis', turnierID, paarung.id, '2', '0')
  }
  const a = erste.a
  const namen = JSON.parse(p.mse(ort, 'turnier', turnierID, '--json')).turnier.teilnehmer
  const wer = (kennung) => namen.find((t) => t.id === kennung).name
  const paarungstext = `${wer(erste.a)} gegen ${wer(erste.b)}`
  const vorher = ersterPlatz(p, ort, turnierID)
  p.zeige(`vorher: ${vorher} · erste Paarung: ${paarungstext} (${a})`)

  await p.starte({ ort })
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Turniere')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Turniere')", '»Turniere« in der Palette')
  await p.warte('das Turnier steht in der Liste', () => p.seite(`return !!${knopf('Öffnen: Freitag')}`), 20000)
  await klicke(p, 'Öffnen: Freitag')
  await p.warte('die Runden stehen', () => p.seite(`return !!${knopf('Runde 1')}`), 20000)

  p.notiz('AK-02 — das Ergebnis der ersten Runde umdrehen')
  // **Die erste Runde ist eingeklappt** — seit UXR-12 (#1920) steht nur die
  // laufende offen, und eine Paarung zeigt ihr Eintragen erst nach
  // *Ergebnis*.
  // **Erst wenn die Seite steht** — beim Öffnen springt der Fokus auf den
  // Namen des Turniers (UXR-12), und ein Klick in den Sprung hinein traf
  // daneben. Gewartet wird auf die Wirkung, nicht auf den Klick.
  await new Promise((los) => setTimeout(los, 800))
  await klicke(p, 'Runde 1')
  await p.warte('Runde 1 ist aufgeklappt', () => p.seite(`return !!${knopf(`Ergebnis: ${paarungstext}`)}`), 10000)
  await klicke(p, `Ergebnis: ${paarungstext}`)
  await p.warte('das Eintragen steht', () => p.seite(`return !!${knopf(`Eintragen: ${paarungstext}`)}`), 10000)
  const felder = (welches) => `[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === ${JSON.stringify(`${welches} — ${paarungstext}`)})`
  for (const [welches, wert] of [['Spiele für den Ersten', '0'], ['Spiele für den Zweiten', '2']]) {
    // **Markiert und überschrieben** — `tippeInsFeld` erwartet ein Anhängen.
    await p.seite(`${felder(welches)}.scrollIntoView({ block: 'center' }); ${felder(welches)}.focus(); ${felder(welches)}.select(); return true`)
    await p.tippe(wert)
    await p.warte(`im Feld steht »${wert}«`, () => p.seite(`return ${felder(welches)}.value === ${JSON.stringify(wert)}`), 10000)
  }
  await klicke(p, `Eintragen: ${paarungstext}`)
  const text = await p.warte('die Frage steht', async () => (await p.seite(`return ${frage}?.innerText ?? ''`)) || false, 15000)
  p.zeige(`Frage: ${text.replace(/\s+/g, ' ').slice(0, 300)}`)
  p.pruefe('die Frage nennt, wer den Platz wechselt', text.includes(wer(erste.b)) && text.includes('Platz'))
  p.pruefe('die Frage sagt, dass Runde 2 bleibt', text.includes('Runde 2 wurde nach dem alten Stand gepaart'))
  p.pruefe('noch ist nichts eingetragen', ersterPlatz(p, ort, turnierID) === vorher, ersterPlatz(p, ort, turnierID))
  await p.bild('frage')

  p.notiz('AK-02 — Nicht eintragen')
  await klicke(p, 'Nicht eintragen')
  await p.warte('die Frage ist fort', async () => !(await p.seite(`return !!${frage}`)), 10000)
  p.pruefe('die Rangliste bleibt', ersterPlatz(p, ort, turnierID) === vorher)

  p.notiz('AK-02 — Trotzdem eintragen')
  await klicke(p, `Eintragen: ${paarungstext}`)
  await p.warte('die Frage steht wieder', () => p.seite(`return !!${frage}`), 15000)
  await klicke(p, 'Trotzdem eintragen')
  await p.warte('die Rangliste ist eine andere', async () => ersterPlatz(p, ort, turnierID) !== vorher, 15000)
  p.pruefe('die Kommandozeile liest die neue Rangliste', true, ersterPlatz(p, ort, turnierID))
  await p.bild('eingetragen')

  p.pruefe('⌘Q beendet', await p.beende())
}
