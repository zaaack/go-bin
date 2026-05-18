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
