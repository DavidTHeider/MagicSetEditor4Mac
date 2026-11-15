# 0047 — Eine zehnte Phase: originalgetreue Karten

**Datum:** 2026-08-24
**Stand:** angenommen
**Betrifft:** `internal/render`, `internal/service` (Planbau), `internal/svg`,
`docs/ROADMAP.md`

---

## Die Entscheidung

**Zwischen Phase Y und Phase Z kommt eine zehnte Phase: Ω —
originalgetreue Karten.** Sie hat drei Gegenstände, und der dritte ist der
Grund für sie:

1. **Geschwindigkeit** — gemessen, mit Benchmarks, die es heute nicht gibt.
2. **Vier Güten** — Ultra-High-Res, High-Res, Medium, Fast, als Wahl der
   Nutzenden und als Angabe an jeder Ausgabe.
3. **Der Abstand zu MSE2** — er ist zu groß, und was ihn ausmacht, ist nicht
   aufgeschlüsselt.

**Sie wird nicht Phase W zugeschlagen.** W misst die *Fähigkeiten* von MSE2 und
trägt sie in eine Paritätsmatrix; hier geht es um die *Treue* einer Fähigkeit,
die längst da ist. §13 verlangt außerdem ausdrücklich, dass eine Phase nicht
stillschweigend größer wird — und der Umfang hier ist eine eigene Phase wert.

## Der Buchstabe sagt, wann sie geschnitten wurde

**Ω, nicht ein Buchstabe zwischen Y und Z** — den gibt es nicht, und die
Buchstaben R bis Z sind vergeben ([0043](0043-neun-phasen-bis-0-2-0.md)).

Das ist dieselbe Übereinkunft wie bei Phase 9
([0009](0009-phase-9-fremdformate-und-schnittstelle.md)): *ihre Nummer sagt,
wann sie geschnitten wurde, nicht wann sie gebaut wird.* Ω ist die zuletzt
geschnittene und wird als vorletzte gebaut — vor Z, weil Z die Freigabe ist und
in ihr nichts Neues mehr entsteht.

## Was gemessen wurde, bevor diese Phase geschnitten wurde

**Am 2026-08-24**, an diesem Rechner, mit den Kommandos daneben.

### Der Abstand zu MSE2 an den vier Referenzkarten

```bash
MSE_FREMDPAKETE=/Users/david/full-magic-pack/data \
MSE_REFERENZBILDER="$HOME/Library/Application Support/MagicSetEditor4Mac/referenzbilder" \
go test -tags fremdpakete -run 'TestDerAbstandZuMSE2$' -v ./internal/app/
```

| Karte | Abstand ohne Bildfenster | betroffene Bildpunkte |
| --- | --- | --- |
| Firebrand-Incinerator | 31,830 | 103.332 von 123.129 — **83,9 %** |
| Mindwarp-Serpent | 34,950 | 103.324 von 123.129 — **83,9 %** |
| Old-Money | 36,330 | 100.535 von 123.129 — **81,7 %** |
| Progress-Through-Struggle | 38,612 | 102.572 von 123.129 — **83,3 %** |

**Vier Karten sind keine Aussage über das Pack** (§10). Und die Zahl selbst ist
als Verlauf zu lesen, nicht als Betrag — der Kopf von `mse2goldbild_test.go`
nennt drei Gründe, warum sie nicht null werden kann.

**Was auffällt, ist nicht die Höhe, sondern die Breite:** über 80 % aller
Bildpunkte weichen ab, und zwar bei jeder der vier Karten fast gleich weit. Das
ist die Signatur eines **durchgehenden** Unterschieds — ein Versatz, eine
Farbkurve, eine Kantenglättung —, nicht die von ein paar verrutschten Texten.
**Diese Deutung ist eine Vermutung und der erste Messauftrag der Phase**, keine
Feststellung.

### Der Kanonlauf ist heute nicht wiederholbar

`TestDerAbstandZuMSE2AmKanon` misst gegen Bilder, die **MSE2 selbst** gerastert
hat ([0026](0026-mse2-rendert-die-referenz-selbst.md)) — 510 Karten mit
bekannter Herkunft, gegen die 27 Bilder unbekannter Vorlage im anderen Lauf.

