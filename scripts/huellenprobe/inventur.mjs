// **Die Inventur am nativen Fenster** — UX-01.AK-02
// (#1673):
// die native Anwendung, am gepinnten Bau, in einem nachweislich isolierten
// Bestand bedient — mit Abzügen und reproduzierbaren Schritten.
//
// **Je registrierte Seite** (`frontend/src/navigation.ts`, dieselbe Liste
// wie `docs/inventur.json`): über die Befehlspalette geöffnet, wie ein
// Mensch es täte — ⌘K, den Namen tippen, den Treffer klicken. Dann wird gelesen, was dasteht:
// die Überschrift, eine Meldung als `role="alert"`, ob die Seite waagerecht
// rollt. Der Abzug des Fensters geht verkleinert nach
// `docs/belege/huellenprobe/inventur/<seite>.jpg`, und `docs/inventur.json`
// nennt ihn als Beleg der nativen Prüfung.
//
// **Isoliert** heißt dreierlei, und alles drei steht im Kopf des Belegs:
// ein frischer Bestand samt Schlüsselbund je Lauf (`p.ort()`), das Bündel
// unter eigener Kennung (`scripts/bauen.sh abnahme`), und kein Weg ins Netz,
// den die Probe öffnet. **Befüllt** ist der Bestand mit dem Werkstattset aus
// der Saat — die Seiten eines Sets brauchen eines.
//
// **Was sie nicht belegt**: die Zustände Laden, Offline und Nur-Lesen je
// Seite — sie stehen in `docs/inventur.json` mit ihrem eigenen Stand —, und
// ob ein Mensch die Seite versteht.
import { execFileSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { join, resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ablage = join(wurzel, 'docs/belege/huellenprobe/inventur')

// seiten ist Beschriftung in der Palette und Schlüssel in `navigation.ts` —
// in der Reihenfolge der Palette.
export const seiten = [
  ['Start', 'start'], ['Sets', 'sets'], ['Pakete', 'pakete'],
  ['Vorlagen', 'vorlagen'], ['Sammlung', 'sammlung'], ['Decks', 'decks'],
  ['Turniere', 'turniere'], ['Spielen', 'spielen'], ['Handeln', 'handeln'],
  ['Hilfe', 'hilfe'], ['Einstellungen', 'einstellungen'],
  ['Nutzer und Profile', 'nutzer'], ['Karten', 'karten'], ['Bilder', 'bilder'],
  ['Suche', 'suche'], ['Ersetzen', 'ersetzen'], ['Stil', 'stil'],
  ['Setangaben', 'setangaben'], ['Symbol', 'symbol'],
  ['Schlüsselwörter', 'schluesselwoerter'], ['Statistik', 'statistik'],
  ['Zufallspack', 'pack'], ['Konsole', 'konsole'], ['Ausgeben', 'ausgeben'],
  ['Druckbogen', 'druckbogen'],
]

// lage liest, was auf der Seite steht — Überschrift, Meldungen, Rollen.
const lage = "return { titel: (document.querySelector('main h1, main h2')?.innerText ?? '').trim(), meldungen: [...document.querySelectorAll('main [role=alert]')].map((e) => e.innerText.trim()).filter(Boolean), rollt: document.scrollingElement.scrollWidth > innerWidth + 1, knoepfe: [...document.querySelectorAll('main button, main a[href], main input, main select, main textarea')].filter((e) => e.checkVisibility()).length }"

export default async function inventur(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })
  mkdirSync(ablage, { recursive: true })

  p.notiz('Das Set öffnen — die Seiten eines Sets brauchen eines')
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('das Set ist offen', () => p.seite("return (document.querySelector('nav ol')?.innerText ?? '').includes('Werkstatt')"), 20000)

  for (const [beschriftung, schluessel] of seiten) {
    p.notiz(`${beschriftung} (\`${schluessel}\`)`)
    await p.taste('k', 'befehl')
    await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
    // **Getippt, dann geklickt** — wie ein Mensch, der die Liste nicht
    // durchrollt; ein Knopf unten in der Liste läge sonst außerhalb ihres
    // sichtbaren Teils, und der Klick träfe daneben.
    await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
    await p.tippeInsFeld(beschriftung)
    // **Die Zeile der Palette, nicht ein Reiter gleichen Namens** — eine
    // offene Schublade ist auch ein Dialog, und ihre Reiter heißen wie Seiten.
    await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(beschriftung)})`, `»${beschriftung}« in der Palette`)
    await p.warte('die Palette ist zu', () => p.seite("const f = document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]'); return !f || !f.checkVisibility()"))
    await new Promise((los) => setTimeout(los, 1200))
    const stand = await p.seite(lage)
    const roh = await p.bild(`inventur-${schluessel}`)
    execFileSync('sips', ['-s', 'format', 'jpeg', '-s', 'formatOptions', '55', '-Z', '1100', roh, '--out', join(ablage, `${schluessel}.jpg`)], { stdio: 'ignore' })
    p.pruefe(`${beschriftung}: öffnet sich über die Palette`, stand.titel !== '', `Überschrift »${stand.titel}«, ${String(stand.knoepfe)} Bedienelemente`)
    p.pruefe(`${beschriftung}: ohne Fehlermeldung`, stand.meldungen.length === 0, stand.meldungen.join(' · ').slice(0, 200))
    p.pruefe(`${beschriftung}: rollt nicht waagerecht`, !stand.rollt)
  }

  p.pruefe('⌘Q beendet', await p.beende())
}
