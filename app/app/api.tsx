const url = new URL('http://localhost:8080')

export const getDir = async (path: string) => {
  url.pathname = `_api/dir${path}`
  const response = await fetch(url)
  if (!response.ok) {
    response.text().then(console.error)
    throw new Error(`Failed to get directory listing for ${path}`)
  }
  return (await response.json()) as {
    name: string
    size: number
    isDir: boolean
  }[]
}

export const remove = async (path: string) => {
  url.pathname = `_api${path}`
  const response = await fetch(url, {
    method: 'DELETE',
  })

  const operation = (await response.json()) as { op: number; id: number }

  return operation
}

export const connectOperationsEventSource = (id: number) => {
  url.pathname = `_api/operations/${id}`
  const es = new EventSource(url)

  es.onmessage = (e) => {
    console.log(e.data)
    const msg = JSON.parse(e.data)

    if ('ok' in msg) {
      console.log('Done! ok:', msg.ok)
      return
    }

    const op = msg as { progress: number }
    console.log({ data: e.data, op })
  }

  es.onerror = (e) => {
    console.error('Failed to connect to event source', e)
    es.close()
  }

  return es
}
