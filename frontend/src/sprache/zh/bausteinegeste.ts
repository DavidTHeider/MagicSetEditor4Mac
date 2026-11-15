// Die Ansagen einer Geste — GESTE (#2358); Ausgangssprache ist Deutsch.
export const geste = {
  gegriffen:
    '已拿起。目标：{ziel}（第 {stelle} 个，共 {von} 个）。用方向键选择目标，按回车放下，按 Esc 取消。',
  ziel: '目标：{ziel}（第 {stelle} 个，共 {von} 个）。',
  abgelegt: '已放到{ziel}。',
  abgebrochen: '已取消。没有移动任何东西。',
  keinZiel: '此对象当前没有目标。',
  zielFehlt: '该目标已不存在，没有移动任何东西。',
  veraltet: '状态已在此期间改变，没有移动任何东西。',
} as const
