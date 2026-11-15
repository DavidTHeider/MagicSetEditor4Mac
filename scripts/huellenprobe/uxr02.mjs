// **Der Rahmen am nativen Fenster** — UXR-02
// (#1910).
//
// `uirevue` misst nur `main`; der Rahmen darum ist dort ungemessen. Hier, je
// Fenstergröße (1280 × 800 und 1000 × 700, wie `uirevue`):
//
// - **die Seitenleiste**: rollt ihre Navigation, steht der letzte Eintrag im
//   Bild, wie hoch ist ihr Inhalt, wie viele Bedienelemente trägt sie;
// - **über den Spalten jeder Arbeitsplatzseite** (Karten, Sammlung, Decks,
//   Pakete, Vorlagen): wie hoch ist die Zeile mit den Regionsknöpfen, steht
//   der Satz „Zu schmal …" darin, wo beginnen die Spalten und wie hoch sind
//   sie, wie hoch die Laschenleiste und der Weg über dem Kopf;
// - **die Palette ohne Eingabe**: wie viele Zeilen, welche Abschnitte tragen
//   eine Überschrift;
// - **die Menüleiste**: in welcher Ordnung die Menüs stehen.
//
// **Gegen einen Stand vor UXR-02 gefahren**, gibt es dieselbe Tabelle als
// „vorher" — die Prüfungen am Ende werden dort rot, und genau das ist der
// Befund. **Was es nicht misst**: ob jemand die Knöpfe über den Spalten
// findet, und wie sich der Rahmen bei „Leiste oben" verhält — das sieht ein
// Mensch (UXR-02.AK-04).
import { oeffnen, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1000, hoehe: 700 },
]

const arbeitsplaetze = ['Karten', 'Sammlung', 'Decks', 'Pakete', 'Vorlagen']

const leiste = `
  const l = document.getElementById('seitenleiste')
  if (!l || !l.checkVisibility()) return null
  const nav = l.querySelector('nav')
  const eintraege = [...nav.querySelectorAll('button')].filter((b) => b.checkVisibility())
  const letzter = eintraege.at(-1).getBoundingClientRect()
  const rand = nav.getBoundingClientRect()
  // **Die Navigation dehnt sich** (\`flex-1\`): ihre eigene Höhe ist der Rest.
  // Was sie bräuchte, sagt der Abstand vom ersten zum letzten Kind samt Rand.
  const kinder = [...nav.children]
  const natur = kinder.at(-1).getBoundingClientRect().bottom - kinder[0].getBoundingClientRect().top + parseFloat(getComputedStyle(nav).paddingTop) + parseFloat(getComputedStyle(nav).paddingBottom)
  return {
    fenster: innerWidth + ' × ' + innerHeight,
    hoehe: l.clientHeight,
    inhalt: Math.round([...l.children].reduce((summe, k) => summe + (k === nav ? natur : k.getBoundingClientRect().height), 0)),
    rollt: nav.scrollHeight > nav.clientHeight + 1,
    letzterImBild: letzter.bottom <= rand.bottom + 1,
    eintragshoehe: Math.round(eintraege[0].getBoundingClientRect().height),
    bedien: [...l.querySelectorAll('button')].filter((b) => b.checkVisibility()).length,
  }
`

// **Gefunden über die Aufschrift**, damit derselbe Ausdruck auch einen Stand
// vor UXR-02 misst: dort trugen die Knöpfe weder Kürzel noch Titel.
const ueberDenSpalten = `
  const knopf = [...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && /^(Navigator|Inspector) (ein|aus)blenden/.test(b.innerText.trim()))
  if (!knopf) return null
  // **Seit P03 (#2174) stehen die Schalter der Kartenseite in der Zeile mit
  // „Karte anlegen"**, als Gruppe außerhalb des Arbeitsplatzes. Gemessen wird
  // dann die Gruppe — eine Zeile kleiner Knöpfe — und die Spalten des
  // Arbeitsplatzes, der dort mit ihnen beginnt.
  const eigene = knopf.closest('[role=region]')
  const region = eigene ?? [...document.querySelectorAll('main [role=region]')].find((r) => r.querySelector('aside'))
  if (!region) return null
  const zeile = (eigene ? region.firstElementChild : knopf.closest('[role=group]')).getBoundingClientRect()
  const spalten = (eigene ? region.children[1] : region.firstElementChild).getBoundingClientRect()
  const hoehe = (wahl) => Math.round(document.querySelector(wahl)?.getBoundingClientRect().height ?? 0)
  return {
    zeile: Math.round(zeile.height),
    satz: [...region.querySelectorAll('span, p')].some((e) => e.checkVisibility() && /^Zu schmal für alle drei Spalten/.test(e.innerText.trim())),
    spaltenOben: Math.round(spalten.top),
    spaltenHoehe: Math.round(spalten.height),
    laschen: hoehe('main [role=tablist]'),
    weg: hoehe('nav[aria-label="Wo du bist"]'),
    waagerecht: document.scrollingElement.scrollWidth > innerWidth + 1,
  }
`

