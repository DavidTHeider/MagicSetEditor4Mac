# 0034 — Assistiertes Erstellen von Paketen gehört nicht zu 0.1

**Stand:** angenommen
**Datum:** 2026-08-20
**Berührt:** [0005](0005-keine-fremden-vorlagen.md) und
[0025](0025-die-vorlagen-duerfen-mitgeliefert-werden.md) — das Argument, das
diesen Punkt einmal dringend machte, ist dort entfallen

## Lage

#243 will ein
eigenes Spiel und einen eigenen Stil anlegen können, **ohne eine Textdatei von
Hand zu schreiben** — samt Grafiken: Manasymbole, Kartenränder, Hintergründe.

**Sein stärkstes Argument war einmal ein rechtliches**: fremde Vorlagen dürfen
nicht mitgeliefert werden (§11), und wer selbst welche erzeugen kann, braucht
sie nicht.

**Dieses Argument trägt nicht mehr.** Seit dem 2026-08-17 dürfen die Vorlagen
mitgeliefert werden ([0025](0025-die-vorlagen-duerfen-mitgeliefert-werden.md));
sie werden es nur nicht, weil sie über ein Gigabyte wiegen. Bezogen werden sie
über `internal/bezug` an einem Stand, mit einem Klick, aus vier Quellen — und
`saat/` liegt daneben als kleine, selbst erstellte Vorlage für den Fall, dass
jemand gar nichts bezieht.

## Entscheidung

**Es gehört nicht zu 0.1, und es ist keine Lücke, sondern eine Erweiterung.**

Der Unterschied ist der, den das Ziel der Iteration macht
([0031](0031-sieben-phasen-bis-zur-freigabe.md)): *in der Liste steht nur noch,
was besser werden soll, nicht was fehlt.* Ein Karteneditor, der Pakete liest,
verwaltet, bezieht und eine eigene Saat mitbringt, ist benutzbar. Einer, der
Pakete auch **erzeugt**, ist besser.

**Der Zuschnitt wird hier trotzdem entschieden**, damit der spätere Entwurf
nicht bei null anfängt:

| Frage aus #243 | Entscheidung |
| --- | --- |
| Wieviel Assistenz? | **Ein Formular über die Felder**, kein Zeichenwerkzeug. Ein Spiel und ein Stil sind Strukturen; ein Manasymbol ist Grafik, und dafür gibt es bessere Programme |
| Erzeugt oder gezeichnet? | **Erzeugt, aus einer Rechenvorschrift** — wie `saat/` es tut (`internal/werkzeuge/saatgrafik_test.go`). Reproduzierbar und rechtlich unbedenklich |
| Was ist die Ausgabe? | **Ein Paket, das MSE2 öffnet.** Alles andere wäre ein Format, das nur dieses Programm liest — und damit das Gegenteil dessen, wofür es dieses Programm gibt |

## Verworfen

**Es in Phase L bis Q hineinnehmen.** Das wäre der Weg, auf dem eine Phase
still größer wird (§13). Der Umfang ist nicht klein: ein Formular über die
Felder eines Spiels, eine Erzeugung für Symbole, ein Schreiber für zwei
Paketformate — und jedes davon mit der Rundlaufzusage im Nacken.

**Das Issue offen lassen, ohne es zu entscheiden.** Dann stünde in der Liste
etwas, das aussieht wie eine Lücke, und die Abnahme von Phase K wäre nicht zu
halten, ohne sie zu verbiegen.

**Es ersatzlos streichen.** Nein: der Wunsch ist berechtigt, und er hat einen
Zuschnitt bekommen. Er wird als **Erweiterung** geführt und nicht als Mangel.

## Folgen

- **#243 wird geschlossen**, und der Zuschnitt steht hier statt in einem
  Entwurf, den niemand liest.
- **Wer es baut, beginnt mit `docs/superpowers/specs/`** und mit den drei
  Zeilen der Tabelle oben.
- **`saat/` bleibt, was es ist**: ein Prüfstück, kein Werkzeug — und der
  Beleg, dass die Erzeugung aus einer Rechenvorschrift geht.
