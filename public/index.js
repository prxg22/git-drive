
const getDir = async (path) => {
  const response = await fetch(`/_api/dir${path}`)
  if (!response.ok) {
    throw new Error(`Failed to get directory listing for ${path}`)
  }
  return await response.json()
}


const div = document.createElement('div')
div.textContent = 'Loading...'
document.body.appendChild(div)

getDir('/').then(files => {
  div.textContent = ''
  files.forEach(f => {
    const container = document.createElement('div')
    const link = document.createElement('a')
    link.textContent = f.name
    link.href = f.name
    container.appendChild(link)
    div.appendChild(container)
  })
})