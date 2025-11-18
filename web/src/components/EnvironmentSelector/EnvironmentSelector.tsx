import { Language, LANGUAGE_CONFIGS, Standard } from '@/types/api.ts'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/select'
import { Code2, Settings2 } from 'lucide-react'

interface EnvironmentSelectorProps {
  selectedLanguage: Language
  selectedStandard?: Standard
  onLanguageChange: (language: Language) => void
  onStandardChange?: (standard: Standard) => void
  disabled?: boolean
}

/**
 * Environment selector component with shadcn/ui Select
 */
export function EnvironmentSelector({
  selectedLanguage,
  selectedStandard,
  onLanguageChange,
  onStandardChange,
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
