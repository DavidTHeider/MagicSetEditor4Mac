import { createContext, use, useMemo, type ReactNode } from 'react'
import { MantineProvider } from '@mantine/core'

import { mantinethema, mantineumgebung } from '../mantine'

import { kataloge, type Katalog, type Sprache } from './kataloge'

// Deutsch ist die Ausgangssprache und damit auch der Rückfall (AGENTS.md §9).
const SprachKontext = createContext<Sprache>('de')

// **Eine eigene Übersetzung liegt als ganzer Katalog darüber** — LOC-02
// (#1688). Ohne sie gilt der Katalog der Sprache.
const EigenerKatalog = createContext<Katalog | undefined>(undefined)

// **Die Darstellung und die Zone** — LOC-03.AK-03
// (#1689).
// Ohne Angabe gilt die Sprache und die Zone des Rechners — so verhielt sich
// das Fenster bisher, und so bleibt es für jeden, der nichts einstellt.
const DarstellungsKontext = createContext<{
  darstellung: string | undefined
  zeitzone: string | undefined
}>({ darstellung: undefined, zeitzone: undefined })

// Sprachraum ist der **Raum, in dem diese Anwendung läuft**.
//
// # Warum hier auch Mantine steht
//
// **Weil eine Vorrichtung, die großzügiger ist als der Betrieb, nichts
// prüft** (§3). `MantineProvider` legt die Merkmale an, aus denen jedes
// Mantine-Bauteil seine Maße nimmt; ohne ihn wirft schon das erste. Stünde er
// nur in `App.tsx`, zeichneten die 344 Läufe, die diesen Raum aufspannen, in
// einer **anderen** Welt als das Fenster — und jeder Fund über ein Overlay
// wäre eine Aussage über eine Welt, die es nicht gibt.
//
// **Der Name bleibt, und das ist eine Abwägung.** „Sprachraum" sagt weniger,
// als er jetzt tut; ihn umzubenennen kostete 344 Dateien und brächte nichts,
// was dieser Absatz nicht auch sagt. Was er ist: der eine Ort, an dem die
// Anwendung ihre Umgebung bekommt.
//
// # Und `env` ist keine Kosmetik
//
// Ohne `test` bleibt unter jsdom **jedes** Overlay auf `display: none` stehen
// und ist für jede rollenbasierte Abfrage unsichtbar — nachgelesen in
// Mantines `Transition.mjs`, Zeile 22, und gemessen am 2026-09-20.
export function Sprachraum({
  sprache,
  katalog,
  darstellung,
  zeitzone,
  children,
}: {
  sprache: Sprache
  katalog?: Katalog | undefined
  darstellung?: string | undefined
  zeitzone?: string | undefined
  children: ReactNode
}) {
  const ort = useMemo(
    () => ({ darstellung, zeitzone }),
    [darstellung, zeitzone],
  )

  return (
    <MantineProvider
      theme={mantinethema}
      env={mantineumgebung(import.meta.env.MODE)}
    >
      <SprachKontext value={sprache}>
        <EigenerKatalog value={katalog}>
          <DarstellungsKontext value={ort}>{children}</DarstellungsKontext>
        </EigenerKatalog>
      </SprachKontext>
    </MantineProvider>
  )
}

// useTexte gibt den Katalog der eingestellten Sprache.
//
// Als Haken und nicht als Import: eine Komponente, die den Katalog direkt
// importiert, ist beim Sprachwechsel taub — und das fällt erst auf, wenn jemand
// wechselt.
export function useTexte(): Katalog {
  const sprache = use(SprachKontext)
  const eigener = use(EigenerKatalog)

  return useMemo(() => eigener ?? kataloge[sprache], [eigener, sprache])
}

// useSprache gibt die Sprache des Fensters — für die Stellen, an denen `Intl`
// **Wörter** liefert und nicht Zahlen: die Namen der Währungen in den
// Vorschlägen des Währungsfeldes (UX-06,
// #1678).
//
// **Nicht die Darstellung**, und das ist der Unterschied zu
// [useDarstellung]: wie eine Zahl aussieht, entscheidet die Gegend, wie ein
// Wort heißt, die Sprache. Wer Englisch wählt und in Deutschland rechnet,
// liest „US Dollar" und nicht „US-Dollar".
export function useSprache(): Sprache {
  return use(SprachKontext)
}

// useDarstellung gibt die Locale, in der Zahlen, Beträge und Zeitpunkte
// erscheinen — LOC-03.AK-03.
//
// **Für `Intl` und nicht für den Katalog.** Die Wörter kommen aus der
// Sprache; wie eine Zahl aussieht, entscheidet die Gegend.
//
// **Nicht das Sprachkürzel aus `useSprache`**: aus `de-CH` würde `de`.
export function useDarstellung(): string {
  const sprache = use(SprachKontext)

  return use(DarstellungsKontext).darstellung ?? sprache
}

// useZeitzone gibt die gewählte Zone — oder keine, dann gilt die des
// Rechners.
export function useZeitzone(): string | undefined {
  return use(DarstellungsKontext).zeitzone
}
