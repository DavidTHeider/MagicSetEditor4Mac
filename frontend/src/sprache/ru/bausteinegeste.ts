// Die Ansagen einer Geste — GESTE (#2358); Ausgangssprache ist Deutsch.
export const geste = {
  gegriffen:
    'Взято. Цель: {ziel} ({stelle} из {von}). Стрелки выбирают цель, Enter кладёт, Escape отменяет.',
  ziel: 'Цель: {ziel} ({stelle} из {von}).',
  abgelegt: 'Положено: {ziel}.',
  abgebrochen: 'Отменено. Ничего не перемещено.',
  keinZiel: 'Сейчас для этого объекта нет цели.',
  zielFehlt: 'Этой цели больше нет — ничего не перемещено.',
  veraltet: 'Тем временем состояние изменилось — ничего не перемещено.',
} as const
