import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
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
  account: '',
}

const authStorageKey = 'crow.auth'

function createAdminForm() {
  return {
    id: '',
    username: '',
    password: '',
    realName: '',
    roleIds: [],
    status: '1',
    remark: '',
  }
}

function createRoleForm() {
  return {
    id: '',
    roleName: '',
    permissionIds: [],
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

function createCpForm() {
  return {
    id: '',
    cpCode: '',
    cpName: '',
    spIds: [],
    status: '1',
  }
}

function createSpForm() {
  return {
    id: '',
    spCode: '',
    spName: '',
    spConfig: '',
    status: '1',
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

function readJwtAccount(token) {
  if (!token) {
    return ''
  }

  try {
    const [, payload] = token.split('.')
    if (!payload) {
      return ''
    }

    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const decoded = window.atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '='))
    const claims = JSON.parse(decoded)
    return typeof claims.account === 'string' ? claims.account : ''
  } catch {
    return ''
  }
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

function toDateTimestamp(value) {
  if (!value) {
    return Number.NaN
  }
  if (typeof value === 'object') {
    const seconds = Number(value.seconds || 0)
    const nanos = Number(value.nanos || 0)
    const timestamp = seconds * 1000 + Math.floor(nanos / 1e6)
    return timestamp > 0 ? timestamp : Number.NaN
  }
  const date = new Date(value)
  return date.getTime()
}

function parseDateTimeFilter(value) {
  if (!value) {
    return null
  }
  const normalized = String(value).trim()
  if (!normalized) {
    return null
  }
  const candidate = normalized.replace(/\//g, '-').replace(/\s+/, 'T')
  const timestamp = new Date(candidate).getTime()
  return Number.isNaN(timestamp) ? null : timestamp
}

function formatDateTimeFilterValue(value) {
  if (!value) {
    return ''
  }
  return String(value).replace('T', ' ').replace(/-/g, '/')
}

function normalizeIdList(values) {
  return Array.from(
    new Set(
      (Array.isArray(values) ? values : [])
        .map((value) => String(value).trim())
        .filter((value) => value !== '' && value !== '0'),
    ),
  )
}

function maskToken(value) {
  if (!value) {
    return '-'
  }
  if (value.length <= 16) {
    return value
  }
  return `${value.slice(0, 8)}...${value.slice(-6)}`
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

function cpStatusLabel(status) {
  switch (Number(status)) {
    case 0:
      return '禁用'
    case 1:
      return '正常'
    case 2:
      return '冻结'
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
    roleIds: normalized.roleId ? [String(normalized.roleId)] : [],
    status: String(normalized.status ?? 1),
    remark: normalized.remark,
  }
}

function buildAdminPayload(form) {
  const roleIds = normalizeIdList(form.roleIds)
  return {
    id: Number(form.id) || 0,
    username: form.username.trim(),
    password: form.password.trim(),
    real_name: form.realName.trim(),
    role_id: Number(roleIds[0]) || 0,
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
    permissionIds: [],
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

function normalizeCpResponse(cp) {
  if (!cp) {
    return null
  }
  return {
    id: Number(pickValue(cp, ['id'], 0)),
    cpCode: pickValue(cp, ['cpCode', 'cp_code']),
    cpName: pickValue(cp, ['cpName', 'cp_name']),
    status: Number(pickValue(cp, ['status'], 0)),
    createTime: pickValue(cp, ['createTime', 'create_time']),
    updateTime: pickValue(cp, ['updateTime', 'update_time']),
  }
}

function normalizeCpForm(cp) {
  const normalized = normalizeCpResponse(cp)
  if (!normalized) {
    return createCpForm()
  }
  return {
    id: String(normalized.id || ''),
    cpCode: normalized.cpCode,
    cpName: normalized.cpName,
    spIds: [],
    status: String(normalized.status ?? 1),
  }
}

function buildCpPayload(form) {
  return {
    id: Number(form.id) || 0,
    cp_code: form.cpCode.trim(),
    cp_name: form.cpName.trim(),
    status: Number(form.status) || 0,
  }
}

function normalizeSpResponse(sp) {
  if (!sp) {
    return null
  }
  return {
    id: Number(pickValue(sp, ['id'], 0)),
    spCode: pickValue(sp, ['spCode', 'sp_code']),
    spName: pickValue(sp, ['spName', 'sp_name']),
    spConfig: pickValue(sp, ['spConfig', 'sp_config']),
    status: Number(pickValue(sp, ['status'], 0)),
    createTime: pickValue(sp, ['createTime', 'create_time']),
    updateTime: pickValue(sp, ['updateTime', 'update_time']),
  }
}

function normalizeSpForm(sp) {
  const normalized = normalizeSpResponse(sp)
  if (!normalized) {
    return createSpForm()
  }
  return {
    id: String(normalized.id || ''),
    spCode: normalized.spCode,
    spName: normalized.spName,
    spConfig: normalized.spConfig,
    status: String(normalized.status ?? 1),
  }
}

function buildSpPayload(form) {
  return {
    id: Number(form.id) || 0,
    sp_code: form.spCode.trim(),
    sp_name: form.spName.trim(),
    sp_config: form.spConfig.trim(),
    status: Number(form.status) || 0,
  }
}

function normalizeCpSpResponse(item) {
  if (!item) {
    return null
  }
  return {
    id: Number(pickValue(item, ['id'], 0)),
    cpId: Number(pickValue(item, ['cpId', 'cp_id'], 0)),
    spId: Number(pickValue(item, ['spId', 'sp_id'], 0)),
    status: Number(pickValue(item, ['status'], 0)),
    createTime: pickValue(item, ['createTime', 'create_time']),
    updateTime: pickValue(item, ['updateTime', 'update_time']),
  }
}

function normalizeAdminOperationLogResponse(item) {
  if (!item) {
    return null
  }
  return {
    id: Number(pickValue(item, ['id'], 0)),
    adminId: Number(pickValue(item, ['adminId', 'admin_id'], 0)),
    adminName: pickValue(item, ['adminName', 'admin_name']),
    module: pickValue(item, ['module']),
    action: pickValue(item, ['action']),
    description: pickValue(item, ['description']),
    requestMethod: pickValue(item, ['requestMethod', 'request_method']),
    requestUrl: pickValue(item, ['requestUrl', 'request_url']),
    requestParams: pickValue(item, ['requestParams', 'request_params']),
    createTime: pickValue(item, ['createTime', 'create_time']),
  }
}

function normalizeSystemLogResponse(item) {
  if (!item) {
    return null
  }
  return {
    id: Number(pickValue(item, ['id'], 0)),
    logUid: pickValue(item, ['logUid', 'log_uid']),
    logLevel: pickValue(item, ['logLevel', 'log_level']),
    message: pickValue(item, ['message']),
    filePath: pickValue(item, ['filePath', 'file_path']),
    lineNumber: Number(pickValue(item, ['lineNumber', 'line_number'], 0)),
    createTime: pickValue(item, ['createTime', 'create_time']),
  }
}

const resourceCatalog = [
  {
    key: 'cps',
    section: '内容管理',
    navLabel: '内容提供商',
    singularLabel: '内容提供商',
    pluralLabel: '内容提供商',
    basePath: '/cps',
    listEndpoint: '/api/v1/cps?page_size=100',
    createEndpoint: '/api/v1/cps/create',
    updateEndpoint: '/api/v1/cps/update',
    getEndpoint: (id) => `/api/v1/cps/${id}`,
    deleteEndpoint: (id) => `/api/v1/cps/${id}`,
    listKey: ['cps'],
    subtitle: '维护向 CDN 注入媒资的版权/制作方，协议侧 CPID 使用 CP 编码。',
    createDescription: 'CP 编码全局唯一，对应注入协议中的 CPID。可同时勾选多个内容服务商建立注入路由。',
    editDescription: '修改内容提供商资料及绑定的内容服务商后返回列表。',
    createForm: createCpForm,
    normalizeResponse: normalizeCpResponse,
    normalizeForm: normalizeCpForm,
    buildPayload: buildCpPayload,
    buildUpdateMaskPaths: () => ['cp_code', 'cp_name', 'status'],
    canSubmit: (form) => form.cpCode.trim() !== '' && form.cpName.trim() !== '',
    describe: (payload) => payload.cp_name || payload.cp_code || '内容提供商',
    fields: [
      { name: 'cpCode', label: 'CP 编码', type: 'text', placeholder: '协议侧 CPID，全局唯一' },
      { name: 'cpName', label: 'CP 全称', type: 'text', placeholder: '请输入内容提供商全称' },
      { name: 'spIds', label: '绑定内容服务商', type: 'checkbox-group', optionsSource: 'sps' },
      {
        name: 'status',
        label: '状态',
        type: 'select',
        options: [
          { value: '1', label: '正常' },
          { value: '0', label: '禁用' },
          { value: '2', label: '冻结' },
        ],
      },
    ],
    getCardTitle: (item) => item.cpName || item.cpCode,
    getCardBadge: (item) => ({
      text: cpStatusLabel(item.status),
      className: `status-chip status-chip--${Number(item.status)}`,
    }),
    listColumns: [
      {
        key: 'cpName',
        label: '内容提供商',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.cpName || '-'}</strong>
            <span>{item.cpCode || '未填写 CP 编码'}</span>
          </div>
        ),
      },
      {
        key: 'cpCode',
        label: 'CP 编码',
        render: (item) => item.cpCode || '-',
      },
      {
        key: 'status',
        label: '状态',
        render: (item) => (
          <span className={`status-chip status-chip--${Number(item.status)}`}>{cpStatusLabel(item.status)}</span>
        ),
      },
      {
        key: 'createTime',
        label: '创建时间',
        render: (item) => formatDateTime(item.createTime),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
    ],
    getCardMeta: (item) => [
      ['CP 编码', item.cpCode || '-'],
      ['CP 全称', item.cpName || '-'],
      ['状态', cpStatusLabel(item.status)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
  {
    key: 'sps',
    section: '内容管理',
    navLabel: '内容服务商',
    singularLabel: '内容服务商',
    pluralLabel: '内容服务商',
    basePath: '/sps',
    listEndpoint: '/api/v1/sps?page_size=100',
    createEndpoint: '/api/v1/sps/create',
    updateEndpoint: '/api/v1/sps/update',
    getEndpoint: (id) => `/api/v1/sps/${id}`,
    deleteEndpoint: (id) => `/api/v1/sps/${id}`,
    listKey: ['sps'],
    subtitle: '维护从 CDN 拉流给终端的 IPTV/OTT 分发平台，协议侧 SPID/CSPID 使用 SP 编码。',
    createDescription: 'SP 编码全局唯一，对应播放协议中的 SPID/CSPID；配置为可选 JSON。',
    editDescription: '修改内容服务商名称、编码、配置或状态后返回列表。',
    createForm: createSpForm,
    normalizeResponse: normalizeSpResponse,
    normalizeForm: normalizeSpForm,
    buildPayload: buildSpPayload,
    buildUpdateMaskPaths: () => ['sp_code', 'sp_name', 'sp_config', 'status'],
    canSubmit: (form) => form.spCode.trim() !== '' && form.spName.trim() !== '',
    describe: (payload) => payload.sp_name || payload.sp_code || '内容服务商',
    fields: [
      { name: 'spCode', label: 'SP 编码', type: 'text', placeholder: '协议侧 SPID/CSPID，全局唯一' },
      { name: 'spName', label: 'SP 全称', type: 'text', placeholder: '请输入内容服务商全称' },
      {
        name: 'status',
        label: '状态',
        type: 'select',
        options: [
          { value: '1', label: '正常' },
          { value: '0', label: '禁用' },
          { value: '2', label: '冻结' },
        ],
      },
      {
        name: 'spConfig',
        label: 'SP 配置',
        type: 'textarea',
        placeholder: '{"play_domain":"example.com"}',
        rows: 6,
      },
    ],
    getCardTitle: (item) => item.spName || item.spCode,
    getCardBadge: (item) => ({
      text: cpStatusLabel(item.status),
      className: `status-chip status-chip--${Number(item.status)}`,
    }),
    listColumns: [
      {
        key: 'spName',
        label: '内容服务商',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.spName || '-'}</strong>
            <span>{item.spCode || '未填写 SP 编码'}</span>
          </div>
        ),
      },
      {
        key: 'spCode',
        label: 'SP 编码',
        render: (item) => item.spCode || '-',
      },
      {
        key: 'status',
        label: '状态',
        render: (item) => (
          <span className={`status-chip status-chip--${Number(item.status)}`}>{cpStatusLabel(item.status)}</span>
        ),
      },
      {
        key: 'createTime',
        label: '创建时间',
        render: (item) => formatDateTime(item.createTime),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
    ],
    getCardMeta: (item) => [
      ['SP 编码', item.spCode || '-'],
      ['SP 全称', item.spName || '-'],
      ['状态', cpStatusLabel(item.status)],
      ['SP 配置', item.spConfig || '-'],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
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
      { name: 'roleIds', label: '绑定角色', type: 'checkbox-group', optionsSource: 'roles' },
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
    listColumns: [
      {
        key: 'username',
        label: '账号信息',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.username || '-'}</strong>
            <span>{item.realName || '未填写真实姓名'}</span>
          </div>
        ),
      },
      {
        key: 'roleId',
        label: '角色',
        render: (item) => `角色 #${item.roleId ?? 0}`,
      },
      {
        key: 'status',
        label: '状态',
        render: (item) => (
          <span className={`status-chip status-chip--${Number(item.status)}`}>{statusLabel(item.status)}</span>
        ),
      },
      {
        key: 'lastLoginTime',
        label: '最后登录',
        render: (item) => formatDateTime(item.lastLoginTime),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
    ],
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
    fields: [
      { name: 'roleName', label: '角色名称', type: 'text', placeholder: '请输入角色名称' },
      { name: 'permissionIds', label: '绑定权限', type: 'checkbox-group', optionsSource: 'permissions' },
    ],
    listColumns: [
      {
        key: 'roleName',
        label: '角色名称',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.roleName || `角色 #${item.id}`}</strong>
            <span>ID #{item.id}</span>
          </div>
        ),
      },
      {
        key: 'createTime',
        label: '创建时间',
        render: (item) => formatDateTime(item.createTime),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
    ],
    getCardTitle: (item) => item.roleName || `角色 #${item.id}`,
    getCardMeta: (item) => [
      ['角色 ID', String(item.id)],
      ['创建时间', formatDateTime(item.createTime)],
      ['更新时间', formatDateTime(item.updateTime)],
    ],
  },
  {
    key: 'admin-roles',
    hidden: true,
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
    listColumns: [
      {
        key: 'adminId',
        label: '管理员',
        render: (item) => (
          <div className="table-primary">
            <strong>管理员 #{item.adminId}</strong>
            <span>关联 ID #{item.id}</span>
          </div>
        ),
      },
      {
        key: 'roleId',
        label: '角色',
        render: (item) => `角色 #${item.roleId}`,
      },
      {
        key: 'createTime',
        label: '创建时间',
        render: (item) => formatDateTime(item.createTime),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
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
    allowCreate: false,
    allowEdit: false,
    allowDelete: false,
    basePath: '/permissions',
    listEndpoint: '/api/v1/permissions?page_size=100',
    createEndpoint: '/api/v1/permissions/create',
    updateEndpoint: '/api/v1/permissions/update',
    getEndpoint: (id) => `/api/v1/permissions/${id}`,
    deleteEndpoint: (id) => `/api/v1/permissions/${id}`,
    listKey: ['permissions'],
    subtitle: '权限由程序路由自动同步到数据库，仅支持查看。',
    createDescription: '权限由程序路由自动写入数据库，不支持手工新增。',
    editDescription: '权限由程序路由维护，不支持手工编辑。',
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
    listColumns: [
      {
        key: 'title',
        label: '权限节点',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.title || `权限 #${item.id}`}</strong>
            <span>{item.handle || '未设置句柄'}</span>
          </div>
        ),
      },
      {
        key: 'parentId',
        label: '父级 ID',
        render: (item) => String(item.parentId ?? 0),
      },
      {
        key: 'weight',
        label: '权重',
        render: (item) => String(item.weight ?? 0),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
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
    key: 'admin-operation-logs',
    section: '系统管理',
    navLabel: '操作日志',
    singularLabel: '操作日志',
    pluralLabel: '操作日志',
    allowCreate: false,
    allowEdit: false,
    allowDelete: false,
    enableTimeRangeSearch: true,
    basePath: '/admin-operation-logs',
    listEndpoint: '/api/v1/admin-operation-logs?page_size=100',
    listKey: ['adminOperationLogs', 'admin_operation_logs'],
    subtitle: '查看管理员登录及后台操作记录。',
    createDescription: '操作日志为系统只读数据。',
    editDescription: '操作日志为系统只读数据。',
    createForm: () => ({}),
    normalizeResponse: normalizeAdminOperationLogResponse,
    normalizeForm: () => ({}),
    buildPayload: () => ({}),
    buildUpdateMaskPaths: () => [],
    canSubmit: () => false,
    describe: () => '操作日志',
    fields: [],
    listColumns: [
      {
        key: 'adminName',
        label: '操作人',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.adminName || '-'}</strong>
            <span>ID #{item.adminId || 0}</span>
          </div>
        ),
      },
      {
        key: 'module',
        label: '模块 / 动作',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.module || '-'}</strong>
            <span>{item.action || '-'}</span>
          </div>
        ),
      },
      {
        key: 'description',
        label: '描述',
        render: (item) => item.description || '-',
      },
      {
        key: 'requestUrl',
        label: '请求',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.requestMethod || '-'}</strong>
            <span>{item.requestUrl || '-'}</span>
          </div>
        ),
      },
      {
        key: 'createTime',
        label: '时间',
        render: (item) => formatDateTime(item.createTime),
      },
    ],
    getCardTitle: (item) => item.description || '操作日志',
    getCardMeta: (item) => [
      ['操作人', `${item.adminName || '-'} (#${item.adminId || 0})`],
      ['模块', item.module || '-'],
      ['动作', item.action || '-'],
      ['请求', `${item.requestMethod || '-'} ${item.requestUrl || '-'}`],
      ['参数', item.requestParams || '-'],
      ['时间', formatDateTime(item.createTime)],
    ],
  },
  {
    key: 'system-logs',
    section: '系统管理',
    navLabel: '系统日志',
    singularLabel: '系统日志',
    pluralLabel: '系统日志',
    allowCreate: false,
    allowEdit: false,
    allowDelete: false,
    enableTimeRangeSearch: true,
    basePath: '/system-logs',
    listEndpoint: '/api/v1/system-logs?page_size=100',
    listKey: ['systemLogs', 'system_logs'],
    subtitle: '查看服务运行时输出的系统日志。',
    createDescription: '系统日志为只读数据。',
    editDescription: '系统日志为只读数据。',
    createForm: () => ({}),
    normalizeResponse: normalizeSystemLogResponse,
    normalizeForm: () => ({}),
    buildPayload: () => ({}),
    buildUpdateMaskPaths: () => [],
    canSubmit: () => false,
    describe: () => '系统日志',
    fields: [],
    listColumns: [
      {
        key: 'logLevel',
        label: '等级',
        render: (item) => item.logLevel || '-',
      },
      {
        key: 'message',
        label: '消息',
        render: (item) => item.message || '-',
      },
      {
        key: 'filePath',
        label: '文件位置',
        render: (item) => (
          <div className="table-primary">
            <strong>{item.filePath || '-'}</strong>
            <span>{item.lineNumber ? `Line ${item.lineNumber}` : '未记录行号'}</span>
          </div>
        ),
      },
      {
        key: 'logUid',
        label: '日志 UID',
        render: (item) => item.logUid || '-',
      },
      {
        key: 'createTime',
        label: '时间',
        render: (item) => formatDateTime(item.createTime),
      },
    ],
    getCardTitle: (item) => item.message || '系统日志',
    getCardMeta: (item) => [
      ['日志 UID', item.logUid || '-'],
      ['等级', item.logLevel || '-'],
      ['消息', item.message || '-'],
      ['文件', item.filePath || '-'],
      ['行号', item.lineNumber ? String(item.lineNumber) : '-'],
      ['时间', formatDateTime(item.createTime)],
    ],
  },
  {
    key: 'group-permissions',
    hidden: true,
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
    listColumns: [
      {
        key: 'groupId',
        label: '分组',
        render: (item) => (
          <div className="table-primary">
            <strong>分组 #{item.groupId}</strong>
            <span>关联 ID #{item.id}</span>
          </div>
        ),
      },
      {
        key: 'permissionId',
        label: '权限',
        render: (item) => `权限 #${item.permissionId}`,
      },
      {
        key: 'createTime',
        label: '创建时间',
        render: (item) => formatDateTime(item.createTime),
      },
      {
        key: 'updateTime',
        label: '更新时间',
        render: (item) => formatDateTime(item.updateTime),
      },
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

function getResourceViewLabel(pathname, resource) {
  if (pathname.endsWith('/new')) {
    return `新建${resource?.singularLabel || ''}`
  }
  if (pathname.endsWith('/edit')) {
    return `编辑${resource?.singularLabel || ''}`
  }
  return `${resource?.pluralLabel || '管理'}列表`
}

function renderColumnContent(column, item) {
  if (typeof column.render === 'function') {
    return column.render(item)
  }

  const value = item?.[column.key]
  if (value === undefined || value === null || value === '') {
    return '-'
  }
  return value
}

function buildSearchText(resource, item) {
  const title = resource.getCardTitle ? String(resource.getCardTitle(item) || '') : ''
  const meta = resource.getCardMeta
    ? resource
        .getCardMeta(item)
        .map(([label, value]) => `${label} ${value || ''}`)
        .join(' ')
    : Object.values(item || {}).join(' ')

  return `${title} ${meta}`.toLowerCase()
}

function mapRoleOptions(roles) {
  return roles.map((role) => ({
    value: String(role.id),
    label: role.roleName || `角色 #${role.id}`,
    description: `角色 ID：${role.id}`,
  }))
}

function mapSpOptions(sps) {
  return sps.map((sp) => ({
    value: String(sp.id),
    label: sp.spName || `内容服务商 #${sp.id}`,
    description: `${sp.spCode || '未填写编码'} · ${cpStatusLabel(sp.status)}`,
  }))
}

function mapPermissionOptions(permissions) {
  const sorted = [...permissions].sort((left, right) => {
    if (left.parentId !== right.parentId) {
      return left.parentId - right.parentId
    }
    if (left.weight !== right.weight) {
      return left.weight - right.weight
    }
    return left.id - right.id
  })

  const childrenMap = new Map()
  sorted.forEach((permission) => {
    const parentId = permission.parentId || 0
    const siblings = childrenMap.get(parentId) || []
    siblings.push(permission)
    childrenMap.set(parentId, siblings)
  })

  const flattened = []
  const walk = (parentId, depth) => {
    const children = childrenMap.get(parentId) || []
    children.forEach((permission) => {
      flattened.push({
        value: String(permission.id),
        label: permission.title || `权限 #${permission.id}`,
        description: `${permission.handle || '未设置路由'} · 父级 ${permission.parentId}`,
        parentValue: permission.parentId > 0 ? String(permission.parentId) : '',
        depth,
      })
      walk(permission.id, depth + 1)
    })
  }

  walk(0, 0)
  return flattened
}

function collectPermissionDescendants(value, options) {
  const childrenMap = new Map()
  options.forEach((option) => {
    const parentValue = option.parentValue || ''
    const siblings = childrenMap.get(parentValue) || []
    siblings.push(option.value)
    childrenMap.set(parentValue, siblings)
  })

  const descendants = []
  const walk = (currentValue) => {
    const children = childrenMap.get(currentValue) || []
    children.forEach((childValue) => {
      descendants.push(childValue)
      walk(childValue)
    })
  }

  walk(value)
  return descendants
}

function normalizePermissionSelection(selectedValues, options) {
  const selected = new Set(normalizeIdList(selectedValues))
  const childrenMap = new Map()

  options.forEach((option) => {
    const parentValue = option.parentValue || ''
    const siblings = childrenMap.get(parentValue) || []
    siblings.push(option.value)
    childrenMap.set(parentValue, siblings)
  })

  let changed = true
  while (changed) {
    changed = false
    options.forEach((option) => {
      const children = childrenMap.get(option.value) || []
      if (children.length === 0) {
        return
      }

      const allChildrenSelected = children.every((childValue) => selected.has(childValue))
      if (allChildrenSelected && !selected.has(option.value)) {
        selected.add(option.value)
        changed = true
      }
      if (!allChildrenSelected && selected.has(option.value)) {
        selected.delete(option.value)
        changed = true
      }
    })
  }

  return [...selected]
}

async function syncAdminRoleBindings(adminId, selectedRoleIds) {
  const current = await requestJson('/api/v1/admin-roles?page_size=1000', {
    headers: buildAuthHeaders(),
  })
  const currentItems = getListItems(current, ['adminRoles', 'admin_roles'])
    .map(normalizeAdminRoleResponse)
    .filter((item) => item && item.adminId === adminId)

  const targetIds = new Set(normalizeIdList(selectedRoleIds).map(Number))
  const currentIds = new Set(currentItems.map((item) => item.roleId))

  const createTasks = [...targetIds]
    .filter((roleId) => !currentIds.has(roleId))
    .map((roleId) =>
      requestJson('/api/v1/admin-roles/create', {
        method: 'POST',
        headers: buildAuthHeaders(),
        body: JSON.stringify({
          admin_id: adminId,
          role_id: roleId,
        }),
      }),
    )

  const deleteTasks = currentItems
    .filter((item) => !targetIds.has(item.roleId))
    .map((item) =>
      requestJson(`/api/v1/admin-roles/${item.id}`, {
        method: 'DELETE',
        headers: buildAuthHeaders(),
      }),
    )

  await Promise.all([...createTasks, ...deleteTasks])
}

async function syncRolePermissionBindings(roleId, selectedPermissionIds) {
  const current = await requestJson('/api/v1/group-permissions?page_size=1000', {
    headers: buildAuthHeaders(),
  })
  const currentItems = getListItems(current, ['groupPermissions', 'group_permissions'])
    .map(normalizeGroupPermissionResponse)
    .filter((item) => item && item.groupId === roleId)

  const targetIds = new Set(normalizeIdList(selectedPermissionIds).map(Number))
  const currentIds = new Set(currentItems.map((item) => item.permissionId))

  const createTasks = [...targetIds]
    .filter((permissionId) => !currentIds.has(permissionId))
    .map((permissionId) =>
      requestJson('/api/v1/group-permissions/create', {
        method: 'POST',
        headers: buildAuthHeaders(),
        body: JSON.stringify({
          group_id: roleId,
          permission_id: permissionId,
        }),
      }),
    )

  const deleteTasks = currentItems
    .filter((item) => !targetIds.has(item.permissionId))
    .map((item) =>
      requestJson(`/api/v1/group-permissions/${item.id}`, {
        method: 'DELETE',
        headers: buildAuthHeaders(),
      }),
    )

  await Promise.all([...createTasks, ...deleteTasks])
}

async function syncCpSpBindings(cpId, selectedSpIds) {
  const current = await requestJson('/api/v1/cp-sps?page_size=1000', {
    headers: buildAuthHeaders(),
  })
  const currentItems = getListItems(current, ['cpSps', 'cp_sps'])
    .map(normalizeCpSpResponse)
    .filter((item) => item && item.cpId === cpId)

  const targetIds = new Set(normalizeIdList(selectedSpIds).map(Number))
  const currentIds = new Set(currentItems.map((item) => item.spId))

  const createTasks = [...targetIds]
    .filter((spId) => !currentIds.has(spId))
    .map((spId) =>
      requestJson('/api/v1/cp-sps/create', {
        method: 'POST',
        headers: buildAuthHeaders(),
        body: JSON.stringify({
          cp_id: cpId,
          sp_id: spId,
          status: 1,
        }),
      }),
    )

  const deleteTasks = currentItems
    .filter((item) => !targetIds.has(item.spId))
    .map((item) =>
      requestJson(`/api/v1/cp-sps/${item.id}`, {
        method: 'DELETE',
        headers: buildAuthHeaders(),
      }),
    )

  await Promise.all([...createTasks, ...deleteTasks])
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
        account: form.account.trim(),
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
  const breadcrumbTail = getResourceViewLabel(location.pathname, currentResource)
  const navSections = ['内容管理', '系统管理', '关联管理']
    .map((title) => ({
      title,
      items: resourceCatalog.filter((resource) => resource.section === title && !resource.hidden),
    }))
    .filter((section) => section.items.length > 0)

  const handleLogout = () => {
    clearAuth()
    navigate('/', { replace: true })
  }

  return (
    <main className="console-page">
      <aside className="sidebar">
        <div className="sidebar__brand">
          <p className="eyebrow">Crow</p>
          <h1>播控管理中心</h1>
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
                </NavLink>
              ))}
            </nav>
          </section>
        ))}

        <section className="sidebar__session">
          <div className="sidebar__session-card">
            <div className="sidebar__session-header">
              <div>
                <p className="eyebrow">Session</p>
                <h3>当前登录状态</h3>
              </div>
              <span className="status-dot">
                <span />
                已连接
              </span>
            </div>

            <dl className="session-grid">
              <div>
                <dt>用户名</dt>
                <dd>{auth.account || readJwtAccount(auth.accessToken) || '-'}</dd>
              </div>
              <div>
                <dt>令牌类型</dt>
                <dd>{auth.tokenType || '-'}</dd>
              </div>
              <div>
                <dt>有效期</dt>
                <dd>{String(auth.expiresIn || '-')}</dd>
              </div>
              <div>
                <dt>访问令牌</dt>
                <dd>{maskToken(auth.accessToken)}</dd>
              </div>
            </dl>

            <button className="ghost-button ghost-button--dark" onClick={handleLogout} type="button">
              退出登录
            </button>
          </div>
        </section>
      </aside>

      <section className="workspace">

        <section className="content-stack">
          <Outlet />
        </section>
      </section>
    </main>
  )
}

