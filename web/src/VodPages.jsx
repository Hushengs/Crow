import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'

const authKey = 'crow.auth'

function headers(includeContentType = true) {
  let token = ''
  try {
    token = JSON.parse(sessionStorage.getItem(authKey) || '{}').accessToken || ''
  } catch {
    /* noop */
  }
  return { ...(includeContentType ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) }
}

async function api(url, options = {}) {
  const hasFormDataBody = options.body instanceof FormData
  const response = await fetch(url, { ...options, headers: { ...headers(!hasFormDataBody), ...options.headers } })
  const data = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(data.message || '请求失败')
  return data
}

function pick(source, keys, fallback = '') {
  for (const key of keys) {
    if (source?.[key] !== undefined && source?.[key] !== null) {
      return source[key]
    }
  }
  return fallback
}

function normalizeCategory(item) {
  if (!item) return null
  return {
    id: Number(pick(item, ['id'], 0)),
    parentId: Number(pick(item, ['parentId', 'parent_id'], 0)),
    name: pick(item, ['name']),
    sortOrder: Number(pick(item, ['sortOrder', 'sort_order'], 0)),
    status: Number(pick(item, ['status'], 1)),
  }
}

function normalizeVideo(item) {
  if (!item) return null
  return {
    id: Number(pick(item, ['id'], 0)),
    categoryId: Number(pick(item, ['categoryId', 'category_id'], 0)),
    videoCode: pick(item, ['videoCode', 'video_code']),
    title: pick(item, ['title']),
    subtitle: pick(item, ['subtitle']),
    videoType: Number(pick(item, ['videoType', 'video_type'], 1)),
    posterVerticalUrl: pick(item, ['posterVerticalUrl', 'poster_vertical_url']),
    posterHorizontalUrl: pick(item, ['posterHorizontalUrl', 'poster_horizontal_url']),
    thumbnailUrl: pick(item, ['thumbnailUrl', 'thumbnail_url']),
    description: pick(item, ['description']),
    year: Number(pick(item, ['year'], 0)),
    duration: Number(pick(item, ['duration'], 0)),
    status: Number(pick(item, ['status'], 1)),
  }
}

function normalizeEpisode(item) {
  if (!item) return null
  return {
    id: Number(pick(item, ['id'], 0)),
    videoId: Number(pick(item, ['videoId', 'video_id'], 0)),
    episodeNo: Number(pick(item, ['episodeNo', 'episode_no'], 1)),
    title: pick(item, ['title']),
    duration: Number(pick(item, ['duration'], 0)),
    description: pick(item, ['description']),
    status: Number(pick(item, ['status'], 1)),
  }
}

function normalizeMedia(item) {
  if (!item) return null
  return {
    id: Number(pick(item, ['id'], 0)),
    videoId: Number(pick(item, ['videoId', 'video_id'], 0)),
    episodeId: Number(pick(item, ['episodeId', 'episode_id'], 0)),
    mediaId: pick(item, ['mediaId', 'media_id']),
    mediaUrl: pick(item, ['mediaUrl', 'media_url']),
    fileFormat: pick(item, ['fileFormat', 'file_format']),
    bitrate: Number(pick(item, ['bitrate'], 0)),
    resolution: pick(item, ['resolution']),
    fileSize: Number(pick(item, ['fileSize', 'file_size'], 0)),
    duration: Number(pick(item, ['duration'], 0)),
    status: Number(pick(item, ['status'], 1)),
  }
}

function createVideoForm(video = null) {
  return {
    id: String(video?.id || ''),
    categoryId: String(video?.categoryId || ''),
    videoCode: video?.videoCode || '',
    title: video?.title || '',
    subtitle: video?.subtitle || '',
    videoType: String(video?.videoType || 1),
    posterVerticalUrl: video?.posterVerticalUrl || '',
    posterHorizontalUrl: video?.posterHorizontalUrl || '',
    thumbnailUrl: video?.thumbnailUrl || '',
    description: video?.description || '',
    year: String(video?.year || ''),
    duration: String(video?.duration || ''),
    status: String(video?.status ?? 1),
  }
}

function buildVideoPayload(form) {
  return {
    id: Number(form.id) || 0,
    category_id: Number(form.categoryId),
    video_code: form.videoCode.trim(),
    title: form.title.trim(),
    subtitle: form.subtitle.trim(),
    video_type: Number(form.videoType),
    poster_vertical_url: form.posterVerticalUrl.trim(),
    poster_horizontal_url: form.posterHorizontalUrl.trim(),
    thumbnail_url: form.thumbnailUrl.trim(),
    description: form.description.trim(),
    year: Number(form.year) || 0,
    duration: Number(form.duration) || 0,
    status: Number(form.status),
  }
}

