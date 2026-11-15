// **Wie tief ein Kartenfeld liegt** — UX-05.AK-01
// (#1677).
//
// # Was das Kriterium verlangt
//
// *„Normal zeigt die für die gewählte Vorlage wesentlichen Karteneingaben,
// Erweitert Gestaltung und Zusatzdaten, Experte technische Felder und
// Skriptoptionen; die Zuordnung folgt Feldmetadaten und ist dokumentiert."*
//
// # Die Zuordnung, und woher sie kommt
//
// **MSE2 kennt keine Tiefe.** Nachgelesen am 2026-09-21 in
// `src/data/field.hpp`, Zeilen 46–59, Fassung f9d9356 vom 2021-07-06: ein
// `Field` trägt `editable`, `save_value`, `show_statistics`, `identifying`,
// `card_list_visible` und `card_list_allow` — **kein** Merkmal für
// Komplexität. MSE2 zeigt alle Felder in einem Bogen. Die Einteilung ist
// damit **neu** (§15) und keine Portierung; gesucht wurde in `field.hpp`,
// `card.hpp` und `card_editor.cpp`.
//
// **Sie stützt sich trotzdem auf Metadaten und nicht auf eine Liste von
// Feldnamen.** Eine Liste veraltete beim ersten fremden Spielpaket, und sie
// wäre genau die erfundene Fachlogik, die §15 verhindert.
//
// | Tiefe | Kriterium | Quelle |
// | --- | --- | --- |
// | `normal` | das Spielpaket erklärt das Feld, es ist bearbeitbar und wird gespeichert | `card field:` im Spiel |
// | `erweitert` | erklärt, aber `editable: false` oder `save value: false` — ein Skript füllt es | `field.hpp` Z. 50–51 |
// | `experte` | das Spielpaket erklärt es nicht: `stylesheet`, `styling data` und ihresgleichen | `card.hpp` Z. 44–50 |
//
// # Und warum sie ineinanderliegen
//
// **AK-02**: *„Der Moduswechsel ändert keine Kartendaten, verliert keine
// versteckten Werte."* Das wird hier nicht versprochen, sondern durch die
// Ordnung ausgeschlossen: `normal ⊆ erweitert ⊆ experte`. Wer tiefer geht,
// sieht mehr — nie anderes.
import type { FeldbeschreibungDTO } from '../api'

// feldtiefen sind die drei Bedienmodi, von flach nach tief.
//
// **Zur Laufzeit und nicht nur im Typ**: der Wechsler zählt sie durch, und
// eine zweite Aufzählung daneben liefe auseinander.
export const feldtiefen = ['normal', 'erweitert', 'experte'] as const

export type Feldtiefe = (typeof feldtiefen)[number]

// tiefeVon sagt, ab welcher Tiefe ein Feld steht.
//
// **`undefined` heißt: das Spielpaket erklärt es nicht.** Das ist kein
// Fehlerfall — ein Set aus MTG JSON trägt solche Felder, und sie zu
// verstecken hieße, sie stillschweigend zu verlieren.
export function tiefeVon(
  beschreibung: FeldbeschreibungDTO | undefined,
): Feldtiefe {
  if (beschreibung === undefined) {
    return 'experte'
  }

  if (!beschreibung.editierbar || !beschreibung.gespeichert) {
    return 'erweitert'
  }

  return 'normal'
}

// stehtIn sagt, ob ein Feld in dieser Tiefe zu sehen ist.
export function stehtIn(gewaehlt: Feldtiefe, feld: Feldtiefe): boolean {
  return feldtiefen.indexOf(feld) <= feldtiefen.indexOf(gewaehlt)
}

// Befundstelle ist ein Problem, so weit diese Rechnung es braucht.
//
// **Nicht das ganze `Problem`**, damit sie ohne die Problemliste prüfbar
// bleibt — dieselbe Überlegung wie bei `Gelegtes` am Tisch.
export interface Befundstelle {
  feld?: string | undefined
}

// verdeckterBefund nennt die Tiefe, in der ein Befund liegt, den die gewählte
// gerade nicht zeigt — oder nichts.
//
// # Wofür
//
// **AK-02** verlangt, dass ein Fehler in einem verdeckten Feld „mit einer
// direkten Öffnungsaktion" dasteht. Ohne das stünde eine Karte da, die sich
// nicht speichern lässt, und niemand sähe, woran es liegt (§6: ein Fehler
// erscheint am verursachenden Feld).
//
// **Die flachste von mehreren**: der Sprung führt dorthin, wo am wenigsten
// dazukommt — wer tiefer will, geht weiter.
export function verdeckterBefund(
  gewaehlt: Feldtiefe,
  probleme: readonly Befundstelle[],
  nachName: ReadonlyMap<string, FeldbeschreibungDTO>,
): Feldtiefe | undefined {
  let flachste: Feldtiefe | undefined

  for (const problem of probleme) {
    // **Ein Befund ohne Feld gehört nirgendwohin.** Ein Rasterbefund nennt
    // nicht immer eines, und ein Sprung ins Blaue wäre schlechter als keiner.
    if (problem.feld === undefined) {
      continue
    }

    const tiefe = tiefeVon(nachName.get(problem.feld))

    if (stehtIn(gewaehlt, tiefe)) {
      continue
    }

    if (
      flachste === undefined ||
      feldtiefen.indexOf(tiefe) < feldtiefen.indexOf(flachste)
    ) {
      flachste = tiefe
    }
  }

  return flachste
}

// wirksameTiefe sagt, welche Tiefe wirklich gilt.
//
// # Der Fall, für den es sie gibt
//
// **Ein Set, dessen Spielpaket nicht dasteht.** Dann erklärt niemand ein
// Feld, alle fallen auf `experte` — und `normal` zeigte ein **leeres
// Formular**. Eine Karte, die sich nicht ausfüllen lässt, ist schlechter als
// eine mit zu vielen Feldern.
//
// **Das ist kein Sonderfall, sondern die ehrliche Antwort**: wo die
// Einteilung nichts zu sagen hat, sagt sie nichts — und es gilt, was vor
// UX-05 galt.
export function wirksameTiefe(
  gewaehlt: Feldtiefe,
  felder: readonly { feld: string }[],
  nachName: ReadonlyMap<string, FeldbeschreibungDTO>,
): Feldtiefe {
  const hatNormale = felder.some(
    (feld) => tiefeVon(nachName.get(feld.feld)) === 'normal',
  )

  return hatNormale ? gewaehlt : 'experte'
}
