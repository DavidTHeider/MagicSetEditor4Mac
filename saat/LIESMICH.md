# Die Saat

**Selbst geschrieben, rechtlich unbedenklich, und mitgeliefert.**

`docs/entscheidungen/0005` sagt, warum fremde Vorlagensammlungen nicht in dieses
Repository kommen: sie stehen unter keiner Lizenz und enthalten geschützte
Grafiken. Wer die Anwendung startet, soll trotzdem etwas sehen können — dafür
ist die Saat da.

| Paket | Was es ist |
| --- | --- |
| `probe.mse-game` | ein Spiel mit vier Feldern: `name`, `text` (mehrzeilig — FP-02, #2156), `farbe`, `kunst` |
| `probe-schlicht.mse-style` | ein Stil dazu: zwei Textfelder, eine Schrift, ein Feldskript — und **drei Stiloptionen** (unten) |
| `probe-bild.mse-style` | derselbe Stil mit Rahmen und Maske — `image:` und `mask:` |
| `probe-kunst.mse-style` | derselbe Stil mit **Kunstfenster** — dem Feld, an dem die Karte ihr eigenes Bild mitbringt (#380) |
| `probe-liste.mse-export-template` | eine **Exportvorlage**: eine Zeile je Karte, mit zwei Wahlmöglichkeiten (#680) |
| `probe-werkstatt.mse-set` | die **Testkarten der Vorlagenwerkstatt** — fünf Karten auf `probe-kunst`: Langtext, Sonderlayout, Zweiseitig, Ohne Bild, Sonderzeichen. Sie erscheinen unter „Beispielsets"; wer eine Vorlage ändert, wählt sie im Inspector der Werkstatt als Testkarte (#1418) |

**Warum eine Exportvorlage dazugehört.** Ohne sie war einer der fünf Ausgänge
auf einem frischen Rechner nicht zu erreichen: der Abnahmeweg „ein Set
hinausgeben" musste ihn überspringen, und wer die Anwendung neu öffnete,
konnte nicht sehen, was eine Exportvorlage überhaupt ist. Sie ist zwanzig
Zeilen und die einfachste, die etwas Nützliches tut.

**`kunst` ist ein Feld der Art `image`, und es bringt keine Grafik mit.** Das
Bild einer Karte liegt im **Set**, nicht in der Vorlage — deshalb steht in
`probe-kunst.mse-style` ein Kasten und sonst nichts. Wer sehen will, wie er
gefüllt wird, liest `internal/app/kunstnaht_intern_test.go`: der Lauf schreibt
sich sein Set samt Illustration selbst.

**Die Stiloptionen von `probe-schlicht`, und was sie tun.** Wer die Anwendung
öffnet, soll auf der Stilseite nicht nur Bedienelemente sehen, sondern auch,
dass sie wirken:

| Option | Art | Was sie tut |
| --- | --- | --- |
| `titelfarbe` | `color`, mit `default: "black"` | färbt die Überschrift, `color: { styling.titelfarbe }`. Zugleich der einzige Ort der Saat, an dem sich der **Vorgabeeintrag** vorführen lässt (#764) |
| `hintergrund` | `choice` | füllt den Kartengrund — `hell` ist Weiß und damit das Bild von vorher, `dunkel` ein Schiefer |
| `rahmen zeigen` | `boolean` | schaltet einen acht Punkte breiten Rand um die ganze Karte |

**Die letzten beiden brauchen zwei Felder, die der Stil selbst mitbringt** —
`extra card field:` vom Typ `color`, gerechnet und nicht auszufüllen
(`save value: false`, `editable: false`). So macht MSE es auch: der schwarze
Rand jeder echten Karte ist ein Farbfeld, und `card stamp border` in
`magic-m15.mse-style` ist genau so ein Zusatzfeld.

**Bis zum 2026-08-22 taten alle drei nichts** — drei `styling field:` gegen
null `styling.`-Nutzungen, in einer ausgelieferten Vorlage
(#770). Wer sie
umstellte, sah keine Änderung.

**Und eine vierte kommt nicht mehr dazu.**
`internal/werkzeuge/saatbedienelemente_test.go` hält die Klasse: jedes
deklarierte Bedienelement wird gelesen — eine Stiloption als `styling.<name>`,
eine Ausfuhroption als `options.<name>`. Was er **nicht** hält, ist die
Sichtbarkeit der Wirkung; dafür zeichnet
`app.TestEineStiloptionDerSaatAendertDasBild` die Karte je Option zweimal.

**`titelfarbe` hieß bis zum 2026-08-21 `randfarbe`** und beschrieb sich als
„Die Farbe des Kartenrands" — ein Name, der etwas zusagte, das nicht geschah.
Er ist umbenannt worden, nicht bloß verdrahtet.

**Sie geht mit ins Bündel** (#58).
`scripts/saat.sh` legt dieses Verzeichnis beim Bauen unter
`Contents/Resources/saat` ab; gerufen wird es über `postBuildHooks` in
`wails.json`, und damit auf demselben Weg für `wails build` und `wails dev`.
Die laufende Anwendung findet es über `app.Saatverzeichnis` neben ihrem eigenen
Programm — bezogene Pakete kommen davor, die Saat danach.

**Die Grafik ist gezeichnet, nicht gefunden.** `rahmen.png` und
`rand_maske.png` entstehen aus einer Rechenvorschrift im Repository:
`internal/werkzeuge/saatgrafik_test.go` zeichnet beide und hält die Dateien
dagegen. Wer wissen will, woher sie kommen, liest dreißig Zeilen; wer sie
austauscht, wird rot.

**Warum das eigens dasteht:** ein PNG sieht einem PNG gleich, und „selbst
gemacht" ist eine Behauptung, die niemand nachprüfen kann.
[0005](../docs/entscheidungen/0005-keine-fremden-vorlagen.md) hält fremde
Vorlagen aus dem Repository — eine Grafik, die einfach dalag, ließe sich davon
nicht unterscheiden.

**Zwei Rechtecke und eine abgerundete Maske, mehr nicht.** Die Saat zeigt, dass
der Weg vom Stil zum Bild trägt — nicht, wie eine schöne Karte aussieht. Jede
Linie mehr wäre eine Linie, die der Goldbildvergleich mitschleppt.

**Die Schrift ist `Helvetica`**, weil macOS sie mitbringt und `x/image` ihre
Sammlung lesen kann (`internal/service/vorschau.go` sagt, warum nicht
`Helvetica Neue`). Eine mitgelieferte Schriftdatei hätte wieder einen Urheber.
