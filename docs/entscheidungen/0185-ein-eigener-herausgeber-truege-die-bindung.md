# 0185 — Ein eigener Herausgeber trüge die Bindung, und gebaut wird sie mit einem verwahrten Schlüssel

**Stand:** angenommen am 2026-10-11 · I01
(#2259).
**Ergänzt** [0181](0181-der-schluesselbund-bleibt-ungebunden-solange-ad-hoc-signiert-wird.md)
und ersetzt dessen ersten Grund; der Bau stand in
#2295.
**Ergänzt durch [0187](0187-kein-eigener-herausgeber.md)** (2026-10-11): die
Verwahrung ist entschieden — kein eigener Herausgeber, weil eine Bindung
ohne Härtung fremden Code nicht aufhält (E5) und ein Schlüssel, der ohne
Menschen signiert, für jedes Programm desselben Nutzers signiert.

## Die Frage

0181 ließ die Einträge dieser Anwendung im Schlüsselbund ungebunden, vor
allem, **weil eine ad hoc signierte Bindung keine Fassung überlebt**: die
ACL trägt den cdhash, und jede Fassung hat einen neuen (SEC-05, E3). Offen
blieb als Hypothese, ob ein **selbst ausgestellter Herausgeberanker** das
änderte. I01 verlangt den Lauf in einem Wegwerf-Gast — E1 bis E4, die
Gatekeeper-Wirkung getrennt — und eine Entscheidung, die 0181 bestätigt oder
ersetzt.

## Was gemessen ist — am 2026-10-11

Im Lume-Gast mit macOS 15.2, an einem künstlichen Bund, mit zwei im Gast
erzeugten und mit dem Bund gelöschten Herausgebern A und B
([`belege/i01-bindung.md`](../belege/i01-bindung.md),
`node scripts/bindunggast.mjs`). Der Leser ist ein kleines Programm über
`SecItemCopyMatching`, in vier Fassungen signiert:

| Fall | Ergebnis |
| --- | --- |
| **E1** heute, ohne `-T` | `security` liest still; der Leser müsste fragen |
| **E2** gebunden an Fassung 1 von A | `security` bekommt nichts; Fassung 1 liest |
| **E3** Fassung 2 **desselben** Herausgebers A, anderer cdhash | **liest still** |
| **E3** gleicher Bezeichner, Herausgeber **B** | müsste fragen |
| **E3** gleicher Bezeichner, **ad hoc** | müsste fragen |
| **E4** gesperrt | `security` bekommt nichts; der Leser müsste fragen |

Die Anforderung eines vom Herausgeber signierten Programms lautet
`identifier "…" and certificate leaf = H"…"` — **am Zeugnis verankert, nicht
am cdhash**. Gatekeeper ändert sich nicht: `spctl` weist alle vier ab, die
signierten mit `origin=` des Herausgebers.

## Die Entscheidung

**Die Bindung ist tragfähig — mit einem eigenen Herausgeber, dessen
Schlüssel verwahrt bleibt.** Der erste Grund aus 0181 gilt nur für ad hoc
signierte Programme.

**Für 0.8.0 bleibt der Weg über `/usr/bin/security`**, und die Begründung
ist jetzt eine andere:

- **Ein Signierschlüssel ist ein Geheimnis, das jemand verwahren muss** —
  wo er liegt, wer signiert, was bei Verlust (jede Bindung bricht, Zugänge
  sind neu einzutragen) und bei Diebstahl geschieht (wer ihn hat, signiert
  Programme, die die Einträge lesen). Das entscheidet ein Mensch, nicht
  dieser Auftrag; **ohne diese Entscheidung wird auf dem Wirt kein
  Herausgeberzeugnis eingerichtet**.
- **Die Programme lesen heute über `/usr/bin/security`.** Dann ist der
  Aufrufer `security`, und eine Bindung an das Programm greift nicht. Es
  braucht die native API, ohne cgo — eine Änderung an allen vier Programmen.

Der Weg dahin steht in #2295,
Schritt für Schritt, beginnend mit der Verwahrung.

## Was bis dahin gilt — unverändert

- **Das Risiko ist benannt** (`docs/HAERTUNG.md`): ein Programm desselben
  Nutzers liest die Einträge über `security`, ohne dass macOS fragt.
- **Ein gesperrter Bund schützt** (E4).
- **Die Zusage bleibt so eng wie das Gemessene** — keine Aussage, die Einträge
  seien an die Anwendung gebunden.

## Verworfen

- **Jetzt mit einem Probeherausgeber binden.** Ein Schlüssel ohne Verwahrung
  ist entweder verloren (jede Bindung bricht beim nächsten Bau) oder liegt
  ungeschützt neben dem Code — beides schlechter als heute.
- **Die Hypothese ohne Lauf übernehmen.** Sie stand als Hypothese in 0181;
  erst E3 mit Herausgeber A und B trennt »liest still« von »müsste fragen«.