async function uploadImageAsset(file) {
  const formData = new FormData()
  formData.append('file', file)
  const payload = await api('/api/v1/videos/posters/upload', {
    method: 'POST',
    body: formData,
  })
  const url = pick(payload, ['url'])
  if (!url) {
    throw new Error('上传成功但未返回图片地址')
  }
  return url
}

function field(name, label, value, onChange, options = {}) {
  return (
    <label className={options.full ? 'field field--full' : 'field'}>
      <span>{label}</span>
      {options.type === 'textarea' ? (
        <textarea name={name} onChange={onChange} rows={options.rows || 4} value={value} placeholder={options.placeholder} />
      ) : options.choices ? (
        <select name={name} onChange={onChange} value={value}>
          {options.choices.map((item) => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>
      ) : (
        <input name={name} onChange={onChange} type={options.type || 'text'} value={value} placeholder={options.placeholder} />
      )}
    </label>
  )
}

function ImageUploadField({ label, helper, shape = 'portrait', value, onChange }) {
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')

  async function handleFileChange(event) {
    const file = event.target.files?.[0]
    if (!file) return
    setUploading(true)
    setError('')
    try {
      onChange(await uploadImageAsset(file))
    } catch (err) {
      setError(err.message || '图片上传失败')
    } finally {
      setUploading(false)
      event.target.value = ''
    }
  }

  return (
    <div className="field field--full poster-upload">
      <div className="poster-upload__header">
        <span>{label}</span>
        <small>{helper}</small>
      </div>
      <div className="poster-upload__body">
        <div className={`poster-upload__frame ${shape === 'landscape' ? 'poster-upload__frame--landscape' : ''}`}>
          {value ? <img alt={label} src={value} /> : <span>{shape === 'landscape' ? '16:9' : '3:4'}</span>}
        </div>
        <div className="poster-upload__actions">
          <label className={`ghost-button poster-upload__button ${uploading ? 'is-disabled' : ''}`}>
            <input accept="image/png,image/jpeg,image/webp,image/gif" disabled={uploading} onChange={handleFileChange} type="file" />
            {uploading ? '上传中...' : value ? '重新上传' : '上传图片'}
          </label>
          <button className="ghost-button" disabled={!value || uploading} onClick={() => onChange('')} type="button">
            清空
          </button>
          <code className="poster-upload__path">{value || '尚未上传图片'}</code>
          {error ? <p className="message message--error">{error}</p> : null}
        </div>
      </div>
    </div>
  )
}

function buildTree(items) {
  const children = new Map()
  items.forEach((item) => {
    const parentId = item.parentId || 0
    children.set(parentId, [...(children.get(parentId) || []), item])
  })
  const walk = (parentId, depth = 0) =>
    (children.get(parentId) || []).flatMap((item) => [{ ...item, depth }, ...walk(item.id, depth + 1)])
  return walk(0)
}

const videoTypeLabel = ['', '电影', '剧集', '综艺', '其他']

export function VideoLibraryPage() {
  const [categories, setCategories] = useState([])
  const [videos, setVideos] = useState([])
  const [categoryId, setCategoryId] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [categoryName, setCategoryName] = useState('')
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    try {
      setError('')
      const query = new URLSearchParams()
      if (categoryId) query.set('category_id', String(categoryId))
      if (keyword.trim()) query.set('keyword', keyword.trim())
      const [categoryData, videoData] = await Promise.all([
        api('/api/v1/video-categories'),
        api(`/api/v1/videos?${query}`),
      ])
      setCategories((categoryData.categories || []).map(normalizeCategory).filter(Boolean))
      setVideos((videoData.videos || []).map(normalizeVideo).filter(Boolean))
    } catch (err) {
      setError(err.message)
    }
  }, [categoryId, keyword])

  useEffect(() => {
    load()
  }, [load])

  const tree = useMemo(() => buildTree(categories), [categories])

  async function createCategory(event) {
    event.preventDefault()
    if (!categoryName.trim()) return
    try {
      await api('/api/v1/video-categories/create', {
        method: 'POST',
        body: JSON.stringify({ parent_id: categoryId, name: categoryName.trim(), status: 1 }),
      })
      setCategoryName('')
      await load()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <section className="vod-library">
      <aside className="panel vod-tree">
        <div className="vod-section-title">
          <div>
            <span className="side-card__label">CATALOG</span>
            <h2>影片分类</h2>
          </div>
          <span className="vod-count">{categories.length}</span>
        </div>
        <button className={`vod-tree__node ${categoryId === 0 ? 'is-active' : ''}`} onClick={() => setCategoryId(0)} type="button">
          <span>全部影片</span>
        </button>
        {tree.map((item) => (
          <button
            className={`vod-tree__node ${categoryId === item.id ? 'is-active' : ''}`}
            key={item.id}
            onClick={() => setCategoryId(item.id)}
            style={{ paddingLeft: `${18 + item.depth * 20}px` }}
            type="button"
          >
            <span>
              {item.depth ? '└ ' : '▸ '}
              {item.name}
            </span>
          </button>
        ))}
        <form className="vod-category-form" onSubmit={createCategory}>
          <input
            onChange={(event) => setCategoryName(event.target.value)}
            placeholder={categoryId ? '新增子分类' : '新增根分类'}
            value={categoryName}
          />
          <button className="ghost-button" type="submit">
            添加
          </button>
        </form>
      </aside>

      <section className="panel vod-content">
        <div className="panel__header">
          <div className="card__header">
            <h2>影片库</h2>
            <p className="subtitle">按分类管理影片，并进入详情维护节目与媒体资源。</p>
          </div>
          <Link className="submit submit--compact action-link" to="/videos/new">
            添加影片
          </Link>
        </div>
        <div className="table-toolbar">
          <label className="search-field">
            <span>检索影片</span>
            <input onChange={(e) => setKeyword(e.target.value)} placeholder="片名 / 副标题 / 影片编码" value={keyword} />
          </label>
        </div>
        {error && <p className="message message--error">{error}</p>}
        <div className="vod-grid">
          {videos.map((video) => (
            <Link className="vod-card" key={video.id} to={`/videos/${video.id}`}>
              <div className="vod-card__poster">
                {video.posterVerticalUrl ? <img alt="" src={video.posterVerticalUrl} /> : <span>{(video.title || '?').slice(0, 1)}</span>}
                <i>{videoTypeLabel[video.videoType] || '影片'}</i>
              </div>
              <div className="vod-card__body">
                <h3>{video.title}</h3>
                <p>{video.subtitle || video.videoCode}</p>
                <div>
                  <span>{video.year || '年份未知'}</span>
                  <span>{Math.round((video.duration || 0) / 60)} 分钟</span>
                </div>
              </div>
            </Link>
          ))}
          {!videos.length && <p className="empty-state">当前分类暂无影片，点击“添加影片”开始录入。</p>}
        </div>
      </section>
    </section>
  )
}

export function VideoCreatePage() {
  const navigate = useNavigate()
  const [categories, setCategories] = useState([])
  const [error, setError] = useState('')
  const [form, setForm] = useState(createVideoForm())

  useEffect(() => {
    api('/api/v1/video-categories')
      .then((data) => setCategories(buildTree((data.categories || []).map(normalizeCategory).filter(Boolean))))
      .catch((err) => setError(err.message))
  }, [])

  const change = (e) => setForm((current) => ({ ...current, [e.target.name]: e.target.value }))
  const changePoster = (name, value) => setForm((current) => ({ ...current, [name]: value }))

  async function submit(event) {
    event.preventDefault()
    try {
      const saved = await api('/api/v1/videos/create', {
        method: 'POST',
        body: JSON.stringify(buildVideoPayload(form)),
      })
      navigate(`/videos/${saved.id}`)
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <section className="panel panel--form">
      <div className="panel__header">
        <div className="card__header">
          <h2>添加影片</h2>
          <p className="subtitle">录入标题级元数据，保存后继续添加节目。</p>
        </div>
        <Link className="ghost-button action-link" to="/videos">
          返回影片库
        </Link>
      </div>
      {error && <p className="message message--error">{error}</p>}
      <form className="form form--spacious" onSubmit={submit}>
        <div className="form-grid">
          {field('categoryId', '所属分类', form.categoryId, change, {
            choices: [{ value: '', label: '请选择分类' }, ...categories.map((c) => ({ value: c.id, label: `${'　'.repeat(c.depth)}${c.name}` }))],
          })}
          {field('videoType', '影片类型', form.videoType, change, {
            choices: [
              { value: '1', label: '电影' },
              { value: '2', label: '剧集' },
              { value: '3', label: '综艺' },
              { value: '4', label: '其他' },
            ],
          })}
          {field('videoCode', '影片编码', form.videoCode, change, { placeholder: '全局唯一业务编码' })}
          {field('title', '影片标题', form.title, change)}
          {field('subtitle', '副标题', form.subtitle, change)}
          <ImageUploadField
            helper="上传后自动生成海报地址，适合列表卡片与详情首屏。"
            label="竖版海报"
            onChange={(value) => changePoster('posterVerticalUrl', value)}
            value={form.posterVerticalUrl}
          />
          <ImageUploadField
            helper="建议上传横版宣传图，详情页或运营位可直接复用。"
            label="横版海报"
            onChange={(value) => changePoster('posterHorizontalUrl', value)}
            shape="landscape"
            value={form.posterHorizontalUrl}
          />
          <ImageUploadField
            helper="上传后自动写入缩略图地址，适合列表、推荐位和轻量封面。"
            label="缩略图"
            onChange={(value) => changePoster('thumbnailUrl', value)}
            shape="landscape"
            value={form.thumbnailUrl}
          />
          {field('year', '出品年份', form.year, change, { type: 'number' })}
          {field('duration', '总时长（秒）', form.duration, change, { type: 'number' })}
          {field('status', '状态', form.status, change, {
            choices: [
              { value: '1', label: '正常' },
              { value: '0', label: '禁用' },
            ],
          })}
          {field('description', '影片简介', form.description, change, { type: 'textarea', full: true })}
        </div>
        <div className="button-row">
          <button className="submit submit--compact" disabled={!form.categoryId || !form.videoCode.trim() || !form.title.trim()} type="submit">
            保存并添加节目
          </button>
        </div>
      </form>
    </section>
  )
}

function MediaPanel({ episode, videoId }) {
  const [media, setMedia] = useState([])
  const [open, setOpen] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    mediaId: '',
    mediaUrl: '',
    fileFormat: 'mp4',
    bitrate: '',
    resolution: '',
    fileSize: '',
    duration: '',
    status: '1',
  })

  const load = useCallback(
    () =>
      api(`/api/v1/media?episode_id=${episode.id}`).then((data) =>
        setMedia((data.media || []).map(normalizeMedia).filter(Boolean)),
      ),
    [episode.id],
  )

  useEffect(() => {
    load().catch((err) => setError(err.message))
  }, [load])

  const change = (e) => setForm((current) => ({ ...current, [e.target.name]: e.target.value }))

  async function submit(event) {
    event.preventDefault()
    setError('')
    try {
      await api('/api/v1/media/create', {
        method: 'POST',
        body: JSON.stringify({
          video_id: videoId,
          episode_id: episode.id,
          media_id: form.mediaId.trim(),
          media_url: form.mediaUrl.trim(),
          file_format: form.fileFormat.trim(),
          bitrate: Number(form.bitrate) || 0,
          resolution: form.resolution.trim(),
          file_size: Number(form.fileSize) || 0,
          duration: Number(form.duration) || 0,
          status: Number(form.status),
        }),
      })
      setForm({
        mediaId: '',
        mediaUrl: '',
        fileFormat: 'mp4',
        bitrate: '',
        resolution: '',
        fileSize: '',
        duration: '',
        status: '1',
      })
      setOpen(false)
      await load()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <article className="episode-card">
      <header>
        <div>
          <span>EP {String(episode.episodeNo).padStart(2, '0')}</span>
          <h3>{episode.title}</h3>
          <p>{episode.description || '暂无节目简介'}</p>
        </div>
        <button className="ghost-button" onClick={() => setOpen((v) => !v)} type="button">
          {open ? '收起' : '添加媒体'}
        </button>
      </header>
      {error && <p className="message message--error">{error}</p>}
      <div className="media-list">
        {media.map((item) => (
          <div className="media-row" key={item.id}>
            <strong>{item.mediaId}</strong>
            <span>
              {item.fileFormat || '-'} · {item.resolution || '-'} · {item.bitrate || 0} kbps
            </span>
            <a href={item.mediaUrl} rel="noreferrer" target="_blank">
              打开媒体
            </a>
          </div>
        ))}
        {!media.length && <p>尚未添加可播媒体。</p>}
      </div>
      {open && (
        <form className="media-form" onSubmit={submit}>
          <input name="mediaId" onChange={change} placeholder="媒体编码 *" value={form.mediaId} />
          <input name="mediaUrl" onChange={change} placeholder="媒体 URL *" value={form.mediaUrl} />
          <input name="fileFormat" onChange={change} placeholder="格式" value={form.fileFormat} />
          <input name="resolution" onChange={change} placeholder="分辨率" value={form.resolution} />
          <input name="bitrate" onChange={change} placeholder="码率 kbps" type="number" value={form.bitrate} />
          <input name="duration" onChange={change} placeholder="时长 秒" type="number" value={form.duration} />
          <input name="fileSize" onChange={change} placeholder="文件大小 字节" type="number" value={form.fileSize} />
          <button className="submit submit--compact" disabled={!form.mediaId.trim() || !form.mediaUrl.trim()} type="submit">
            保存媒体
          </button>
        </form>
      )}
    </article>
  )
}

export function VideoDetailPage() {
  const { id } = useParams()
  const [video, setVideo] = useState(null)
  const [categories, setCategories] = useState([])
  const [episodes, setEpisodes] = useState([])
  const [error, setError] = useState('')
  const [videoForm, setVideoForm] = useState(createVideoForm())
  const [videoFormError, setVideoFormError] = useState('')
  const [videoFormSuccess, setVideoFormSuccess] = useState('')
  const [videoSubmitting, setVideoSubmitting] = useState(false)
  const [form, setForm] = useState({ episodeNo: '1', title: '', duration: '', description: '', status: '1' })

  const load = useCallback(async () => {
    const [categoryData, videoData, episodeData] = await Promise.all([
      api('/api/v1/video-categories'),
      api(`/api/v1/videos/${id}`),
      api(`/api/v1/episodes?video_id=${id}`),
    ])
    const nextVideo = normalizeVideo(videoData)
    setCategories(buildTree((categoryData.categories || []).map(normalizeCategory).filter(Boolean)))
    setVideo(nextVideo)
    setVideoForm(createVideoForm(nextVideo))
    setEpisodes((episodeData.episodes || []).map(normalizeEpisode).filter(Boolean))
  }, [id])

  useEffect(() => {
    load().catch((err) => setError(err.message))
  }, [load])

  const change = (e) => setForm((current) => ({ ...current, [e.target.name]: e.target.value }))
  const changeVideo = (e) => {
    setVideoFormSuccess('')
    setVideoForm((current) => ({ ...current, [e.target.name]: e.target.value }))
  }
  const changeVideoPoster = (name, value) => {
    setVideoFormSuccess('')
    setVideoForm((current) => ({ ...current, [name]: value }))
  }

  async function submitVideo(event) {
    event.preventDefault()
    setVideoSubmitting(true)
    setVideoFormError('')
    setVideoFormSuccess('')
    try {
      const saved = normalizeVideo(
        await api('/api/v1/videos/update', {
          method: 'PUT',
          body: JSON.stringify(buildVideoPayload(videoForm)),
        }),
      )
      setVideo(saved)
      setVideoForm(createVideoForm(saved))
      setVideoFormSuccess('影片基础信息已更新。')
    } catch (err) {
      setVideoFormError(err.message || '影片信息更新失败')
    } finally {
      setVideoSubmitting(false)
    }
  }

  async function submit(event) {
    event.preventDefault()
    try {
      await api('/api/v1/episodes/create', {
        method: 'POST',
        body: JSON.stringify({
          video_id: Number(id),
          episode_no: Number(form.episodeNo),
          title: form.title.trim(),
          duration: Number(form.duration) || 0,
          description: form.description.trim(),
          status: Number(form.status),
        }),
      })
      setForm((current) => ({
        ...current,
        episodeNo: String(Number(current.episodeNo) + 1),
        title: '',
        duration: '',
        description: '',
      }))
      await load()
    } catch (err) {
      setError(err.message)
    }
  }

  if (!video) {
    return (
      <section className="panel">
        <p className="empty-state">{error || '正在加载影片详情...'}</p>
      </section>
    )
  }

  return (
    <section className="vod-detail">
      <header className="panel vod-hero">
        <div className="vod-hero__poster">
          {video.posterVerticalUrl ? <img alt="" src={video.posterVerticalUrl} /> : <span>{(video.title || '?').slice(0, 1)}</span>}
        </div>
        <div>
          <Link className="side-card__label" to="/videos">
            ← 返回影片库
          </Link>
          <h1>{video.title}</h1>
          <p>{video.subtitle || video.videoCode}</p>
          <div className="vod-hero__meta">
            <span>{video.year || '年份未知'}</span>
            <span>{Math.round((video.duration || 0) / 60)} 分钟</span>
            <span>{episodes.length} 个节目</span>
          </div>
          <p>{video.description || '暂无影片简介。'}</p>
        </div>
      </header>
      {error && <p className="message message--error">{error}</p>}
      <div className="vod-detail__grid">
        <section className="episode-stack">
          <div className="vod-section-title">
            <div>
              <span className="side-card__label">PROGRAMS</span>
              <h2>节目与媒体</h2>
            </div>
            <span className="vod-count">{episodes.length}</span>
          </div>
          {episodes.map((episode) => (
            <MediaPanel episode={episode} key={episode.id} videoId={Number(id)} />
          ))}
          {!episodes.length && <p className="panel empty-state">还没有节目，请先在右侧添加。</p>}
        </section>
        <aside className="vod-detail__side">
          <section className="panel episode-create">
            <h2>编辑影片</h2>
            <p className="subtitle">海报改为直接上传图片，保存后立即更新当前影片资料。</p>
            {videoFormError ? <p className="message message--error">{videoFormError}</p> : null}
            {videoFormSuccess ? <p className="message message--success">{videoFormSuccess}</p> : null}
            <form className="form form--spacious" onSubmit={submitVideo}>
              {field('categoryId', '所属分类', videoForm.categoryId, changeVideo, {
                choices: [{ value: '', label: '请选择分类' }, ...categories.map((c) => ({ value: c.id, label: `${'　'.repeat(c.depth)}${c.name}` }))],
              })}
              {field('videoType', '影片类型', videoForm.videoType, changeVideo, {
                choices: [
                  { value: '1', label: '电影' },
                  { value: '2', label: '剧集' },
                  { value: '3', label: '综艺' },
                  { value: '4', label: '其他' },
                ],
              })}
              {field('videoCode', '影片编码', videoForm.videoCode, changeVideo)}
              {field('title', '影片标题', videoForm.title, changeVideo)}
              {field('subtitle', '副标题', videoForm.subtitle, changeVideo)}
              <ImageUploadField
                helper="用于竖版卡片封面。"
                label="竖版海报"
                onChange={(value) => changeVideoPoster('posterVerticalUrl', value)}
                value={videoForm.posterVerticalUrl}
              />
              <ImageUploadField
                helper="用于横版宣传位或详情延展位。"
                label="横版海报"
                onChange={(value) => changeVideoPoster('posterHorizontalUrl', value)}
                shape="landscape"
                value={videoForm.posterHorizontalUrl}
              />
              <ImageUploadField
                helper="用于推荐位和轻量封面展示。"
                label="缩略图"
                onChange={(value) => changeVideoPoster('thumbnailUrl', value)}
                shape="landscape"
                value={videoForm.thumbnailUrl}
              />
              {field('year', '出品年份', videoForm.year, changeVideo, { type: 'number' })}
              {field('duration', '总时长（秒）', videoForm.duration, changeVideo, { type: 'number' })}
              {field('status', '状态', videoForm.status, changeVideo, {
                choices: [
                  { value: '1', label: '正常' },
                  { value: '0', label: '禁用' },
                ],
              })}
              {field('description', '影片简介', videoForm.description, changeVideo, { type: 'textarea' })}
              <button
                className="submit submit--compact"
                disabled={!videoForm.categoryId || !videoForm.videoCode.trim() || !videoForm.title.trim() || videoSubmitting}
                type="submit"
              >
                {videoSubmitting ? '保存中...' : '保存影片'}
              </button>
            </form>
          </section>

          <section className="panel episode-create">
            <h2>添加节目</h2>
            <p className="subtitle">节目保存后，可在节目卡片中继续添加不同码率、格式的媒体。</p>
            <form className="form" onSubmit={submit}>
              {field('episodeNo', '集序号', form.episodeNo, change, { type: 'number' })}
              {field('title', '节目标题', form.title, change)}
              {field('duration', '时长（秒）', form.duration, change, { type: 'number' })}
              {field('description', '节目简介', form.description, change, { type: 'textarea' })}
              <button className="submit submit--compact" disabled={!form.title.trim() || Number(form.episodeNo) < 1} type="submit">
                保存节目
              </button>
            </form>
          </section>
        </aside>
      </div>
    </section>
  )
}
