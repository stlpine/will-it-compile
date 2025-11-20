import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import tailwindcss from "@tailwindcss/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    // Enable minification
    minify: 'esbuild',

    // Target modern browsers for smaller bundles
    target: 'es2020',

    // Optimize chunks for better caching
    rollupOptions: {
      output: {
        manualChunks: {
          // Separate vendor chunks for better caching
          'react-vendor': ['react', 'react-dom'],
          'monaco-vendor': ['@monaco-editor/react'],
        },
      },
    },

    // Reduce chunk size warning limit (500kb default is too high)
    chunkSizeWarningLimit: 500,

    // Enable CSS code splitting
    cssCodeSplit: true,

    // Generate sourcemaps for production debugging (can disable to save ~30% size)
    sourcemap: false,

    // Optimize asset inlining (inline small assets as base64)
    assetsInlineLimit: 4096, // 4kb
  },

  // Optimize dependencies
  optimizeDeps: {
    include: ['react', 'react-dom', '@monaco-editor/react'],
  },
})