```bash
ls "$HOME/Library/Application Support/MagicSetEditor4Mac/mse2-ausgabe"   # leer
```

Das Verzeichnis ist leer. Die belastbare Messgröße muss deshalb **zuerst
wiederhergestellt** werden, bevor irgendeine Verbesserung behauptet wird.

### Geschwindigkeit

**Null Benchmarks** — so steht es schon im Audit über 0.1.0
(`AUDIT-0.2.0.md`), und daran hat sich nichts geändert. Es gibt keine Zahl,
gegen die eine Beschleunigung gemessen werden könnte.

### `scale down to:` ist portiert

Der Verdacht, die längenabhängige Anpassung von Regel- und Flavortext sei ein
Fehler, trifft **nicht** den Umstand selbst: MSE2 tut genau das.

```
double FontTextElement::minScale() const {
  return min(font->size(), font->scale_down_to) / max(0.01, font->size());
}
```

`src/render/text/font.cpp`, Zeilen 51–53, Fassung f9d9356 vom 2021-07-06; die
Suche über die Maßstäbe steht in `prepareLinesTryScales`,
`src/render/text/viewer.cpp`, Zeilen 444–500. Portiert in
#498, belegt
in `internal/render/verkleinerung.go` und `verkleinerung_test.go`.

**Offen ist deshalb nicht das Ob, sondern das Wie**: bei welchem Maßstab
verkleinert wird, in welchen Schritten gesucht wird, und ob die Schranke aus
demselben Feld gelesen wird wie dort. Das ist ein Messauftrag, kein Umbau.

## Die vier Güten

| Güte | Wofür |
| --- | --- |
| **Fast** | die Vorschau beim Tippen — sie darf gröber sein, sie muss sofort da sein |
| **Medium** | die gewöhnliche Anzeige |
| **High-Res** | Ausfuhr als Bild, der Druckbogen |
| **Ultra-High-Res** | Druck in voller Auflösung |

**Sie sind eine Wahl über den Aufwand, nicht über die Treue.** Eine Karte in
`Fast` soll dieselbe Karte sein, nur billiger gerechnet — kein anderes Layout,
keine andere Zeilenumbrechung. Wo eine Güte etwas *weglässt*, sagt sie es.

## Was diese Phase ausdrücklich nicht tut

- **Sie erfindet keine Renderlogik.** §15 gilt unverändert: was MSE2 tut, wird
  dort nachgelesen und portiert.
- **Sie ändert die Semantik nicht, um eine Zahl zu senken.** Eine Messgröße,
  auf die hin optimiert wird, hört auf, eine Messgröße zu sein.
- **Sie lädt keine fremden Bilder ins Repositorium** (§11). Referenzbilder
  liegen neben dem Baum, wie die bezogenen Pakete.

## Der Vergleich gegen Scryfall

**Er misst etwas anderes als der Kanonlauf, und das ist der Grund für ihn.**

| | misst |
| --- | --- |
| Kanonlauf gegen MSE2 | ob **wir** rechnen wie **MSE2** |
| Vergleich gegen Scryfall | ob eine Karte aussieht wie eine **gedruckte** |

Die zweite Frage hat bisher niemand gestellt. Sie ist die, die im Ziel steht:
*originalgetreu*, nicht *MSE2-getreu*.

**Bedingungen** (§16): Bilder von Scryfall werden über den vorhandenen Klienten
geholt, mit Drossel und Zwischenspeicher, liegen **außerhalb** des
Repositoriums, und die Illustration wird beim Vergleich ausgespart — verglichen
werden Rahmen, Textlage und Schrift, nicht die Kunst.

## Verworfen

**Die Güten an die Vergrößerung zu hängen.** `Vorschau` nimmt schon eine
`vergroesserung`; eine Güte daraus abzuleiten wäre bequem und falsch — eine
kleine Vorschau kann sehr genau sein wollen (der Vergleichslauf), eine große
sehr schnell (das Ziehen am Zuschnitt).

**Den Abstand als Abnahmekriterium der ganzen Phase zu nehmen.** Eine einzelne
Zahl über 510 Karten sagt nicht, *was* besser wurde. Abgenommen wird je
Ursache, mit einer Karte, an der sie sichtbar ist.
