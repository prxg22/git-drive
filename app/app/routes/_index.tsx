import { useNavigate } from '@remix-run/react'
import { useEffect } from 'react'

export default function Index() {
  const navigate = useNavigate()
  console.log('navigating to /d')
  useEffect(() => {
    navigate('/d')
  }, [])
  return null
}
