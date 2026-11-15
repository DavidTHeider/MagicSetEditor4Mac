// **Die Aufklapps der Kartenlasche in Tests öffnen** —
// #1240.
//
// # Warum es diesen Helfer gibt
//
// Seit die Lasche umgruppiert ist, stehen die Handlungen einer Karte hinter
// **einem** Knopf statt nebeneinander über der Karte. Jeder Lauf, der eine
// davon anklickt, muss ihn vorher öffnen — und das an **einer** Stelle, damit
// eine spätere Umbenennung nicht durch acht Dateien wandert.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { kataloge } from '../sprache/kataloge'

// oeffneKartenhandlungen klappt »Diese Karte« auf.
//
// **`findBy` und nicht `getBy`**: der Knopf steht erst da, wenn eine Karte
// aufgeschlagen ist, und das ist eine Antwort der Fassade.
export async function oeffneKartenhandlungen(sprache: 'de' | 'en' = 'de') {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge[sprache].seiten.karten.dieseKarte,
    }),
  )
}

// waehleKartenhandlung wählt einen Eintrag aus »Diese Karte«.
//
// **Seit UXR-06 (#1914) stehen Zwischenablage, Bezug und Teilen dort** — bis
// dahin im Aufklapper »Ablage und Bezug« und als Knopf daneben. Ein Lauf,
// der kopiert, wählt den Eintrag; der Name trägt die Zahl („Kopieren (2)").
export async function waehleKartenhandlung(
  name: string,
  sprache: 'de' | 'en' = 'de',
) {
  await oeffneKartenhandlungen(sprache)
  await userEvent.click(await screen.findByRole('menuitem', { name }))
}

// oeffneBezug schlägt den Kasten »Bezug zu einem Druck« auf.
export async function oeffneBezug(sprache: 'de' | 'en' = 'de') {
  await waehleKartenhandlung(
    kataloge[sprache].seiten.karten.bezugAufklapp,
    sprache,
  )
}

// oeffneAnlegen schlägt »Aus Scryfall anlegen …« auf.
//
// **Seit UXR-05 (#1913) über das Menü „Mehr"** neben „Karte anlegen" — der
// Weg stand bis dahin als eigener Aufklapper offen über dem Arbeitsplatz.
export async function oeffneAnlegen(sprache: 'de' | 'en' = 'de') {
  const karten = kataloge[sprache].seiten.karten

  await userEvent.click(
    await screen.findByRole('button', { name: karten.mehr }),
  )
  await userEvent.click(
    await screen.findByRole('menuitem', { name: karten.ausScryfallAufklapp }),
  )
}

// waehleMehrere schaltet die Kartenliste in den Auswahlmodus.
//
// **Seit UXR-05 (#1913) tragen ihre Zeilen erst dann ein Kästchen** — bei
// einem großen Set waren es rund zwei Dutzend Felder, die niemand braucht,
// der eine einzelne Karte bearbeitet. Ein Lauf, der anhakt, schaltet zuerst
// um.
export async function waehleMehrere(sprache: 'de' | 'en' = 'de') {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge[sprache].seiten.karten.massen.mehrereWaehlen,
    }),
  )
}
