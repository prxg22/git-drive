import type { MetaFunction } from '@remix-run/node'
import type {
  ClientActionFunctionArgs,
  ClientLoaderFunctionArgs,
} from '@remix-run/react'
import { Form, Link, useLoaderData } from '@remix-run/react'

import { getDir, remove } from '../api'

export const meta: MetaFunction = () => {
  return [
    { title: 'git-drive' },
    { name: 'description', content: 'Welcome to Remix (SPA Mode)!' },
  ]
}

export const clientLoader = async ({ params }: ClientLoaderFunctionArgs) => {
  const path = params['*'] ? '/' + params['*'] : '/'

  try {
    const filesInfos = await getDir(path)
    return { filesInfos, path }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export const clientAction = async ({ request }: ClientActionFunctionArgs) => {
  const formData = await request.formData()

  const path = formData.get('path')?.toString()
  if (!path) return { error: 'No path provided' }

  try {
    const ok = remove(path)
    if (!ok) {
      return { error: 'Failed to remove file' }
    }

    return { deleted: { path } }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export default function Dir() {
  const { filesInfos, error, path } = useLoaderData<typeof clientLoader>()
  if (error) {
    return <div>{error}</div>
  }
  const pwd = `/d${path}`
  const breadcrumbs = path?.split('/').slice(0, -1)
  console.log({ path, pwd, breadcrumbs, filesInfos })
  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', lineHeight: '1.8' }}>
      <h1>git-drive</h1>
      <div>
        {breadcrumbs?.map((dir, i, arr) => {
          return (
            <>
              <Link to={`/d${arr.slice(0, i + 1).join('/')}/`} key={dir}>
                {dir || 'Home'}
              </Link>
              {' / '}
            </>
          )
        })}
      </div>

      {filesInfos?.map((info, i) => {
        return (
          <>
            <Form method="post" id={`form_${i}`} action={pwd}>
              <input type="hidden" name="path" value={path + info.name} />
            </Form>
            <div key={info.name}>
              {info.isDir ? '📁' : '📄'}{' '}
              {info.isDir ? (
                <Link to={`${pwd}${info.name}/`}>{info.name}</Link>
              ) : (
                <>{info.name}</>
              )}
              <button type="submit" form={`form_${i}`}>
                remove
              </button>
            </div>
          </>
        )
      })}
    </div>
  )
}
