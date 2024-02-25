import type { MetaFunction } from '@remix-run/node'
import type { ClientLoaderFunctionArgs } from '@remix-run/react'
import { Link, useLoaderData } from '@remix-run/react'

export const meta: MetaFunction = () => {
  return [
    { title: 'git-drive' },
    { name: 'description', content: 'Welcome to Remix (SPA Mode)!' },
  ]
}

const getDir = async (path: string) => {
  const response = await fetch(`http://localhost:8080/_api/dir${path}`)
  if (!response.ok) {
    response.text().then(console.error)
    throw new Error(`Failed to get directory listing for ${path}`)
  }
  return (await response.json()) as { name: string }[]
}

export const clientLoader = async ({ params }: ClientLoaderFunctionArgs) => {
  const path = params['*'] ? '/' + params['*'] : '/'

  try {
    const files = await getDir(path)
    return { files }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export default function Dir() {
  const { files, error } = useLoaderData<typeof clientLoader>()
  if (error) {
    return <div>{error}</div>
  }
  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', lineHeight: '1.8' }}>
      <h1>git-drive</h1>
      {files?.map((file) => {
        return (
          <div key={file.name}>
            <Link to={`${file.name}`}>{file.name}</Link>
          </div>
        )
      })}
    </div>
  )
}
