# 0145 — Zusatzdaten werden nicht in die Setdatei eingebettet

**Stand:** entschieden am 2026-09-29 · **Betrifft:** FMT-05
(#1694),
AK-02 · **Ergänzt** [0144](0144-keine-neue-endung-fuer-deck-und-sammlung.md),
[0052](0052-die-sammlung-ist-nicht-das-set.md)

## Die Frage

Preise, Exemplare, Marken und die Herkunft einzelner Werte gehören zu einem
Set, stehen aber nicht in MSE2s Datenmodell. Sollen sie **in** die
`.mse-set`-Datei (eingebettete Felder) oder **neben** sie (eine Datei daneben)?

## Gemessen am 2026-09-29

Mit dem echten MSE2 (Fassung `f9d9356` vom 2021-07-06): ein Set mit zwei
eingebetteten Feldern und einer Datei daneben öffnen, über `write_set_file`
wieder speichern, vergleichen ([`scripts/mse2-zusatzdaten.sh`](../../scripts/mse2-zusatzdaten.sh),
Beleg [`docs/belege/mse2-zusatzdaten.md`](../belege/mse2-zusatzdaten.md)):

| Ansatz | Beim Öffnen | Nach dem Speichern durch MSE2 |
| --- | --- | --- |
| **eingebettet** (`mse4mac preisstand:` im Setkopf, `mse4mac exemplar:` in einer Karte) | je eine Warnung „Unexpected key" | **verloren** — beide |
| **Datei daneben** (`probe.mse4mac.json`) | nichts | liegt unverändert da; **MSE2 nimmt sie nicht mit** — die gespeicherte Datei trägt nur `set` und die Karte |
| MSE2-eigenes `notes:` einer Karte (zum Vergleich) | nichts | bleibt |

**Gelesen** dazu: `Reader::unknownKey` (`src/util/io/reader.cpp`, Zeilen
231–246) überspringt einen unbekannten Schlüssel mit einer Warnung; der
Schreibweg serialisiert nur, was die Klassen erklären.

## Die Entscheidung

1. **Nichts wird eingebettet.** Eingebettete Felder machten jedes Öffnen in
   MSE2 laut (Warnungen) und gingen beim ersten Speichern dort verloren —
   genau die stille Beeinträchtigung, die FMT-05 ausschließt.
2. **Keine Datei daneben.** Sie überlebt MSE2, aber sie reist nicht mit: ein
   „Speichern unter" in MSE2 lässt sie zurück, und danach beschreibt sie ein
   Set, das es so nicht mehr gibt.
3. **Zusatzdaten leben im Bestand** und reisen in **eigenen, versionierten
   Behältern** mit Namensraum: dem Umschlag (`austausch.Fassung`, Manifest mit
   Objektart) und der Sammlungs-CSV (Druckverweis mit `domain.Verweisfassung`
   je Zeile) — beide in [`FASSUNGEN.md`](../FASSUNGEN.md).
4. **Die Sicherung sagt, was sie nicht trägt** — unter *Ausgeben → Als
   Setdatei sichern* und in [`DATEIOPERATIONEN.md`](../DATEIOPERATIONEN.md).

## Verworfen

- **Einbetten mit MSE2-verträglichem Namen** (etwa in `notes:`) — `notes` ist
  ein Menschenfeld; Maschinendaten darin wären eine zweite Bedeutung desselben
  Feldes und gingen verloren, sobald jemand die Notiz bearbeitet.
- **Eine Datei daneben mit Rückverweis** — MSE2 kennt ihn nicht und pflegt ihn
  nicht; die Datei veraltete trotzdem.
