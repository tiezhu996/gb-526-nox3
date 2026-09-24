export const formatGas = (mix: { o2: number; he: number; n2: number }) =>
  `O2 ${(mix.o2 * 100).toFixed(0)} / He ${(mix.he * 100).toFixed(0)} / N2 ${(mix.n2 * 100).toFixed(0)}`

export const formatTimestamp = (value: string) => new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium', timeStyle: 'short',
}).format(new Date(value))
