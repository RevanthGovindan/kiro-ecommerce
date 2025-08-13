import React from 'react';
import { render, screen } from '@testing-library/react';
import { ErrorBoundary } from '../ErrorBoundary';

// Mock fetch for error logging
global.fetch = jest.fn(() => Promise.resolve({
  ok: true,
  json: () => Promise.resolve({}),
})) as jest.Mock;

// Component that throws an error
const ThrowError = ({ shouldThrow }: { shouldThrow: boolean }) => {
  if (shouldThrow) {
    throw new Error('Test error');
  }
  return <div>No error</div>;
};

// Mock console.error to avoid noise in tests
const originalError = console.error;
beforeAll(() => {
  console.error = jest.fn();
});

afterAll(() => {
  console.error = originalError;
});

beforeEach(() => {
  (fetch as jest.Mock).mockClear();
});

describe('ErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByText('No error')).toBeInTheDocument();
  });

  it('renders error fallback when there is an error', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Oops! Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Try Again')).toBeInTheDocument();
    expect(screen.getByText('Go back to homepage')).toBeInTheDocument();
  });

  it('logs error to monitoring service', () => {
    (fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => ({}),
    });

    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(fetch).toHaveBeenCalledWith('/api/errors/client', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: expect.stringContaining('Test error'),
    });
  });

  it('handles fetch error gracefully', () => {
    (fetch as jest.Mock).mockRejectedValue(new Error('Network error'));

    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    // Should still render error boundary even if logging fails
    expect(screen.getByText('Oops! Something went wrong')).toBeInTheDocument();
  });

  it('uses custom fallback component when provided', () => {
    const CustomFallback = () => <div>Custom error message</div>;

    render(
      <ErrorBoundary fallback={CustomFallback}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Custom error message')).toBeInTheDocument();
  });

  it('resets error state when resetError is called', () => {
    const { rerender } = render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Oops! Something went wrong')).toBeInTheDocument();

    // Click try again button
    const tryAgainButton = screen.getByText('Try Again');
    tryAgainButton.click();

    // Re-render with no error
    rerender(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByText('No error')).toBeInTheDocument();
  });
});