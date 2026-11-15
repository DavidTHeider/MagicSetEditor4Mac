# 0196 — Jede Sprache spricht die Lesenden vertraut an, in einer Form

**Stand:** angenommen am 2026-10-17 · #2368 (AP-11, Befund B-27). Wächter: [`frontend/src/sprache/anrede.test.ts`](../../frontend/src/sprache/anrede.test.ts).

## Lage

Die deutsche Ausgangssprache duzt. Drei Sätze siezten trotzdem („Hier
arbeiten Sie gerade“, „tippen Sie, um einzugrenzen“, „was Sie anhaken“),
und die Übersetzungen mischten, gezählt am 2026-10-17:

| Sprache | förmlich | vertraut |
| --- | ---: | ---: |
| Französisch | 440 „vous/votre/vos“ | 27 „tu/ton/ta/tes“ |
| Russisch | 325 „вы/вас/ваш…“ | 21 „ты/тебя/твой…“ |
| Chinesisch | 60 „您“ | 292 „你“ |
| Spanisch | keine — „su“ und „puede“ stehen dort in dritter Person, durchgesehen | durchgehend „tú“ |
| Italienisch | keine | durchgehend „tu“ |
| Japanisch | keine Ehrerbietung gegenüber den Lesenden | — |

Das Abnahmekriterium von AP-11 verlangt „Du-Ansprache in acht Katalogen“.

## Entscheidung

1. **Jede Sprache spricht die einzelne Leserin, den einzelnen Leser
   vertraut an, und nur so:** Deutsch „du“, Französisch „tu“, Spanisch „tú“
   (die Fassung für Spanien, wie „ordenador“ zeigt), Italienisch „tu“,
   Russisch „ты“, Chinesisch „你“. Englisch kennt keine zwei Formen.
2. **Japanisch bleibt höflich-neutral** (です・ます, „ください“) und **ohne
   Ehrerbietung gegenüber den Lesenden** (kein „お客様“, „あなた様“,
   „いただく“). Eine vertraute Form im Sinne von „du“ gibt es dort in einer
   Oberfläche nicht; die neutrale Höflichkeit ist ihre Entsprechung.
3. **Mehrere Menschen zugleich** — zwei Spielende, zwei Handelnde —
   spricht die Mehrzahl an: „ihr“, im Französischen „vous“, im Russischen
   „вы“. Diese Stellen stehen mit Grund in der Ausnahmeliste des Wächters.
4. **Ein Wächter hält es**, je Sprache mit den Formen, die eine
   Rückkehr erkennen lassen. Was er nicht sieht, steht in seinem Kopf: im
   Spanischen eine förmliche Verbform ohne Marke, im Deutschen ein „Sie“ am
   Satzanfang.

## Verworfen

- **Französisch und Russisch förmlich lassen**, weil Software dort oft
  siezt: dann spräche dieselbe Anwendung in zwei Sprachen verschieden nah,
  und die Mischung bliebe — die vertrauten Stellen waren ja schon da.
- **Einen Wächter nur für Deutsch**: die Mischung war in den Übersetzungen
  größer als im Deutschen.

## Rückweg

Die Kataloge sind Text; eine andere Regel je Sprache ist eine neue Fassung
dieses Vermerks und ein neuer Satz im Wächter. Ein redaktionell geprüfter
Katalog hat Vorrang vor der maschinellen Übertragung von heute.
