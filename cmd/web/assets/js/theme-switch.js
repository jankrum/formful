;(() => {
  // Constants
  const BUTTON_ID = 'theme-switch-button'
  const BUTTON_CLASS = 'button'
  const BUTTON_ATTRIBUTE_KEY = 'aria-label'
  const BUTTON_ATTRIBUTE_VALUE = 'Toggle theme'
  const MOON_ICON = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-moon-icon lucide-moon"><path d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"/></svg>`
  const SUN_ICON = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-sun-icon lucide-sun"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg>`
  const STORAGE_KEY = 'theme'
  const STORAGE_VALUE_LIGHT = 'light'
  const STORAGE_VALUE_DARK = 'dark'

  function main() {
    // Find or create the button
    const existing = document.getElementById(BUTTON_ID)
    const button = existing || document.createElement('button')
    if (!existing) {
      button.id = BUTTON_ID
      button.className = BUTTON_CLASS
      button.setAttribute(BUTTON_ATTRIBUTE_KEY, BUTTON_ATTRIBUTE_VALUE)
    }

    // Functions
    function isDark() {
      const theme = document.documentElement.getAttribute('data-theme')
      if (theme) return theme === STORAGE_VALUE_DARK
      return window.matchMedia('(prefers-color-scheme: dark)').matches
    }

    function updateIcon() {
      button.innerHTML = isDark() ? SUN_ICON : MOON_ICON
    }

    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) document.documentElement.setAttribute('data-theme', stored)

    button.addEventListener('click', () => {
      const next = isDark() ? STORAGE_VALUE_LIGHT : STORAGE_VALUE_DARK
      document.documentElement.setAttribute('data-theme', next)
      localStorage.setItem(STORAGE_KEY, next)
      updateIcon()
    })

    window.addEventListener('storage', event => {
      if (event.key !== STORAGE_KEY) return
      if (event.newValue) {
        document.documentElement.setAttribute('data-theme', event.newValue)
        updateIcon()
      } else {
        document.documentElement.removeAttribute('data-theme')
        updateIcon()
      }
    })

    updateIcon()
    if (!existing) document.body.append(button)
  }

  // Initialize on DOMContentLoaded
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', main)
  } else {
    main()
  }
})()
