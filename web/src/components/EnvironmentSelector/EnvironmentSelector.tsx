import { Language, LANGUAGE_CONFIGS, Standard } from '../../types/api'

interface EnvironmentSelectorProps {
  selectedLanguage: Language
  selectedStandard?: Standard
  onLanguageChange: (language: Language) => void
  onStandardChange?: (standard: Standard) => void
  disabled?: boolean
}

/**
 * Environment selector component for choosing language and standard
 */
export function EnvironmentSelector({
  selectedLanguage,
  selectedStandard,
  onLanguageChange,
  onStandardChange,
  disabled = false,
}: EnvironmentSelectorProps) {
  const languages = Object.keys(LANGUAGE_CONFIGS)
  const showStandardSelector = selectedLanguage === 'cpp' || selectedLanguage === 'c++'

  const cppStandards: Standard[] = ['c++11', 'c++14', 'c++17', 'c++20', 'c++23']

  return (
    <div className="flex flex-col sm:flex-row gap-4">
      {/* Language Selector */}
      <div className="flex-1">
        <label
          htmlFor="language-select"
          className="block text-sm font-medium text-gray-700 mb-1"
        >
          Language
        </label>
        <select
          id="language-select"
          value={selectedLanguage}
          onChange={(e) => onLanguageChange(e.target.value as Language)}
          disabled={disabled}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
        >
          {languages.map((lang) => (
            <option key={lang} value={lang}>
              {LANGUAGE_CONFIGS[lang].label}
            </option>
          ))}
        </select>
      </div>

      {/* Standard Selector (C++ only) */}
      {showStandardSelector && onStandardChange && (
        <div className="flex-1">
          <label
            htmlFor="standard-select"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            C++ Standard
          </label>
          <select
            id="standard-select"
            value={selectedStandard || 'c++20'}
            onChange={(e) => onStandardChange(e.target.value as Standard)}
            disabled={disabled}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
          >
            {cppStandards.map((std) => (
              <option key={std} value={std}>
                {std}
              </option>
            ))}
          </select>
        </div>
      )}
    </div>
  )
}
