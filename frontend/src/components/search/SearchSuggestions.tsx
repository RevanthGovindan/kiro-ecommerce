'use client';

import React from 'react';

interface SearchSuggestionsProps {
  suggestions: string[];
  onSelect: (suggestion: string) => void;
}

export const SearchSuggestions: React.FC<SearchSuggestionsProps> = ({
  suggestions,
  onSelect,
}) => {
  if (suggestions.length === 0) {
    return null;
  }

  return (
    <div className="absolute top-full left-0 right-0 bg-white border border-gray-300 rounded-b-lg shadow-lg z-50 max-h-60 overflow-y-auto">
      {suggestions.map((suggestion, index) => (
        <button
          key={index}
          onClick={() => onSelect(suggestion)}
          className="w-full px-4 py-3 text-left hover:bg-gray-50 focus:bg-gray-50 focus:outline-none border-b border-gray-100 last:border-b-0"
        >
          <div className="flex items-center">
            <svg
              className="h-4 w-4 text-gray-400 mr-3"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <span className="text-gray-900">{suggestion}</span>
          </div>
        </button>
      ))}
    </div>
  );
};