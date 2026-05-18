;(() => {
  const STORAGE_KEY = 'theme'

  function main() {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) document.documentElement.setAttribute('data-theme', stored)

    window.addEventListener('storage', event => {
      if (event.key !== STORAGE_KEY) return
      if (event.newValue) {
        document.documentElement.setAttribute('data-theme', event.newValue)
      } else {
        document.documentElement.removeAttribute('data-theme')
      }
    })
  }

  // Initialize on DOMContentLoaded
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', main)
  } else {
    main()
  }
})()
