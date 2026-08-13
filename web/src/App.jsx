import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  BrowserRouter,
  Link,
  Navigate,
  NavLink,
  Outlet,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
} from 'react-router-dom'

const initialLoginForm = {
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

function createAdminForm() {
  return {
    id: '',
    username: '',
    password: '',
    realName: '',
    roleId: '0',
    status: '1',
    remark: '',
  }
}

function createRoleForm() {
  return {
    id: '',
    roleName: '',
  }
}

function createAdminRoleForm() {
  return {
    id: '',
    adminId: '',
    roleId: '',
  }
}

function createPermissionForm() {
  return {
    id: '',
    parentId: '0',
    title: '',
    handle: '',
    weight: '0',
  }
}

function createGroupPermissionForm() {
  return {
    id: '',
    groupId: '',
    permissionId: '',
  }
}

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

async function readJson(response) {
  return response.json().catch(() => ({}))
}

async function requestJson(url, options = {}) {
  const response = await fetch(url, options)
  const payload = await readJson(response)
  if (!response.ok) {
    throw new Error(payload.message || '请求失败，请稍后重试。')
  }
  return payload
}

function buildAuthHeaders() {
  const auth = readStoredAuth()
  const headers = {
    'Content-Type': 'application/json',
  }
  if (auth.accessToken) {
    headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return headers
}

function pickValue(source, keys, fallback = '') {
  for (const key of keys) {
    if (source?.[key] !== undefined && source?.[key] !== null) {
      return source[key]
    }
  }
  return fallback
}

function formatDateTime(value) {
  if (!value) {
    return '-'
  }
  if (typeof value === 'object') {
    const seconds = Number(value.seconds || 0)
    const nanos = Number(value.nanos || 0)
    const timestamp = seconds * 1000 + Math.floor(nanos / 1e6)
    if (timestamp > 0) {
      return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
    }
    return '-'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return String(value)
  }
  return date.toLocaleString('zh-CN', { hour12: false })
}

function statusLabel(status) {
  switch (Number(status)) {
    case 0:
      return '禁用'
    case 1:
      return '正常'
    case 2:
      return '锁定'
    default:
      return `状态 ${status}`
  }
}

function buildUpdateMaskQuery(paths) {
  const searchParams = new URLSearchParams()
  paths.forEach((path) => {
    searchParams.append('updateMask.paths', path)
  })
  return searchParams.toString()
}

function normalizeAdminResponse(admin) {
  if (!admin) {
    return null
  }
  return {
    id: Number(pickValue(admin, ['id'], 0)),
    username: pickValue(admin, ['username']),
    realName: pickValue(admin, ['realName', 'real_name']),
    roleId: Number(pickValue(admin, ['roleId', 'role_id'], 0)),
    status: Number(pickValue(admin, ['status'], 0)),
    remark: pickValue(admin, ['remark']),
    lastLoginTime: pickValue(admin, ['lastLoginTime', 'last_login_time']),
    passwordUpdatedAt: pickValue(admin, ['passwordUpdatedAt', 'password_updated_at']),
    createTime: pickValue(admin, ['createTime', 'create_time']),
    updateTime: pickValue(admin, ['updateTime', 'update_time']),
  }
}

function normalizeAdminForm(admin) {
  const normalized = normalizeAdminResponse(admin)
  if (!normalized) {
    return createAdminForm()
  }
  return {
    id: String(normalized.id || ''),
    username: normalized.username,
    password: '',
    realName: normalized.realName,
    roleId: String(normalized.roleId ?? 0),
    status: String(normalized.status ?? 1),
    remark: normalized.remark,
  }
}

function buildAdminPayload(form) {
  return {
    id: Number(form.id) || 0,
    username: form.username.trim(),
    password: form.password.trim(),
    real_name: form.realName.trim(),
    role_id: Number(form.roleId) || 0,
    status: Number(form.status) || 0,
    remark: form.remark.trim(),
  }
}

function normalizeRoleResponse(role) {
  if (!role) {
    return null
  }
  return {
    id: Number(pickValue(role, ['id'], 0)),
    roleName: pickValue(role, ['roleName', 'role_name']),
    createTime: pickValue(role, ['createTime', 'create_time']),
    updateTime: pickValue(role, ['updateTime', 'update_time']),
  }
}

function normalizeRoleForm(role) {
  const normalized = normalizeRoleResponse(role)
  if (!normalized) {
    return createRoleForm()
  }
  return {
    id: String(normalized.id || ''),
    roleName: normalized.roleName,
  }
}

function buildRolePayload(form) {
  return {
    id: Number(form.id) || 0,
    role_name: form.roleName.trim(),
  }
}

function normalizeAdminRoleResponse(item) {
  if (!item) {
    return null
  }
  return {
    id: Number(pickValue(item, ['id'], 0)),
    adminId: Number(pickValue(item, ['adminId', 'admin_id'], 0)),
    roleId: Number(pickValue(item, ['roleId', 'role_id'], 0)),
    createTime: pickValue(item, ['createTime', 'create_time']),
    updateTime: pickValue(item, ['updateTime', 'update_time']),
  }
}

function normalizeAdminRoleForm(item) {
  const normalized = normalizeAdminRoleResponse(item)
  if (!normalized) {
    return createAdminRoleForm()
  }
  return {
    id: String(normalized.id || ''),
    adminId: String(normalized.adminId || ''),
    roleId: String(normalized.roleId || ''),
  }
}

function buildAdminRolePayload(form) {
  return {
    id: Number(form.id) || 0,
    admin_id: Number(form.adminId) || 0,
    role_id: Number(form.roleId) || 0,
  }
}

function normalizePermissionResponse(permission) {
  if (!permission) {
    return null
  }
  return {
    id: Number(pickValue(permission, ['id'], 0)),
    parentId: Number(pickValue(permission, ['parentId', 'parent_id'], 0)),
    title: pickValue(permission, ['title']),
    handle: pickValue(permission, ['handle']),
    weight: Number(pickValue(permission, ['weight'], 0)),
    createTime: pickValue(permission, ['createTime', 'create_time']),
    updateTime: pickValue(permission, ['updateTime', 'update_time']),
  }
}

function normalizePermissionForm(permission) {
  const normalized = normalizePermissionResponse(permission)
  if (!normalized) {
    return createPermissionForm()
  }
  return {
    id: String(normalized.id || ''),
    parentId: String(normalized.parentId ?? 0),
    title: normalized.title,
    handle: normalized.handle,
    weight: String(normalized.weight ?? 0),
  }
}

function buildPermissionPayload(form) {
  return {
    id: Number(form.id) || 0,
    parent_id: Number(form.parentId) || 0,
    title: form.title.trim(),
    handle: form.handle.trim(),
    weight: Number(form.weight) || 0,
  }
}

function normalizeGroupPermissionResponse(item) {
  if (!item) {
    return null
  }
  return {
    id: Number(pickValue(item, ['id'], 0)),
    groupId: Number(pickValue(item, ['groupId', 'group_id'], 0)),
    permissionId: Number(pickValue(item, ['permissionId', 'permission_id'], 0)),
    createTime: pickValue(item, ['createTime', 'create_time']),
    updateTime: pickValue(item, ['updateTime', 'update_time']),
  }
}

function normalizeGroupPermissionForm(item) {
  const normalized = normalizeGroupPermissionResponse(item)
  if (!normalized) {
    return createGroupPermissionForm()
  }
  return {
    id: String(normalized.id || ''),
    groupId: String(normalized.groupId || ''),
    permissionId: String(normalized.permissionId || ''),
  }
}

function buildGroupPermissionPayload(form) {
  return {
    id: Number(form.id) || 0,
    group_id: Number(form.groupId) || 0,
    permission_id: Number(form.permissionId) || 0,
  }
}

const resourceCatalog = [
  {
    key: 'admins',
    section: '系统管理',
    navLabel: '管理员',
    singularLabel: '管理员',
    pluralLabel: '管理员',
    basePath: '/admins',
    listEndpoint: '/api/v1/admins?page_size=100',
    createEndpoint: '/api/v1/admins/create',
    updateEndpoint: '/api/v1/admins/update',
    getEndpoint: (id) => `/api/v1/admins/${id}`,
    deleteEndpoint: (id) => `/api/v1/admins/${id}`,
    listKey: ['admins'],
    subtitle: '管理后台登录账号、状态与角色关联。',
    createDescription: '新建管理员时必须填写初始密码，编辑时留空密码表示不修改。',
    editDescription: '在独立页面中编辑管理员资料，保存后返回列表。',
    createForm: createAdminForm,
    normalizeResponse: normalizeAdminResponse,
    normalizeForm: normalizeAdminForm,
    buildPayload: buildAdminPayload,
    buildUpdateMaskPaths: (_, payload) => {
      const paths = ['username', 'real_name', 'role_id', 'status', 'remark']
      if (payload.password) {
        paths.push('password')
      }
      return paths
    },
    canSubmit: (form, isEditing) => form.username.trim() !== '' && (isEditing || form.password.trim() !== ''),
    describe: (payload) => payload.username || '管理员',
    fields: [
      { name: 'username', label: '用户名', type: 'text', placeholder: '请输入登录用户名' },
      {
        name: 'password',
        label: (isEditing) => (isEditing ? '新密码（可选）' : '密码'),
        type: 'password',
        placeholder: (isEditing) => (isEditing ? '不修改密码可留空' : '请输入至少 6 位密码'),
      },
      { name: 'realName', label: '真实姓名', type: 'text', placeholder: '请输入真实姓名' },
      { name: 'roleId', label: '角色 ID', type: 'number', placeholder: '请输入角色 ID' },
      {
        name: 'status',
        label: '状态',
        type: 'select',
        options: [
          { value: '1', label: '正常' },
          { value: '0', label: '禁用' },
          { value: '2', label: '锁定' },
        ],
      },
      { name: 'remark', label: '备注', type: 'textarea', placeholder: '输入备注信息', rows: 4 },
    ],
    getCardTitle: (item) => item.realName || item.username,
    getCardBadge: (item) => ({
      text: statusLabel(item.status),
      className: `status-chip status-chip--${Number(item.status)}`,
    }),
    getCardMeta: (item) => [
      ['用户名', item.username || '-'],
      ['角色 ID', String(item.roleId ?? 0)],
      ['备注', item.remark || '-'],
      ['最后登录', formatDateTime(item.lastLoginTime)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
  {
    key: 'roles',
    section: '系统管理',
    navLabel: '角色',
    singularLabel: '角色',
    pluralLabel: '角色',
    basePath: '/roles',
    listEndpoint: '/api/v1/roles?page_size=100',
    createEndpoint: '/api/v1/roles/create',
    updateEndpoint: '/api/v1/roles/update',
    getEndpoint: (id) => `/api/v1/roles/${id}`,
    deleteEndpoint: (id) => `/api/v1/roles/${id}`,
    listKey: ['roles'],
    subtitle: '维护后台角色名称。',
    createDescription: '角色编辑页和列表页分离，便于后续扩展更多角色字段。',
    editDescription: '修改角色名称后返回列表查看结果。',
    createForm: createRoleForm,
    normalizeResponse: normalizeRoleResponse,
    normalizeForm: normalizeRoleForm,
    buildPayload: buildRolePayload,
    buildUpdateMaskPaths: () => ['role_name'],
    canSubmit: (form) => form.roleName.trim() !== '',
    describe: (payload) => payload.role_name || '角色',
    fields: [{ name: 'roleName', label: '角色名称', type: 'text', placeholder: '请输入角色名称' }],
    getCardTitle: (item) => item.roleName || `角色 #${item.id}`,
    getCardMeta: (item) => [
      ['角色 ID', String(item.id)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
  {
    key: 'admin-roles',
    section: '关联管理',
    navLabel: '管理员角色',
    singularLabel: '管理员角色关联',
    pluralLabel: '管理员角色关联',
    basePath: '/admin-roles',
    listEndpoint: '/api/v1/admin-roles?page_size=100',
    createEndpoint: '/api/v1/admin-roles/create',
    updateEndpoint: '/api/v1/admin-roles/update',
    getEndpoint: (id) => `/api/v1/admin-roles/${id}`,
    deleteEndpoint: (id) => `/api/v1/admin-roles/${id}`,
    listKey: ['adminRoles', 'admin_roles'],
    subtitle: '维护管理员和角色之间的关联关系。',
    createDescription: '这是关联表，列表页只负责查看和跳转，编辑在独立页面完成。',
    editDescription: '修改管理员 ID 或角色 ID 后保存。',
    createForm: createAdminRoleForm,
    normalizeResponse: normalizeAdminRoleResponse,
    normalizeForm: normalizeAdminRoleForm,
    buildPayload: buildAdminRolePayload,
    buildUpdateMaskPaths: () => ['admin_id', 'role_id'],
    canSubmit: (form) => form.adminId.trim() !== '' && form.roleId.trim() !== '',
    describe: (payload) => `管理员 ${payload.admin_id} - 角色 ${payload.role_id}`,
    fields: [
      { name: 'adminId', label: '管理员 ID', type: 'number', placeholder: '请输入管理员 ID' },
      { name: 'roleId', label: '角色 ID', type: 'number', placeholder: '请输入角色 ID' },
    ],
    getCardTitle: (item) => `管理员 #${item.adminId} -> 角色 #${item.roleId}`,
    getCardMeta: (item) => [
      ['关联 ID', String(item.id)],
      ['管理员 ID', String(item.adminId)],
      ['角色 ID', String(item.roleId)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
  {
    key: 'permissions',
    section: '系统管理',
    navLabel: '权限',
    singularLabel: '权限',
    pluralLabel: '权限',
    basePath: '/permissions',
    listEndpoint: '/api/v1/permissions?page_size=100',
    createEndpoint: '/api/v1/permissions/create',
    updateEndpoint: '/api/v1/permissions/update',
    getEndpoint: (id) => `/api/v1/permissions/${id}`,
    deleteEndpoint: (id) => `/api/v1/permissions/${id}`,
    listKey: ['permissions'],
    subtitle: '维护权限树节点、路由句柄和排序权重。',
    createDescription: '支持独立新增和编辑，避免和列表页混在一起。',
    editDescription: '可修改父节点、标题、路由和权重。',
    createForm: createPermissionForm,
    normalizeResponse: normalizePermissionResponse,
    normalizeForm: normalizePermissionForm,
    buildPayload: buildPermissionPayload,
    buildUpdateMaskPaths: () => ['parent_id', 'title', 'handle', 'weight'],
    canSubmit: (form) => form.title.trim() !== '' && form.handle.trim() !== '',
    describe: (payload) => payload.title || '权限',
    fields: [
      { name: 'parentId', label: '父级 ID', type: 'number', placeholder: '根节点填 0' },
      { name: 'title', label: '权限名称', type: 'text', placeholder: '请输入权限名称' },
      { name: 'handle', label: '路由句柄', type: 'text', placeholder: '例如 /system/users' },
      { name: 'weight', label: '权重', type: 'number', placeholder: '请输入排序权重' },
    ],
    getCardTitle: (item) => item.title || `权限 #${item.id}`,
    getCardMeta: (item) => [
      ['权限 ID', String(item.id)],
      ['父级 ID', String(item.parentId)],
      ['路由', item.handle || '-'],
      ['权重', String(item.weight)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
  {
    key: 'group-permissions',
    section: '关联管理',
    navLabel: '分组权限',
    singularLabel: '分组权限关联',
    pluralLabel: '分组权限关联',
    basePath: '/group-permissions',
    listEndpoint: '/api/v1/group-permissions?page_size=100',
    createEndpoint: '/api/v1/group-permissions/create',
    updateEndpoint: '/api/v1/group-permissions/update',
    getEndpoint: (id) => `/api/v1/group-permissions/${id}`,
    deleteEndpoint: (id) => `/api/v1/group-permissions/${id}`,
    listKey: ['groupPermissions', 'group_permissions'],
    subtitle: '维护分组和权限之间的关联关系。',
    createDescription: '关联数据同样使用“列表页 + 独立编辑页”结构。',
    editDescription: '修改分组 ID 或权限 ID 后保存。',
    createForm: createGroupPermissionForm,
    normalizeResponse: normalizeGroupPermissionResponse,
    normalizeForm: normalizeGroupPermissionForm,
    buildPayload: buildGroupPermissionPayload,
    buildUpdateMaskPaths: () => ['group_id', 'permission_id'],
    canSubmit: (form) => form.groupId.trim() !== '' && form.permissionId.trim() !== '',
    describe: (payload) => `分组 ${payload.group_id} - 权限 ${payload.permission_id}`,
    fields: [
      { name: 'groupId', label: '分组 ID', type: 'number', placeholder: '请输入分组 ID' },
      { name: 'permissionId', label: '权限 ID', type: 'number', placeholder: '请输入权限 ID' },
    ],
    getCardTitle: (item) => `分组 #${item.groupId} -> 权限 #${item.permissionId}`,
    getCardMeta: (item) => [
      ['关联 ID', String(item.id)],
      ['分组 ID', String(item.groupId)],
      ['权限 ID', String(item.permissionId)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
]

function getListItems(payload, keys) {
  for (const key of keys) {
    const items = payload?.[key]
    if (Array.isArray(items)) {
      return items
    }
  }
  return []
}

function getCurrentResource(pathname) {
  return resourceCatalog.find(
    (resource) => pathname === resource.basePath || pathname.startsWith(`${resource.basePath}/`),
  )
}

function LoginPage() {
  const navigate = useNavigate()
  const [form, setForm] = useState(initialLoginForm)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const isLoggedIn = useMemo(() => Boolean(readStoredAuth().accessToken), [])
  const canSubmit = useMemo(() => form.account.trim() !== '' && form.password.trim() !== '', [form.account, form.password])

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
      const payload = await requestJson('/api/v1/login', {
        method: 'POST',
        headers: buildAuthHeaders(),
        body: JSON.stringify({
          account: form.account.trim(),
          password: form.password,
        }),
      })

      saveAuth({
        accessToken: pickValue(payload, ['accessToken', 'access_token']),
        refreshToken: pickValue(payload, ['refreshToken', 'refresh_token']),
        tokenType: pickValue(payload, ['tokenType', 'token_type']),
        expiresIn: pickValue(payload, ['expiresIn', 'expires_in']),
        userId: pickValue(payload, ['userId', 'user_id']),
      })
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
          <p className="subtitle">输入账号和密码，进入后台管理界面。</p>
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

function RequireAuth() {
  const auth = readStoredAuth()
  if (!auth.accessToken) {
    return <Navigate replace to="/" />
  }
  return <Outlet />
}

function ConsoleLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const auth = readStoredAuth()
  const currentResource = getCurrentResource(location.pathname)
  const breadcrumbTail = location.pathname.endsWith('/new')
    ? `新建${currentResource?.singularLabel || ''}`
    : location.pathname.endsWith('/edit')
      ? `编辑${currentResource?.singularLabel || ''}`
      : `${currentResource?.pluralLabel || '管理'}列表`
  const navSections = [
    { title: '系统管理', items: resourceCatalog.filter((resource) => resource.section === '系统管理') },
    { title: '关联管理', items: resourceCatalog.filter((resource) => resource.section === '关联管理') },
  ]

  const handleLogout = () => {
    clearAuth()
    navigate('/', { replace: true })
  }

  return (
    <main className="console-page">
      <aside className="sidebar">
        <div className="sidebar__brand">
          <p className="eyebrow">Crow</p>
          <h1>运营管理中心</h1>
          <p>后台资源维护</p>
        </div>

        {navSections.map((section) => (
          <section className="sidebar__section" key={section.title}>
            <div className="sidebar__section-title">{section.title}</div>

            <nav className="nav-list">
              {section.items.map((resource) => (
                <NavLink
                  className={({ isActive }) =>
                    `nav-link${isActive || location.pathname.startsWith(`${resource.basePath}/`) ? ' nav-link--active' : ''}`
                  }
                  key={resource.key}
                  to={resource.basePath}
                >
                  <span>{resource.navLabel}</span>
                  <small>{resource.singularLabel}</small>
                </NavLink>
              ))}
            </nav>
          </section>
        ))}

        <section className="result result--standalone sidebar__session">
          <div className="result__title">
            <h3>当前会话</h3>
            <p className="subtitle">登录返回的令牌信息。</p>
          </div>
          <dl>
            <div>
              <dt>accessToken</dt>
              <dd>{auth.accessToken || '-'}</dd>
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
              <dd>{String(auth.expiresIn || '-')}</dd>
            </div>
            <div>
              <dt>userId</dt>
              <dd>{String(auth.userId || '-')}</dd>
            </div>
          </dl>
        </section>
      </aside>

      <section className="workspace">
        <header className="workspace__topbar">
          <div className="workspace__header">
            <p className="workspace__crumbs">工作台 / {currentResource?.navLabel || '管理中心'} / {breadcrumbTail}</p>
            <h2>{breadcrumbTail}</h2>
            <p className="subtitle">{currentResource?.subtitle || '统一维护后台资源数据。'}</p>
          </div>

          <div className="hero__actions">
            <button className="submit submit--compact" onClick={handleLogout} type="button">
              退出登录
            </button>
          </div>
        </header>

        <section className="content-stack">
            <Outlet />
        </section>
      </section>
    </main>
  )
}

function ResourceListPage({ resource }) {
  const location = useLocation()
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(location.state?.message || '')

  const loadItems = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const payload = await requestJson(resource.listEndpoint, {
        headers: buildAuthHeaders(),
      })
      setItems(getListItems(payload, resource.listKey).map(resource.normalizeResponse).filter(Boolean))
    } catch (loadError) {
      setError(loadError.message || `${resource.pluralLabel}列表加载失败。`)
    } finally {
      setLoading(false)
    }
  }, [resource])

  useEffect(() => {
    setSuccess(location.state?.message || '')
  }, [location.state])

  useEffect(() => {
    loadItems()
  }, [loadItems])

  const handleDelete = async (item) => {
    if (!window.confirm(`确定删除「${resource.describe(resource.buildPayload(resource.normalizeForm(item)))}」吗？`)) {
      return
    }
    setError('')
    setSuccess('')
    try {
      await requestJson(resource.deleteEndpoint(item.id), {
        method: 'DELETE',
        headers: buildAuthHeaders(),
      })
      setSuccess(`${resource.singularLabel}已删除。`)
      await loadItems()
    } catch (deleteError) {
      setError(deleteError.message || `删除${resource.singularLabel}失败。`)
    }
  }

  return (
    <section className="panel">
      <div className="panel__header">
        <div className="card__header">
          <h2>{resource.pluralLabel}列表</h2>
          <p className="subtitle">{resource.subtitle}</p>
        </div>

        <div className="panel__actions">
          <button className="ghost-button" onClick={loadItems} type="button">
            {loading ? '刷新中...' : '刷新列表'}
          </button>
          <Link className="submit submit--compact action-link" to={`${resource.basePath}/new`}>
            新建{resource.singularLabel}
          </Link>
        </div>
      </div>

      {error && <p className="message message--error">{error}</p>}
      {success && <p className="message message--success">{success}</p>}

      <div className="resource-list">
        {items.length === 0 && !loading ? (
          <p className="empty-state">暂无{resource.pluralLabel}数据。</p>
        ) : (
          items.map((item) => {
            const badge = resource.getCardBadge ? resource.getCardBadge(item) : null
            return (
              <article className="resource-item" key={item.id}>
                <div className="resource-item__main">
                  <div className="resource-item__title">
                    <h3>{resource.getCardTitle(item)}</h3>
                    {badge ? <span className={badge.className}>{badge.text}</span> : null}
                  </div>

                  <dl className="meta-grid">
                    {resource.getCardMeta(item).map(([label, value]) => (
                      <div key={label}>
                        <dt>{label}</dt>
                        <dd>{value || '-'}</dd>
                      </div>
                    ))}
                  </dl>
                </div>

                <div className="resource-item__actions">
                  <Link className="ghost-button action-link" to={`${resource.basePath}/${item.id}/edit`}>
                    编辑
                  </Link>
                  <button className="danger-button" onClick={() => handleDelete(item)} type="button">
                    删除
                  </button>
                </div>
              </article>
            )
          })
        )}
      </div>
    </section>
  )
}

function ResourceFormPage({ resource, mode }) {
  const navigate = useNavigate()
  const params = useParams()
  const itemId = params.id || ''
  const isEditing = mode === 'edit'
  const [form, setForm] = useState(() => resource.createForm())
  const [loading, setLoading] = useState(isEditing)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function loadItem() {
      if (!isEditing) {
        setForm(resource.createForm())
        setLoading(false)
        return
      }

      setLoading(true)
      setError('')
      try {
        const payload = await requestJson(resource.getEndpoint(itemId), {
          headers: buildAuthHeaders(),
        })
        if (!cancelled) {
          setForm(resource.normalizeForm(payload))
        }
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError.message || `${resource.singularLabel}详情加载失败。`)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadItem()

    return () => {
      cancelled = true
    }
  }, [isEditing, itemId, resource])

  const canSubmit = resource.canSubmit(form, isEditing)

  const handleChange = (event) => {
    const { name, value } = event.target
    setForm((current) => ({
      ...current,
      [name]: value,
    }))
  }

  const handleSubmit = async (event) => {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    const payload = resource.buildPayload(form)

    try {
      if (isEditing) {
        const query = buildUpdateMaskQuery(resource.buildUpdateMaskPaths(form, payload))
        await requestJson(`${resource.updateEndpoint}?${query}`, {
          method: 'PUT',
          headers: buildAuthHeaders(),
          body: JSON.stringify(payload),
        })
      } else {
        await requestJson(resource.createEndpoint, {
          method: 'POST',
          headers: buildAuthHeaders(),
          body: JSON.stringify(payload),
        })
      }

      navigate(resource.basePath, {
        replace: true,
        state: {
          message: `${resource.singularLabel}「${resource.describe(payload)}」${isEditing ? '已更新' : '已创建'}。`,
        },
      })
    } catch (submitError) {
      setError(submitError.message || `保存${resource.singularLabel}失败。`)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section className="panel">
      <div className="panel__header">
        <div className="card__header">
          <h2>{isEditing ? `编辑${resource.singularLabel}` : `新建${resource.singularLabel}`}</h2>
          <p className="subtitle">{isEditing ? resource.editDescription : resource.createDescription}</p>
        </div>

        <div className="panel__actions">
          <Link className="ghost-button action-link" to={resource.basePath}>
            返回列表
          </Link>
        </div>
      </div>

      {error && <p className="message message--error">{error}</p>}

      {loading ? (
        <p className="empty-state">加载中...</p>
      ) : (
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-grid">
            {resource.fields.map((field) => {
              const label = typeof field.label === 'function' ? field.label(isEditing) : field.label
              const placeholder = typeof field.placeholder === 'function' ? field.placeholder(isEditing) : field.placeholder
              const className = field.type === 'textarea' ? 'field field--full' : 'field'

              return (
                <label className={className} key={field.name}>
                  <span>{label}</span>

                  {field.type === 'textarea' ? (
                    <textarea
                      name={field.name}
                      onChange={handleChange}
                      placeholder={placeholder}
                      rows={field.rows || 4}
                      value={form[field.name]}
                    />
                  ) : field.type === 'select' ? (
                    <select name={field.name} onChange={handleChange} value={form[field.name]}>
                      {field.options.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      name={field.name}
                      onChange={handleChange}
                      placeholder={placeholder}
                      type={field.type}
                      value={form[field.name]}
                    />
                  )}
                </label>
              )
            })}
          </div>

          <div className="button-row">
            <button className="submit submit--compact" disabled={!canSubmit || submitting} type="submit">
              {submitting ? '提交中...' : isEditing ? '保存修改' : '确认创建'}
            </button>
            <Link className="ghost-button action-link" to={resource.basePath}>
              取消
            </Link>
          </div>
        </form>
      )}
    </section>
  )
}

function HomeRedirect() {
  const auth = readStoredAuth()
  if (!auth.accessToken) {
    return <Navigate replace to="/" />
  }
  return <Navigate replace to="/admins" />
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<LoginPage />} path="/" />
        <Route element={<HomeRedirect />} path="/home" />

        <Route element={<RequireAuth />} path="/">
          <Route element={<ConsoleLayout />}>
            {resourceCatalog.map((resource) => (
              <Route key={resource.key}>
                <Route element={<ResourceListPage resource={resource} />} path={resource.basePath} />
                <Route element={<ResourceFormPage mode="new" resource={resource} />} path={`${resource.basePath}/new`} />
                <Route
                  element={<ResourceFormPage mode="edit" resource={resource} />}
                  path={`${resource.basePath}/:id/edit`}
                />
              </Route>
            ))}
          </Route>
        </Route>

        <Route element={<Navigate replace to="/" />} path="*" />
      </Routes>
    </BrowserRouter>
  )
}

export default App
