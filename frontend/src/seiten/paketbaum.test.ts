// **Der Paketbaum** — die Portierung von MSE2s `PackageUpdateList::TreeItem`
// (#244).
//
// Die Vorlage steht in `src/gui/package_update_list.cpp`, Zeilen 18–75 und
// 200–212, Fassung f9d9356 vom 2021-07-06. Sie tut drei Dinge:
//
//  1. `add` zerlegt `installer group` an `/` und legt je Stufe einen Knoten an,
//  2. `compare_pos_hint` ordnet Geschwister nach Art, Rang und Aufschrift,
//  3. `toItems` macht daraus wieder eine **flache** Liste mit Stufen — genau
//     die Form, die eine Baumliste zeichnet.
import { describe, expect, it } from 'vitest'

import type { PaketDTO } from '../api'
import { paketbaum, sichtbareZeilen } from './paketbaum'

// paket baut ein Paket mit dem Nötigsten; alles andere ist leer.
function paket(teile: Partial<PaketDTO> & { name: string }): PaketDTO {
  return {
    art: 'mse-style',
    anzeigename: teile.name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 1000000,
    artrang: 30,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
    ...teile,
  }
}

describe('paketbaum', () => {
  // **Der Pfad wird an `/` zerlegt, und jede Stufe wird ein Knoten.**
  //
  //	size_t pos = path.find_first_of(_('/'));
  //	String name = path.substr(0,pos);
  //	String rest = pos == String::npos ? _("") : path.substr(pos+1);
  //
  // `package_update_list.cpp`, Zeilen 35–37.
  //
  // **Und die letzte Stufe trägt das Paket, statt ihm eine eigene zu geben** —
  // MSEs `add` legt es dort ab, sobald der Weg leer ist (Zeilen 39–42).
  //
  // Der erste Anlauf gab jedem Paket ein eigenes Kind und nannte das
  // „dieselbe Anzeige". Das war falsch, und der Zweitblick über Phase G hat es
  // angestrichen: **298 der 329 Gruppen im Pack tragen genau ein Paket** —
  // 90 %. Die Zusatzebene war damit der Normalfall, nicht der Rand.
  it('zerlegt die Gruppe an den Schrägstrichen', () => {
    const zeilen = paketbaum(
      [paket({ name: 'magic-m15', gruppe: 'magic/m15 style/normal cards' })],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => [zeile.stufe, zeile.aufschrift])).toEqual([
      [0, 'magic'],
      [1, 'm15 style'],

      // **Die Aufschrift ist der Paketname, nicht `normal cards`.** Hier
      // weicht die Umsetzung von MSE ab, mit Grund: MSE zeichnet den letzten
      // Wegabschnitt, und bei mehreren Paketen einer Gruppe stehen dann
      // mehrere Zeilen mit **demselben** Wort. Der Paketname beantwortet die
      // Frage, die jemand vor der Liste hat.
      [2, 'magic-m15'],
    ])
  })

  // **Eine Gruppe mit zwei Paketen behält ihre Zeile.**
  //
  // MSE legt das erste Paket auf den Knoten und schiebt für jedes weitere ein
  // Geschwister daneben (Zeilen 43–49) — beide dann mit derselben Aufschrift.
  // Hier bleibt der Gruppenknoten stehen und trägt die Pakete als Kinder: eine
  // Zeile mehr als bei MSE, dafür je Paket ein eigener Name.
  it('behält die Gruppenzeile, wenn mehrere Pakete darin liegen', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'magic-m15', gruppe: 'magic/m15 style' }),
        paket({ name: 'magic-m15-planeswalker', gruppe: 'magic/m15 style' }),
      ],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => [zeile.stufe, zeile.aufschrift])).toEqual([
      [0, 'magic'],
      [1, 'm15 style'],
      [2, 'magic-m15'],
      [2, 'magic-m15-planeswalker'],
    ])
  })

  // **Zwei Pakete derselben Gruppe teilen sich ihre Knoten.**
  //
  //	for (auto it = children.begin() ; it != children.end() ; ++it) {
  //	  if (ti->label == name) { ti->add(package, rest, level + 1); return; }
  //	}
  //
  // Zeilen 40–52 — der Kern der Gruppierung: 194 Stile eines Spiels stehen
  // unter **einem** Knoten und nicht unter 194 gleichnamigen.
  it('fasst gleiche Gruppen zusammen', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'magic-m15', gruppe: 'magic/m15 style' }),
        paket({ name: 'magic-m15-planeswalker', gruppe: 'magic/m15 style' }),
      ],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.filter((zeile) => zeile.paket === undefined)).toHaveLength(2)
    expect(zeilen).toHaveLength(4)
  })

  // **Geordnet wird nach Art, dann Rang, dann Aufschrift.**
  //
  //	if (a->position_type < b->position_type) return true;
  //	…
  //	if (a->position_hint < b->position_hint) return true;
  //	…
  //	return a->label < b->label;
  //
  // Zeilen 61–67. Die drei Stufen werden einzeln geprüft, denn eine
  // Umsetzung, die nur die erste beachtet, bestünde einen Test über alle drei
  // zusammen genauso gut wie eine richtige.
  it('ordnet Geschwister nach dem Rang', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'spaet', gruppe: 'g', rang: 900 }),
        paket({ name: 'frueh', gruppe: 'g', rang: 100 }),
      ],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.slice(1).map((zeile) => zeile.aufschrift)).toEqual([
      'frueh',
      'spaet',
    ])
  })

  it('ordnet bei gleichem Rang nach der Aufschrift', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'b', gruppe: 'g', rang: 100 }),
        paket({ name: 'a', gruppe: 'g', rang: 100 }),
      ],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.slice(1).map((zeile) => zeile.aufschrift)).toEqual(['a', 'b'])
  })

  it('ordnet die Arten vor dem Rang', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'stil', gruppe: 'g', artrang: 30, rang: 100 }),
        paket({ name: 'spiel', gruppe: 'g', artrang: 20, rang: 900 }),
      ],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.slice(1).map((zeile) => zeile.aufschrift)).toEqual([
      'spiel',
      'stil',
    ])
  })

  // **Ein Gruppenknoten erbt die kleinste Einordnung seiner Kinder.**
  //
  //	if (new_type < position_type || (new_type == position_type && new_hint < position_hint)) {
  //	  position_type = new_type;
  //	  position_hint = new_hint;
  //	}
  //
  // Zeilen 21–27 — und ohne diese Zeilen stünde jede Gruppe hinten, weil ein
  // frischer Knoten mit `TYPE_OTHER` und einer Million anfängt.
  //
  // **Mit je zwei Paketen je Gruppe**, damit die Gruppenzeile stehen bleibt:
  // eine Gruppe mit genau einem Paket wird durch dieses ersetzt (siehe
  // [ausgeben]).
  it('gibt einer Gruppe die Einordnung ihres frühesten Kindes', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'spaet', gruppe: 'zweite', rang: 500 }),
        paket({ name: 'spaet-zwei', gruppe: 'zweite', rang: 501 }),
        paket({ name: 'frueh', gruppe: 'erste', rang: 100 }),
        paket({ name: 'frueh-zwei', gruppe: 'erste', rang: 101 }),
      ],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => zeile.aufschrift)).toEqual([
      'erste',
      'frueh',
      'frueh-zwei',
      'zweite',
      'spaet',
      'spaet-zwei',
    ])
  })

  // **Ein Paket ohne Gruppe kommt unter einen Sammelknoten.**
  //
  //	String group = ip->description->installer_group;
  //	if (group.empty()) group = _("custom");
  //
  // Zeilen 204–205. `_(S)` ist **kein** Übersetzungsaufruf, sondern ein Makro
  // für Unicode-Literale (`src/util/string.hpp`, Zeile 58) — MSE zeigt dort
  // wörtlich „Custom", und die Aufschrift steht hier deshalb im Sprachkatalog
  // und nicht in diesem Modul.
  //
  // **Der erste Anlauf wich hier ab** und stellte solche Pakete auf die
  // oberste Stufe. Die Sichtprüfung hat es widerlegt: von den 324 Paketen des
  // installierten `M15-Magic-Pack` tragen **99** keine Gruppe — nach zwei
  // zugeklappten Knoten folgten 99 lose Zeilen, und das ist wieder die Liste,
  // die #244 abschafft.
  it('sammelt Pakete ohne Gruppe unter einem Knoten', () => {
    const zeilen = paketbaum(
      [paket({ name: 'allein' }), paket({ name: 'auch-allein' })],
      'Ohne Gruppe',
      'de',
    )

    expect(zeilen.map((zeile) => [zeile.stufe, zeile.aufschrift])).toEqual([
      [0, 'Ohne Gruppe'],
      [1, 'allein'],
      [1, 'auch-allein'],
    ])
  })

  // **Zwei Pakete mit demselben Pfad bekommen zwei Zeilen.**
  //
  //	if (pos == String::npos && ti->package) {
  //	  // two packages with the same path
  //	  TreeItemP ti2(new TreeItem); …
  //	}
  //
  // Zeilen 43–49. Der Fall kommt vor: `magic-mana-large` und
  // `magic-mana-small` teilen sich Gruppen im Pack. Wer ihn übersieht,
  // **verliert** das zweite Paket still — es ist installiert und steht in
  // keiner Liste.
  it('verliert kein Paket mit gleichem Gruppenpfad', () => {
    const zeilen = paketbaum(
      [
        paket({ name: 'erstes', gruppe: 'magic/mana' }),
        paket({ name: 'zweites', gruppe: 'magic/mana' }),
      ],
      'Ohne Gruppe',
      'de',
    )

    const namen = zeilen
      .filter((zeile) => zeile.paket !== undefined)
      .map((zeile) => zeile.paket?.name)

    expect(namen).toEqual(['erstes', 'zweites'])
  })

  it('gibt für nichts nichts', () => {
    expect(paketbaum([], 'Ohne Gruppe', 'de')).toEqual([])
  })
})

