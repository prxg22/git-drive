import type { MetaFunction } from '@remix-run/node'
import type {
  ClientActionFunctionArgs,
  ClientLoaderFunctionArgs,
} from '@remix-run/react'
import { Form, Link, useActionData, useLoaderData } from '@remix-run/react'
import React, { useEffect, useState } from 'react'

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
    const parts = path.split('/')
    return {
      operation: {
        ...operation,
        path: parts[parts.length - 1],
      },
    }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

const Progress = (props: { id: number; path: string; op: number }) => {
  const { id } = props
  const [progress, setProgress] = useState(0)

  useEffect(() => {
    const es = connectOperationsEventSource(id)
    const close = () => {
      console.log('Connection closed')
      es.close()
    }
    es.onmessage = (e) => {
      if (e.data === 'close') {
        return close()
      }

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
  }, [id])

  return (
    <>
      <span>
        [{props.id} - {props.path} - {props.op}]:
        <progress value={progress} max="100" />
      </span>
    </>
  )
}

export default function Dir() {
  const [ops, setOps] = useState<{ id: number; path: string; op: number }[]>([])
  const { filesInfos, error, path } = useLoaderData<typeof clientLoader>()
  const actionData = useActionData<typeof clientAction>()
  const { operation, error: actionError } = actionData || {}
  if (error || actionError) {
    return <div>{error || actionError}</div>
  }
  const pwd = `/d${path}`
  const breadcrumbs = path?.split('/').slice(0, -1)

  if (operation && !ops.some((o) => o.id === operation.id)) {
    setOps((old) => {
      return [...old, operation]
    })
  }

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', lineHeight: '1.8' }}>
      <h1>git-drive</h1>
      <div>
        {breadcrumbs?.map((dir, i, arr) => {
          const name = dir || 'Home'
          const breadcrumbKey = `${name.replace(/\s/g, '_')}_${i}`
          return (
            <span key={breadcrumbKey}>
              <Link to={`/d${arr.slice(0, i + 1).join('/')}/`}>{name}</Link>
            </span>
          )
        })}
      </div>

      {filesInfos?.map((info, i) => {
        const key = `${info.name.replace(/\s/g, '_')}_${i}`
        return (
          <React.Fragment key={key}>
            <Form method="post" id={`form_${i}`} action={pwd}>
              <input type="hidden" name="path" value={path + info.name} />
            </Form>
            <div>
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
          </React.Fragment>
        )
      })}

      {ops.length > 0 &&
        ops.map((op) => (
          <div key={op.id}>
            <Progress {...op} />
          </div>
        ))}
    </div>
  )
}
