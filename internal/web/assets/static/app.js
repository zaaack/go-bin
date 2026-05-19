function resolveCopyValue(button) {
  if (button.dataset.copyRaw !== undefined) {
    return button.dataset.copyRaw
  }

  const value = button.dataset.copy || ""
  if (button.dataset.copyAbsolute === "true") {
    return new URL(value, window.location.origin).toString()
  }
  return value
}

async function copyText(text) {
  await navigator.clipboard.writeText(text)
}

function toast(message, isError) {
  const node = document.createElement("div")
  node.className = "toast" + (isError ? " error" : "")
  node.textContent = message
  document.body.appendChild(node)
  requestAnimationFrame(() => node.classList.add("show"))
  setTimeout(() => {
    node.classList.remove("show")
    setTimeout(() => node.remove(), 180)
  }, 1800)
}

document.addEventListener("click", async (event) => {
  const button = event.target.closest("[data-copy],[data-copy-raw]")
  if (!button) {
    return
  }

  try {
    await copyText(resolveCopyValue(button))
    toast("已复制")
  } catch (error) {
    toast("复制失败", true)
  }
})

function isSingleURL(value) {
  const text = value.trim()
  if (!text || /\s/.test(text)) {
    return false
  }

  try {
    const parsed = new URL(text)
    return parsed.protocol === "http:" || parsed.protocol === "https:"
  } catch (error) {
    return false
  }
}

function setComposeFile(input, file) {
  const transfer = new DataTransfer()
  transfer.items.add(file)
  input.files = transfer.files
}

function initComposeForm() {
  const form = document.querySelector("[data-compose-form]")
  if (!form) {
    return
  }

  const singleFileMode = form.dataset.singleFile === "true"
  const textarea = form.querySelector("[data-compose-textarea]")
  const fileInput = form.querySelector("[data-compose-file]")
  const picker = form.querySelector("[data-compose-picker]")
  const submit = form.querySelector("[data-compose-submit]")
  const kindInput = form.querySelector("[data-compose-kind]")
  const textInput = form.querySelector("[data-compose-text]")
  const linkInput = form.querySelector("[data-compose-link]")
  const publicInput = form.querySelector("[data-compose-public]")
  const publicHidden = form.querySelector("[data-compose-public-hidden]")
  const expireInput = form.querySelector("[data-compose-expire]")
  const expireHidden = form.querySelector("[data-compose-expire-hidden]")
  const status = form.querySelector("[data-compose-status]")
  const fileList = form.querySelector("[data-compose-file-list]")

  let selectedFiles = []

  function renderFileList() {
    if (!fileList) return
    fileList.innerHTML = ""
    if (selectedFiles.length === 0 || singleFileMode) return

    selectedFiles.forEach((file, index) => {
      const item = document.createElement("div")
      item.className = "file-item"
      item.innerHTML = `
        <span class="file-name">${escapeHtml(file.name)}</span>
        <button type="button" class="button ghost button-sm" data-remove-file="${index}">×</button>
      `
      fileList.appendChild(item)
    })
  }

  function escapeHtml(text) {
    const div = document.createElement("div")
    div.textContent = text
    return div.innerHTML
  }

  function updateFileInput() {
    const transfer = new DataTransfer()
    selectedFiles.forEach(file => transfer.items.add(file))
    fileInput.files = transfer.files
  }

  function removeFile(index) {
    selectedFiles.splice(index, 1)
    updateFileInput()
    renderFileList()
    renderState()
  }

  function addFiles(files) {
    if (singleFileMode) {
      selectedFiles = [files[0]]
    } else {
      for (const file of files) {
        selectedFiles.push(file)
      }
    }
    updateFileInput()
    renderFileList()
    renderState()
  }

  function renderState() {
    const text = textarea.value.trim()

    if (selectedFiles.length > 0) {
      if (singleFileMode) {
        submit.textContent = "上传文件"
        status.textContent = `当前将上传文件：${selectedFiles[0].name}`
      } else {
        submit.textContent = `上传 ${selectedFiles.length} 个文件`
        status.textContent = `当前将上传 ${selectedFiles.length} 个文件`
      }
      return
    }

    if (isSingleURL(text)) {
      submit.textContent = "分享链接"
      status.textContent = "当前将分享链接"
      return
    }

    submit.textContent = "分享文本"
    status.textContent = text ? "当前将分享文本" : "输入文本，或选择 / 拖拽 / 粘贴文件后上传"
  }

  function submitFile(file) {
    selectedFiles = [file]
    updateFileInput()
    renderFileList()
    renderState()
    form.requestSubmit()
  }

  function syncOptions() {
    publicHidden.value = publicInput.checked ? "1" : "0"
    expireHidden.value = expireInput.value
  }

  picker.addEventListener("click", () => fileInput.click())

  fileInput.addEventListener("change", () => {
    if (fileInput.files.length > 0) {
      addFiles(fileInput.files)
      // In single file mode, auto-submit when a file is selected
      if (singleFileMode && selectedFiles.length > 0) {
        form.requestSubmit()
      }
    }
  })

  // Handle remove file button clicks
  if (fileList) {
    fileList.addEventListener("click", (event) => {
      const button = event.target.closest("[data-remove-file]")
      if (!button) return
      const index = parseInt(button.dataset.removeFile, 10)
      removeFile(index)
    })
  }

  textarea.addEventListener("input", renderState)
  publicInput.addEventListener("change", syncOptions)
  expireInput.addEventListener("change", syncOptions)

  textarea.addEventListener("paste", (event) => {
    const clipboardData = event.clipboardData
    if (!clipboardData) return

    // Try files first
    let file = clipboardData.files && clipboardData.files[0]
    
    // If no file, try items (for screenshot paste in some browsers)
    if (!file && clipboardData.items) {
      for (const item of clipboardData.items) {
        if (item.kind === "file") {
          file = item.getAsFile()
          if (file) break
        }
      }
    }

    if (!file) {
      return
    }

    event.preventDefault()
    submitFile(file)
  })

  textarea.addEventListener("dragover", (event) => {
    event.preventDefault()
    textarea.classList.add("drag-over")
  })

  textarea.addEventListener("dragleave", () => {
    textarea.classList.remove("drag-over")
  })

  textarea.addEventListener("drop", (event) => {
    event.preventDefault()
    textarea.classList.remove("drag-over")

    const files = event.dataTransfer && event.dataTransfer.files
    if (!files || files.length === 0) {
      return
    }

    addFiles(files)
    // In single file mode, auto-submit on drop
    if (singleFileMode && selectedFiles.length > 0) {
      form.requestSubmit()
    }
  })

  form.addEventListener("submit", (event) => {
    syncOptions()

    const text = textarea.value
    const trimmed = text.trim()

    if (selectedFiles.length > 0) {
      kindInput.value = "file"
      textInput.value = ""
      linkInput.value = ""
      return
    }

    if (!trimmed) {
      event.preventDefault()
      toast("请输入文本或选择文件", true)
      textarea.focus()
      return
    }

    if (isSingleURL(trimmed)) {
      kindInput.value = "link"
      linkInput.value = trimmed
      textInput.value = ""
      return
    }

    kindInput.value = "text"
    textInput.value = text
    linkInput.value = ""
  })

  renderState()
  syncOptions()
}

initComposeForm()
