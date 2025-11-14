import { Editor } from '@monaco-editor/react'
import { Language } from '../../types/api'

interface CodeEditorProps {
  value: string
  onChange: (value: string) => void
  language: Language
  readOnly?: boolean
  height?: string
}

/**
 * Code editor component using Monaco Editor
 */
export function CodeEditor({
  value,
  onChange,
  language,
  readOnly = false,
  height = '500px',
}: CodeEditorProps) {
  // Map our language types to Monaco's language identifiers
  const monacoLanguage = mapLanguage(language)

  const handleEditorChange = (value: string | undefined) => {
    onChange(value || '')
  }

  return (
    <div className="border border-gray-300 rounded-lg overflow-hidden shadow-sm">
      <Editor
        height={height}
        language={monacoLanguage}
        value={value}
        onChange={handleEditorChange}
        theme="vs-dark"
        options={{
          readOnly,
          minimap: { enabled: true },
          fontSize: 14,
          lineNumbers: 'on',
          roundedSelection: true,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 4,
          insertSpaces: true,
          wordWrap: 'on',
          formatOnPaste: true,
          formatOnType: true,
        }}
        loading={
          <div className="flex items-center justify-center h-full bg-gray-900">
            <div className="text-white">Loading editor...</div>
          </div>
        }
      />
    </div>
  )
}

/**
 * Map our language type to Monaco editor language identifier
 */
function mapLanguage(language: Language): string {
  switch (language) {
    case 'c':
      return 'c'
    case 'cpp':
    case 'c++':
      return 'cpp'
    case 'go':
      return 'go'
    case 'rust':
      return 'rust'
    default:
      return 'plaintext'
  }
}
