# 0028 — Die Fassung in einer Herkunftsangabe ist der Commit-Hash

**Stand:** gültig
**Datum:** 2026-08-18
**Betrifft:** #296

## Lage

`AGENTS.md` §11 verlangt für jede übernommene Regel die Herkunft **an Ort und
Stelle** — „Datei und Fassung, im Kommentar neben dem Code".

Genannt wurden Datei und **Zeile**. Eine Zeilennummer ohne Fassung zeigt bei
der nächsten Änderung an MSE2 auf etwas anderes; bei einer späteren Abweichung
ließe sich nicht bestimmen, gegen welchen Stand geprüft wurde — und genau
dafür steht die Angabe da.

Am 2026-08-18 nachgezählt: **40 Kommentarblöcke** mit einem Verweis auf eine
`.cpp`- oder `.hpp`-Datei trugen keine Fassung, bei 421 Verweisen insgesamt.

## Entscheidung

**Die Fassung ist der Commit-Hash des Klons, mit seinem Datum:**

```
Fassung f9d9356 vom 2021-07-06
```

Nicht ein Release-Tag und nicht eine Versionsnummer. Der Grund ist derselbe
wie bei den GitHub-Actions in §1: ein Tag bezeichnet, was heute darunter
liegt; ein Hash bezeichnet den Stand. Das Datum steht daneben, weil ein Hash
allein niemandem sagt, welcher Stand es ist.

**Sie steht im Kommentar**block, nicht zwingend in derselben Zeile. Ein Block,
der drei Fundstellen nennt, braucht die Fassung einmal — sie gilt für alle
drei, weil es nur einen Klon gibt.

## Der Wächter

`internal/werkzeuge/herkunft_test.go` liest jede Go-Datei unter `internal/`,
sucht Kommentarblöcke mit einem Verweis auf `src/…/….cpp` oder `.hpp` und
verlangt in demselben Block das Wort `Fassung`.

**Was er nicht prüft**, und das steht auch dort:

- **Ob der Hash stimmt.** Er vergleicht keinen Kommentar mit dem Klon; er
  verlangt nur, dass eine Fassung dasteht.
- **Ob die Zeilennummer noch trifft.** Dafür müsste der Klon in der Pipeline
  liegen, und er liegt nach §11 ausdrücklich **neben** dem Repositorium.
- **Verweise ohne Dateipfad.** „`ScriptValue::toColor` sagt …" ohne
  `src/script/value.cpp` fällt durch — der Wächter sieht den Pfad, nicht den
  Anspruch.

## Was das kostet

**Ein zweiter Klon bräuchte eine zweite Angabe.** Kommt `magic-set-generator`
als Quelle dazu (§11 erlaubt es), trägt seine Fundstelle eine eigene Fassung —
der Wächter verlangt heute nur das Wort und unterscheidet die Quellen nicht.
Wer die zweite Quelle in die Prüfung nimmt, erweitert ihn hier.
