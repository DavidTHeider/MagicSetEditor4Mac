// **Eine eigene Karte: Wortlisten, Marken, Bilder, Kartenskript, Verweise** —
// eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Die Gruppe legt **ihre eigene Karte** an und löscht sie am Ende wieder. Sie
// verlässt sich damit weder auf `k.neueKarte` noch auf eine der fünf
// Testkarten des Saatsets: wer anderswo löscht, nähme ihr sonst den Boden.
// Braucht `k.set` aus der ersten Gruppe und den Stil `probe-kunst` (das Feld
// `kunst` ist ein Bildfeld; `a-stil.mjs` stellt ihn nach seinem Lauf wieder ein).
import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { png } from '../../pngbau.mjs'

// wort ist ein Feldwert mit Wortliste, so wie ihn das Feldskript des Spiels
// schreibt (`internal/app/saatwortlisten_intern_test.go`).
const wort = '<word-list-probe-art>Große Kreatur</word-list-probe-art>'

// Das Kunstfenster des Saatstils ist 315 × 230 Punkte groß
// (`saat/probe-kunst.mse-style`); ein Bild im selben Verhältnis braucht keinen
// Zuschnitt, der etwas wegnimmt.
const bildbreite = 63
const bildhoehe = 46
const bild = (k) => join(k.ort.ort, 'vertrag', 'kunst.png')
const auswahl = { x: 0, y: 0, breite: bildbreite, hoehe: bildhoehe }

export const aKarten = {
  name: 'Eine eigene Karte (A)',
  rufe: [
    { name: 'KarteAnlegen', argumente: (k) => [k.set.id], nimm: 'a.karteID', nach: (w, k) => { k.aKarte = w } },

    // --- Wortliste: ein anderer Typ, dann eine Wahl im Untermenü
    { name: 'FeldSetzen', argumente: (k) => [k.set.id, k.aKarte, 'text', wort, []] },
    { name: 'WortWaehlen', argumente: (k) => [k.set.id, k.aKarte, 'text', 0, [2], []] },
    { name: 'WortWaehlen', argumente: (k) => [k.set.id, k.aKarte, 'text', 0, [3, 0], []] },

    // --- Marken: zwei setzen, eine abnehmen, die andere überall wegnehmen
    { name: 'MarkeSetzen', argumente: (k) => [k.set.id, k.aKarte, 'vertrag-karte'] },
    { name: 'MarkeSetzen', argumente: (k) => [k.set.id, k.aKarte, 'vertrag-ueberall'] },
    { name: 'MarkeEntfernen', argumente: (k) => [k.set.id, k.aKarte, 'vertrag-karte'] },
    { name: 'MarkeUeberallEntfernen', argumente: () => ['vertrag-ueberall'] },

    // --- Bilder: eine Datei von der Platte, dann die Karte als Bild
    {
      name: 'BildEinsetzen', argumente: (k) => [k.set.id, k.aKarte, 'kunst', bild(k), auswahl],
      vorher: (_p, k) => {
        mkdirSync(join(k.ort.ort, 'vertrag'), { recursive: true })
        writeFileSync(bild(k), png({ breite: bildbreite, hoehe: bildhoehe }))
      },
    },
    {
      // Mit `ueberschreiben`: ein zweiter Lauf im selben Ordner schriebe sonst
      // `vertragskarte.1.png` und gelänge trotzdem — der Beleg wäre derselbe.
      name: 'BildSchreiben',
      argumente: (k) => [k.aKarte, join(k.ort.ort, 'vertrag', 'bilder'), 'vertragskarte.png', 'png', 0, 'ueberschreiben', ''],
    },

    // --- Kartenskript: legt zwei Karten an, die diese Gruppe wieder löscht
    {
      name: 'Kartenskripte', argumente: (k) => [k.set.id],
      nimm: 'a.skripte.map((s) => s.name)',
      nach: (w, k) => { k.aSkript = w.find((n) => n.includes('Zwei')) ?? w[0] },
    },
    {
      name: 'KartenskriptAusfuehren', argumente: (k) => [k.set.id, k.aSkript],
      nimm: 'a.karten', nach: (w, k) => { k.aSkriptkarten = w },
    },
    { name: 'KarteLoeschen', argumente: (k) => [k.set.id, k.aSkriptkarten[0], true] },
    { name: 'KarteLoeschen', argumente: (k) => [k.set.id, k.aSkriptkarten[1], true] },

    // --- Was einen fremden Dienst braucht
    {
      // **Ein leerer Druck bleibt in der Anwendung**: `Kartenverweise.Setzen`
      // weist ihn vor jedem Ruf ab (`internal/service/kartenverweis.go`, „ohne
      // Kennung ist kein Druck gemeint"). Jede echte Kennung ginge zu Scryfall
      // hinaus, und der Verweis entsteht erst aus dessen Antwort — in einer
      // Anwendung ohne Netz kann er nicht entstehen.
      name: 'KarteVerweisSetzen', argumente: (k) => [k.aKarte, ''], erwartet: 'fachfehler',
      grund: 'ein Verweis entsteht nur aus der Antwort von Scryfall (kartenverweis.go, Setzen)',
    },
    {
      // Wo kein Verweis war, ist das Lösen ein Nein — und ohne Scryfall kann
      // keiner entstanden sein (`Kartenverweise.Loesen`).
      name: 'KarteVerweisLoesen', argumente: (k) => [k.aKarte], erwartet: 'fachfehler',
      grund: 'die Karte hat keinen Verweis, und ohne Scryfall entsteht keiner (kartenverweis.go, Loesen)',
    },
    {
      // Ein wartender Bildvorschlag entsteht nur aus `KIBildVorschlagen`, das
      // einen Bildanbieter fragt (`internal/service/bildvorrat.go`: der Vorrat
      // wird allein von dort gefüllt). Karte und Feld sind gültig, damit der
      // Ruf **bis zum Vorrat** kommt und dort scheitert.
      name: 'KIBildEinsetzen', argumente: (k) => [k.set.id, k.aKarte, 'kunst', 'vertrag-ohne-vorschlag', auswahl],
      erwartet: 'fachfehler', grund: 'es wartet kein Bildvorschlag; er entstünde nur aus einem Bildmodell (bildvorrat.go)',
    },
    {
      name: 'KIBildVerwerfen', argumente: () => ['vertrag-ohne-vorschlag'],
      erwartet: 'fachfehler', grund: 'es wartet kein Bildvorschlag (Bildvorrat.Verwerfen: keinWartendesBild)',
    },

    // Ans Ende, was wegnimmt.
    { name: 'KarteLoeschen', argumente: (k) => [k.set.id, k.aKarte, true] },
  ],
}
