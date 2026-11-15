// **Was eine Hüllenprobe hinterlässt: Bild und Beleg** — aus
// `huellenprobe.mjs` herausgelöst, als der Vermerk fehlender Bilder
// (DOM-04, #1699) die Datei über die Zeilengrenze aus §2 brachte.
import { execFileSync } from 'node:child_process'
import { mkdirSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

// fensterbild nimmt ein Fenster auf und sagt, ob es gelang.
//
// **Schläft der Bildschirm, gibt es keinen Abzug** — `screencapture` meldet
// dann »could not create image from window«. Das ist kein Befund an der
// Anwendung und kein Grund, die Prüfungen danach nicht zu fahren; der
// Aufrufer vermerkt es als **Lücke**, damit niemand ein fehlendes Bild für
// ein gesehenes hält. Jeder andere Fehler bleibt einer.
export function fensterbild(nummer, pfad) {
  try {
    execFileSync('screencapture', ['-x', '-o', '-l', String(nummer), pfad], { stdio: ['ignore', 'ignore', 'pipe'] })
  } catch (fehler) {
    if (!String(fehler.stderr ?? '').includes('could not create image')) throw fehler

    return false
  }

  return true
}

// schreibeBeleg legt den Beleg einer Probe unter `docs/belege/huellenprobe`
// ab — mit Commit, Maschine und System, und **mit der Zahl der Lücken im
// Kopf**, nicht nur in ihren Zeilen.
// offeneZeilen nennt, was am Baum ungesichert ist; offeneAenderungen zählt es.
//
// **Ein Baum mit ungesicherten Änderungen wird genannt** — sonst nennt der
// Beleg einen Commit, aus dem das Bündel nicht gebaut ist. Gefunden bei
// FMT-03: zwei Läufe von `fmt03b` trugen denselben Commit und prüften
// verschiedene Stände. **Was der Lauf selbst schreibt, zählt nicht**: die
// Belege und die README-Bilder aus `readmebilder` (B06, #2225) — ohne die
// zweite Ausnahme trug am 2026-10-09 jeder Beleg nach `readmebilder` „mit 64
// ungesicherten Änderung(en)".
export function offeneZeilen(wurzel) {
  return execFileSync('git', ['-C', wurzel, 'status', '--porcelain', '--', '.', ':!docs/belege', ':!docs/bilder'])
    .toString().split('\n').filter((zeile) => zeile.trim() !== '')
}

export const offeneAenderungen = (wurzel) => offeneZeilen(wurzel).length

export function schreibeBeleg(probe, wurzel) {
  const commit = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
  const offen = offeneAenderungen(wurzel)
  const maschine = execFileSync('sysctl', ['-n', 'hw.model']).toString().trim()
  const system = execFileSync('sw_vers', ['-productVersion']).toString().trim()
  const kopf = [
    `# Hüllenprobe: ${probe.name}`,
    '',
    `**${probe.fehler === 0 ? 'bestanden' : `${String(probe.fehler)} Befund(e)`}** — Commit \`${commit}\`${offen > 0 ? ` **mit ${String(offen)} ungesicherten Änderung(en)**` : ''},`,
    `${new Date().toISOString().slice(0, 16).replace('T', ' ')} UTC, ${maschine}, macOS ${system},`,
    'das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.',
    // **Welcher Lauf, und aus welchem Stand das Bündel** — T06 (#2099); das
    // Manifest daneben nennt Programme, Oberfläche und Umwelt.
    ...(probe.lauf ? [`Lauf \`${probe.lauf.kennung}\`, Bauabdruck \`${probe.lauf.bauabdruck.slice(0, 12)}\`${probe.lauf.veraltet ? ' — **das Bündel stammt aus einem anderen Stand**' : ''}.`] : []),
    ...(probe.ohneBild > 0 ? [`**${String(probe.ohneBild)} Bild(er) fehlen** — der Bildschirm schlief; geprüft ist am Fenster, gesehen nicht.`] : []),
    '',
  ]
  const ziel = join(wurzel, 'docs/belege/huellenprobe', `${probe.name}.md`)
  mkdirSync(dirname(ziel), { recursive: true })
  writeFileSync(ziel, `${[...kopf, ...probe.zeilen].join('\n')}\n`)

  return ziel
}
