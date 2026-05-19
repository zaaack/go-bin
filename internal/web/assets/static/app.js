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

  function renderState() {
    const file = fileInput.files && fileInput.files[0]
    const text = textarea.value.trim()

    if (file) {
      submit.textContent = "上传文件"
      status.textContent = `当前将上传文件：${file.name}`
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
    setComposeFile(fileInput, file)
    renderState()
    form.requestSubmit()
  }

  function syncOptions() {
    publicHidden.value = publicInput.checked ? "1" : "0"
    expireHidden.value = expireInput.value
  }

  function firstFile(fileList) {
    if (!fileList || fileList.length === 0) {
      return null
    }
    return fileList[0]
  }

  picker.addEventListener("click", () => fileInput.click())

  fileInput.addEventListener("change", () => {
    renderState()
    if (firstFile(fileInput.files)) {
      form.requestSubmit()
    }
  })

  textarea.addEventListener("input", renderState)
  publicInput.addEventListener("change", syncOptions)
  expireInput.addEventListener("change", syncOptions)

  textarea.addEventListener("paste", (event) => {
    const file = firstFile(event.clipboardData && event.clipboardData.files)
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

    const file = firstFile(event.dataTransfer && event.dataTransfer.files)
    if (!file) {
      return
    }

    submitFile(file)
  })

  form.addEventListener("submit", (event) => {
    syncOptions()

    const file = firstFile(fileInput.files)
    const text = textarea.value
    const trimmed = text.trim()

    if (file) {
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