// **Was zugeklappt ist, bleibt verborgen** — und mit ihm alles darunter.
//
// **Zugeklappt ist die Ausnahme, nicht die Regel**: MSE2 ruft beim Öffnen
// seines Paketfensters `package_list->expandAll()`
// (`src/gui/packages_window.cpp`, Zeile 222, Fassung f9d9356 vom 2021-07-06).
// Das Prädikat hier sagt für jeden Schlüssel, ob er offen ist; wer ihn füllt,
// entscheidet der Aufrufer.
describe('sichtbareZeilen', () => {
  // **Zwei Pakete je Gruppe**, damit die Gruppenzeilen stehen bleiben — eine
  // Gruppe mit genau einem Paket wird durch dieses ersetzt (siehe [ausgeben]).
  const zeilen = paketbaum(
    [
      paket({ name: 'magic-m15', gruppe: 'magic/m15 style' }),
      paket({ name: 'magic-m15-pw', gruppe: 'magic/m15 style' }),
      paket({ name: 'vtes', gruppe: 'vtes' }),
      paket({ name: 'vtes-zwei', gruppe: 'vtes' }),
    ],
    'Ohne Gruppe',
    'de',
  )

  it('zeigt ohne Ausklappen nur die oberste Stufe', () => {
    expect(
      sichtbareZeilen(zeilen, () => false).map((zeile) => zeile.aufschrift),
    ).toEqual(['magic', 'vtes'])
  })

  it('zeigt die Kinder einer ausgeklappten Gruppe', () => {
    expect(
      sichtbareZeilen(zeilen, (k) => k === 'magic').map(
        (zeile) => zeile.aufschrift,
      ),
    ).toEqual(['magic', 'm15 style', 'vtes'])
  })

  // **Ein Enkel braucht jeden Vorfahren.** Wer nur die tiefe Gruppe ausklappt,
  // ohne die darüber, sähe eine Zeile ohne ihren Weg — und das ist der Fall,
  // der entsteht, wenn jemand eine Gruppe zuklappt, deren Kind offen war.
  it('zeigt einen Enkel erst, wenn jeder Vorfahr offen ist', () => {
    expect(
      sichtbareZeilen(zeilen, (k) => k === 'magic/m15 style').map(
        (zeile) => zeile.aufschrift,
      ),
    ).toEqual(['magic', 'vtes'])

    expect(
      sichtbareZeilen(zeilen, (k) => k.startsWith('magic')).map(
        (zeile) => zeile.aufschrift,
      ),
    ).toEqual(['magic', 'm15 style', 'magic-m15', 'magic-m15-pw', 'vtes'])
  })
})
