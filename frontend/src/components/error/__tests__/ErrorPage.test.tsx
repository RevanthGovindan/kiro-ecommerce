import React from 'react';
import { render, screen } from '@testing-library/react';
import { ErrorPage } from '../ErrorPage';

// Mock Next.js Link component
jest.mock('next/link', () => {
  return ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
});

describe('ErrorPage', () => {
  it('renders default 500 error page', () => {
    render(<ErrorPage />);

    expect(screen.getByText('500')).toBeInTheDocument();
    expect(screen.getByText('Internal Server Error')).toBeInTheDocument();
    expect(screen.getByText('We\'re experiencing technical difficulties. Please try again later.')).toBeInTheDocument();
    expect(screen.getByText('Go back home')).toBeInTheDocument();
    expect(screen.getByText('Contact support')).toBeInTheDocument();
  });

  it('renders 404 error page', () => {
    render(<ErrorPage statusCode={404} />);

    expect(screen.getByText('404')).toBeInTheDocument();
    expect(screen.getByText('Page Not Found')).toBeInTheDocument();
    expect(screen.getByText('The page you\'re looking for doesn\'t exist.')).toBeInTheDocument();
  });

  it('renders 401 error page', () => {
    render(<ErrorPage statusCode={401} />);

    expect(screen.getByText('401')).toBeInTheDocument();
    expect(screen.getByText('Unauthorized')).toBeInTheDocument();
    expect(screen.getByText('You need to sign in to access this page.')).toBeInTheDocument();
  });

  it('renders 403 error page', () => {
    render(<ErrorPage statusCode={403} />);

    expect(screen.getByText('403')).toBeInTheDocument();
    expect(screen.getByText('Forbidden')).toBeInTheDocument();
    expect(screen.getByText('You don\'t have permission to access this resource.')).toBeInTheDocument();
  });

  it('renders 429 error page', () => {
    render(<ErrorPage statusCode={429} />);

    expect(screen.getByText('429')).toBeInTheDocument();
    expect(screen.getByText('Too Many Requests')).toBeInTheDocument();
    expect(screen.getByText('You\'ve made too many requests. Please wait a moment and try again.')).toBeInTheDocument();
  });

  it('renders custom title and message', () => {
    render(
      <ErrorPage
        statusCode={500}
        title="Custom Error"
        message="This is a custom error message"
      />
    );

    expect(screen.getByText('Custom Error')).toBeInTheDocument();
    expect(screen.getByText('This is a custom error message')).toBeInTheDocument();
  });

  it('renders retry button when showRetry is true', () => {
    const mockRetry = jest.fn();
    
    render(
      <ErrorPage
        statusCode={500}
        showRetry={true}
        onRetry={mockRetry}
      />
    );

    const retryButton = screen.getByText('Try again');
    expect(retryButton).toBeInTheDocument();
    
    retryButton.click();
    expect(mockRetry).toHaveBeenCalledTimes(1);
  });

  it('does not render retry button when showRetry is false', () => {
    render(<ErrorPage showRetry={false} />);

    expect(screen.queryByText('Try again')).not.toBeInTheDocument();
  });

  it('renders correct links', () => {
    render(<ErrorPage />);

    const homeLink = screen.getByText('Go back home');
    expect(homeLink.closest('a')).toHaveAttribute('href', '/');

    const supportLink = screen.getByText('Contact support');
    expect(supportLink.closest('a')).toHaveAttribute('href', '/contact');
  });

  it('renders different icons for different error types', () => {
    const { rerender } = render(<ErrorPage statusCode={404} />);
    
    // 404 should have a different icon (we can't easily test the actual SVG content,
    // but we can verify the component renders without errors)
    expect(screen.getByText('404')).toBeInTheDocument();

    rerender(<ErrorPage statusCode={500} />);
    expect(screen.getByText('500')).toBeInTheDocument();
  });

  it('handles unknown status codes gracefully', () => {
    render(<ErrorPage statusCode={999} />);

    expect(screen.getByText('999')).toBeInTheDocument();
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('An unexpected error occurred. Please try again later.')).toBeInTheDocument();
  });
});