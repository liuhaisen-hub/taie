import { useCallback, useEffect, useState } from 'react'

type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'theme-mode'

const getInitialTheme = (): ThemeMode => {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') {
        return stored
    }
    const attr = document.body.getAttribute('theme-mode')
    return attr === 'light' ? 'light' : 'dark'
}

export const useTheme = () => {
    const [theme, setTheme] = useState<ThemeMode>(getInitialTheme)

    useEffect(() => {
        document.body.setAttribute('theme-mode', theme)
        localStorage.setItem(STORAGE_KEY, theme)
    }, [theme])

    const toggleTheme = useCallback(() => {
        setTheme(prev => (prev === 'dark' ? 'light' : 'dark'))
    }, [])

    return { theme, toggleTheme }
}
