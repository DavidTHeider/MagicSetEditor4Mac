# 0074 — Wer arbeitet, steht fest

**Entschieden am 2026-09-06**, Werkstück 17 von Phase 4.12
(#1400) ·
**Ergänzt** [0070](0070-rechte-rollen-und-sichtbarkeit.md) ·
**Voraussetzung für** [0073](0073-ein-leser-darf-wirklich-lesen.md)

## Der Anlass

**Beim Entwurf von Werkstück 16 fiel auf, dass es keinen angemeldeten Nutzer
gibt.** Wer „arbeitet", wurde aus dem *gezeigten Profil* abgeleitet:

```go
gezeigt, fehler := v.Gezeigt(ctx)
…
verlangt := v.freigaben.Verlangt(ctx, gezeigt.NutzerID, ziel, RolleMitarbeiter)
```

Nach einem Wechsel in ein fremdes Profil ist `gezeigt.NutzerID` der **fremde**
Eigner. Wer hineingewechselt war, wurde damit zu ihm — mit allem, was er darf.

## Was gemessen ist

**Am 2026-09-06 an `internal/service`**, drei Nutzer, zwei Freigaben:

| Schritt | Freigabe | Ergebnis |
| --- | --- | --- |
| Erster → Zweites Profil | vorhanden | glückt |
| **Erster → Drittes Profil** | **keine** | **glückt** — geprüft wurde als *Zweiter* |

Der Zeiger stand danach auf dem dritten Profil, ohne dass der Erste je eine
Freigabe darauf hatte.

**Die Reichweite dieser Messung** (§10): gemessen am Dienst, die Daten kamen
aus Doppeln, die geprüfte Logik ist `freigabeReicht` selbst. **Ohne zweiten
Faktor** — `Zutritt` verlangt zusätzlich den des Zieleigners, und die Kette
bricht dort, wo einer geführt wird. Das ist die Voreinstellung nicht.

## Die Entscheidung

**„Wer arbeitet" steht im Bestand und ändert sich durch einen Profilwechsel
nicht.**

### 1. Ein eigener Zeiger neben dem Profilzeiger

`nutzer_zeiger` neben `profil_zeiger`, beide im `bestand`. Zwei Fragen, zwei
Einträge: *wer* arbeitet und *worin*.

**Ohne Eintrag gilt der Eigner des gezeigten Profils** — das Verhalten von
vorher. Jeder Bestand, der vor diesem Werkstück entstand, hat keinen, und beim
ersten Start soll sich nichts ändern, nicht einmal für einen Augenblick.

**Festgeschrieben wird er in dem Augenblick, in dem der Rückfall sonst
driftete**: wenn der Zeiger sich in ein **fremdes** Profil bewegt. Für einen
Bestand mit einem einzigen Menschen entsteht damit nie eine Zeile.

### 2. Ein anderer Nutzer wird man **ausdrücklich**

Der Auftraggeber hat sich für die getrennte Geste entschieden. Zwei Wege
standen zur Wahl:

| | |
| --- | --- |
| **Gewählt** | Ein eigener Befehl *Nutzer wechseln*, der den zweiten Faktor des **Ziels** verlangt. Ein Profilwechsel ändert die Identität **nie** |
| Verworfen | Ein Profilwechsel ohne Freigabe *wird* zum Nutzerwechsel. Kein neuer Befehl — aber zwei Bedeutungen an einer Geste, und niemand sähe, wann die Identität wechselt |

**Der Ausweis ist der des Ziels und nicht der eigene.** Wer den eigenen
fragte, ließe jeden zu jedem werden, der seinen eigenen Code kennt — die
Prüfung schützte dann den Falschen.

**Wer sich selbst wählt, wird nicht gefragt.** Ein Ausweis für das, was ohnehin
gilt, wäre eine Hürde ohne Zweck — und eine, die jemanden aussperrte, dessen
Faktor gerade klemmt.

**Der Zeiger zieht mit.** Wer ein anderer Mensch wird, arbeitet in dessen
erstem Profil. Bliebe er stehen, stünde der neue Mensch in einem Profil, für
das er womöglich keine Freigabe hat.

## Was das **nicht** löst

**Einen Bestand ohne zweite Faktoren schützt es nicht.** Die Regel aus
Werkstück 11 gilt unverändert: „ausweisen **oder** freigegeben sein", und wo
niemand einen Faktor führt, sperrt sie niemanden aus. Der Nutzerwechsel ist
dann frei — und die Kette entsteht wieder, nur sichtbar und als eigene
Handlung.

**Das ist kein Versehen, sondern die Zusage des zweiten Faktors selbst.** Wer
sich auf einem geteilten Rechner schützen will, richtet einen ein; das steht
seit Werkstück 7 so da, und es steht in der Hilfe.

## Was es kostet

- **Ein neuer Befehl an allen vier Flächen** und eine Bindung mehr am Fenster.
- **Ein Begriff mehr**, den jede Fläche kennen muss: „wer arbeitet" ist nicht
  „wem das gezeigte Profil gehört". Wo die beiden verwechselt werden, zeigt
  eine Seite die Rolle des Falschen.
- **Phase 4.12 wächst auf siebzehn Werkstücke**, und 0.4.0 rückt nach hinten.
  §13 verlangt die sichtbare Begründung; sie steht in `docs/ROADMAP.md`.
