// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': 'Разрешить обмен и торговлю',
  'richtlinie.handel.markt': 'Разрешить торговлю через стороннего поставщика',
  'richtlinie.handel.menge': 'Максимальное количество на позицию',
  'richtlinie.handel.wert':
    'Максимальная стоимость предложения в наименьшей единице (для EUR — центы)',
  'richtlinie.handel.frist': 'Наибольший срок удержания в минутах',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    'Одновременно отключает передачу, одалживание колод, обмен и рынок. **Ограничение, а не предпочтение**: то, что здесь выключено, нельзя снова включить ни в одном профиле.',
  'richtlinie.handel.markt':
    'По умолчанию **выкл.** Рынок обращается к стороннему поставщику и к деньгам — всё остальное остаётся между двумя людьми.',
  'richtlinie.handel.menge':
    'Сколько экземпляров может называть отдельная строка предложения. **Пусто — без ограничения.**',
  'richtlinie.handel.wert':
    'Какой может быть суммарная стоимость предложения. Действует только там, где оценка **является частью договора** — простой обзор рынка её не ограничивает. **Пусто — без ограничения.**',
  'richtlinie.handel.frist':
    'Как долго обязательство может удерживать экземпляры. **Пусто — без ограничения.**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': 'Выкл.',
  'richtlinie.handel.an': 'Вкл.',
  'richtlinie.handel.markt.aus': 'Выкл.',
  'richtlinie.handel.markt.an': 'Вкл.',
}
