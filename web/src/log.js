function quote(value) {
  return JSON.stringify(String(value))
}

export function logfmt(level, event, fields = {}) {
  const values = {
    time: new Date().toISOString(),
    level,
    component: 'web',
    event,
    ...fields,
  }
  const line = Object.entries(values)
    .map(([key, value]) => `${key}=${quote(value)}`)
    .join(' ')
  console.log(line)
}
