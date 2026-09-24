import React from 'react'
import ReactDOM from 'react-dom/client'
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material'
import App from './App'
import './styles.css'

const theme = createTheme({
  palette: { mode: 'light', primary: { main: '#176c64' }, secondary: { main: '#a55835' }, background: { default: '#eef1ed', paper: '#f8faf7' }, text: { primary: '#1c2826', secondary: '#5e6966' }, error: { main: '#ad3d36' }, warning: { main: '#a76b12' } },
  shape: { borderRadius: 6 },
  typography: { fontFamily: '"Avenir Next", "Trebuchet MS", sans-serif', button: { textTransform: 'none', fontWeight: 700, letterSpacing: 0 } },
  components: {
    MuiButton: { defaultProps: { disableElevation: true }, styleOverrides: { root: { minHeight: 38, borderRadius: 4 } } },
    MuiPaper: { styleOverrides: { root: { backgroundImage: 'none', boxShadow: 'none' } } },
    MuiTextField: { defaultProps: { size: 'small' } },
    MuiSelect: { defaultProps: { size: 'small' } },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(<React.StrictMode><ThemeProvider theme={theme}><CssBaseline /><App /></ThemeProvider></React.StrictMode>)
