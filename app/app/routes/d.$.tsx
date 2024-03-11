import type { MetaFunction } from '@remix-run/node'
import type {
  ClientActionFunctionArgs,
  ClientLoaderFunctionArgs,
} from '@remix-run/react'
import { Form, Link, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useMemo, useState } from 'react'

import { connectOperationsEventSource, getDir, remove } from '../api'

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
    const operation = await remove(path)

    return { operation }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

const Progress = (props: { id: number }) => {
  const { id } = props
  const [progress, setProgress] = useState(0)
  const es = useMemo(() => connectOperationsEventSource(id), [id])

  useEffect(() => {
    es.onmessage = (e) => {
      console.log(e.data)
      const msg = JSON.parse(e.data)

      if ('ok' in msg) {
        console.log('Done! ok:', msg.ok)
        return
      }

      const op = msg as { progress: number }
      setProgress(op.progress)
      console.log({ data: e.data, op })
    }

    es.onopen = () => {
      console.log(`oppened connection with ${id}`)
    }

    // return () => {
    //   console.log(`closing connection with ${id}`)
    //   onClose()
    //   es.close()
    // }
  }, [es, id])

  return (
    <>
      <span>
        [{props.id}]:
        <progress value={progress} max="100" />
      </span>
    </>
  )
}

export default function Dir() {
  const { filesInfos, error, path } = useLoaderData<typeof clientLoader>()
  const actionData = useActionData<typeof clientAction>()
  const { operation, error: actionError } = actionData || {}
  if (error || actionError) {
    return <div>{error || actionError}</div>
  }
  const pwd = `/d${path}`
  const breadcrumbs = path?.split('/').slice(0, -1)
  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', lineHeight: '1.8' }}>
      <h1>git-drive</h1>
      <div>
        {breadcrumbs?.map((dir, i, arr) => {
          const name = dir || 'Home'
          const breadcrumbKey = `${name.replace(/\s/g, '_')}_${i}`
          return (
            <>
              <Link
                to={`/d${arr.slice(0, i + 1).join('/')}/`}
                key={breadcrumbKey}
              >
                {name}
              </Link>
              {' / '}
            </>
          )
        })}
      </div>

      {filesInfos?.map((info, i) => {
        const fileKey = `${info.name.replace(/\s/g, '_')}_${i}`
        const formKey = `${info.name.replace(/\s/g, '_')}-form`
        return (
          <>
            <Form method="post" id={`form_${i}`} action={pwd} key={formKey}>
              <input type="hidden" name="path" value={path + info.name} />
            </Form>
            <div key={fileKey}>
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

      {operation && <Progress id={operation.id} key={operation.id} />}
    </div>
  )
}
