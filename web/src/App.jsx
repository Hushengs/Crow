import { BrowserRouter, Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import { useMemo, useState } from 'react'

const initialForm = {
  account: '',
  password: '',
}

const initialAuth = {
  accessToken: '',
  refreshToken: '',
  tokenType: '',
  expiresIn: '',
  userId: '',
}

const authStorageKey = 'crow.auth'

function readStoredAuth() {
  const raw = window.sessionStorage.getItem(authStorageKey)
  if (!raw) {
    return initialAuth
  }

  try {
    return {
      ...initialAuth,
      ...JSON.parse(raw),
    }
  } catch {
    window.sessionStorage.removeItem(authStorageKey)
    return initialAuth
  }
}

function saveAuth(auth) {
  window.sessionStorage.setItem(authStorageKey, JSON.stringify(auth))
}

function clearAuth() {
  window.sessionStorage.removeItem(authStorageKey)
}

function LoginPage() {
  const navigate = useNavigate()
  const [form, setForm] = useState(initialForm)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const isLoggedIn = useMemo(() => {
    return Boolean(readStoredAuth().accessToken)
  }, [])

  const canSubmit = useMemo(() => {
    return form.account.trim() !== '' && form.password.trim() !== ''
  }, [form.account, form.password])

  const handleChange = (event) => {
    const { name, value } = event.target
    setForm((current) => ({
      ...current,
      [name]: value,
    }))
  }

  const handleSubmit = async (event) => {
    event.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await fetch('/api/v1/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          account: form.account.trim(),
          password: form.password,
        }),
      })

      const payload = await response.json().catch(() => ({}))

      if (!response.ok) {
        throw new Error(payload.message || '登录失败，请检查账号或密码。')
      }

      const nextAuth = {
        accessToken: payload.accessToken || '',
        refreshToken: payload.refreshToken || '',
        tokenType: payload.tokenType || '',
        expiresIn: payload.expiresIn || '',
        userId: payload.userId || '',
      }
      saveAuth(nextAuth)
      navigate('/home', { replace: true })
    } catch (submitError) {
      setError(submitError.message || '请求失败，请稍后重试。')
    } finally {
      setLoading(false)
    }
  }

  if (isLoggedIn) {
    return <Navigate replace to="/home" />
  }

  return (
    <main className="page">
      <section className="card">
        <div className="card__header">
          <p className="eyebrow">Crow</p>
          <h1>登录系统</h1>
          <p className="subtitle">输入账号和密码，调用后端登录接口获取令牌。</p>
        </div>

        <form className="form" onSubmit={handleSubmit}>
          <label className="field">
            <span>账号</span>
            <input
              autoComplete="username"
              name="account"
              onChange={handleChange}
              placeholder="请输入账号"
              value={form.account}
            />
          </label>

          <label className="field">
            <span>密码</span>
            <input
              autoComplete="current-password"
              name="password"
              onChange={handleChange}
              placeholder="请输入密码"
              type="password"
              value={form.password}
            />
          </label>

          <button className="submit" disabled={!canSubmit || loading} type="submit">
            {loading ? '登录中...' : '立即登录'}
          </button>
        </form>

        {error && <p className="message message--error">{error}</p>}
      </section>
    </main>
  )
}

function HomePage() {
  const navigate = useNavigate()
  const auth = readStoredAuth()

  if (!auth.accessToken) {
    return <Navigate replace to="/" />
  }

  const handleLogout = () => {
    clearAuth()
    navigate('/', { replace: true })
  }

  return (
    <main className="page">
      <section className="card">
        <div className="card__header">
          <p className="eyebrow">Crow</p>
          <h1>首页</h1>
          <p className="subtitle">你已完成登录，下面是当前会话的令牌信息。</p>
        </div>

        <section className="result result--standalone">
          <h2>当前会话</h2>
          <dl>
            <div>
              <dt>accessToken</dt>
              <dd>{auth.accessToken}</dd>
            </div>
            <div>
              <dt>refreshToken</dt>
              <dd>{auth.refreshToken || '-'}</dd>
            </div>
            <div>
              <dt>tokenType</dt>
              <dd>{auth.tokenType || '-'}</dd>
            </div>
            <div>
              <dt>expiresIn</dt>
              <dd>{String(auth.expiresIn)}</dd>
            </div>
            <div>
              <dt>userId</dt>
              <dd>{String(auth.userId)}</dd>
            </div>
          </dl>
        </section>

        <button className="submit" onClick={handleLogout} type="button">
          退出登录
        </button>
      </section>
    </main>
  )
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<LoginPage />} path="/" />
        <Route element={<HomePage />} path="/home" />
      </Routes>
    </BrowserRouter>
  )
}

export default App
