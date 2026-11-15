// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} / {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'カード', viele: 'カード' },
    set: { eine: 'セット', viele: 'セット' },
    paket: { eine: 'パッケージ', viele: 'パッケージ' },
    bild: { eine: '画像', viele: '画像' },
    schritt: { eine: 'ステップ', viele: 'ステップ' },
  },
  einheiten: {
    karte: { eine: 'カード', viele: 'カード' },
    set: { eine: 'セット', viele: 'セット' },
    paket: { eine: 'パッケージ', viele: 'パッケージ' },
    bild: { eine: '画像', viele: '画像' },
    schritt: { eine: 'ステップ', viele: 'ステップ' },
  },
  laeuft: '実行中…',
  abbrechen: 'キャンセル',
  wirdAngehalten: '停止しています…',
  abgebrochen:
    'キャンセルしました。すでに完了した分は残ります。中途半端に書き込まれたものはありません。',
  bisDahin: 'ここまで: {stand}。',
} as const
