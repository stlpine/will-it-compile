import {
  Language,
  LANGUAGE_CONFIGS,
  Standard,
  Compiler,
  Environment,
  getCompilersForLanguage,
  CompilerInfo,
} from '@/types/api.ts'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/select'
import { Code2, Settings2, Cpu } from 'lucide-react'

interface EnvironmentSelectorProps {
  selectedLanguage: Language
  selectedCompiler?: Compiler
  selectedStandard?: Standard
  onLanguageChange: (language: Language) => void
  onCompilerChange?: (compiler: Compiler) => void
  onStandardChange?: (standard: Standard) => void
  environments?: Environment[]
  disabled?: boolean
}

/**
 * Environment selector component with shadcn/ui Select
 */
export function EnvironmentSelector({
  selectedLanguage,
  selectedCompiler,
  selectedStandard,
  onLanguageChange,
  onCompilerChange,
  onStandardChange,
  environments = [],
  disabled = false,
}: EnvironmentSelectorProps) {
  const languages = Object.keys(LANGUAGE_CONFIGS)
  const showStandardSelector =
    selectedLanguage === 'cpp' || selectedLanguage === 'c++' || selectedLanguage === 'c'

  const cppStandards: Standard[] = ['c++11', 'c++14', 'c++17', 'c++20', 'c++23']
  const cStandards: Standard[] = ['c89', 'c99', 'c11', 'c17', 'c23']

  // Determine which standards to show based on language
  const standards =
    selectedLanguage === 'c' ? cStandards : cppStandards
  const standardLabel =
    selectedLanguage === 'c' ? 'C Standard' : 'C++ Standard'
  const defaultStandard =
    selectedLanguage === 'c' ? 'c17' : 'c++20'

  // Get available compilers for the selected language
  const availableCompilers: CompilerInfo[] = getCompilersForLanguage(
    environments,
    selectedLanguage
  )

  // Get default compiler for the language
  const defaultCompiler = LANGUAGE_CONFIGS[selectedLanguage]?.compiler || ''

  return (
    <div className="flex flex-col sm:flex-row gap-4">
      {/* Language Selector */}
      <div className="flex-1">
        <label className="block text-sm font-medium mb-2 flex items-center gap-2">
          <Code2 className="h-4 w-4" />
          Language
        </label>
        <Select
          value={selectedLanguage}
          onValueChange={(value) => onLanguageChange(value as Language)}
          disabled={disabled}
        >
          <SelectTrigger className="w-full">
            <SelectValue placeholder="Select a language" />
          </SelectTrigger>
          <SelectContent>
            {languages.map((lang) => (
              <SelectItem key={lang} value={lang}>
                {LANGUAGE_CONFIGS[lang].label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Compiler Version Selector */}
      {onCompilerChange && availableCompilers.length > 0 && (
        <div className="flex-1">
          <label className="block text-sm font-medium mb-2 flex items-center gap-2">
            <Cpu className="h-4 w-4" />
            Compiler Version
          </label>
          <Select
            value={selectedCompiler || defaultCompiler}
            onValueChange={(value) => onCompilerChange(value as Compiler)}
            disabled={disabled}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select compiler version" />
            </SelectTrigger>
            <SelectContent>
              {availableCompilers.map((compiler) => (
                <SelectItem key={compiler.id} value={compiler.id}>
                  {compiler.version}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      {/* Standard Selector (C/C++ only) */}
      {showStandardSelector && onStandardChange && (
        <div className="flex-1">
          <label className="block text-sm font-medium mb-2 flex items-center gap-2">
            <Settings2 className="h-4 w-4" />
            {standardLabel}
          </label>
          <Select
            value={selectedStandard || defaultStandard}
            onValueChange={(value) => onStandardChange(value as Standard)}
            disabled={disabled}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={`Select ${standardLabel.toLowerCase()}`} />
            </SelectTrigger>
            <SelectContent>
              {standards.map((std) => (
                <SelectItem key={std} value={std}>
                  {std}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}
    </div>
  )
}
