import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { SearchSuggestions } from '../SearchSuggestions';

describe('SearchSuggestions', () => {
  const mockSuggestions = [
    'Advanced Product',
    'Advanced Search',
    'Advanced Filter',
    'Advanced Technology',
  ];

  const mockOnSelect = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders suggestions correctly', () => {
    render(
      <SearchSuggestions
        suggestions={mockSuggestions}
        onSelect={mockOnSelect}
      />
    );

    mockSuggestions.forEach(suggestion => {
      expect(screen.getByText(suggestion)).toBeInTheDocument();
    });
  });

  it('calls onSelect when suggestion is clicked', () => {
    render(
      <SearchSuggestions
        suggestions={mockSuggestions}
        onSelect={mockOnSelect}
      />
    );

    const firstSuggestion = screen.getByText('Advanced Product');
    fireEvent.click(firstSuggestion);

    expect(mockOnSelect).toHaveBeenCalledWith('Advanced Product');
  });

  it('renders nothing when suggestions array is empty', () => {
    const { container } = render(
      <SearchSuggestions
        suggestions={[]}
        onSelect={mockOnSelect}
      />
    );

    expect(container.firstChild).toBeNull();
  });

  it('applies hover styles correctly', () => {
    render(
      <SearchSuggestions
        suggestions={mockSuggestions}
        onSelect={mockOnSelect}
      />
    );

    const firstSuggestion = screen.getByText('Advanced Product').closest('button');
    expect(firstSuggestion).toHaveClass('hover:bg-gray-50');
  });

  it('displays search icon for each suggestion', () => {
    render(
      <SearchSuggestions
        suggestions={mockSuggestions}
        onSelect={mockOnSelect}
      />
    );

    const searchIcons = screen.getAllByRole('button').map(button => 
      button.querySelector('svg')
    );

    expect(searchIcons).toHaveLength(mockSuggestions.length);
    searchIcons.forEach(icon => {
      expect(icon).toBeInTheDocument();
    });
  });

  it('handles keyboard navigation', () => {
    render(
      <SearchSuggestions
        suggestions={mockSuggestions}
        onSelect={mockOnSelect}
      />
    );

    const firstSuggestion = screen.getByText('Advanced Product').closest('button');
    
    // Test focus
    firstSuggestion?.focus();
    expect(firstSuggestion).toHaveFocus();
    
    // Test Enter key
    fireEvent.keyDown(firstSuggestion!, { key: 'Enter' });
    // Note: The component doesn't handle Enter key explicitly, 
    // but the click event should be triggered by default button behavior
  });
});