function TimeRangeField({ label, onChange, placeholder, value }) {
  const inputRef = useRef(null)

  const handleClick = () => {
    inputRef.current?.showPicker?.()
    inputRef.current?.focus()
  }

  return (
    <label className="search-field search-field--compact">
      <span>{label}</span>
      <div className="time-range-field" onClick={handleClick} role="presentation">
        <input
          className="search-field__input--placeholder time-range-field__display"
          placeholder={placeholder}
          readOnly
          tabIndex={-1}
          type="text"
          value={formatDateTimeFilterValue(value)}
        />
        <input
          ref={inputRef}
          className="time-range-field__native"
          onChange={onChange}
          type="datetime-local"
          value={value}
        />
      </div>
    </label>
  )
}

function ResourceListPage({ resource }) {
  const location = useLocation()
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(location.state?.message || '')
  const [searchQuery, setSearchQuery] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const filteredItems = useMemo(() => {
    const keyword = searchQuery.trim().toLowerCase()
    const startTimestamp = parseDateTimeFilter(startTime)
    const endTimestamp = parseDateTimeFilter(endTime)

    return items.filter((item) => {
      if (keyword && !buildSearchText(resource, item).includes(keyword)) {
        return false
      }
      if (resource.enableTimeRangeSearch) {
        const createdAt = toDateTimestamp(item.createTime)
        if (startTimestamp && (Number.isNaN(createdAt) || createdAt < startTimestamp)) {
          return false
        }
        if (endTimestamp && (Number.isNaN(createdAt) || createdAt > endTimestamp)) {
          return false
        }
      }
      return true
    })
  }, [endTime, items, resource, searchQuery, startTime])

  const totalPages = Math.max(1, Math.ceil(filteredItems.length / pageSize))
  const paginatedItems = useMemo(() => {
    const start = (page - 1) * pageSize
    return filteredItems.slice(start, start + pageSize)
  }, [filteredItems, page, pageSize])

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
    setSearchQuery('')
    setStartTime('')
    setEndTime('')
    setPage(1)
  }, [resource.key])

  useEffect(() => {
    setPage(1)
  }, [resource.key, searchQuery, startTime, endTime, pageSize])

  useEffect(() => {
    if (page > totalPages) {
      setPage(totalPages)
    }
  }, [page, totalPages])

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
          {resource.allowCreate === false ? null : (
            <Link className="submit submit--compact action-link" to={`${resource.basePath}/new`}>
              新建{resource.singularLabel}
            </Link>
          )}
        </div>
      </div>

      {error && <p className="message message--error">{error}</p>}
      {success && <p className="message message--success">{success}</p>}

      <div className="table-toolbar">
        <label className="search-field">
          <span>搜索</span>
          <input
            onChange={(event) => setSearchQuery(event.target.value)}
            placeholder={`搜索${resource.singularLabel}名称、ID、状态等信息`}
            value={searchQuery}
          />
        </label>

        {resource.enableTimeRangeSearch ? (
          <>
            <TimeRangeField
              label="开始时间"
              onChange={(event) => setStartTime(event.target.value)}
              placeholder="YYYY/MM/DD HH:mm"
              value={startTime}
            />
            <TimeRangeField
              label="结束时间"
              onChange={(event) => setEndTime(event.target.value)}
              placeholder="YYYY/MM/DD HH:mm"
              value={endTime}
            />
          </>
        ) : null}
      </div>

      <div className="table-card">
        {loading ? (
          <p className="empty-state">加载中...</p>
        ) : filteredItems.length === 0 ? (
          <p className="empty-state">没有匹配的{resource.pluralLabel}数据。</p>
        ) : (
          <div className="table-scroll">
            <table className="admin-table">
              <thead>
                <tr>
                  <th scope="col">ID</th>
                  {resource.listColumns.map((column) => (
                    <th key={column.label} scope="col">
                      {column.label}
                    </th>
                  ))}
                  {resource.allowEdit === false && resource.allowDelete === false ? null : (
                    <th className="admin-table__actions" scope="col">
                      操作
                    </th>
                  )}
                </tr>
              </thead>
              <tbody>
                {paginatedItems.map((item) => (
                  <tr key={item.id}>
                    <td>{item.id}</td>
                    {resource.listColumns.map((column) => (
                      <td key={column.label}>{renderColumnContent(column, item)}</td>
                    ))}
                    {resource.allowEdit === false && resource.allowDelete === false ? null : (
                      <td className="admin-table__actions-cell">
                        <div className="table-actions">
                          {resource.allowEdit === false ? null : (
                            <Link className="ghost-button action-link action-link--inline" to={`${resource.basePath}/${item.id}/edit`}>
                              编辑
                            </Link>
                          )}
                          {resource.allowDelete === false ? null : (
                            <button className="danger-button" onClick={() => handleDelete(item)} type="button">
                              删除
                            </button>
                          )}
                        </div>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="pagination">
        <div className="pagination__summary">
          共 {filteredItems.length} 条，当前第 {page} / {totalPages} 页
        </div>

        <div className="pagination__controls">
          <label className="pagination__page-size">
            <span>每页</span>
            <select onChange={(event) => setPageSize(Number(event.target.value))} value={pageSize}>
              {[10, 20, 50].map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
            </select>
          </label>

          <button className="ghost-button" disabled={page <= 1} onClick={() => setPage((current) => Math.max(1, current - 1))} type="button">
            上一页
          </button>
          <span className="pagination__current">{page}</span>
          <button
            className="ghost-button"
            disabled={page >= totalPages}
            onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
            type="button"
          >
            下一页
          </button>
        </div>
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
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [referenceOptions, setReferenceOptions] = useState({
    roles: [],
    permissions: [],
    sps: [],
  })

  useEffect(() => {
    let cancelled = false

    async function loadItem() {
      setLoading(true)
      setError('')

      try {
        const headers = buildAuthHeaders()
        const requests = [
          isEditing
            ? requestJson(resource.getEndpoint(itemId), {
                headers,
              })
            : Promise.resolve(null),
          resource.key === 'admins'
            ? requestJson('/api/v1/roles?page_size=100', {
                headers,
              })
            : Promise.resolve(null),
          resource.key === 'admins' && isEditing
            ? requestJson('/api/v1/admin-roles?page_size=1000', {
                headers,
              })
            : Promise.resolve(null),
          resource.key === 'roles'
            ? requestJson('/api/v1/permissions?page_size=1000', {
                headers,
              })
            : Promise.resolve(null),
          resource.key === 'roles' && isEditing
            ? requestJson('/api/v1/group-permissions?page_size=1000', {
                headers,
              })
            : Promise.resolve(null),
          resource.key === 'cps'
            ? requestJson('/api/v1/sps?page_size=1000', {
                headers,
              })
            : Promise.resolve(null),
          resource.key === 'cps' && isEditing
            ? requestJson('/api/v1/cp-sps?page_size=1000', {
                headers,
              })
            : Promise.resolve(null),
        ]

        const [itemPayload, rolesPayload, adminRolesPayload, permissionsPayload, groupPermissionsPayload, spsPayload, cpSpsPayload] = await Promise.all(requests)

        if (!cancelled) {
          const nextForm = isEditing ? resource.normalizeForm(itemPayload) : resource.createForm()
          const nextReferenceOptions = {
            roles: [],
            permissions: [],
            sps: [],
          }

          if (rolesPayload) {
            const roles = getListItems(rolesPayload, ['roles']).map(normalizeRoleResponse).filter(Boolean)
            nextReferenceOptions.roles = mapRoleOptions(roles)

            if (resource.key === 'admins') {
              const selectedRoleIds = adminRolesPayload
                ? getListItems(adminRolesPayload, ['adminRoles', 'admin_roles'])
                    .map(normalizeAdminRoleResponse)
                    .filter((relation) => relation && relation.adminId === Number(itemId))
                    .map((relation) => String(relation.roleId))
                : []

              if (selectedRoleIds.length > 0) {
                nextForm.roleIds = normalizeIdList(selectedRoleIds)
              }
            }
          }

          if (permissionsPayload) {
            const permissions = getListItems(permissionsPayload, ['permissions']).map(normalizePermissionResponse).filter(Boolean)
            nextReferenceOptions.permissions = mapPermissionOptions(permissions)

            if (resource.key === 'roles' && groupPermissionsPayload) {
              nextForm.permissionIds = normalizeIdList(
                getListItems(groupPermissionsPayload, ['groupPermissions', 'group_permissions'])
                  .map(normalizeGroupPermissionResponse)
                  .filter((relation) => relation && relation.groupId === Number(itemId))
                  .map((relation) => String(relation.permissionId)),
              )
            }
          }

          if (spsPayload) {
            const sps = getListItems(spsPayload, ['sps']).map(normalizeSpResponse).filter(Boolean)
            nextReferenceOptions.sps = mapSpOptions(sps)

            if (resource.key === 'cps' && cpSpsPayload) {
              nextForm.spIds = normalizeIdList(
                getListItems(cpSpsPayload, ['cpSps', 'cp_sps'])
                  .map(normalizeCpSpResponse)
                  .filter((relation) => relation && relation.cpId === Number(itemId))
                  .map((relation) => String(relation.spId)),
              )
            }
          }

          setReferenceOptions(nextReferenceOptions)
          setForm(nextForm)
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

  const getFieldOptions = (field) => {
    if (field.optionsSource === 'roles') {
      return referenceOptions.roles
    }
    if (field.optionsSource === 'permissions') {
      return referenceOptions.permissions
    }
    if (field.optionsSource === 'sps') {
      return referenceOptions.sps
    }
    return field.options || []
  }

  const handleChange = (event) => {
    const { name, value } = event.target
    setForm((current) => ({
      ...current,
      [name]: value,
    }))
  }

  const handleCheckboxGroupChange = (name, value, checked) => {
    setForm((current) => {
      const currentValues = normalizeIdList(current[name])
      const field = resource.fields.find((item) => item.name === name)
      const options = getFieldOptions(field || {})
      const descendants = field?.optionsSource === 'permissions' ? collectPermissionDescendants(value, options) : []
      const touchedValues = [value, ...descendants]
      const nextValues = checked
        ? normalizeIdList([...currentValues, ...touchedValues])
        : currentValues.filter((item) => !touchedValues.includes(item))

      return {
        ...current,
        [name]: field?.optionsSource === 'permissions' ? normalizePermissionSelection(nextValues, options) : normalizeIdList(nextValues),
      }
    })
  }

  const handleSubmit = async (event) => {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    const payload = resource.buildPayload(form)

    try {
      let savedResource = null

      if (isEditing) {
        const query = buildUpdateMaskQuery(resource.buildUpdateMaskPaths(form, payload))
        savedResource = await requestJson(`${resource.updateEndpoint}?${query}`, {
          method: 'PUT',
          headers: buildAuthHeaders(),
          body: JSON.stringify(payload),
        })
      } else {
        savedResource = await requestJson(resource.createEndpoint, {
          method: 'POST',
          headers: buildAuthHeaders(),
          body: JSON.stringify(payload),
        })
      }

      const savedID = Number(pickValue(savedResource, ['id'], payload.id || itemId || 0))
      if (resource.key === 'admins' && savedID > 0) {
        await syncAdminRoleBindings(savedID, form.roleIds)
      }
      if (resource.key === 'roles' && savedID > 0) {
        await syncRolePermissionBindings(savedID, form.permissionIds)
      }
      if (resource.key === 'cps' && savedID > 0) {
        await syncCpSpBindings(savedID, form.spIds)
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
    <section className="form-page-grid">
      <div className="panel panel--form">
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
          <form className="form form--spacious" onSubmit={handleSubmit}>
            <div className="form-section">
              <div className="form-section__header">
                <h3>基础信息</h3>
                <p>按照标准后台表单布局填写并保存当前资源。</p>
              </div>

              <div className="form-grid">
                {resource.fields.map((field) => {
                  const label = typeof field.label === 'function' ? field.label(isEditing) : field.label
                  const placeholder = typeof field.placeholder === 'function' ? field.placeholder(isEditing) : field.placeholder
                  const className = field.type === 'textarea' || field.type === 'checkbox-group' ? 'field field--full' : 'field'
                  const options = getFieldOptions(field)

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
                          {options.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                      ) : field.type === 'checkbox-group' ? (
                        <div className="checkbox-group">
                          {options.length === 0 ? (
                            <p className="checkbox-group__empty">暂无可选项</p>
                          ) : (
                            options.map((option) => {
                              const checked = Array.isArray(form[field.name]) && form[field.name].includes(option.value)

                              return (
                                <label
                                  className="checkbox-option"
                                  key={option.value}
                                  style={option.depth ? { paddingLeft: `${12 + option.depth * 18}px` } : undefined}
                                >
                                  <input
                                    checked={checked}
                                    onChange={(event) => handleCheckboxGroupChange(field.name, option.value, event.target.checked)}
                                    type="checkbox"
                                  />
                                  <div>
                                    <strong>{option.label}</strong>
                                    {option.description ? <small>{option.description}</small> : null}
                                  </div>
                                </label>
                              )
                            })
                          )}
                        </div>
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
      </div>

      <aside className="side-stack">
        <article className="panel side-card">
          <span className="side-card__label">当前操作</span>
          <strong>{isEditing ? '编辑模式' : '创建模式'}</strong>
          <p>{resource.section}</p>
        </article>

        <article className="panel side-card">
          <span className="side-card__label">字段概览</span>
          <strong>{resource.fields.length} 个字段</strong>
          <ul className="side-card__list">
            {resource.fields.slice(0, 5).map((field) => {
              const label = typeof field.label === 'function' ? field.label(isEditing) : field.label
              return <li key={field.name}>{label}</li>
            })}
          </ul>
        </article>

        <article className="panel side-card">
          <span className="side-card__label">填写说明</span>
          <strong>保存前校验必填项</strong>
          <p>该页保持标准后台表单结构，适合连续录入和修改资源信息。</p>
        </article>
      </aside>
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
                {resource.allowCreate === false ? null : (
                  <Route element={<ResourceFormPage mode="new" resource={resource} />} path={`${resource.basePath}/new`} />
                )}
                {resource.allowEdit === false ? null : (
                  <Route
                    element={<ResourceFormPage mode="edit" resource={resource} />}
                    path={`${resource.basePath}/:id/edit`}
                  />
                )}
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
