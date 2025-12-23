import { useState } from 'react'
import { uploadImage, verifyImportAccount, importImages } from '../api/images'
import type { ImageMeta } from '../types'
import { useImageListStore } from '../store/imageListStore'
import * as EXIF from 'exif-js'
import { format } from 'date-fns'
import './UploadPage.css'

const UploadPage = () => {
  const setHasNewImages = useImageListStore((state) => state.setHasNewImages)
  
  // 上传相关状态
  const [files, setFiles] = useState<File[]>([])
  const [uploadMode, setUploadMode] = useState<'single' | 'multiple' | 'folder'>('single')
  const [tags, setTags] = useState('')
  const [autoTags, setAutoTags] = useState<string[]>([])
  const [message, setMessage] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState<{ current: number; total: number } | null>(null)
  const [useAI, setUseAI] = useState(true) // 默认启用AI自动生成标签
  
  // 导入相关状态
  const [activeTab, setActiveTab] = useState<'upload' | 'import'>('upload')
  const [importUsername, setImportUsername] = useState('')
  const [importPassword, setImportPassword] = useState('')
  const [importImagesList, setImportImagesList] = useState<ImageMeta[]>([])
  const [selectedImageIds, setSelectedImageIds] = useState<Set<number>>(new Set())
  const [importLoading, setImportLoading] = useState(false)
  const [importMessage, setImportMessage] = useState<string | null>(null)

  const extractEXIFTags = (file: File): Promise<string[]> => {
    return new Promise((resolve) => {
      EXIF.getData(file as any, function(this: any) {
        const tags: string[] = []
        
        // 相机品牌
        const make = EXIF.getTag(this, 'Make')
        if (make) {
          tags.push(make.trim())
        }
        
        // 相机型号
        const model = EXIF.getTag(this, 'Model')
        if (model) {
          tags.push(model.trim())
        }
        
        resolve(tags)
      })
    })
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(e.target.files || [])
    
    // 过滤出图片文件
    const imageFiles = selectedFiles.filter(file => file.type.startsWith('image/'))
    
    if (imageFiles.length === 0) {
      setMessage('请选择图片文件')
      // 重置input，允许重新选择
      e.target.value = ''
      return
    }
    
    // 单文件模式下，只保留第一个文件
    const finalFiles = uploadMode === 'single' ? [imageFiles[0]] : imageFiles
    
    setFiles(finalFiles)
    setAutoTags([])
    setMessage(null)
    
    // 如果是单文件模式，提取第一个文件的EXIF标签
    if (uploadMode === 'single' && finalFiles.length > 0) {
      try {
        const extractedTags = await extractEXIFTags(finalFiles[0])
        setAutoTags(extractedTags)
      } catch (err) {
        console.error('Failed to extract EXIF tags:', err)
      }
    }
    
    // 重置input，允许重新选择
    e.target.value = ''
  }
  
  const handleRemoveFile = (index: number) => {
    const newFiles = files.filter((_, i) => i !== index)
    setFiles(newFiles)
  }
  
  const handleClearAll = () => {
    setFiles([])
    setAutoTags([])
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (files.length === 0) {
      setMessage('请选择文件')
      return
    }
    setLoading(true)
    setMessage(null)
    setUploadProgress({ current: 0, total: files.length })
    
    try {
      // 解析标签字符串，支持中英文逗号分隔
      // 先统一将中文逗号替换为英文逗号，然后按英文逗号分割
      const normalizedTags = tags.replace(/，/g, ',')
      const manualTags = normalizedTags
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean)
      
      // 合并自动生成的标签和手动输入的标签，去重
      const allTags = [...new Set([...autoTags, ...manualTags])]
      
      // 批量上传
      let successCount = 0
      let failCount = 0
      
      for (let i = 0; i < files.length; i++) {
        try {
          await uploadImage(files[i], allTags, useAI)
          successCount++
        } catch (err: any) {
          console.error(`上传文件 ${files[i].name} 失败:`, err)
          failCount++
        }
        setUploadProgress({ current: i + 1, total: files.length })
      }
      
      if (failCount === 0) {
        setMessage(`成功上传 ${successCount} 张图片`)
      } else {
        setMessage(`上传完成：成功 ${successCount} 张，失败 ${failCount} 张`)
      }
      
      // 清空状态
      setFiles([])
      setTags('')
      setAutoTags([])
      setUploadProgress(null)
      // 标记有新图片上传，当用户切换到图片列表页面时会自动刷新
      setHasNewImages(true)
    } catch (err: any) {
      setMessage(err.response?.data?.message ?? '上传失败')
      setUploadProgress(null)
    } finally {
      setLoading(false)
    }
  }

  // 导入相关处理函数
  const handleVerifyAccount = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!importUsername || !importPassword) {
      setImportMessage('请输入用户名和密码')
      return
    }
    setImportLoading(true)
    setImportMessage(null)
    try {
      const result = await verifyImportAccount(importUsername, importPassword)
      setImportImagesList(result.images)
      // 默认选中所有图片
      setSelectedImageIds(new Set(result.images.map((img) => img.id)))
      setImportMessage(`找到 ${result.images.length} 张图片`)
    } catch (err: any) {
      setImportMessage(err.response?.data?.message ?? '验证失败')
      setImportImagesList([])
      setSelectedImageIds(new Set())
    } finally {
      setImportLoading(false)
    }
  }

  const handleToggleImageSelection = (imageId: number) => {
    const newSelected = new Set(selectedImageIds)
    if (newSelected.has(imageId)) {
      newSelected.delete(imageId)
    } else {
      newSelected.add(imageId)
    }
    setSelectedImageIds(newSelected)
  }

  const handleSelectAll = () => {
    if (selectedImageIds.size === importImagesList.length) {
      setSelectedImageIds(new Set())
    } else {
      setSelectedImageIds(new Set(importImagesList.map((img) => img.id)))
    }
  }

  const handleImport = async () => {
    if (selectedImageIds.size === 0) {
      setImportMessage('请至少选择一张图片')
      return
    }
    if (!importUsername || !importPassword) {
      setImportMessage('请输入用户名和密码')
      return
    }
    setImportLoading(true)
    setImportMessage(null)
    try {
      const result = await importImages(importUsername, importPassword, Array.from(selectedImageIds))
      setImportMessage(result.message)
      // 清空状态
      setImportImagesList([])
      setSelectedImageIds(new Set())
      setImportUsername('')
      setImportPassword('')
      // 标记有新图片导入，当用户切换到图片列表页面时会自动刷新
      setHasNewImages(true)
    } catch (err: any) {
      setImportMessage(err.response?.data?.message ?? '导入失败')
    } finally {
      setImportLoading(false)
    }
  }

  return (
    <div className="upload-page">
      <div className="tab-container">
        <button
          className={`tab-button ${activeTab === 'upload' ? 'active' : ''}`}
          onClick={() => setActiveTab('upload')}
        >
          上传图片
        </button>
        <button
          className={`tab-button ${activeTab === 'import' ? 'active' : ''}`}
          onClick={() => setActiveTab('import')}
        >
          从其他账户导入
        </button>
      </div>

      {activeTab === 'upload' && (
        <form className="upload-card" onSubmit={handleSubmit}>
          <h2>上传图片</h2>
          {message && <div className="upload-message">{message}</div>}
          
          {uploadProgress && (
            <div className="upload-progress">
              <div className="progress-bar">
                <div 
                  className="progress-fill" 
                  style={{ width: `${(uploadProgress.current / uploadProgress.total) * 100}%` }}
                />
              </div>
              <span className="progress-text">
                上传进度: {uploadProgress.current} / {uploadProgress.total}
              </span>
            </div>
          )}

          <div className="upload-mode-selector">
            <label className="mode-option">
              <input
                type="radio"
                name="uploadMode"
                value="single"
                checked={uploadMode === 'single'}
                onChange={(e) => {
                  setUploadMode(e.target.value as 'single')
                  setFiles([])
                }}
              />
              <span>单张图片</span>
            </label>
            <label className="mode-option">
              <input
                type="radio"
                name="uploadMode"
                value="multiple"
                checked={uploadMode === 'multiple'}
                onChange={(e) => {
                  setUploadMode(e.target.value as 'multiple')
                  setFiles([])
                }}
              />
              <span>多张图片</span>
            </label>
            <label className="mode-option">
              <input
                type="radio"
                name="uploadMode"
                value="folder"
                checked={uploadMode === 'folder'}
                onChange={(e) => {
                  setUploadMode(e.target.value as 'folder')
                  setFiles([])
                }}
              />
              <span>整个文件夹</span>
            </label>
          </div>

          <label className="file-drop">
            <input
              key={uploadMode} // 当模式改变时，强制重新渲染input以重置状态
              type="file"
              accept="image/*"
              multiple={uploadMode !== 'single'}
              {...(uploadMode === 'folder' ? { 
                webkitdirectory: 'true' as any,
                directory: true as any 
              } : {})}
              onChange={handleFileChange}
            />
            {files.length === 0 ? (
              <span>
                {uploadMode === 'single' && '点击或拖拽图片至此（最大10MB，支持 JPEG, PNG, GIF, BMP, TIFF, WebP）'}
                {uploadMode === 'multiple' && '点击选择多张图片或拖拽图片至此'}
                {uploadMode === 'folder' && '点击选择文件夹（将上传文件夹下所有图片）'}
              </span>
            ) : (
              <span>已选择 {files.length} 个文件</span>
            )}
          </label>

          {files.length > 0 && (
            <div className="file-list">
              <div className="file-list-header">
                <span>已选择的文件 ({files.length})</span>
                <button
                  type="button"
                  className="clear-all-btn"
                  onClick={handleClearAll}
                >
                  清空
                </button>
              </div>
              <div className="file-list-content">
                {files.map((file, index) => (
                  <div key={index} className="file-item">
                    <span className="file-name" title={file.name}>{file.name}</span>
                    <span className="file-size">
                      {(file.size / 1024 / 1024).toFixed(2)} MB
                    </span>
                    <button
                      type="button"
                      className="remove-file-btn"
                      onClick={() => handleRemoveFile(index)}
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {autoTags.length > 0 && (
            <div className="auto-tags-info">
              <span>根据EXIF信息自动添加标签：</span>
              <div className="auto-tags">
                {autoTags.map((tag, index) => (
                  <span key={index} className="auto-tag">{tag}</span>
                ))}
              </div>
            </div>
          )}

          <label className="ai-toggle-label">
            <input
              type="checkbox"
              checked={useAI}
              onChange={(e) => setUseAI(e.target.checked)}
              className="ai-toggle"
            />
            <span>使用AI自动生成标签</span>
          </label>

          <label>
            自定义标签（用逗号分隔，支持中英文逗号）
            <input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="例：旅行, 北京, 朋友 或 旅行，北京，朋友"
            />
          </label>

          <button type="submit" disabled={loading}>
            {loading ? '上传中...' : '上传'}
          </button>
        </form>
      )}

      {activeTab === 'import' && (
        <div className="import-card">
          <h2>从其他账户导入图片</h2>
          {importMessage && <div className={`upload-message ${importMessage.includes('成功') ? 'success' : ''}`}>{importMessage}</div>}

          <form onSubmit={handleVerifyAccount}>
            <label>
              用户名或邮箱
              <input
                type="text"
                value={importUsername}
                onChange={(e) => setImportUsername(e.target.value)}
                placeholder="请输入要导入的账户用户名或邮箱"
              />
            </label>
            <label>
              密码
              <input
                type="password"
                value={importPassword}
                onChange={(e) => setImportPassword(e.target.value)}
                placeholder="请输入账户密码"
              />
            </label>
            <button type="submit" disabled={importLoading}>
              {importLoading ? '验证中...' : '验证并查看图片'}
            </button>
          </form>

          {importImagesList.length > 0 && (
            <div className="import-images-section">
              <div className="import-controls">
                <button
                  type="button"
                  className="select-all-btn"
                  onClick={handleSelectAll}
                >
                  {selectedImageIds.size === importImagesList.length ? '取消全选' : '全选'}
                </button>
                <span className="selection-count">
                  已选择 {selectedImageIds.size} / {importImagesList.length} 张
                </span>
                <button
                  type="button"
                  className="import-btn"
                  onClick={handleImport}
                  disabled={importLoading || selectedImageIds.size === 0}
                >
                  {importLoading ? '导入中...' : '导入选中图片'}
                </button>
              </div>
              <div className="import-image-grid">
                {importImagesList.map((image) => {
                  const thumbnailUrl = `${import.meta.env.VITE_API_BASE_URL ?? '/api/v1'}/images/${image.id}/thumbnail`
                  const isSelected = selectedImageIds.has(image.id)
                  return (
                    <div
                      key={image.id}
                      className={`import-image-item ${isSelected ? 'selected' : ''}`}
                      onClick={() => handleToggleImageSelection(image.id)}
                    >
                      <div className="import-image-wrapper">
                        <img src={thumbnailUrl} alt={image.originalFilename} />
                        <div className="import-checkbox">
                          {isSelected && <span className="checkmark">✓</span>}
                        </div>
                      </div>
                      <div className="import-image-info">
                        <div className="import-image-name" title={image.originalFilename}>
                          {image.originalFilename}
                        </div>
                        <div className="import-image-date">
                          {format(new Date(image.createdAt), 'yyyy-MM-dd HH:mm')}
                        </div>
                        {image.tags && image.tags.length > 0 && (
                          <div className="import-image-tags">
                            {image.tags.slice(0, 3).map((tag) => (
                              <span
                                key={tag.id}
                                className="import-tag-badge"
                                style={{ backgroundColor: tag.color ?? '#e0e7ff' }}
                              >
                                {tag.name}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default UploadPage