const palette = `
  const d = [...document.querySelectorAll('[role=dialog]')].find((e) => e.checkVisibility())
  if (!d) return null
  return {
    zeilen: [...d.querySelectorAll('[data-palettenzeile]')].filter((b) => b.checkVisibility()).length,
    // **\`textContent\` und nicht \`innerText\`**: die Überschriften stehen in
    // Großbuchstaben (\`uppercase\`), und \`innerText\` gibt, was gezeichnet ist.
    ueberschriften: [...d.querySelectorAll('p')].filter((e) => e.checkVisibility()).map((e) => e.textContent.trim()),
  }
`

export default async function uxr02(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  const zeilen = []
  for (const groesse of groessen) {
    await oeffnen(p, ort, groesse)

    p.notiz(`${groesse.name} — die Seitenleiste`)
    const l = await p.seite(leiste)
    p.zeige(`Seitenleiste: ${JSON.stringify(l)}`)
    p.pruefe(`die Seitenleiste rollt nicht (${groesse.name})`, !!l && !l.rollt && l.letzterImBild, JSON.stringify(l))

    p.notiz(`${groesse.name} — die Palette ohne Eingabe`)
    await p.taste('k', 'befehl')
    const pa = await p.warte('die Palette steht', () => p.seite(palette))
    p.zeige(`Palette: ${JSON.stringify(pa)}`)
    p.pruefe(`die Seiten tragen eine Überschrift (${groesse.name})`, pa.ueberschriften.includes('Seiten'), pa.ueberschriften.join(' · '))
    await p.taste('Escape')

    for (const name of arbeitsplaetze) {
      try {
        await zurSeite(p, name)
        const wert = await p.warte(`die Spalten von ${name} stehen`, () => p.seite(ueberDenSpalten))
        zeilen.push({ groesse: groesse.name, name, ...wert })
      } catch (grund) {
        p.zeige(`${name} (${groesse.name}): nicht gemessen — ${String(grund.message ?? grund)}`)
        await p.taste('Escape')
      }
    }
    await p.bild(`rahmen-${groesse.name}`)
    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  p.notiz('Über den Spalten')
  p.zeige('| Größe | Seite | Zeile über den Spalten | Satz darin | Spalten ab | Spalten hoch | Laschenleiste | Weg | waagerecht |')
  p.zeige('| --- | --- | ---: | --- | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) {
    p.zeige(`| ${z.groesse} | ${z.name} | ${String(z.zeile)} | ${z.satz ? 'ja' : 'nein'} | ${String(z.spaltenOben)} | ${String(z.spaltenHoehe)} | ${String(z.laschen)} | ${String(z.weg)} | ${z.waagerecht ? 'ja' : 'nein'} |`)
  }
  p.pruefe('jede Arbeitsplatzseite gemessen', zeilen.length === groessen.length * arbeitsplaetze.length, `${String(zeilen.length)} von ${String(groessen.length * arbeitsplaetze.length)}`)
  const mitSatz = zeilen.filter((z) => z.satz).map((z) => `${z.name} (${z.groesse})`)
  p.pruefe('kein Satz „Zu schmal …" in der Zeile', mitSatz.length === 0, mitSatz.join(', '))
  // **Niedriger als ein großer Knopf**: die Zeile in voller Größe maß 40,
  // umgebrochen 88. Ein kleiner Knopf (`py-1` um `text-sm`, Rand) ist 30
  // gerechnet — **am Fenster maß die Zeile 33** (2026-10-02); hier stand
  // zuerst „höchstens 32", geschätzt statt gemessen. Was die Prüfung halten
  // soll, ist: eine Zeile kleiner Knöpfe, nicht umgebrochen.
  const hoch = zeilen.filter((z) => z.zeile >= 40).map((z) => `${z.name} (${z.groesse}): ${String(z.zeile)}`)
  p.pruefe('die Zeile über den Spalten ist niedriger als ein großer Knopf (40)', hoch.length === 0, hoch.join(', '))
  const rollt = zeilen.filter((z) => z.waagerecht).map((z) => `${z.name} (${z.groesse})`)
  p.pruefe('R5: nichts rollt waagerecht', rollt.length === 0, rollt.join(', '))

  p.notiz('Die Menüleiste')
  await p.starte({ ort })
  const oben = (await p.menue()).filter((m) => !m.trenner && !m.verborgen).map((m) => m.titel)
  p.zeige(`oben: ${oben.join(' · ')}`)
  p.pruefe('die eigenen Menüs zwischen Darstellung und Fenster', oben.join(' · ') === 'MagicSetEditor4Mac · Ablage · Bearbeiten · Darstellung · Set · Fenster · Hilfe', oben.join(' · '))
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
