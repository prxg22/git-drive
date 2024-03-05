export const getDir = async (path: string) => {
  const response = await fetch(`http://localhost:8080/_api/dir${path}`)
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
  const response = await fetch(`http://localhost:8080/_api${path}`, {
    method: 'DELETE',
  })

  return response.ok
}
