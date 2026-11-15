// Die Ansagen einer Geste — GESTE (#2358); Ausgangssprache ist Deutsch.
export const geste = {
  gegriffen:
    'つかみました。移動先: {ziel}（{von}件中{stelle}件目）。矢印キーで移動先を選び、Returnで置き、Escapeで取り消します。',
  ziel: '移動先: {ziel}（{von}件中{stelle}件目）。',
  abgelegt: '{ziel}に置きました。',
  abgebrochen: '取り消しました。何も移動していません。',
  keinZiel: 'このオブジェクトには今、移動先がありません。',
  zielFehlt: 'その移動先はもうありません。何も移動していません。',
  veraltet: 'その間に状態が変わりました。何も移動していません。',
} as const
